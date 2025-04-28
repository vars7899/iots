package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/vars7899/iots/internal/domain/model"
	"github.com/vars7899/iots/pkg/auth/token"
)

type AuthTokenService interface {
	// JWT
	ParseRefreshToken(tokenStr string) (*token.RefreshTokenClaims, error)
	ValidateRefreshToken(refreshToken string) (*token.RefreshTokenClaims, error)
	ParseAccessToken(tokenStr string) (*token.AccessTokenClaims, error)
	IssueAuthTokenSet(ctx context.Context, userID uuid.UUID, roles []string) (*token.AuthTokenSet, error)
	// JTI
	RevokeJTI(ctx context.Context, jtiStr string, expiresAt time.Time) error
	IsJTIRevoked(ctx context.Context, jti string) (bool, error)
}

type AccessControlService interface {
	Enforce(subject string, object string, action string) (bool, error)
	LoadPolicy() error
	AddRoleForUser(user string, role string) (bool, error)
	DeleteRoleForUser(user string, role string) (bool, error)
	AddPolicy(role string, resource string, action string) (bool, error)
	RemovePolicy(role string, resource string, action string) (bool, error)
	CheckPermission(userID uuid.UUID, resource string, action string) (bool, error)
	SyncUserRoles(user *model.User) error
}
