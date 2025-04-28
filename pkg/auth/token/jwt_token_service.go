package token

import (
	"fmt"
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
	Roles []string `json:"roles"`

	jwt.RegisteredClaims // includes exp, iat, iss, etc.
}

type RefreshTokenClaims struct {
	jwt.RegisteredClaims // includes exp, iat, iss, etc.
}

type AuthTokenSet struct {
	TokenType        string        `json:"token_type"`
	AccessToken      string        `json:"access_token"`
	RefreshToken     string        `json:"refresh_token"`
	RefreshTokenJTI  string        `json:"refresh_token_jti"`
	AccessExpiresAt  time.Time     `json:"access_expires_at"`
	RefreshExpiresAt time.Time     `json:"refresh_expires_at"`
	AccessExpiresIn  time.Duration `json:"access_expires_in"`
	RefreshExpiresIn time.Duration `json:"refresh_expires_in"`
}

func NewJwtTokenService(cfg *config.JwtConfig, baseLogger *zap.Logger) TokenService {
	return &JwtTokenService{
		config: cfg,
		logger: logger.Named(baseLogger, "JwtTokenService"),
	}
}

func (s *JwtTokenService) GenerateAuthTokenSet(userID uuid.UUID, roles []string) (*AuthTokenSet, error) {
	refreshJTI := uuid.NewString()

	accessToken, err := s.GenerateAccessToken(refreshJTI, userID, roles)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.GenerateRefreshToken(refreshJTI, userID)
	if err != nil {
		return nil, err
	}

	return &AuthTokenSet{
		TokenType:        "Bearer",
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		RefreshTokenJTI:  refreshJTI,
		AccessExpiresAt:  time.Now().Add(s.config.AccessTokenTTL),
		RefreshExpiresAt: time.Now().Add(s.config.RefreshTokenTTL),
		AccessExpiresIn:  s.config.AccessTokenTTL,
		RefreshExpiresIn: s.config.RefreshTokenTTL,
	}, nil
}

func (j *JwtTokenService) GenerateAccessToken(jti string, userID uuid.UUID, roles []string) (string, error) {
	claims := AccessTokenClaims{
		Roles: roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.config.AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID.String(),
			ID:        jti, // store JTI in claims for blacklisting

		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.AccessSecret))
}

func (j *JwtTokenService) GenerateRefreshToken(jti string, userID uuid.UUID) (string, error) {
	claims := RefreshTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.config.RefreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID.String(),
			ID:        jti, // store JTI in claims for blacklisting
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(j.config.RefreshSecret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
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
		fmt.Println(err)
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
