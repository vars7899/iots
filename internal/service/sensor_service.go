package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/vars7899/iots/internal/domain/sensor"
	"github.com/vars7899/iots/internal/repository"
	"github.com/vars7899/iots/pkg/validator"
	"gorm.io/gorm"
)

type SensorService struct {
	repo repository.SensorRepository
}

func NewSensorService(r repository.SensorRepository) *SensorService {
	return &SensorService{repo: r}
}

func (s *SensorService) CreateSensor(ctx context.Context, sensor *sensor.Sensor) error {
	if err := validator.ValidateSensor(sensor); err != nil {
		return fmt.Errorf("invalid sensor data: %v", err)
	}
	return s.repo.Create(ctx, sensor)
}

func (s *SensorService) GetSensor(ctx context.Context, sensorID string) (*sensor.Sensor, error) {
	if err := validator.ValidateSensorID(sensorID); err != nil {
		return nil, err
	}
	sensorData, err := s.repo.GetByID(ctx, sensor.SensorID(sensorID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sensor.ErrSensorNotFound
		}
		return nil, err
	}
	return sensorData, nil
}

func (s *SensorService) UpdateSensor(ctx context.Context, sensor *sensor.Sensor) error {
	if err := validator.ValidateSensor(sensor); err != nil {
		return fmt.Errorf("invalid sensor data: %v", err)
	}
	return s.repo.Update(ctx, sensor)
}

func (s *SensorService) DeleteSensor(ctx context.Context, sensorID string) error {
	if err := validator.ValidateSensorID(string(sensorID)); err != nil {
		return fmt.Errorf("invalid sensor id: %w", err)
	}
	return s.repo.Delete(ctx, sensor.SensorID(sensorID))
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
