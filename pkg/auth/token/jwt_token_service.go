package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/vars7899/iots/internal/errorz"
)

type JwtTokenService struct {
	name          string
	accessSecret  string
	refreshSecret string
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

type AccessTokenClaims struct {
	UserID string   `json:"sub"`
	Roles  []string `json:"roles"`

	jwt.RegisteredClaims // includes exp, iat, iss, etc.
}

func NewJwtTokenService(accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) TokenService {
	return &JwtTokenService{
		name:          "token.jwt_token_service",
		accessSecret:  accessSecret,
		refreshSecret: refreshSecret,
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}

func (j *JwtTokenService) GenerateAccessToken(userID uuid.UUID, roles []string) (string, error) {
	_claims := AccessTokenClaims{
		UserID: userID.String(),
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID.String(),
		},
	}

	_token := jwt.NewWithClaims(jwt.SigningMethodHS256, _claims)
	return _token.SignedString([]byte(j.accessSecret))
}

func (j *JwtTokenService) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	_claims := AccessTokenClaims{
		UserID: userID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.refreshTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID.String(),
		},
	}

	_token := jwt.NewWithClaims(jwt.SigningMethodHS256, _claims)
	return _token.SignedString([]byte(j.refreshSecret))
}

func (j *JwtTokenService) ValidateAccessToken(tokenStr string) (*jwt.Token, error) {
	return jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(j.accessSecret), nil
	})
}

func (j *JwtTokenService) ValidateRefreshToken(tokenStr string) (*jwt.Token, error) {
	return jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(j.refreshSecret), nil
	})
}

func (j *JwtTokenService) ParseAccessToken(tokenStr string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &AccessTokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(j.accessSecret), nil
	})
	if err != nil {
		return nil, j.wrapError("parse access token", err)
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok || !token.Valid {
		return nil, j.wrapError("parse access token", errors.New("invalid or expired token"))
	}

	return claims, nil
}

func (j *JwtTokenService) GetAccessTTL() time.Duration {
	return j.accessTTL
}
func (j *JwtTokenService) GetRefreshTTL() time.Duration {
	return j.refreshTTL
}
func (j *JwtTokenService) wrapError(action string, err error) error {
	return errorz.WrapError(j.name, action, err)
}
