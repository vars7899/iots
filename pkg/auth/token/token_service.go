package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenService interface {
	GenerateAuthTokenSet(userID uuid.UUID, roles []string) (*AuthTokenSet, error)
	GenerateAccessToken(jti string, userID uuid.UUID, roles []string) (string, error)
	GenerateRefreshToken(jti string, userID uuid.UUID) (string, error)
	ValidateAccessToken(tokenStr string) (*jwt.Token, error)
	ValidateRefreshToken(tokenStr string) (*jwt.Token, error)
	ParseAccessToken(tokenStr string) (*AccessTokenClaims, error)
	ParseRefreshToken(tokenStr string) (*RefreshTokenClaims, error)
	GetAccessTTL() time.Duration
	GetRefreshTTL() time.Duration
}
