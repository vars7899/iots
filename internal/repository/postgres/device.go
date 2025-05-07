package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/vars7899/iots/internal/domain"
	"github.com/vars7899/iots/internal/domain/model"
	"github.com/vars7899/iots/internal/repository"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/logger"
	"github.com/vars7899/iots/pkg/pagination"
	"github.com/vars7899/iots/pkg/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DeviceRepositoryPostgres struct {
	db           *gorm.DB
	logger       *zap.Logger
	maxBatchSize int
}

func NewDeviceRepositoryPostgres(db *gorm.DB, baseLogger *zap.Logger) repository.DeviceRepository {
	return &DeviceRepositoryPostgres{
		db:           db,
		logger:       logger.Named(baseLogger, "DeviceRepositoryPostgres"),
		maxBatchSize: 100,
	}
}

func (r *DeviceRepositoryPostgres) Transaction(ctx context.Context, fn func(txRepo repository.DeviceRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &DeviceRepositoryPostgres{db: tx, logger: r.logger, maxBatchSize: r.maxBatchSize}
		return fn(txRepo)
	})
}

func (r *DeviceRepositoryPostgres) Create(ctx context.Context, device *model.Device) (*model.Device, error) {
	device.Status = model.DeviceStatusPendingProvision
	device.OwnerID = nil

	if err := r.db.WithContext(ctx).Model(&model.Device{}).Create(&device).Error; err != nil {
		r.logger.Debug("Failed to create new device", zap.Error(err))
		return nil, apperror.MapDBError(err, domain.EntityDevice)
	}
	return device, nil
}

func (r *DeviceRepositoryPostgres) GetByID(ctx context.Context, deviceID uuid.UUID) (*model.Device, error) {
	var device model.Device
	if err := r.db.WithContext(ctx).Where("id = ?", deviceID).First(&device).Error; err != nil {
		r.logger.Debug("Failed to get device", zap.String("device_id", deviceID.String()), zap.Error(err))
		return nil, apperror.MapDBError(err, domain.EntityDevice)
	}
	return &device, nil
}

func (r *DeviceRepositoryPostgres) GetByIDWithSensors(ctx context.Context, deviceID uuid.UUID) (*model.Device, error) {
	var device model.Device
	if err := r.db.WithContext(ctx).Preload("Sensors").Where("id = ?", deviceID).First(&device).Error; err != nil {
		r.logger.Debug("Failed to get device with sensors", zap.String("device_id", deviceID.String()), zap.Error(err))
		return nil, apperror.MapDBError(err, domain.EntityDevice)
	}
	return &device, nil
}

func (r *DeviceRepositoryPostgres) Update(ctx context.Context, device *model.Device) (*model.Device, error) {
	var updatedDevice model.Device
	tx := r.db.WithContext(ctx).Model(&model.Device{}).Clauses(clause.Returning{}).Where("id = ?", device.ID).Updates(device).Scan(&updatedDevice)
	if tx.Error != nil {
		r.logger.Debug("Failed to update device", zap.String("device_id", device.ID.String()), zap.Error(tx.Error))
		return nil, apperror.MapDBError(tx.Error, domain.EntityDevice)
	}
	if tx.RowsAffected == 0 {
		r.logger.Debug("Failed to update device: no matching record found", zap.String("device_id", device.ID.String()))
		return nil, apperror.ErrNotFound.WithMessagef("update operation failed: no matching %s found", domain.EntityDevice)
	}
	return &updatedDevice, nil
}

func (r *DeviceRepositoryPostgres) HardDelete(ctx context.Context, deviceID uuid.UUID) error {
	tx := r.db.WithContext(ctx).Unscoped().Where("id = ?", deviceID).Delete(&model.Device{})
	if tx.Error != nil {
		r.logger.Debug("Failed to hard delete", zap.String("device_id", deviceID.String()), zap.Error(tx.Error))
		return apperror.MapDBError(tx.Error, domain.EntityDevice)
	}
	if tx.RowsAffected == 0 {
		r.logger.Debug("Failed to hard delete device: no matching record found", zap.String("device_id", deviceID.String()))
		return apperror.ErrNotFound.WithMessagef("hard delete operation failed: no matching %s found", domain.EntityDevice)
	}
	return nil
}

