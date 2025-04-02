package postgres

import (
	"context"

	"github.com/vars7899/iots/internal/domain/sensor"
	"github.com/vars7899/iots/internal/repository"
	"gorm.io/gorm"
)

type SensorRepositoryPostgres struct {
	db *gorm.DB
}

func NewSensorRepositoryPostgres(db *gorm.DB) repository.SensorRepository {
	return &SensorRepositoryPostgres{db: db}
}

func (r SensorRepositoryPostgres) Create(ctx context.Context, s *sensor.Sensor) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r SensorRepositoryPostgres) GetByID(ctx context.Context, sid sensor.SensorID) (*sensor.Sensor, error) {
	var s sensor.Sensor
	err := r.db.WithContext(ctx).First(&s, "id = ?", sid).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r SensorRepositoryPostgres) Delete(ctx context.Context, sid sensor.SensorID) error {
	return r.db.WithContext(ctx).Where("id = ?", sid).Delete(&sensor.Sensor{}).Error
}

func (r SensorRepositoryPostgres) Update(ctx context.Context, s *sensor.Sensor) error {
	return r.db.WithContext(ctx).Save(s).Error
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
