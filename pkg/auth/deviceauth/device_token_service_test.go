package deviceauth_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/vars7899/iots/config"
	"github.com/vars7899/iots/pkg/auth/deviceauth"
	"go.uber.org/zap"
)

func TestGenerateAndValidateToke(t *testing.T) {
	config := &config.JwtConfig{
		AccessSecret:             "this_is_top_secret",
		RefreshSecret:            "this_is_top_secret_as_well",
		DeviceRefreshTokenTTL:    1 * time.Minute,
		DeviceConnectionTokenTTL: 30 * time.Second,
		Leeway:                   15 * time.Second,
	}
	logger := zap.NewNop()
	service := deviceauth.NewDeviceConnectionTokenService(config, logger)

	deviceID := uuid.New()

	// generate
	genTokens, err := service.GenerateTokens(deviceID)
	assert.NoError(t, err)
	assert.NotEmpty(t, genTokens.ConnectionToken)
	assert.NotEmpty(t, genTokens.RefreshToken)
	assert.NotEmpty(t, genTokens.ConnectionTokenJTI)
	assert.NotEmpty(t, genTokens.RefreshTokenJTI)

	// validate connection claims
	connClaims, err := service.ParseConnectionToken(genTokens.ConnectionToken)
	assert.NoError(t, err)
	claimID, err := connClaims.DeviceID()
	assert.NoError(t, err)
	assert.Equal(t, deviceID.String(), claimID.String())

	// validate refresh claims
	refreshClaims, err := service.ParseRefreshToken(genTokens.RefreshToken)
	assert.NoError(t, err)
	claimID2, err := refreshClaims.DeviceID()
	assert.NoError(t, err)
	assert.Equal(t, deviceID.String(), claimID2.String())
}

func TestExpiredToken(t *testing.T) {
	cfg := &config.JwtConfig{
		DeviceConnectionTokenTTL: -1 * time.Minute, // Already expired
		DeviceRefreshTokenTTL:    30 * time.Minute,
		AccessSecret:             "access-secret-key",
		RefreshSecret:            "refresh-secret-key",
	}
	logger := zap.NewNop()
	service := deviceauth.NewDeviceConnectionTokenService(cfg, logger)

	deviceID := uuid.New()
	tokens, err := service.GenerateTokens(deviceID)
	assert.NoError(t, err)

	_, err = service.ParseConnectionToken(tokens.ConnectionToken)
	assert.Error(t, err)
}
