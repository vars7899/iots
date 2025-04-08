package postgres

import (
	"context"
	"fmt"

	"github.com/vars7899/iots/internal/domain/sensor"
	"github.com/vars7899/iots/internal/repository"
	"github.com/vars7899/iots/pkg/validatorutils"
	"gorm.io/gorm"
)

type SensorRepositoryPostgres struct {
	db *gorm.DB
}

func NewSensorRepositoryPostgres(db *gorm.DB) repository.SensorRepository {
	return &SensorRepositoryPostgres{db: db}
}

func (r SensorRepositoryPostgres) Create(ctx context.Context, s *sensor.Sensor) error {
	if err := r.db.WithContext(ctx).Create(s).Error; err != nil {
		if validatorutils.IsPgDuplicateKeyError(err) {
			return repository.ErrDuplicateKey
		}
		return fmt.Errorf("failed to create sensor: %w", err)
	}
	return nil
}

func (r SensorRepositoryPostgres) GetByID(ctx context.Context, sid string) (*sensor.Sensor, error) {
	var s sensor.Sensor

	if err := r.db.WithContext(ctx).First(&s, "id = ?", sid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get sensor: %w", err)
	}
	return &s, nil
}

func (r SensorRepositoryPostgres) Delete(ctx context.Context, sid string) error {
	result := r.db.WithContext(ctx).Where("id = ?", sid).Delete(&sensor.Sensor{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete sensor: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r SensorRepositoryPostgres) Update(ctx context.Context, s *sensor.Sensor) error {
	result := r.db.WithContext(ctx).Model(&sensor.Sensor{}).Where("id = ?", s.ID).Select("*").Updates(s)
	if result.Error != nil {
		return fmt.Errorf("failed to update sensor: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r SensorRepositoryPostgres) List(ctx context.Context, filter sensor.SensorFilter) ([]*sensor.Sensor, error) {
	var sensorList []sensor.Sensor
	query := r.db.WithContext(ctx)

	if filter.DeviceID != nil {
		query.Where("device_id = ?", *filter.DeviceID)
	}
	if filter.Status != nil {
		query.Where("status = ?", *filter.Status)
	}
	if filter.Type != nil {
		query.Where("type = ?", *filter.Type)
	}
	if filter.CreatedAt != nil {
		query.Where("created_at = ?", *filter.CreatedAt)
	}

	if err := query.Find(&sensorList).Error; err != nil {
		return nil, err
	}

	// convert the sensor value to sensor pointers
	sensors := make([]*sensor.Sensor, len(sensorList))
	for i := range sensorList {
		sensors[i] = &sensorList[i]
	}

	return sensors, nil
}