func (r *DeviceRepositoryPostgres) Delete(ctx context.Context, deviceID uuid.UUID) error {
	tx := r.db.WithContext(ctx).Where("id = ?", deviceID).Delete(&model.Device{})
	if tx.Error != nil {
		r.logger.Debug("Failed to soft delete", zap.String("device_id", deviceID.String()), zap.Error(tx.Error))
		return apperror.MapDBError(tx.Error, domain.EntityDevice)
	}
	if tx.RowsAffected == 0 {
		var exist bool
		r.db.WithContext(ctx).Unscoped().Model(&model.Device{}).Select("count(*) > 0").Where("id = ?", deviceID).Find(&exist)
		if !exist {
			r.logger.Debug("Failed to soft delete device: no matching record found", zap.String("device_id", deviceID.String()))
			return apperror.ErrNotFound.WithMessagef("delete operation failed: %s with ID %s not found", domain.EntityDevice, deviceID)
		}
		r.logger.Debug("Failed to soft delete device: device not found or already deleted", zap.String("device_id", deviceID.String()))
		return apperror.ErrNotFound.WithMessagef("delete operation failed: %s not found or already soft deleted", domain.EntityDevice)
	}
	return nil
}

func (r *DeviceRepositoryPostgres) Recover(ctx context.Context, deviceID uuid.UUID) error {
	tx := r.db.WithContext(ctx).Model(&model.Device{}).Where("id = ?", deviceID).Update("deleted_at", nil)
	if tx.Error != nil {
		r.logger.Debug("Failed to recover deleted device", zap.String("device_id", deviceID.String()), zap.Error(tx.Error))
		return apperror.MapDBError(tx.Error, domain.EntityDevice)
	}
	if tx.RowsAffected == 0 {
		r.logger.Debug("Failed to recover deleted device: no matching record found", zap.String("device_id", deviceID.String()))
		return apperror.ErrNotFound.WithMessagef("delete operation failed: no matching %s found", domain.EntityDevice)
	}
	return nil
}

func (r *DeviceRepositoryPostgres) List(ctx context.Context, filter *domain.DeviceFilter, opt ...*pagination.Pagination) ([]*model.Device, int64, error) {
	var paginationConfig *pagination.Pagination
	if len(opt) > 0 {
		paginationConfig = opt[0]
	}

	queryBuilder := func(tx *gorm.DB) *gorm.DB {
		if filter.Name != nil {
			tx = tx.Where("name ILIKE ?", filter.Name)
		}
		if filter.Status != nil {
			tx = tx.Where("status = ?", filter.Status)
		}
		return tx
	}

	if paginationConfig != nil {
		devices, totalDevice, err := FindWithPagination[model.Device](ctx, r.db, paginationConfig, queryBuilder, r.logger)
		if err != nil {
			r.logger.Debug("Failed to list devices with pagination", zap.Error(err))
			return nil, 0, apperror.MapDBError(err, domain.EntityDevice)
		}
		return utils.ConvertVectorToPointerVector(devices), totalDevice, nil
	}

	var devices []model.Device
	if err := queryBuilder(r.db.WithContext(ctx)).Find(&devices).Error; err != nil {
		r.logger.Debug("Failed to list devices without pagination", zap.Error(err))
		return nil, 0, err
	}

	return utils.ConvertVectorToPointerVector(devices), int64(len(devices)), nil
}

func (r *DeviceRepositoryPostgres) GetByOwnerID(ctx context.Context, ownerID uuid.UUID, paginationConfig *pagination.Pagination) ([]*model.Device, int64, error) {
	queryBuilder := func(tx *gorm.DB) *gorm.DB {
		return tx.Where("owner_id = ?", ownerID)
	}

	devices, totalDevice, err := FindWithPagination[model.Device](ctx, r.db, paginationConfig, queryBuilder, r.logger)
	if err != nil {
		r.logger.Debug("Failed to list devices by owner ID with pagination", zap.Error(err))
		return nil, 0, apperror.MapDBError(err, domain.EntityDevice)
	}
	return utils.ConvertVectorToPointerVector(devices), totalDevice, nil
}

