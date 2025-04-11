package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/vars7899/iots/internal/domain/user"
)

type UserRepository interface {
	// Basic CRUD
	Create(ctx context.Context, user *user.User) (*user.User, error)                   // ✅
	GetByID(ctx context.Context, userID uuid.UUID) (*user.User, error)                 // ✅
	Update(ctx context.Context, userID uuid.UUID, user *user.User) (*user.User, error) // ✅
	Delete(ctx context.Context, userID uuid.UUID) error                                // ✅
	SoftDelete(ctx context.Context, userID uuid.UUID) error                            // ✅
	Restore(ctx context.Context, userID uuid.UUID) error                               // ✅
	FindByEmail(ctx context.Context, email string) (*user.User, error)                 // ✅
	FindByUserName(ctx context.Context, userName string) (*user.User, error)           // ✅
	FindByPhoneNumber(ctx context.Context, userName string) (*user.User, error)        // ✅
	FindByRoles(ctx context.Context, userID uuid.UUID) (*user.User, error)             // ✅
	SetLastLogin(ctx context.Context, userID uuid.UUID, timestamp time.Time) error     // ✅
	ExistByEmail(ctx context.Context, email string) (bool, error)                      // ✅
	ExistByUserName(ctx context.Context, userName string) (bool, error)                // ✅
	AssignRole(ctx context.Context, userID, roleID uuid.UUID) (*user.User, error)      // ✅
	AssignRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error      // ✅
	RemoveRole(ctx context.Context, userID, roleID uuid.UUID) (*user.User, error)      // ✅
	RemoveRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error      // ✅
	ReplaceRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error     // ✅
	GetRoles(ctx context.Context, userID uuid.UUID) ([]user.Role, error)               // ✅
	GetPermissions(ctx context.Context, userID uuid.UUID) ([]user.Permission, error)   // ✅
	List(ctx context.Context, filter user.UserFilter) ([]*user.User, error)            // ✅
	Count(ctx context.Context, filter user.UserFilter) (int64, error)                  // ✅
	// SetPassword(ctx context.Context, userID uuid.UUID, newHashedPassword string) error
}
