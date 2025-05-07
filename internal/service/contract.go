package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/vars7899/iots/internal/api/v1/dto"
	"github.com/vars7899/iots/internal/domain/model"
	"github.com/vars7899/iots/pkg/auth/deviceauth"
	"github.com/vars7899/iots/pkg/auth/token"
)

type RoleService interface {
	GetDefaultRoleID(ctx context.Context) (uuid.UUID, error)
	GetRoleBySlug(ctx context.Context, slug string) (*model.Role, error)
	GetAllRolesWithPermissions(ctx context.Context) ([]*model.Role, error)
}

type ResetPasswordTokenService interface {
	CreateToken(ctx context.Context, userID uuid.UUID, expiresIn time.Duration) (*model.ResetPasswordToken, error)
	ValidateToken(ctx context.Context, tokenStr string) (*model.ResetPasswordToken, error)
	DeleteTokensByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteExpiredTokens(ctx context.Context) error
}

type UserService interface {
	CreateUser(ctx context.Context, userData *model.User) (*model.User, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, userData *model.User) (*model.User, error)
	SetPassword(ctx context.Context, userID uuid.UUID, newPasswordHash string) (*model.User, error)
	HardDeleteUser(ctx context.Context, userID uuid.UUID) error
	AssignUserRoles(ctx context.Context, userID uuid.UUID, roles []uuid.UUID) (*model.User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*model.User, error)
	// GetUsers(ctx context.Context, filter dto.UserFilter) ([]*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByUserName(ctx context.Context, username string) (*model.User, error)
	FindByPhoneNumber(ctx context.Context, phoneNumber string) (*model.User, error)
	SetLastLogin(ctx context.Context, userID uuid.UUID) error
}

type AuthService interface {
	RegisterUser(ctx context.Context, userData *model.User) (*model.User, *token.AuthTokenSet, error)
	LoginUser(ctx context.Context, identifiers *dto.LoginCredentials) (*model.User, *token.AuthTokenSet, error)
	RefreshAuthTokens(ctx context.Context, refreshTokenStr string) (*model.User, *token.AuthTokenSet, error)
	LogoutUser(ctx context.Context, userID *uuid.UUID, claims *token.AccessTokenClaims, refreshTokenStr string) error
	RequestPasswordReset(ctx context.Context, email string) (*model.ResetPasswordToken, *string, error)
	ResetPassword(ctx context.Context, resetToken, newRawPassword string) error
}

type DeviceService interface {
	CreateDevice(ctx context.Context, device *model.Device) (*model.Device, error)
	ProvisionDevice(ctx context.Context, idStr string, provisionCode string) (*deviceauth.DeviceConnectionTokens, error)
}