func (r *DeviceRepositoryPostgres) GetByStatus(ctx context.Context, status model.DeviceStatus, paginationConfig *pagination.Pagination) ([]*model.Device, int64, error) {
	queryBuilder := func(tx *gorm.DB) *gorm.DB {
		return tx.Where("status = ?", string(status))
	}

	devices, totalDevice, err := FindWithPagination[model.Device](ctx, r.db, paginationConfig, queryBuilder, r.logger)
	if err != nil {
		r.logger.Debug("Failed to list devices by status with pagination", zap.Error(err))
		return nil, 0, apperror.MapDBError(err, domain.EntityDevice)
	}
	return utils.ConvertVectorToPointerVector(devices), totalDevice, nil
}

func (r *DeviceRepositoryPostgres) GetDeleted(ctx context.Context, paginationConfig *pagination.Pagination) ([]*model.Device, int64, error) {
	queryBuilder := func(tx *gorm.DB) *gorm.DB {
		return tx.Unscoped().Where("deleted_at IS NOT NULL")
	}

	devices, totalDevice, err := FindWithPagination[model.Device](ctx, r.db, paginationConfig, queryBuilder, r.logger)
	if err != nil {
		r.logger.Debug("Failed to list deleted by  with pagination", zap.Error(err))
		return nil, 0, apperror.MapDBError(err, domain.EntityDevice)
	}
	return utils.ConvertVectorToPointerVector(devices), totalDevice, nil
}

func (r *DeviceRepositoryPostgres) SearchByName(ctx context.Context, searchStr string, paginationConfig *pagination.Pagination) ([]*model.Device, int64, error) {
	queryBuilder := func(tx *gorm.DB) *gorm.DB {
		return tx.Where("name ILIKE ?", "%"+searchStr+"%")
	}

	devices, totalDevice, err := FindWithPagination[model.Device](ctx, r.db, paginationConfig, queryBuilder, r.logger)
	if err != nil {
		r.logger.Debug("Failed to search devices by name with pagination", zap.Error(err))
		return nil, 0, apperror.MapDBError(err, domain.EntityDevice)
	}
	return utils.ConvertVectorToPointerVector(devices), totalDevice, nil
}

func (r *DeviceRepositoryPostgres) SearchByTags(ctx context.Context, tags []string, paginationConfig *pagination.Pagination) ([]*model.Device, int64, error) {
	queryBuilder := func(tx *gorm.DB) *gorm.DB {
		return tx.Where("tags @> ?", tags)
	}

	devices, totalDevice, err := FindWithPagination[model.Device](ctx, r.db, paginationConfig, queryBuilder, r.logger)
	if err != nil {
		r.logger.Debug("Failed to search devices by tags with pagination", zap.Error(err))
		return nil, 0, apperror.MapDBError(err, domain.EntityDevice)
	}
	return utils.ConvertVectorToPointerVector(devices), totalDevice, nil
}

func (r *DeviceRepositoryPostgres) SearchByCapabilities(ctx context.Context, capabilities []string, paginationConfig *pagination.Pagination) ([]*model.Device, int64, error) {
	queryBuilder := func(tx *gorm.DB) *gorm.DB {
		return tx.Where("capabilities @> ?", capabilities)
	}

	devices, totalDevice, err := FindWithPagination[model.Device](ctx, r.db, paginationConfig, queryBuilder, r.logger)
	if err != nil {
		r.logger.Debug("Failed to search devices by capabilities with pagination", zap.Error(err))
		return nil, 0, apperror.MapDBError(err, domain.EntityDevice)
	}
	return utils.ConvertVectorToPointerVector(devices), totalDevice, nil
}

func (r *DeviceRepositoryPostgres) AssignOwner(ctx context.Context, deviceID uuid.UUID, ownerID uuid.UUID) error {
	tx := r.db.WithContext(ctx).Model(&model.Device{}).Where("id = ?", deviceID).Update("owner_id", ownerID)
	if tx.Error != nil {
		r.logger.Debug("Failed to assign owner to device", zap.Error(tx.Error), zap.String("device_id", deviceID.String()), zap.String("owner_id", ownerID.String()))
		return apperror.MapDBError(tx.Error, domain.EntityDevice)
	}
	if tx.RowsAffected == 0 {
		r.logger.Debug("Failed to assign owner: no matching record found", zap.String("device_id", deviceID.String()), zap.String("owner_id", ownerID.String()))
		return apperror.ErrNotFound.WithMessagef("assign owner operation failed: no matching %s found", domain.EntityDevice)
	}
	return nil
}

