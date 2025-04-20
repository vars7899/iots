package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/vars7899/iots/internal/api/v1/dto"
	"github.com/vars7899/iots/internal/domain/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) (*model.User, error)                            // ✅
	GetByID(ctx context.Context, userID uuid.UUID) (*model.User, error)                           // ✅
	Update(ctx context.Context, user *model.User) (*model.User, error)                            // ✅
	HardDelete(ctx context.Context, userID uuid.UUID) error                                       // ✅
	Delete(ctx context.Context, userID uuid.UUID) error                                           // ✅
	Restore(ctx context.Context, userID uuid.UUID) error                                          // ✅
	FindByEmail(ctx context.Context, email string) (*model.User, error)                           // ✅
	FindByUserName(ctx context.Context, userName string) (*model.User, error)                     // ✅
	FindByPhoneNumber(ctx context.Context, userName string) (*model.User, error)                  // ✅
	FindByRoles(ctx context.Context, userID uuid.UUID) (*model.User, error)                       // ✅
	SetLastLogin(ctx context.Context, userID uuid.UUID, timestamp time.Time) error                // ✅
	ExistByEmail(ctx context.Context, email string) (bool, error)                                 // ✅
	ExistByUserName(ctx context.Context, userName string) (bool, error)                           // ✅
	ExistByPhoneNumber(ctx context.Context, phoneNumber string) (bool, error)                     // ✅
	AssignRole(ctx context.Context, userID, roleID uuid.UUID) (*model.User, error)                // ✅
	AssignRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) (*model.User, error)  // ✅
	RemoveRole(ctx context.Context, userID, roleID uuid.UUID) (*model.User, error)                // ✅
	RemoveRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) (*model.User, error)  // ✅
	ReplaceRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) (*model.User, error) // ✅
	GetRoles(ctx context.Context, userID uuid.UUID) ([]model.Role, error)                         // ✅
	GetPermissions(ctx context.Context, userID uuid.UUID) ([]*model.Permission, error)            // ✅
	List(ctx context.Context, filter dto.UserFilter) ([]*model.User, error)                       // ✅
	Count(ctx context.Context, filter dto.UserFilter) (int64, error)                              // ✅
	// SetPassword(ctx context.Context, userID uuid.UUID, newHashedPassword string) error
}
