package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/vars7899/iots/internal/domain"
	"github.com/vars7899/iots/internal/domain/model"
	"github.com/vars7899/iots/internal/repository"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
)

type DeviceService struct {
	deviceRepo repository.DeviceRepository
	log        *zap.Logger
}

type DeviceServiceOpts struct {
}

func NewDeviceService(repo repository.DeviceRepository, baseLogger *zap.Logger) *DeviceService {
	return &DeviceService{deviceRepo: repo, log: logger.NewNamedZapLogger(baseLogger, "service.DeviceService")}
}

func (s *DeviceService) CreateDevice(ctx context.Context, d *model.Device) (*model.Device, error) {
	createdDevice, err := s.deviceRepo.Create(ctx, d)
	if err != nil {
		return nil, ServiceError(err, apperror.ErrCodeDBInsert)
	}
	return createdDevice, nil
}

func (s *DeviceService) GetDeviceByID(ctx context.Context, deviceID uuid.UUID) (*model.Device, error) {
	deviceExist, err := s.deviceRepo.GetByID(ctx, deviceID)
	if err != nil {
		return nil, apperror.ErrorHandler(err, apperror.ErrCodeDBQuery, fmt.Sprintf("failed to retrieve device with ID %s", deviceID))
	}
	if deviceExist == nil {
		return nil, apperror.ErrNotFound.WithMessagef("device with ID %s not found", deviceID)
	}
	return deviceExist, nil
}

func (s *DeviceService) UpdateDevice(ctx context.Context, deviceID uuid.UUID, deviceUpdates *model.Device) (*model.Device, error) {
	updatedDevice, err := s.deviceRepo.Update(ctx, deviceID, deviceUpdates)
	if err != nil {
		return nil, ServiceError(err, apperror.ErrDBUpdate)
	}
	return updatedDevice, nil
}

func (s *DeviceService) DeleteDevice(ctx context.Context, deviceID uuid.UUID) error {
	deviceExist, err := s.deviceRepo.GetByID(ctx, deviceID)
	if err != nil {
		return ServiceError(err, apperror.ErrDBQuery)
	}
	if deviceExist == nil {
		return ServiceError(apperror.ErrNotFound, fmt.Sprintf("failed to delete device: no device found with id: %s", deviceID))
	}

	if err = s.deviceRepo.SoftDelete(ctx, deviceID); err != nil {
		return ServiceError(err, apperror.ErrDBDelete)
	}
	return nil
}

func (s *DeviceService) BulkCreateDevices(ctx context.Context, devices []*model.Device) ([]*model.Device, error) {
	return s.deviceRepo.BulkCreate(ctx, devices)
}

func (s *DeviceService) BulkDeleteDevices(ctx context.Context, ids []uuid.UUID) error {
	return s.deviceRepo.BulkDelete(ctx, ids)
}

func (s *DeviceService) BulkUpdateDevices(ctx context.Context, devices []*model.Device) ([]*model.Device, error) {
	return s.deviceRepo.BulkUpdate(ctx, devices)
}

func (s *DeviceService) MarkDeviceAsOnline(ctx context.Context, deviceID uuid.UUID) (*model.Device, error) {
	return s.deviceRepo.MarkOnline(ctx, deviceID)
}
func (s *DeviceService) MarkDeviceAsOffline(ctx context.Context, deviceID uuid.UUID) (*model.Device, error) {
	return s.deviceRepo.MarkOffline(ctx, deviceID)
}
func (s *DeviceService) UpdateDeviceStatus(ctx context.Context, deviceID uuid.UUID, status domain.Status) error {
	return s.deviceRepo.UpdateStatus(ctx, deviceID, status)
}