func (r *DeviceRepositoryPostgres) UpdateStatus(ctx context.Context, deviceID uuid.UUID, newStatus model.DeviceStatus) error {
	tx := r.db.WithContext(ctx).Model(&model.Device{}).Where("id = ?", deviceID).Update("status", newStatus)
	if tx.Error != nil {
		r.logger.Debug("Failed to update device status", zap.Error(tx.Error), zap.String("device_id", deviceID.String()), zap.String("status", string(newStatus)))
		return apperror.MapDBError(tx.Error, domain.EntityDevice)
	}
	if tx.RowsAffected == 0 {
		r.logger.Debug("Failed to update device status: no matching record found", zap.String("device_id", deviceID.String()), zap.String("status", string(newStatus)))
		return apperror.ErrNotFound.WithMessagef("update device status operation failed: no matching %s found", domain.EntityDevice)
	}
	return nil
}

func (r *DeviceRepositoryPostgres) UpdateLastConnected(ctx context.Context, deviceID uuid.UUID, timestamp time.Time) error {
	tx := r.db.WithContext(ctx).Model(&model.Device{}).Where("id = ?", deviceID).Update("last_connected", timestamp)
	if tx.Error != nil {
		r.logger.Debug("Failed to update last connected", zap.Error(tx.Error), zap.String("device_id", deviceID.String()))
		return apperror.MapDBError(tx.Error, domain.EntityDevice)
	}
	if tx.RowsAffected == 0 {
		r.logger.Debug("Failed to update last connected: no matching record found", zap.String("device_id", deviceID.String()))
		return apperror.ErrNotFound.WithMessagef("update last connected operation failed: no matching %s found", domain.EntityDevice)
	}
	return nil
}

func (r *DeviceRepositoryPostgres) CountByStatus(ctx context.Context) (map[model.DeviceStatus]int64, error) {
	type countResult struct {
		StatusName model.DeviceStatus
		Count      int64
	}

	var results []countResult
	if err := r.db.WithContext(ctx).Model(&model.Device{}).Select("status, count(*) as count").Group("status").Find(&results).Error; err != nil {
		r.logger.Debug("Failed to count by status")
		return nil, apperror.MapDBError(err, domain.EntityDevice)
	}

	// restructure
	StatusCountMap := make(map[model.DeviceStatus]int64, len(results))
	for _, r := range results {
		StatusCountMap[r.StatusName] = r.Count
	}
	return StatusCountMap, nil
}

func (r *DeviceRepositoryPostgres) FindByMACAddr(ctx context.Context, macAddr string) (*model.Device, error) {
	var device model.Device
	if err := r.db.WithContext(ctx).Model(&model.Device{}).Where("mac_address = ?", macAddr).First(&device).Error; err != nil {
		return nil, apperror.MapDBError(err, domain.EntityDevice)
	}
	return &device, nil
}

func (r *DeviceRepositoryPostgres) ExistByMACAddr(ctx context.Context, macAddr string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Device{}).Where("mac_address = ?", macAddr).Count(&count).Error; err != nil {
		return false, apperror.MapDBError(err, domain.EntityDevice)
	}
	return count > 1, nil
}

func (r *DeviceRepositoryPostgres) MarkAsProvisioned(ctx context.Context, deviceID uuid.UUID) error {
	tx := r.db.WithContext(ctx).Model(&model.Device{}).Where("id = ?", deviceID).Update("status", model.DeviceStatusProvisioned)
	if tx.Error != nil {
		r.logger.Debug("Failed to mark device as provisioned", zap.Error(tx.Error), zap.String("device_id", deviceID.String()))
		return apperror.MapDBError(tx.Error, domain.EntityDevice)
	}
	if tx.RowsAffected == 0 {
		r.logger.Debug("Failed to mark device as provisioned: no matching record found", zap.String("device_id", deviceID.String()))
		return apperror.ErrNotFound.WithMessagef("mark device as provisioned operation failed: no matching %s found", domain.EntityDevice)
	}
	return nil
}

// func (r *DeviceRepositoryPostgres) GetDeviceCountTransaction(tx *gorm.DB) (int64, error) {
// 	var count int64
// 	if err := tx.Model(&model.Device{}).Count(&count).Error; err != nil {
// 		return 0, err
// 	}
// 	return count, nil
// }

