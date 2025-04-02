package service

import (
	"context"
	"fmt"

	"github.com/vars7899/iots/internal/domain/sensor"
	"github.com/vars7899/iots/internal/repository"
	"github.com/vars7899/iots/pkg/validator"
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
	if sensorID == "" {
		return nil, fmt.Errorf("invalid sensor id: cannot be empty")
	}
	return s.repo.GetByID(ctx, sensor.SensorID(sensorID))
}

func (s *SensorService) UpdateSensor(ctx context.Context, sensor *sensor.Sensor) error {
	if err := validator.ValidateSensor(sensor); err != nil {
		return fmt.Errorf("invalid sensor data: %v", err)
	}
	return s.repo.Update(ctx, sensor)
}

func (s *SensorService) DeleteSensor(ctx context.Context, sensorID string) error {
	if sensorID == "" {
		return fmt.Errorf("invalid sensor id: cannot be empty")
	}
	return s.repo.Delete(ctx, sensor.SensorID(sensorID))
}

func (s *SensorService) ListSensor(ctx context.Context, sensorFilter sensor.SensorFilter) ([]*sensor.Sensor, error) {
	return s.repo.List(ctx, sensorFilter)
}
