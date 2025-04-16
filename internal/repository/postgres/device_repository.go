package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
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
	log          *zap.Logger
	queryTimeout time.Duration
	maxBatchSize int
}

func NewDeviceRepositoryPostgres(db *gorm.DB) repository.DeviceRepository {
	log := logger.Lgr.Named("postgres.DeviceRepositoryPostgres")
	return &DeviceRepositoryPostgres{db: db, log: log, queryTimeout: time.Millisecond * 10000, maxBatchSize: 100}
}

func (r *DeviceRepositoryPostgres) Create(ctx context.Context, deviceData *model.Device) (*model.Device, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Device{}).Clauses(clause.Returning{}).Select("*").Create(&deviceData).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, repository.HandleRepoError("create device resource", err, apperror.ErrDBInsert, r.log)
	}

	return deviceData, nil
}

func (r *DeviceRepositoryPostgres) GetByID(ctx context.Context, deviceID uuid.UUID) (*model.Device, error) {
	var d model.Device
	if err := r.db.WithContext(ctx).Where("id = ?", deviceID).First(&d).Error; err != nil {
		return nil, apperror.HandleDBError(err)
	}
	return &d, nil
}

func (r *DeviceRepositoryPostgres) Update(ctx context.Context, deviceData *model.Device) (*model.Device, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	tx := r.db.WithContext(ctx).Model(&model.Device{}).Clauses(clause.Returning{}).Where("id = ?", deviceData.ID).Updates(&deviceData)
	if tx.Error != nil {
		return nil, repository.HandleRepoError("update resource", tx.Error, apperror.ErrDBUpdate, r.log)
	}
	if tx.RowsAffected == 0 {
		return nil, apperror.ErrDBUpdate.WithMessage("resource not found to update")
	}
	return deviceData, nil
}

func (r *DeviceRepositoryPostgres) SoftDelete(ctx context.Context, deviceID uuid.UUID) error {
	tx := r.db.WithContext(ctx).Where("id = ?", deviceID).Delete(&model.Device{})
	if tx.Error != nil {
		return repository.HandleRepoError("delete resource", tx.Error, apperror.ErrDBDelete, r.log)
	}
	if tx.RowsAffected == 0 {
		return apperror.ErrDBUpdate.WithMessage("resource not found to delete")
	}
	return nil
}

func (r *DeviceRepositoryPostgres) FindAll(ctx context.Context, p *pagination.Pagination) ([]*model.Device, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	var (
		deviceList       []model.Device
		totalDeviceCount int64
	)
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Device{}).Count(&totalDeviceCount).Error; err != nil {
			return err
		}
		query := tx.Offset(p.GetOffset()).Limit(p.GetLimit())
		if orderClause := p.GetSortOrderClause(); orderClause != "" {
			query.Order(orderClause)
		}

		return query.Find(&deviceList).Error
	})
	if err != nil {
		return nil, 0, repository.HandleRepoError("find all (transaction)", err, apperror.ErrDBQuery, r.log)
	}
	devicesPtrList := utils.ConvertVectorToPointerVector(deviceList)
	return devicesPtrList, totalDeviceCount, nil
}

func (r *DeviceRepositoryPostgres) FindByOwnerID(ctx context.Context, ownerID uuid.UUID, p *pagination.Pagination) ([]*model.Device, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	var (
		deviceList       []model.Device
		totalDeviceCount int64
	)

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Device{}).Where("owner_id = ?", ownerID).Count(&totalDeviceCount).Error; err != nil {
			return err
		}

		query := tx.Offset(p.GetOffset()).Limit(p.GetLimit())
		if orderClause := p.GetSortOrderClause(); orderClause != "" {
			query = query.Order(orderClause)
		}

		return query.Where("owner_id = ?", ownerID).Find(&deviceList).Error
	})
	if err != nil {
		return nil, 0, repository.HandleRepoError("find by owner id (transaction)", err, apperror.ErrDBQuery, r.log)
	}
	devicesPtrList := utils.ConvertVectorToPointerVector(deviceList)
	return devicesPtrList, totalDeviceCount, nil

}

func (r *DeviceRepositoryPostgres) GetDeviceCountTransaction(tx *gorm.DB) (int64, error) {
	var count int64
	if err := tx.Model(&model.Device{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *DeviceRepositoryPostgres) UpdateStatus(ctx context.Context, deviceID uuid.UUID, status string) error {
	r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := r.UpdateStatusTx(ctx, tx, deviceID, status); err != nil {
			return err
		}
		return nil
	})
	return nil
}

