package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vars7899/iots/config"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
)

type JTIStore struct {
	client     *redis.Client
	prefix     string
	defaultTTL time.Duration
	logger     *zap.Logger
}

func NewRedisJTIStore(cfg *config.RedisConfig, baseLogger *zap.Logger) *JTIStore {
	return &JTIStore{
		client:     NewRedisClient(cfg),
		prefix:     cfg.Prefix,
		defaultTTL: cfg.TTL,
		logger:     logger.Named(baseLogger, "RedisStore"),
	}
}

const (
	JTI_ACTIVE_PREFIX  = "active"
	JTI_REVOKED_PREFIX = "revoked"
)

func (s *JTIStore) activeJTIKey(jti string) string {
	return fmt.Sprintf("%s:%s:%s", s.prefix, JTI_ACTIVE_PREFIX, jti)
}

func (s *JTIStore) revokedJTIKey(jti string) string {
	return fmt.Sprintf("%s:%s:%s", s.prefix, JTI_REVOKED_PREFIX, jti)
}

func (s *JTIStore) RecordJTI(ctx context.Context, jti string, expiresAt time.Time) error {
	// Calculate TTL based on token expiration
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		ttl = s.defaultTTL // Token is already expired, use default TTL as fallback
	}

	// Add a buffer for clock skew
	ttl += 5 * time.Minute

	s.logger.Debug("Recording active JTI",
		zap.String("jti", jti),
		zap.Time("expiresAt", expiresAt),
		zap.Duration("ttl", ttl))

	// Store the JTI with its issue timestamp
	key := s.activeJTIKey(jti)
	err := s.client.Set(ctx, key, time.Now().Unix(), ttl).Err()
	if err != nil {
		s.logger.Error("Failed to record JTI", zap.String("jti", jti), zap.Error(err))
		return apperror.ErrInternal.WithMessage("Failed to record token").Wrap(err)
	}

	return nil
}

// RevokeJTI marks a JTI as explicitly revoked
// This should be called during logout or token invalidation
func (s *JTIStore) RevokeJTI(ctx context.Context, jti string, expiresAt time.Time) error {
	// Calculate TTL based on token expiration
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		// Token is already expired, use default TTL as fallback
		ttl = s.defaultTTL
	}

	// Add a buffer for clock skew
	ttl += 5 * time.Minute

	s.logger.Debug("Revoking JTI",
		zap.String("jti", jti),
		zap.Time("expiresAt", expiresAt),
		zap.Duration("ttl", ttl))

	// First check if the JTI exists
	exists, err := s.client.Exists(ctx, s.activeJTIKey(jti)).Result()
	if err != nil {
		s.logger.Error("Failed to check JTI existence", zap.String("jti", jti), zap.Error(err))
		return apperror.ErrInternal.WithMessage("Failed to check token status").Wrap(err)
	}

	fmt.Println("ppppp", exists)

	// If not found, log but don't fail
	if exists == 0 {
		s.logger.Warn("Attempting to revoke unknown JTI", zap.String("jti", jti))
	}

	// Store the JTI in revoked set with revocation timestamp
	key := s.revokedJTIKey(jti)
	err = s.client.Set(ctx, key, time.Now().Unix(), ttl).Err()
	if err != nil {
		s.logger.Error("Failed to revoke JTI", zap.String("jti", jti), zap.Error(err))
		return apperror.ErrInternal.WithMessage("Failed to revoke token").Wrap(err)
	}

	// Optionally delete from active set if it exists
	if exists > 0 {
		if err := s.client.Del(ctx, s.activeJTIKey(jti)).Err(); err != nil {
			s.logger.Warn("Failed to remove JTI from active set", zap.String("jti", jti), zap.Error(err))
			// Continue, not critical
		}
	}

	return nil
}

// IsJTIRevoked checks if a JTI has been revoked
// Returns true if the JTI is revoked or not found in active JTIs
func (s *JTIStore) IsJTIRevoked(ctx context.Context, jti string) (bool, error) {
	// Check if JTI is explicitly revoked
	isRevoked, err := s.client.Exists(ctx, s.revokedJTIKey(jti)).Result()

	if err != nil {
		s.logger.Error("Failed to check if JTI is revoked", zap.String("jti", jti), zap.Error(err))
		return false, apperror.ErrInternal.WithMessage("Token validation failed").Wrap(err)
	}

	if isRevoked > 0 {
		s.logger.Debug("JTI is explicitly revoked", zap.String("jti", jti))
		return true, nil
	}

	// If not explicitly revoked, check if it's an active token
	isActive, err := s.client.Exists(ctx, s.activeJTIKey(jti)).Result()
	if err != nil {
		s.logger.Error("Failed to check if JTI is active", zap.String("jti", jti), zap.Error(err))
		return false, apperror.ErrInternal.WithMessage("Token validation failed").Wrap(err)
	}

	// If token is not in active list, consider it revoked (or unknown)
	if isActive == 0 {
		s.logger.Debug("JTI not found in active tokens", zap.String("jti", jti))
		return true, nil
	}

	return false, nil
}

// CleanupJTIs removes expired JTIs from Redis
// This is optional, as Redis will automatically expire keys based on TTL
func (s *JTIStore) CleanupJTIs(ctx context.Context) error {
	// This is not strictly necessary with Redis TTL, but can be used for maintenance
	s.logger.Debug("JTI cleanup not needed - Redis TTL handles expiration automatically")
	return nil
}

// GetStats retrieves statistics about the JTI store
func (s *JTIStore) GetStats(ctx context.Context) (map[string]interface{}, error) {
	activePattern := fmt.Sprintf("%s:%s:*", s.prefix, JTI_ACTIVE_PREFIX)
	revokedPattern := fmt.Sprintf("%s:%s:*", s.prefix, JTI_REVOKED_PREFIX)

	// Count active JTIs
	var activeCount int64
	var cursor uint64
	var err error
	for {
		var keys []string
		keys, cursor, err = s.client.Scan(ctx, cursor, activePattern, 100).Result()
		if err != nil {
			return nil, err
		}
		activeCount += int64(len(keys))
		if cursor == 0 {
			break
		}
	}

	// Count revoked JTIs
	var revokedCount int64
	cursor = 0
	for {
		var keys []string
		keys, cursor, err = s.client.Scan(ctx, cursor, revokedPattern, 100).Result()
		if err != nil {
			return nil, err
		}
		revokedCount += int64(len(keys))
		if cursor == 0 {
			break
		}
	}

	return map[string]interface{}{
		"active_jtis_count":  activeCount,
		"revoked_jtis_count": revokedCount,
		"total_jtis_count":   activeCount + revokedCount,
	}, nil
}
