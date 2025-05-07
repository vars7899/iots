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

type deviceAuthService struct {
	deviceTokenService DeviceTokenService
	jtiStoreService    redis.JTIStore
	logger             *zap.Logger
}

func NewDeviceAuthManager(tokenSrv DeviceTokenService, store redis.JTIStore, baseLogger *zap.Logger) DeviceAuthService {
	return &deviceAuthService{
		deviceTokenService: tokenSrv,
		jtiStoreService:    store,
		logger:             logger.Named(baseLogger, "DeviceAuthService"),
	}
}

func (s *deviceAuthService) IssueTokens(ctx context.Context, deviceID uuid.UUID) (*DeviceConnectionTokens, error) {
	genTokens, err := s.deviceTokenService.GenerateTokens(deviceID)
	if err != nil {
		return nil, apperror.ErrInternal.Wrap(err)
	}

	leeway := s.deviceTokenService.GetLeeway()
	if genTokens.ConnectionTokenJTI == "" || genTokens.RefreshTokenJTI == "" {
		return nil, apperror.ErrInternal.WithMessage("something went wrong while issuing device tokens")
	}
	if err := s.jtiStoreService.RecordJTI(ctx, genTokens.ConnectionTokenJTI, genTokens.ConnectionTokenExpiresAt.Add(leeway)); err != nil {
		return nil, apperror.ErrInternal.Wrap(err)
	}
	if err := s.jtiStoreService.RecordJTI(ctx, genTokens.RefreshTokenJTI, genTokens.RefreshTokenExpiresAt.Add(leeway)); err != nil {
		return nil, apperror.ErrInternal.Wrap(err)
	}
	return genTokens, nil
}

func (s *deviceAuthService) RotateTokens(ctx context.Context, connTokenStr string, refreshTokenStr string) (*DeviceConnectionTokens, error) {
	connClaims, err := s.ParseDeviceConnectionTokens(ctx, connTokenStr)
	if err != nil {
		return nil, err
	}
	refreshClaims, err := s.ValidateDeviceRefreshTokens(ctx, refreshTokenStr)
	if err != nil {
		return nil, err
	}

	if connClaims.Subject != refreshClaims.Subject {
		return nil, apperror.ErrUnauthorized.WithMessage("mismatched device token")
	}

	if err := s.RevokeJTI(ctx, connClaims.ID, connClaims.ExpiresAt.Time); err != nil {
		return nil, err
	}
	if err := s.RevokeJTI(ctx, refreshClaims.ID, refreshClaims.ExpiresAt.Time); err != nil {
		return nil, err
	}

	deviceID, err := refreshClaims.DeviceID()
	if err != nil {
		return nil, apperror.ErrInternal.Wrap(err)
	}

	return s.IssueTokens(ctx, deviceID)
}

func (s *deviceAuthService) ParseDeviceConnectionTokens(ctx context.Context, connTokenStr string) (*DeviceConnectionClaims, error) {
	claims, err := s.deviceTokenService.ParseConnectionToken(connTokenStr)
	if err != nil {
		return nil, apperror.ErrInternal.Wrap(err)
	}
	return claims, nil
}

func (s *deviceAuthService) ValidateDeviceConnectionTokens(ctx context.Context, connTokenStr string) (*DeviceConnectionClaims, error) {
	claims, err := s.deviceTokenService.ParseConnectionToken(connTokenStr)
	if err != nil {
		return nil, apperror.ErrInternal.Wrap(err)
	}
	revoked, err := s.IsJTIRevoked(ctx, claims.ID)
	if err != nil {
		return nil, err
	}
	if revoked {
		return nil, apperror.ErrExpiredToken.WithMessage("device connection token revoked")
	}
	return claims, nil
}

func (s *deviceAuthService) ParseDeviceRefreshTokens(ctx context.Context, refreshTokenStr string) (*DeviceRefreshClaims, error) {
	claims, err := s.deviceTokenService.ParseRefreshToken(refreshTokenStr)
	if err != nil {
		return nil, apperror.ErrInternal.Wrap(err)
	}
	return claims, nil
}

func (s *deviceAuthService) ValidateDeviceRefreshTokens(ctx context.Context, refreshTokenStr string) (*DeviceRefreshClaims, error) {
	claims, err := s.deviceTokenService.ParseRefreshToken(refreshTokenStr)
	if err != nil {
		return nil, apperror.ErrInternal.Wrap(err)
	}
	revoked, err := s.IsJTIRevoked(ctx, claims.ID)
	if err != nil {
		return nil, err
	}
	if revoked {
		return nil, apperror.ErrExpiredToken.WithMessage("device refresh token revoked")
	}
	return claims, nil
}

func (s *deviceAuthService) RevokeJTI(ctx context.Context, jti string, expiresAt time.Time) error {
	if jti == "" {
		return nil
	}
	if err := s.jtiStoreService.RevokeJTI(ctx, jti, expiresAt); err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeDBInsert, "failed to revoke jti")
	}
	return nil
}

func (s *deviceAuthService) IsJTIRevoked(ctx context.Context, jti string) (bool, error) {
	if jti == "" {
		return false, nil
	}
	isRevoked, err := s.jtiStoreService.IsJTIRevoked(ctx, jti)
	if err != nil {
		return false, apperror.ErrorHandler(err, apperror.ErrCodeDBInsert)
	}
	return isRevoked, nil
}
