package deviceauth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/vars7899/iots/internal/cache/redis"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
)

type deviceAuthManager struct {
	deviceTokenService DeviceTokenService
	jtiStoreService    redis.JTIStore
	logger             *zap.Logger
}

func NewDeviceAuthManager(tokenSrv DeviceTokenService, store redis.JTIStore, baseLogger *zap.Logger) DeviceAuthManager {
	return &deviceAuthManager{
		deviceTokenService: tokenSrv,
		jtiStoreService:    store,
		logger:             logger.Named(baseLogger, "DeviceAuthManager"),
	}
}

func (m *deviceAuthManager) IssueTokens(ctx context.Context, deviceID uuid.UUID) (*DeviceConnectionTokens, error) {
	genTokens, err := m.deviceTokenService.GenerateTokens(deviceID)
	if err != nil {
		return nil, apperror.ErrInternal.Wrap(err)
	}

	leeway := m.deviceTokenService.GetLeeway()
	if genTokens.ConnectionTokenJTI == "" || genTokens.RefreshTokenJTI == "" {
		return nil, apperror.ErrInternal.WithMessage("something went wrong while issuing device tokens")
	}
	if err := m.jtiStoreService.RecordJTI(ctx, genTokens.ConnectionTokenJTI, genTokens.ConnectionTokenExpiresAt.Add(leeway)); err != nil {
		return nil, apperror.ErrInternal.Wrap(err)
	}
	if err := m.jtiStoreService.RecordJTI(ctx, genTokens.RefreshTokenJTI, genTokens.RefreshTokenExpiresAt.Add(leeway)); err != nil {
		return nil, apperror.ErrInternal.Wrap(err)
	}
	return genTokens, nil
}

func (m *deviceAuthManager) RotateTokens(ctx context.Context, connTokenStr string, refreshTokenStr string) (*DeviceConnectionTokens, error) {
	connClaims, err := m.ParseDeviceConnectionTokens(ctx, connTokenStr)
	if err != nil {
		return nil, err
	}
	refreshClaims, err := m.ValidateDeviceRefreshTokens(ctx, refreshTokenStr)
	if err != nil {
		return nil, err
	}

	if connClaims.Subject != refreshClaims.Subject {
		return nil, apperror.ErrUnauthorized.WithMessage("mismatched device token")
	}

	if err := m.RevokeJTI(ctx, connClaims.ID, connClaims.ExpiresAt.Time); err != nil {
		return nil, err
	}
	if err := m.RevokeJTI(ctx, refreshClaims.ID, refreshClaims.ExpiresAt.Time); err != nil {
		return nil, err
	}

	deviceID, err := refreshClaims.DeviceID()
	if err != nil {
		return nil, apperror.ErrInternal.Wrap(err)
	}

	return m.IssueTokens(ctx, deviceID)
}

func (m *deviceAuthManager) ParseDeviceConnectionTokens(ctx context.Context, connTokenStr string) (*DeviceConnectionClaims, error) {
	claims, err := m.deviceTokenService.ParseConnectionToken(connTokenStr)
	if err != nil {
		return nil, apperror.ErrInternal.Wrap(err)
	}
	return claims, nil
}

func (m *deviceAuthManager) ValidateDeviceConnectionTokens(ctx context.Context, connTokenStr string) (*DeviceConnectionClaims, error) {
	claims, err := m.deviceTokenService.ParseConnectionToken(connTokenStr)
	if err != nil {
		return nil, apperror.ErrInternal.Wrap(err)
	}
	revoked, err := m.IsJTIRevoked(ctx, claims.ID)
	if err != nil {
		return nil, err
	}
	if revoked {
		return nil, apperror.ErrExpiredToken.WithMessage("device connection token revoked")
	}
	return claims, nil
}

func (m *deviceAuthManager) ParseDeviceRefreshTokens(ctx context.Context, refreshTokenStr string) (*DeviceRefreshClaims, error) {
	claims, err := m.deviceTokenService.ParseRefreshToken(refreshTokenStr)
	if err != nil {
		return nil, apperror.ErrInternal.Wrap(err)
	}
	return claims, nil
}

func (m *deviceAuthManager) ValidateDeviceRefreshTokens(ctx context.Context, refreshTokenStr string) (*DeviceRefreshClaims, error) {
	claims, err := m.deviceTokenService.ParseRefreshToken(refreshTokenStr)
	if err != nil {
		return nil, apperror.ErrInternal.Wrap(err)
	}
	revoked, err := m.IsJTIRevoked(ctx, claims.ID)
	if err != nil {
		return nil, err
	}
	if revoked {
		return nil, apperror.ErrExpiredToken.WithMessage("device refresh token revoked")
	}
	return claims, nil
}

func (m *deviceAuthManager) RevokeJTI(ctx context.Context, jti string, expiresAt time.Time) error {
	if jti == "" {
		return nil
	}
	if err := m.jtiStoreService.RevokeJTI(ctx, jti, expiresAt); err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeDBInsert, "failed to revoke jti")
	}
	return nil
}

func (m *deviceAuthManager) IsJTIRevoked(ctx context.Context, jti string) (bool, error) {
	if jti == "" {
		return false, nil
	}
	isRevoked, err := m.jtiStoreService.IsJTIRevoked(ctx, jti)
	if err != nil {
		return false, apperror.ErrorHandler(err, apperror.ErrCodeDBInsert)
	}
	return isRevoked, nil
}