// func (r *DeviceRepositoryPostgres) UpdateStatus(ctx context.Context, deviceID uuid.UUID, status domain.Status) error {
// 	r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
// 		if err := r.UpdateStatusTx(ctx, tx, deviceID, status); err != nil {
// 			return err
// 		}
// 		return nil
// 	})
// 	return nil
// }

// func (r *DeviceRepositoryPostgres) UpdateStatusTx(ctx context.Context, tx *gorm.DB, deviceID uuid.UUID, status domain.Status) error {
// 	op := "UpdateStatusTx"
// 	var deviceData model.Device
// 	if err := tx.WithContext(ctx).Model(&model.Device{}).Where("id = ?", deviceID).Update("status", status).First(&deviceData).Error; err != nil {
// 		return repository.HandleRepoError(op, err, apperror.ErrDBQuery, r.log)
// 	}
// 	return nil
// }

// func (r *DeviceRepositoryPostgres) MarkOnline(ctx context.Context, deviceID uuid.UUID) (*model.Device, error) {
// 	var deviceData *model.Device
// 	r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
// 		var err error
// 		deviceData, err = r.markOnlineTx(ctx, tx, deviceID)
// 		if err != nil {
// 			return err
// 		}
// 		return nil
// 	})
// 	return deviceData, nil
// }

// func (r *DeviceRepositoryPostgres) MarkOffline(ctx context.Context, deviceID uuid.UUID) (*model.Device, error) {
// 	var deviceData *model.Device
// 	r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
// 		var err error
// 		deviceData, err = r.markOfflineTx(ctx, tx, deviceID)
// 		if err != nil {
// 			return err
// 		}
// 		return nil
// 	})
// 	return deviceData, nil
// }

// // Bulk Operations
// func (r *DeviceRepositoryPostgres) BulkCreate(ctx context.Context, input []*model.Device) ([]*model.Device, error) {
// 	deviceList := make([]*model.Device, 0, len(input))
// 	var err error
// 	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
// 		deviceList, err = r.bulkCreateTx(ctx, tx, input)
// 		return err
// 	})
// 	if err != nil {
// 		r.log.Error("Failed to bulk create device", zap.Int("count", len(input)), zap.Error(err))
// 		return nil, err
// 	}
// 	r.log.Debug("bulk devices created", zap.Int("count", len(input)))
// 	return deviceList, nil
// }

// func (r *DeviceRepositoryPostgres) BulkUpdate(ctx context.Context, input []*model.Device) ([]*model.Device, error) {
// 	deviceList := make([]*model.Device, 0, len(input))

// 	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
// 		d, err := r.bulkUpdateTx(ctx, tx, input)
// 		deviceList = d
// 		return err
// 	})
// 	if err != nil {
// 		return nil, apperror.ErrorHandler(err, apperror.ErrCodeDBUpdate, "").Wrap(err)
// 	}
// 	r.log.Debug("bulk devices updated", zap.Int("count", len(input)))
// 	return deviceList, nil
// }

// func (r *DeviceRepositoryPostgres) BulkDelete(ctx context.Context, input []uuid.UUID) error {
// 	if len(input) == 0 {
// 		return apperror.ErrBadRequest.WithMessage("no device IDs provided for bulk delete")
// 	}
// 	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
// 		return r.bulkDeleteTx(ctx, tx, input)
// 	})
// 	if err != nil {
// 		return apperror.ErrDBDelete.WithMessage("Failed to delete devices in bulk")
// 	}

// 	r.log.Debug("bulk devices deleted", zap.Int("count", len(input)))
// 	return nil
// }

// func (r *DeviceRepositoryPostgres) markOnlineTx(ctx context.Context, tx *gorm.DB, deviceID uuid.UUID) (*model.Device, error) {
// 	op := "MarkOnlineTx"
// 	return r.updateIsOnlineTx(ctx, op, tx, deviceID, true)
// }

// func (r *DeviceRepositoryPostgres) markOfflineTx(ctx context.Context, tx *gorm.DB, deviceID uuid.UUID) (*model.Device, error) {
// 	op := "MarkOfflineTx"
// 	return r.updateIsOnlineTx(ctx, op, tx, deviceID, false)
// }