func (r *DeviceRepositoryPostgres) UpdateStatusTx(ctx context.Context, tx *gorm.DB, deviceID uuid.UUID, status string) error {
	op := "UpdateStatusTx"
	var deviceData model.Device
	if err := tx.WithContext(ctx).Where("id = ?", deviceID).First(&deviceData).Error; err != nil {
		return repository.HandleRepoError(op, err, apperror.ErrDBQuery, r.log)
	}
	return nil
}

func (r *DeviceRepositoryPostgres) MarkOnline(ctx context.Context, deviceID uuid.UUID) (*model.Device, error) {
	var deviceData *model.Device
	r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		deviceData, err = r.markOnlineTx(ctx, tx, deviceID)
		if err != nil {
			return err
		}
		return nil
	})
	return deviceData, nil
}

func (r *DeviceRepositoryPostgres) MarkOffline(ctx context.Context, deviceID uuid.UUID) (*model.Device, error) {
	var deviceData *model.Device
	r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		deviceData, err = r.markOfflineTx(ctx, tx, deviceID)
		if err != nil {
			return err
		}
		return nil
	})
	return deviceData, nil
}

// Bulk Operations
func (r *DeviceRepositoryPostgres) BulkCreate(ctx context.Context, input []*model.Device) ([]*model.Device, error) {
	deviceList := make([]*model.Device, 0, len(input))
	var err error
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		deviceList, err = r.bulkCreateTx(ctx, tx, input)
		return err
	})
	if err != nil {
		r.log.Error("failed to bulk create device", zap.Int("count", len(input)), zap.Error(err))
		return nil, err
	}
	r.log.Debug("bulk devices created", zap.Int("count", len(input)))
	return deviceList, nil
}

func (r *DeviceRepositoryPostgres) BulkUpdate(ctx context.Context, input []*model.Device) ([]*model.Device, error) {
	deviceList := make([]*model.Device, 0, len(input))
	var err error

	r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		deviceList, err = r.bulkUpdateTx(ctx, tx, input)
		r.log.Error("failed to bulk update device", zap.Int("count", len(input)), zap.Error(err))
		return err
	})

	r.log.Debug("bulk devices updated", zap.Int("count", len(input)))
	return deviceList, nil
}

func (r *DeviceRepositoryPostgres) BulkDelete(ctx context.Context, input []uuid.UUID) error {
	r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := r.bulkDeleteTx(ctx, tx, input)
		r.log.Error("failed to bulk update device", zap.Int("count", len(input)), zap.Error(err))
		return err
	})

	r.log.Debug("bulk devices deleted", zap.Int("count", len(input)))
	return nil
}

func (r *DeviceRepositoryPostgres) markOnlineTx(ctx context.Context, tx *gorm.DB, deviceID uuid.UUID) (*model.Device, error) {
	op := "MarkOnlineTx"
	return r.updateIsOnlineTx(ctx, op, tx, deviceID, true)
}

func (r *DeviceRepositoryPostgres) markOfflineTx(ctx context.Context, tx *gorm.DB, deviceID uuid.UUID) (*model.Device, error) {
	op := "MarkOfflineTx"
	return r.updateIsOnlineTx(ctx, op, tx, deviceID, false)
}

func (r *DeviceRepositoryPostgres) updateIsOnlineTx(ctx context.Context, op string, tx *gorm.DB, deviceID uuid.UUID, newState bool) (*model.Device, error) {
	var deviceData model.Device
	result := tx.WithContext(ctx).Clauses(clause.Returning{}).Where("id = ?", deviceID).Update("is_online", newState).First(&deviceData)
	if result.Error != nil {
		return nil, repository.HandleRepoError(op, result.Error, apperror.ErrDBUpdate, r.log)
	}
	if result.RowsAffected == 0 {
		return nil, apperror.ErrDBUpdate.WithMessage("resource not found to update")
	}
	return &deviceData, nil
}

