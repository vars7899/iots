package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/vars7899/iots/internal/domain/model"
	"github.com/vars7899/iots/internal/repository"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/auth/deviceauth"
	"github.com/vars7899/iots/pkg/logger"
	"github.com/vars7899/iots/pkg/utils"
	"go.uber.org/zap"
)

type deviceService struct {
	deviceRepo        repository.DeviceRepository
	deviceAuthService deviceauth.DeviceAuthService
	log               *zap.Logger
}

type DeviceServiceOpts struct {
}

func NewDeviceService(deviceRepo repository.DeviceRepository, deviceAuthService deviceauth.DeviceAuthService, baseLogger *zap.Logger) DeviceService {
	return &deviceService{
		deviceRepo:        deviceRepo,
		deviceAuthService: deviceAuthService,
		log:               logger.Named(baseLogger, "DeviceService"),
	}
}

func (s *deviceService) PreRegister(ctx context.Context, device *model.Device) (*model.Device, error) {
	return nil, nil
}

func (s *deviceService) CreateDevice(ctx context.Context, device *model.Device) (*model.Device, error) {
	exist, err := s.deviceRepo.ExistByMACAddr(ctx, device.MACAddress)
	if err != nil {
		return nil, apperror.ErrorHandler(err, apperror.ErrCodeInternal, "failed to validate mac address")
	}
	if exist {
		return nil, apperror.ErrDBInsert.WithMessagef("failed to create device: device with %s is either protected or already present", device.MACAddress)
	}

	// // Generate secure initial connection token
	// token, err := utils.GenerateSecureToken(64)
	// if err != nil {
	// 	return nil, apperror.ErrInternal

	// }
	// device.HashConnectionToken(token)

	provisionCode, err := utils.GenerateSecureToken(32)
	if err != nil {
		return nil, apperror.ErrInternal
	}
	s.log.Debug("provision code", zap.String("raw_provision_code", provisionCode))
	device.StoreProvisionCode(provisionCode)

	createdDevice, err := s.deviceRepo.Create(ctx, device)
	if err != nil {
		return nil, ServiceError(err, apperror.ErrCodeDBInsert)
	}
	return createdDevice, nil
}

func (s *deviceService) ProvisionDevice(ctx context.Context, idStr string, provisionCode string) (*deviceauth.DeviceConnectionTokens, error) {
	deviceID, err := uuid.Parse(idStr)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessagef("invalid device ID format: %s", idStr)
	}

	var connectionTokens *deviceauth.DeviceConnectionTokens

	err = s.deviceRepo.Transaction(ctx, func(txRepo repository.DeviceRepository) error {
		device, err := txRepo.GetByID(ctx, deviceID)
		if err != nil {
			return apperror.ErrorHandler(err, apperror.ErrCodeDBQuery)
		}

		if err := device.CompareProvisionCode(provisionCode); err != nil {
			return apperror.ErrInvalidCredentials.WithMessage("invalid provision credentials")
		}

		// update device status as 'provisioned'
		if err := txRepo.MarkAsProvisioned(ctx, device.ID); err != nil {
			return apperror.ErrorHandler(err, apperror.ErrCodeInternal)
		}
		// generate device authentication tokens
		tokens, err := s.deviceAuthService.IssueTokens(ctx, device.ID)
		if err != nil {
			return apperror.ErrorHandler(err, apperror.ErrCodeInternal)
		}
		connectionTokens = tokens
		return nil
	})

	if err != nil {
		return nil, err
	}

	return connectionTokens, nil
}

func (s *deviceService) RefreshDeviceTokens(ctx context.Context, connectionTokenStr string, refreshTokenStr string) (*deviceauth.DeviceConnectionTokens, error) {
	return s.deviceAuthService.RotateTokens(ctx, connectionTokenStr, refreshTokenStr)
}

func (s *deviceService) GetDeviceByID(ctx context.Context, deviceID uuid.UUID) (*model.Device, error) {
	deviceExist, err := s.deviceRepo.GetByID(ctx, deviceID)
	if err != nil {
		return nil, apperror.ErrorHandler(err, apperror.ErrCodeDBQuery, fmt.Sprintf("failed to retrieve device with ID %s", deviceID))
	}
	if deviceExist == nil {
		return nil, apperror.ErrNotFound.WithMessagef("device with ID %s not found", deviceID)
	}
	return deviceExist, nil
}

func (s *deviceService) UpdateDevice(ctx context.Context, deviceUpdates *model.Device) (*model.Device, error) {
	updatedDevice, err := s.deviceRepo.Update(ctx, deviceUpdates)
	if err != nil {
		return nil, ServiceError(err, apperror.ErrDBUpdate)
	}
	return updatedDevice, nil
}

func (s *deviceService) DeleteDevice(ctx context.Context, deviceID uuid.UUID) error {
	deviceExist, err := s.deviceRepo.GetByID(ctx, deviceID)
	if err != nil {
		return ServiceError(err, apperror.ErrDBQuery)
	}
	if deviceExist == nil {
		return ServiceError(apperror.ErrNotFound, fmt.Sprintf("failed to delete device: no device found with id: %s", deviceID))
	}

	if err = s.deviceRepo.Delete(ctx, deviceID); err != nil {
		return ServiceError(err, apperror.ErrDBDelete)
	}
	return nil
}

// func (s *DeviceService) BulkCreateDevices(ctx context.Context, devices []*model.Device) ([]*model.Device, error) {
// 	return s.deviceRepo.BulkCreate(ctx, devices)
// }

// func (s *DeviceService) BulkDeleteDevices(ctx context.Context, ids []uuid.UUID) error {
// 	return s.deviceRepo.BulkDelete(ctx, ids)
// }

// func (s *DeviceService) BulkUpdateDevices(ctx context.Context, devices []*model.Device) ([]*model.Device, error) {
// 	return s.deviceRepo.BulkUpdate(ctx, devices)
// }

// func (s *DeviceService) MarkDeviceAsOnline(ctx context.Context, deviceID uuid.UUID) (*model.Device, error) {
// 	return s.deviceRepo.MarkOnline(ctx, deviceID)
// }
// func (s *DeviceService) MarkDeviceAsOffline(ctx context.Context, deviceID uuid.UUID) (*model.Device, error) {
// 	return s.deviceRepo.MarkOffline(ctx, deviceID)
// }
// func (s *DeviceService) UpdateDeviceStatus(ctx context.Context, deviceID uuid.UUID, status domain.Status) error {
// 	return s.deviceRepo.UpdateStatus(ctx, deviceID, status)
// }
