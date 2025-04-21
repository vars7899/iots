package cache

import (
	"context"
	"time"
)

type JTIStore interface {
	RecordJTI(ctx context.Context, jti string, expiresAt time.Time) error
	RevokeJTI(ctx context.Context, jti string, expiresAt time.Time) error
	IsJTIRevoked(ctx context.Context, jti string) (bool, error)
	CleanupJTIs(ctx context.Context) error
	GetStats(ctx context.Context) (map[string]interface{}, error)
}

// type JTIStore interface {
// 	RecordJTI(ctx context.Context, jti string, ttl time.Duration) error
// 	RevokeJTI(ctx context.Context, jti string) error
// 	IsJTIRevoked(ctx context.Context, jti string) (bool, error)
// 	ListUsedJTIs(ctx context.Context) ([]string, error)
// 	ListRevokedJTIs(ctx context.Context) ([]string, error)
// 	IsTokenRevoked(ctx context.Context, jti string) (bool, error)
// 	RevokeToken(ctx context.Context, jti string, expiryTime time.Time) error
// }
