package deviceauth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/vars7899/iots/config"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
)

type deviceConnectionTokenService struct {
	config *config.JwtConfig
	logger *zap.Logger
}

func NewDeviceConnectionTokenService(config *config.JwtConfig, baseLogger *zap.Logger) DeviceTokenService {
	return &deviceConnectionTokenService{
		config: config,
		logger: logger.Named(baseLogger, "DeviceTokenService"),
	}
}

func (s *deviceConnectionTokenService) GenerateTokens(deviceID uuid.UUID) (*DeviceConnectionTokens, error) {
	connectionToken, connectionJTI, err := s.generateConnectionToken(deviceID)
	if err != nil {
		return nil, err
	}
	refreshToken, refreshJTI, err := s.generateRefreshToken(deviceID)
	if err != nil {
		return nil, err
	}

	return &DeviceConnectionTokens{
		ConnectionToken:          connectionToken,
		ConnectionTokenJTI:       connectionJTI,
		ConnectionTokenExpiresAt: time.Now().Add(s.config.DeviceConnectionTokenTTL),
		RefreshToken:             refreshToken,
		RefreshTokenJTI:          refreshJTI,
		RefreshTokenExpiresAt:    time.Now().Add(s.config.DeviceRefreshTokenTTL),
	}, nil
}

func (s *deviceConnectionTokenService) ParseConnectionToken(tokenStr string) (*DeviceConnectionClaims, error) {
	unsigned, err := jwt.ParseWithClaims(tokenStr, &DeviceConnectionClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, apperror.ErrInvalidToken.WithMessage("unexpected signing method")
		}
		return []byte(s.config.AccessSecret), nil
	}, jwt.WithLeeway(s.config.Leeway))
	if err != nil {
		return nil, apperror.ErrInvalidToken.WithMessage("invalid device connection token format").Wrap(err)
	}

	claims, ok := unsigned.Claims.(*DeviceConnectionClaims)
	if !ok || !unsigned.Valid {
		return nil, apperror.ErrExpiredToken.WithMessage("invalid or expired device connection token")
	}
	return claims, nil
}

func (s *deviceConnectionTokenService) ParseRefreshToken(tokenStr string) (*DeviceRefreshClaims, error) {
	unsigned, err := jwt.ParseWithClaims(tokenStr, &DeviceRefreshClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, apperror.ErrInvalidToken.WithMessage("unexpected signing method")
		}
		return []byte(s.config.RefreshSecret), nil
	}, jwt.WithLeeway(s.config.Leeway))
	if err != nil {
		return nil, apperror.ErrInvalidToken.WithMessage("invalid device refresh token format").Wrap(err)
	}

	claims, ok := unsigned.Claims.(*DeviceRefreshClaims)
	if !ok || !unsigned.Valid {
		return nil, apperror.ErrExpiredToken.WithMessage("invalid or expired device refresh token")
	}
	return claims, nil
}

func (s *deviceConnectionTokenService) generateConnectionToken(deviceID uuid.UUID) (string, string, error) {
	jti := uuid.NewString()

	claims := DeviceConnectionClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.DeviceConnectionTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   deviceID.String(),
			ID:        jti,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString([]byte(s.config.AccessSecret))
	if err != nil {
		return "", "", err
	}

	return signed, jti, nil
}

func (s *deviceConnectionTokenService) generateRefreshToken(deviceID uuid.UUID) (string, string, error) {
	jti := uuid.NewString()

	claims := DeviceRefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.DeviceRefreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   deviceID.String(),
			ID:        jti,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString([]byte(s.config.RefreshSecret))
	if err != nil {
		return "", "", err
	}

	return signed, jti, nil
}

func (s *deviceConnectionTokenService) GetConnectionTTL() time.Duration {
	return s.config.DeviceConnectionTokenTTL
}

func (s *deviceConnectionTokenService) GetRefreshTTL() time.Duration {
	return s.config.DeviceRefreshTokenTTL
}

func (s *deviceConnectionTokenService) GetLeeway() time.Duration {
	return s.config.Leeway
}
