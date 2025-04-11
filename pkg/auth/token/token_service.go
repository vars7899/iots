package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenService interface {
	GenerateAccessToken(userID uuid.UUID, roles []string) (string, error)
	GenerateRefreshToken(userID uuid.UUID) (string, error)
	ValidateAccessToken(tokenStr string) (*jwt.Token, error)
	ValidateRefreshToken(tokenStr string) (*jwt.Token, error)
	ParseAccessToken(tokenStr string) (*AccessTokenClaims, error)
	GetAccessTTL() time.Duration
	GetRefreshTTL() time.Duration
}