// func (r *DeviceRepositoryPostgres) updateIsOnlineTx(ctx context.Context, op string, tx *gorm.DB, deviceID uuid.UUID, newState bool) (*model.Device, error) {
// 	var deviceData model.Device
// 	result := tx.WithContext(ctx).Model(&model.Device{}).Clauses(clause.Returning{}).Where("id = ?", deviceID).Update("is_online", newState).First(&deviceData)
// 	if result.Error != nil {
// 		return nil, repository.HandleRepoError(op, result.Error, apperror.ErrDBUpdate, r.log)
// 	}
// 	if result.RowsAffected == 0 {
// 		return nil, apperror.ErrDBUpdate.WithMessage("resource not found to update")
// 	}
// 	return &deviceData, nil
// }

// func (r *DeviceRepositoryPostgres) bulkCreateTx(ctx context.Context, tx *gorm.DB, deviceList []*model.Device) ([]*model.Device, error) {
// 	op := "bulkCreateTx"
// 	if len(deviceList) == 0 {
// 		return nil, nil
// 	}

// 	batchSize := r.maxBatchSize
// 	totalDevices := len(deviceList)

// 	devices := make([]*model.Device, 0, totalDevices)

// 	for i := 0; i < totalDevices; i += batchSize {
// 		end := i + batchSize
// 		if end > totalDevices {
// 			end = totalDevices
// 		}

// 		deviceBatch := deviceList[i:end]

// 		if err := tx.WithContext(ctx).Model(&model.Device{}).Clauses(clause.Returning{}).Select("*").Create(&deviceBatch).Error; err != nil {
// 			return nil, repository.HandleRepoError(op, err, apperror.ErrDBInsert, r.log)
// 		}
// 		devices = append(devices, deviceBatch...)
// 	}

// 	return devices, nil
// }

// func (r *DeviceRepositoryPostgres) bulkUpdateTx(ctx context.Context, tx *gorm.DB, deviceList []*model.Device) ([]*model.Device, error) {
// 	op := "bulkUpdateTx"
// 	if len(deviceList) == 0 {
// 		return nil, nil
// 	}

// 	// check for available input ids
// 	inputIDs := make([]string, 0, len(deviceList))

// 	for _, d := range deviceList {
// 		inputIDs = append(inputIDs, d.ID.String())
// 	}

// 	var existingIDs []string
// 	if err := tx.Raw(`
// 			SELECT id FROM devices
// 			WHERE id IN (?)
// 		`, inputIDs).Scan(&existingIDs).Error; err != nil {
// 		return nil, repository.HandleRepoError(op, err, apperror.ErrDBQuery, r.log)
// 	}

// 	missing := findMissingIDs(inputIDs, existingIDs)
// 	if len(missing) > 0 {
// 		return nil, apperror.ErrInvalidUUID.WithMessage(fmt.Sprintf("some device IDs do not exist: %v", missing))
// 	}

// 	updatableFields := []string{"name", "description", "manufacturer", "model_number", "serial_number", "firmware_version", "ip_address", "mac_address", "connection_type"}
// 	caseMap := make(map[string][]string)
// 	idSet := make([]string, 0, len(deviceList))

// 	for _, device := range deviceList {
// 		idStr := device.ID.String()
// 		idSet = append(idSet, fmt.Sprintf("'%s'", idStr))

// 		val := reflect.ValueOf(device).Elem()
// 		for _, field := range updatableFields {
// 			fieldVal := val.FieldByName(field)
// 			if isZeroValue(fieldVal) {
// 				continue
// 			}

// 			fieldName := strings.ToLower(field)
// 			caseMap[fieldName] = append(caseMap[fieldName], fmt.Sprintf("WHEN '%s' THEN '%v'", idStr, escape(fmt.Sprintf("%v", fieldVal.Interface()))))
// 		}
// 	}

// 	setClauses := make([]string, 0)
// 	for field, cases := range caseMap {
// 		if len(cases) == 0 {
// 			continue
// 		}
// 		caseSQL := fmt.Sprintf("%s = CASE id\n    %s\nEND", field, strings.Join(cases, "\n    "))
// 		setClauses = append(setClauses, caseSQL)
// 	}

// 	if len(setClauses) == 0 {
// 		r.log.Warn("no valid fields to update in bulk")
// 		return nil, nil
// 	}

// 	query := fmt.Sprintf(`
// 		UPDATE devices
// 		SET %s
// 		WHERE id IN (%s)
// 		RETURNING *;
// 	`, strings.Join(setClauses, ",\n"), strings.Join(idSet, ","))

