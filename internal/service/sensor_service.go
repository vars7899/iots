package service

import (
	"context"
	"errors"

	"github.com/vars7899/iots/internal/domain/sensor"
	"github.com/vars7899/iots/internal/repository"
	"gorm.io/gorm"
)

type SensorService struct {
	repo repository.SensorRepository
}

func NewSensorService(r repository.SensorRepository) *SensorService {
	return &SensorService{repo: r}
}

func (s *SensorService) CreateSensor(ctx context.Context, sensor *sensor.Sensor) error {
	sensor.StampNew()
	return s.repo.Create(ctx, sensor)
}

func (s *SensorService) GetSensor(ctx context.Context, sensorID string) (*sensor.Sensor, error) {
	sensorData, err := s.repo.GetByID(ctx, sensorID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sensor.ErrSensorNotFound
		}
		return nil, err
	}
	return sensorData, nil
}

func (s *SensorService) UpdateSensor(ctx context.Context, sensor *sensor.Sensor) error {
	sensor.StampUpdate()
	return s.repo.Update(ctx, sensor)
}

func (s *SensorService) DeleteSensor(ctx context.Context, sensorID string) error {
	if err := s.repo.Delete(ctx, sensorID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return sensor.ErrSensorNotFound
		}
	}
	return nil
}

func (s *SensorService) ListSensor(ctx context.Context, sensorFilter sensor.SensorFilter) ([]*sensor.Sensor, error) {
	sensorList, err := s.repo.List(ctx, sensorFilter)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sensor.ErrSensorNotFound
		}
		return nil, err
	}
	return sensorList, nil
}
