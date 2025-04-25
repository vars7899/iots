package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vars7899/iots/internal/api/v1/dto"
	"github.com/vars7899/iots/internal/domain/model"
	"github.com/vars7899/iots/internal/repository"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
)

type UserService struct {
	userRepo      repository.UserRepository
	roleService   RoleService
	casbinService CasbinService
	logger        *zap.Logger
}

func NewUserService(r repository.UserRepository, roleService RoleService, casbinService CasbinService, baseLogger *zap.Logger) *UserService {
	return &UserService{
		userRepo:      r,
		roleService:   roleService,
		casbinService: casbinService,
		logger:        logger.Named(baseLogger, "UserService"),
	}
}

func (s *UserService) CreateUser(ctx context.Context, userData *model.User) (*model.User, error) {
	// check email
	emailAlreadyTaken, err := s.userRepo.ExistByEmail(ctx, userData.Email)
	if err != nil {
		return nil, apperror.ErrorHandler(err, apperror.ErrCodeDBQuery, "failed to create user")
	}
	if emailAlreadyTaken {
		return nil, apperror.ErrEmailAlreadyExist.WithMessage("failed to create user: email already taken")
	}
	// check username
	usernameAlreadyTaken, err := s.userRepo.ExistByUserName(ctx, userData.Username)
	if err != nil {
		return nil, apperror.ErrorHandler(err, apperror.ErrCodeDBQuery, "failed to create user")
	}
	if usernameAlreadyTaken {
		return nil, apperror.ErrUsernameAlreadyExists.WithMessage("failed to create user: username already taken")
	}
	// check phone number
	phoneNumberAlreadyTaken, err := s.userRepo.ExistByPhoneNumber(ctx, userData.PhoneNumber)
	if err != nil {
		return nil, apperror.ErrorHandler(err, apperror.ErrCodeDBQuery, "failed to create user")
	}
	if phoneNumberAlreadyTaken {
		return nil, apperror.ErrPhoneNumberAlreadyExists.WithMessage("failed to create user: phone number already taken")
	}
	// hash password
	if err := userData.SetPassword(userData.Password); err != nil {
		return nil, apperror.ErrorHandler(err, apperror.ErrCodeInternal, "failed to create user: unable to store credentials")
	}
	// create user row
	createdUser, err := s.userRepo.Create(ctx, userData)
	if err != nil {
		return nil, apperror.ErrorHandler(err, apperror.ErrCodeInternal, "failed to create user")
	}

	defaultRoleID, err := s.roleService.GetDefaultRoleID(ctx)
	if err != nil {
		s.logger.Error("Failed to assign default role", zap.String("userID", createdUser.ID.String()), zap.Error(err))
		rollbackErr := s.userRepo.HardDelete(ctx, createdUser.ID)
		if rollbackErr != nil {
			s.logger.Error("Failed to rollback user creation after role assignment error", zap.String("userID", createdUser.ID.String()), zap.Error(rollbackErr))
		}
		return nil, apperror.ErrInternal.Wrap(err).WithMessage("failed to assign default role to user")
	}

	userWithRoles, err := s.userRepo.AssignRoles(ctx, createdUser.ID, []uuid.UUID{defaultRoleID})
	if err != nil {
		s.logger.Error("Failed to assign default role in repository", zap.String("userID", createdUser.ID.String()), zap.String("roleID", defaultRoleID.String()), zap.Error(err))
		rollbackErr := s.userRepo.HardDelete(ctx, createdUser.ID) // Example: attempt rollback
		if rollbackErr != nil {
			s.logger.Error("Failed to rollback user creation after role assignment error", zap.String("userID", createdUser.ID.String()), zap.Error(rollbackErr))
		}
		return nil, apperror.ErrorHandler(err, apperror.ErrCodeDBInsert, "failed to assign default role to user")
	}

	if err := s.casbinService.SyncUserRoles(userWithRoles); err != nil {
		s.logger.Error("Failed to sync user roles with Casbin after creation",
			zap.String("userID", userWithRoles.ID.String()),
			zap.Error(err))
		// Decide how to handle this error. It means the user is created and has roles in DB,
		// but Casbin's view might be out of sync. This is serious for auth.
		// You might want to log, alert, and potentially disable the user account until fixed.
		// For now, we'll log and let the user creation proceed, but be aware of the auth risk.
		// A background worker could periodically sync or you could implement retry logic.
	}

	return userWithRoles, nil
}

func (s *UserService) GetUserByID(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	userData, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, apperror.ErrorHandler(err, apperror.ErrCodeDBQuery, fmt.Sprintf("failed to fetch user data with ID %s", userID))
	}
	return userData, nil
}

func (s *UserService) UpdateUser(ctx context.Context, userID uuid.UUID, userData *model.User) (*model.User, error) {
	userData, err := s.userRepo.Update(ctx, userData)
	if err != nil {
		return nil, apperror.ErrorHandler(err, apperror.ErrCodeDBUpdate, "failed to update user data")
	}
	return userData, nil
}

func (s *UserService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeDBQuery, fmt.Sprintf("failed to delete user with ID %s", userID))
	}
	return nil
}

func (s *UserService) GetUser(ctx context.Context, filter dto.UserFilter) ([]*model.User, error) {
	userList, err := s.userRepo.List(ctx, filter)
	if err != nil {
		return nil, apperror.ErrorHandler(err, apperror.ErrCodeDBQuery, "failed to get user")
	}
	return userList, nil
}

func (s *UserService) SetLastLogin(ctx context.Context, userID uuid.UUID) error {
	if err := s.userRepo.SetLastLogin(ctx, userID, time.Now()); err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeDBQuery, fmt.Sprintf("failed to set last login for user with ID %s", userID))
	}
	return nil
}

type LoginIdentifier struct {
	Email       string
	Username    string
	PhoneNumber string
}

// UserService.go
func (s *UserService) FindByLoginIdentifier(ctx context.Context, identifiers LoginIdentifier) (*model.User, error) {
	if identifiers.Email != "" {
		return s.userRepo.FindByEmail(ctx, identifiers.Email)
	}
	if identifiers.PhoneNumber != "" {
		return s.userRepo.FindByPhoneNumber(ctx, identifiers.PhoneNumber)
	}
	if identifiers.Username != "" {
		return s.userRepo.FindByUserName(ctx, identifiers.Username)
	}
	return nil, apperror.ErrInvalidCredentials.WithMessage("missing login identifier")
}