func (r *DeviceRepositoryPostgres) bulkCreateTx(ctx context.Context, tx *gorm.DB, deviceList []*model.Device) ([]*model.Device, error) {
	op := "bulkCreateTx"
	if len(deviceList) == 0 {
		return nil, nil
	}

	batchSize := r.maxBatchSize
	totalDevices := len(deviceList)

	devices := make([]*model.Device, 0, totalDevices)

	for i := 0; i < totalDevices; i += batchSize {
		end := i + batchSize
		if end > totalDevices {
			end = totalDevices
		}

		deviceBatch := deviceList[i:end]

		if err := tx.WithContext(ctx).Model(&model.Device{}).Clauses(clause.Returning{}).Select("*").Create(&deviceBatch).Error; err != nil {
			return nil, repository.HandleRepoError(op, err, apperror.ErrDBInsert, r.log)
		}
		devices = append(devices, deviceBatch...)
	}

	return devices, nil
}

func (r *DeviceRepositoryPostgres) bulkUpdateTx(ctx context.Context, tx *gorm.DB, deviceList []*model.Device) ([]*model.Device, error) {
	op := "bulkUpdateTx"

	batchSize := r.maxBatchSize
	totalDevices := len(deviceList)

	devices := make([]*model.Device, 0, totalDevices)
	for i := 0; i < totalDevices; i += batchSize {
		end := i + batchSize
		if end < batchSize {
			end = totalDevices
		}

		deviceBatch := deviceList[i:end]
		if err := tx.WithContext(ctx).Model(&model.Device{}).Clauses(clause.Returning{}).Select("*").Updates(&deviceBatch).Error; err != nil {
			return nil, repository.HandleRepoError(op, err, apperror.ErrDBUpdate, r.log)
		}

		devices = append(devices, deviceBatch...)
	}

	r.log.Debug("bulk device updated successfully", zap.Int("devices updated", totalDevices))
	return devices, nil
}

func (r *DeviceRepositoryPostgres) bulkDeleteTx(ctx context.Context, tx *gorm.DB, deviceIDs []uuid.UUID) error {
	op := "bulkDeleteTx"

	batchSize := r.maxBatchSize
	totalDevices := len(deviceIDs)

	for i := 0; i < totalDevices; i += batchSize {
		end := i + batchSize
		if end > totalDevices {
			end = totalDevices
		}
		deviceBatch := deviceIDs[i:end]

		if err := tx.WithContext(ctx).Model(&model.Device{}).Delete(deviceBatch).Error; err != nil {
			return repository.HandleRepoError(op, err, apperror.ErrDBDelete, r.log)
		}
	}
	r.log.Debug("bulk device deleted successfully", zap.Int("device deleted", totalDevices))
	return nil
}

func (r *DeviceRepositoryPostgres) updateDevicesStatusTx(ctx context.Context, tx *gorm.DB, updates []model.DeviceUpdate, batchSize int) ([]*model.Device, error) {
	op := "updateDeviceStatusTx"
	totalUpdates := len(updates)
	updatedDevices := make([]*model.Device, 0, totalUpdates)

	for i := 0; i < totalUpdates; i += batchSize {
		end := i + batchSize
		if end > totalUpdates {
			end = totalUpdates
		}
		batchUpdate := updates[i:end]

		// collect device ids
		ids := make([]uuid.UUID, 0, len(batchUpdate))
		for idx, update := range batchUpdate {
			ids[idx] = update.ID
		}

		caseExp := "CASE id "
		params := make([]interface{}, 0, len(batchUpdate)*2)

		for _, update := range batchUpdate {
			status, ok := update.Updates.(string)
			if !ok {
				return nil, repository.HandleRepoError(op, errors.New("failed to batch update status, invalid data"), apperror.ErrDBUpdate, r.log)
			}

			caseExp += "WHEN ? THEN ?"
			params = append(params, update.ID, status)
		}
		caseExp += "ELSE status END"

		updateQuery := "UPDATE devices SET status = " + caseExp + ", updated_at = NOW() WHERE id IN ?"
		params = append(params, ids)

		if err := tx.WithContext(ctx).Exec(updateQuery, params...).Error; err != nil {
			return nil, repository.HandleRepoError(op, err, apperror.ErrDBUpdate, r.log)
		}

		// Fetch updated devices
		var devices []*model.Device
		if err := tx.WithContext(ctx).Where("id IN ?", ids).Find(&devices).Error; err != nil {
			return nil, repository.HandleRepoError(op, err, apperror.ErrDBUpdate, r.log)
		}

		updatedDevices = append(updatedDevices, devices...)
	}
	return updatedDevices, nil
}