// 	var updated []*model.Device
// 	if err := tx.WithContext(ctx).Raw(query).Scan(&updated).Error; err != nil {
// 		return nil, repository.HandleRepoError(op, err, apperror.ErrDBUpdate, r.log)
// 	}

// 	r.log.Debug("bulk update completed using CASE WHEN", zap.Int("updated", len(updated)))
// 	return updated, nil
// }

// func (r *DeviceRepositoryPostgres) BulkDelete(ctx context.Context, deviceIDs []uuid.UUID) error {
// 	batchSize := r.maxBatchSize
// 	totalDevices := len(deviceIDs)

// 	for i := 0; i < totalDevices; i += batchSize {
// 		end := i + batchSize
// 		if end > totalDevices {
// 			end = totalDevices
// 		}

// 		deviceBatch := deviceIDs[i:end]
// 		if err := tx.WithContext(ctx).Where("id IN (?)", deviceBatch).Delete(&model.Device{}).Error; err != nil {
// 			return repository.HandleRepoError(op, err, apperror.ErrDBDelete, r.log)
// 		}
// 	}
// 	r.log.Debug("bulk device deleted successfully", zap.Int("device deleted", totalDevices))
// 	return nil
// }

// func (r *DeviceRepositoryPostgres) updateDevicesStatusTx(ctx context.Context, tx *gorm.DB, updates []model.DeviceUpdate, batchSize int) ([]*model.Device, error) {
// 	op := "updateDeviceStatusTx"
// 	totalUpdates := len(updates)
// 	updatedDevices := make([]*model.Device, 0, totalUpdates)

// 	for i := 0; i < totalUpdates; i += batchSize {
// 		end := i + batchSize
// 		if end > totalUpdates {
// 			end = totalUpdates
// 		}
// 		batchUpdate := updates[i:end]

// 		// collect device ids
// 		ids := make([]uuid.UUID, 0, len(batchUpdate))
// 		for idx, update := range batchUpdate {
// 			ids[idx] = update.ID
// 		}

// 		caseExp := "CASE id "
// 		params := make([]interface{}, 0, len(batchUpdate)*2)

// 		for _, update := range batchUpdate {
// 			status, ok := update.Updates.(string)
// 			if !ok {
// 				return nil, repository.HandleRepoError(op, errors.New("Failed to batch update status, invalid data"), apperror.ErrDBUpdate, r.log)
// 			}

// 			caseExp += "WHEN ? THEN ?"
// 			params = append(params, update.ID, status)
// 		}
// 		caseExp += "ELSE status END"

// 		updateQuery := "UPDATE devices SET status = " + caseExp + ", updated_at = NOW() WHERE id IN ?"
// 		params = append(params, ids)

// 		if err := tx.WithContext(ctx).Exec(updateQuery, params...).Error; err != nil {
// 			return nil, repository.HandleRepoError(op, err, apperror.ErrDBUpdate, r.log)
// 		}

// 		// Fetch updated devices
// 		var devices []*model.Device
// 		if err := tx.WithContext(ctx).Where("id IN ?", ids).Find(&devices).Error; err != nil {
// 			return nil, repository.HandleRepoError(op, err, apperror.ErrDBUpdate, r.log)
// 		}

// 		updatedDevices = append(updatedDevices, devices...)
// 	}
// 	return updatedDevices, nil
// }

// func escape(s string) string {
// 	// This is a quick patch for single quote escaping.
// 	return strings.ReplaceAll(s, "'", "''")
// }

// func isZeroValue(v reflect.Value) bool {
// 	// Handles invalid (unset) fields
// 	if !v.IsValid() {
// 		return true
// 	}

// 	// Handle pointers: dereference them
// 	if v.Kind() == reflect.Ptr {
// 		return v.IsNil()
// 	}

// 	// Use DeepEqual to compare with zero value of the type
// 	zero := reflect.Zero(v.Type())
// 	return reflect.DeepEqual(v.Interface(), zero.Interface())
// }

// func findMissingIDs(all, existing []string) []string {
// 	exists := map[string]struct{}{}
// 	for _, id := range existing {
// 		exists[id] = struct{}{}
// 	}
// 	var missing []string
// 	for _, id := range all {
// 		if _, found := exists[id]; !found {
// 			missing = append(missing, id)
// 		}
// 	}
// 	return missing
// }
