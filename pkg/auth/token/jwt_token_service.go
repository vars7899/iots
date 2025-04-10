package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JwtTokenService struct {
	accessSecret  string
	refreshSecret string
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewJwtTokenService(accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) TokenService {
	return &JwtTokenService{
		accessSecret:  accessSecret,
		refreshSecret: refreshSecret,
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}

func (j *JwtTokenService) GenerateAccessToken(userID uuid.UUID, roles []string) (string, error) {
	_claims := jwt.MapClaims{
		"sub":   userID.String(),
		"roles": roles,
		"exp":   time.Now().Add(j.accessTTL).Unix(),
		"iat":   time.Now().Unix(),
	}
	_token := jwt.NewWithClaims(jwt.SigningMethodHS256, _claims)
	return _token.SignedString([]byte(j.accessSecret))
}

func (j *JwtTokenService) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	_claims := jwt.MapClaims{
		"sub": userID.String(),
		"exp": time.Now().Add(j.refreshTTL).Unix(),
		"iat": time.Now().Unix(),
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

func (j *JwtTokenService) GetAccessTTL() time.Duration {
	return j.accessTTL
}
