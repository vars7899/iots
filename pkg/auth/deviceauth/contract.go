package deviceauth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type DeviceTokenService interface {
	GenerateTokens(deviceID uuid.UUID) (*DeviceConnectionTokens, error)
	ParseConnectionToken(tokenStr string) (*DeviceConnectionClaims, error)
	ParseRefreshToken(tokenStr string) (*DeviceRefreshClaims, error)
	GetConnectionTTL() time.Duration
	GetRefreshTTL() time.Duration
	GetLeeway() time.Duration
}

type DeviceAuthManager interface {
	IssueTokens(ctx context.Context, deviceID uuid.UUID) (*DeviceConnectionTokens, error)
	RotateTokens(ctx context.Context, connectionToken string, refreshToken string) (*DeviceConnectionTokens, error)
	ParseDeviceConnectionTokens(ctx context.Context, connectionToken string) (*DeviceConnectionClaims, error)
	ValidateDeviceConnectionTokens(ctx context.Context, connectionToken string) (*DeviceConnectionClaims, error)
	ParseDeviceRefreshTokens(ctx context.Context, refreshToken string) (*DeviceRefreshClaims, error)
	ValidateDeviceRefreshTokens(ctx context.Context, refreshToken string) (*DeviceRefreshClaims, error)
	RevokeJTI(ctx context.Context, jti string, expiresAt time.Time) error
	IsJTIRevoked(ctx context.Context, jti string) (bool, error)
}
