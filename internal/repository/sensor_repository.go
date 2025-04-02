package repository

import (
	"context"

	"github.com/vars7899/iots/internal/domain/sensor"
)

type SensorRepository interface {
	Create(ctx context.Context, s *sensor.Sensor) error
	GetByID(ctx context.Context, sid sensor.SensorID) (*sensor.Sensor, error)
	Update(ctx context.Context, s *sensor.Sensor) error
	Delete(ctx context.Context, sid sensor.SensorID) error
	List(ctx context.Context, filter sensor.SensorFilter) ([]*sensor.Sensor, error)
}
