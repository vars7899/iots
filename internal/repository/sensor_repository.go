package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/vars7899/iots/internal/api/v1/dto"
	"github.com/vars7899/iots/internal/domain/model"
	"github.com/vars7899/iots/pkg/pagination"
)

type SensorRepository interface {
	Create(ctx context.Context, sensorData *model.Sensor) (*model.Sensor, error)
	GetByID(ctx context.Context, sensorID uuid.UUID) (*model.Sensor, error)
	Update(ctx context.Context, sensorData *model.Sensor) (*model.Sensor, error)
	Delete(ctx context.Context, sensorID uuid.UUID) error
	List(ctx context.Context, filter *dto.SensorFilter, paginationOpt ...*pagination.Pagination) ([]*model.Sensor, error)
}
