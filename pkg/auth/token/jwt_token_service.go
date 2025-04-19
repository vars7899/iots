package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/vars7899/iots/config"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
)

type JwtTokenService struct {
	config *config.JwtConfig
	logger *zap.Logger
}

type AccessTokenClaims struct {
	UserID string   `json:"sub"`
	Roles  []string `json:"roles"`

	jwt.RegisteredClaims // includes exp, iat, iss, etc.
}

type RefreshTokenClaims struct {
	UserID string `json:"sub"`

	jwt.RegisteredClaims // includes exp, iat, iss, etc.
}

func NewJwtTokenService(cfg *config.JwtConfig, baseLogger *zap.Logger) TokenService {
	return &JwtTokenService{
		config: cfg,
		logger: logger.Named(baseLogger, "JwtTokenService"),
	}
}

func (j *JwtTokenService) GenerateAccessToken(userID uuid.UUID, roles []string) (string, error) {
	claims := AccessTokenClaims{
		UserID: userID.String(),
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.config.AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.AccessSecret))
}

func (j *JwtTokenService) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	claims := AccessTokenClaims{
		UserID: userID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.config.RefreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.RefreshSecret))
}

func (j *JwtTokenService) ValidateAccessToken(tokenStr string) (*jwt.Token, error) {
	return jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(j.config.AccessSecret), nil
	})
}

func (j *JwtTokenService) ValidateRefreshToken(tokenStr string) (*jwt.Token, error) {
	return jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(j.config.RefreshSecret), nil
	})
}

func (j *JwtTokenService) ParseAccessToken(tokenStr string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &AccessTokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(j.config.AccessSecret), nil
	})
	if err != nil {
		return nil, apperror.ErrInvalidToken.WithMessage("failed to parse access token").Wrap(err)
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok || !token.Valid {
		return nil, apperror.ErrExpiredToken.WithMessage("invalid or expired access token")
	}

	return claims, nil
}

func (j *JwtTokenService) ParseRefreshToken(tokenStr string) (*RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &RefreshTokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(j.config.RefreshSecret), nil
	})
	if err != nil {
		return nil, apperror.ErrInvalidToken.WithMessage("failed to parse refresh token").Wrap(err)
	}

	claims, ok := token.Claims.(*RefreshTokenClaims)
	if !ok || !token.Valid {
		return nil, apperror.ErrExpiredToken.WithMessage("invalid or expired refresh token")
	}

	return claims, nil
}

func (j *JwtTokenService) GetAccessTTL() time.Duration {
	return j.config.AccessTokenTTL
}
func (j *JwtTokenService) GetRefreshTTL() time.Duration {
	return j.config.RefreshTokenTTL
}
