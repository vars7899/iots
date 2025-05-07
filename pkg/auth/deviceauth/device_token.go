package deviceauth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type DeviceConnectionTokens struct {
	ConnectionToken          string    `json:"connection_token"`
	RefreshToken             string    `json:"refresh_token"`
	ConnectionTokenExpiresAt time.Time `json:"connection_token_expires_at"`
	RefreshTokenExpiresAt    time.Time `json:"refresh_token_expires_at"`
	ConnectionTokenJTI       string    `json:"connection_token_jti"`
	RefreshTokenJTI          string    `json:"refresh_token_jti"`
}

type DeviceConnectionClaims struct {
	jwt.RegisteredClaims
}

func (c *DeviceConnectionClaims) DeviceID() (uuid.UUID, error) {
	return uuid.Parse(c.Subject)
}

type DeviceRefreshClaims struct {
	jwt.RegisteredClaims
}

func (c *DeviceRefreshClaims) DeviceID() (uuid.UUID, error) {
	return uuid.Parse(c.Subject)
}
