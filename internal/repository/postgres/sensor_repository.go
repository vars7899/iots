package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/vars7899/iots/internal/api/v1/dto"
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

type SensorRepositoryPostgres struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewSensorRepositoryPostgres(db *gorm.DB, baseLogger *zap.Logger) repository.SensorRepository {
	return &SensorRepositoryPostgres{db: db, logger: logger.NewNamedZapLogger(baseLogger, "SensorRepositoryPostgres")}
}

func (r SensorRepositoryPostgres) Create(ctx context.Context, sensorData *model.Sensor) (*model.Sensor, error) {
	if err := r.db.WithContext(ctx).Clauses(clause.Returning{}).Create(&sensorData).Error; err != nil {
		return nil, apperror.MapDBError(err, domain.EntitySensor)
	}
	return sensorData, nil
}

func (r SensorRepositoryPostgres) GetByID(ctx context.Context, sensorID uuid.UUID) (*model.Sensor, error) {
	var sensorExist model.Sensor
	if err := r.db.WithContext(ctx).First(&sensorExist, "id = ?", sensorID).Error; err != nil {
		return nil, apperror.MapDBError(err, domain.EntitySensor)
	}
	return &sensorExist, nil
}

func (r SensorRepositoryPostgres) Delete(ctx context.Context, sensorID uuid.UUID) error {
	tx := r.db.WithContext(ctx).Where("id = ?", sensorID).Delete(&model.Sensor{})
	if tx.Error != nil {
		return apperror.MapDBError(tx.Error, domain.EntitySensor)
	}
	if tx.RowsAffected == 0 {
		return apperror.ErrNotFound.WithMessagef("cannot delete %s: no matching record found", domain.EntitySensor)
	}
	return nil
}

func (r SensorRepositoryPostgres) Update(ctx context.Context, sensorData *model.Sensor) (*model.Sensor, error) {
	var updatedSensor model.Sensor

	tx := r.db.WithContext(ctx).Model(&model.Sensor{}).Clauses(clause.Returning{}).Where("id = ?", sensorData.ID).Updates(sensorData).Scan(&updatedSensor)
	if tx.Error != nil {
		return nil, apperror.MapDBError(tx.Error, domain.EntitySensor)
	}
	if tx.RowsAffected == 0 {
		return nil, apperror.ErrNotFound.WithMessagef("error encountered while performing %s update operation, please retry", domain.EntitySensor)
	}
	return &updatedSensor, nil
}

func (r SensorRepositoryPostgres) List(ctx context.Context, filter *dto.SensorFilter, paginationOpt ...*pagination.Pagination) ([]*model.Sensor, error) {
	var paginationConfig *pagination.Pagination
	if len(paginationOpt) > 0 && paginationOpt[0] != nil {
		paginationConfig = paginationOpt[0]
	}

	queryBuilder := func(tx *gorm.DB) *gorm.DB {
		if filter.DeviceID != nil {
			tx = tx.Where("device_id = ?", *filter.DeviceID)
		}
		if filter.Status != nil {
			tx = tx.Where("status = ?", *filter.Status)
		}
		if filter.Type != nil {
			tx = tx.Where("type = ?", *filter.Type)
		}
		if filter.CreatedAt != nil {
			tx = tx.Where("created_at = ?", *filter.CreatedAt)
		}
		if filter.Name != nil {
			tx = tx.Where("name ILIKE ?", "%"+*filter.Name+"%")
		}
		if filter.Location != nil {
			tx = tx.Where("location ILIKE ?", "%"+*filter.Location+"%")
		}
		return tx
	}

	// If pagination config is provided, do paginated query
	if paginationConfig != nil {
		sensors, _, err := FindWithPagination[model.Sensor](ctx, r.db, paginationConfig, queryBuilder, r.logger, &QueryOptions{SkipCount: true})
		if err != nil {
			return nil, err
		}
		return utils.ConvertVectorToPointerVector(sensors), nil
	}

	// Otherwise: normal (non-paginated) query
	var sensors []model.Sensor
	if err := queryBuilder(r.db.WithContext(ctx)).Find(&sensors).Error; err != nil {
		return nil, err
	}

	return utils.ConvertVectorToPointerVector(sensors), nil
}

// paginationConfig := &pagination.Pagination{
// 	PageSize:  filter.Limit,
// 	Page:      (filter.Offset / filter.Limit) + 1, // or just use filter.Offset directly in your Pagination struct
// 	SortBy:    filter.SortBy,
// 	SortOrder: filter.SortOrder,
// }

// func (r SensorRepositoryPostgres) List(ctx context.Context, filter sensor.SensorFilter) ([]*sensor.Sensor, error) {
// 	var sensorList []sensor.Sensor
// 	query := r.db.WithContext(ctx).Model(&sensor.Sensor{})

// 	// Filters
// 	if filter.DeviceID != nil {
// 		query = query.Where("device_id = ?", *filter.DeviceID)
// 	}
// 	if filter.Status != nil {
// 		query = query.Where("status = ?", *filter.Status)
// 	}
// 	if filter.Type != nil {
// 		query = query.Where("type = ?", *filter.Type)
// 	}
// 	if filter.CreatedAt != nil {
// 		query = query.Where("created_at = ?", *filter.CreatedAt)
// 	}
// 	if filter.Name != nil {
// 		query = query.Where("name ILIKE ?", "%"+*filter.Name+"%")
// 	}
// 	if filter.Location != nil {
// 		query = query.Where("location ILIKE ?", "%"+*filter.Location+"%")
// 	}

// 	// Sorting
// 	sortBy := "created_at"
// 	if filter.SortBy != "" {
// 		sortBy = filter.SortBy
// 	}
// 	sortOrder := "desc"
// 	if filter.SortOrder == "asc" {
// 		sortOrder = "asc"
// 	}
// 	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

// 	// Pagination
// 	if filter.Limit > 0 {
// 		query = query.Limit(filter.Limit)
// 	}
// 	if filter.Offset > 0 {
// 		query = query.Offset(filter.Offset)
// 	}

// 	// Execute
// 	if err := query.Find(&sensorList).Error; err != nil {
// 		return nil, err
// 	}

// 	// Convert to []*sensor.Sensor
// 	sensors := make([]*sensor.Sensor, len(sensorList))
// 	for i := range sensorList {
// 		sensors[i] = &sensorList[i]
// 	}

// 	return sensors, nil
// }
