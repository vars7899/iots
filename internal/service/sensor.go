package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/vars7899/iots/internal/api/v1/dto"
	"github.com/vars7899/iots/internal/domain"
	"github.com/vars7899/iots/internal/domain/model"
	"github.com/vars7899/iots/internal/repository"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/logger"
	"github.com/vars7899/iots/pkg/pagination"
	"go.uber.org/zap"
)

type SensorService struct {
	sensorRepo repository.SensorRepository
	logger     *zap.Logger
}

func NewSensorService(r repository.SensorRepository, baseLogger *zap.Logger) *SensorService {
	return &SensorService{sensorRepo: r, logger: logger.Named(baseLogger, "SensorService")}
}

func (s *SensorService) CreateSensor(ctx context.Context, sensor *model.Sensor) (*model.Sensor, error) {
	if sensor.ID != uuid.Nil {
		return nil, apperror.ErrBadRequest.WithMessagef("cannot specify ID when creating a %s", domain.EntitySensor)
	}
	sensor.Status = domain.Pending
	return s.sensorRepo.Create(ctx, sensor)
}

func (s *SensorService) GetSensor(ctx context.Context, sensorID uuid.UUID) (*model.Sensor, error) {
	sensorData, err := s.sensorRepo.GetByID(ctx, sensorID)
	if err != nil {
		return nil, apperror.ErrorHandler(err, apperror.ErrCodeDBQuery, fmt.Sprintf("failed to fetch %s with ID %s", domain.EntitySensor, sensorID))
	}
	return sensorData, nil
}

func (s *SensorService) UpdateSensor(ctx context.Context, sensorData *model.Sensor) (*model.Sensor, error) {
	return s.sensorRepo.Update(ctx, sensorData)
}

func (s *SensorService) DeleteSensor(ctx context.Context, sensorID uuid.UUID) error {
	return s.sensorRepo.Delete(ctx, sensorID)
}

func (s *SensorService) ListSensor(ctx context.Context, sensorFilter *dto.SensorFilter, paginationConfig *pagination.Pagination) ([]*model.Sensor, error) {
	sensorList, err := s.sensorRepo.List(ctx, sensorFilter, paginationConfig)
	if err != nil {
		return nil, err
	}
	return sensorList, nil
}
