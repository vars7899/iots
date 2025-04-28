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

type userService struct {
	userRepo repository.UserRepository
	logger   *zap.Logger
}

func NewUserService(userRepo repository.UserRepository, baseLogger *zap.Logger) UserService {
	return &userService{
		userRepo: userRepo,
		logger:   logger.Named(baseLogger, "UserService"),
	}
}

func (s *userService) CreateUser(ctx context.Context, user *model.User) (*model.User, error) {
	if err := s.checkEmailNotTaken(ctx, user.Email); err != nil {
		return nil, err
	}
	if err := s.checkUsernameNotTaken(ctx, user.Username); err != nil {
		return nil, err
	}
	if err := s.checkPhoneNumberNotTaken(ctx, user.PhoneNumber); err != nil {
		return nil, err
	}
	if err := user.HashPassword(user.Password); err != nil {
		return nil, apperror.ErrorHandler(err, apperror.ErrCodeInternal, "failed to set user password")
	}
	createdUser, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, apperror.ErrorHandler(err, apperror.ErrCodeDBInsert, "failed to create user in database")
	}
	return createdUser, nil
}

func (s *userService) GetUserByID(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	userData, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, apperror.ErrorHandler(err, apperror.ErrCodeDBQuery, fmt.Sprintf("failed to fetch user data with ID %s", userID))
	}
	return userData, nil
}

func (s *userService) UpdateUser(ctx context.Context, userID uuid.UUID, userData *model.User) (*model.User, error) {
	userData, err := s.userRepo.Update(ctx, userData)
	if err != nil {
		return nil, apperror.ErrorHandler(err, apperror.ErrCodeDBUpdate, "failed to update user data")
	}
	return userData, nil
}

func (s *userService) SetPassword(ctx context.Context, userID uuid.UUID, newHashPassword string) (*model.User, error) {
	userData, err := s.userRepo.SetPassword(ctx, userID, newHashPassword)
	if err != nil {
		return nil, apperror.ErrorHandler(err, apperror.ErrCodeDBUpdate, "failed to set new password")
	}
	return userData, nil
}

func (s *userService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeDBQuery, fmt.Sprintf("failed to delete user with ID %s", userID))
	}
	return nil
}

func (s *userService) GetUser(ctx context.Context, filter dto.UserFilter) ([]*model.User, error) {
	userList, err := s.userRepo.List(ctx, filter)
	if err != nil {
		return nil, apperror.ErrorHandler(err, apperror.ErrCodeDBQuery, "failed to get user")
	}
	return userList, nil
}

func (s *userService) SetLastLogin(ctx context.Context, userID uuid.UUID) error {
	if err := s.userRepo.SetLastLogin(ctx, userID, time.Now()); err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeDBQuery, fmt.Sprintf("failed to set last login for user with ID %s", userID))
	}
	return nil
}

func (s *userService) AssignUserRoles(ctx context.Context, userID uuid.UUID, roles []uuid.UUID) (*model.User, error) {
	return s.userRepo.AssignRoles(ctx, userID, roles)
}

// func (s *userService) RequestPasswordReset(ctx context.Context, email string) (*model.ResetPasswordToken, error) {
// 	user, err := s.userRepo.FindByEmail(ctx, email)
// 	if err != nil {
// 		return nil, apperror.ErrorHandler(err, apperror.ErrCodeDBQuery, "failed to fetch user by email")
// 	}
// 	if user == nil {
// 		return nil, apperror.ErrNotFound.WithMessage("user not found")
// 	}

// 	token, err := s.resetPasswordTokenService.CreateToken(ctx, user.ID, 15*time.Minute)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// later you can send email to user with token.Token here
// 	return token, nil
// }

func (s *userService) HardDeleteUser(ctx context.Context, userID uuid.UUID) error {
	return s.userRepo.HardDelete(ctx, userID)
}

func (s *userService) checkEmailNotTaken(ctx context.Context, email string) error {
	taken, err := s.userRepo.ExistByEmail(ctx, email)
	if err != nil {
		// Wrap the DB error here with context
		return apperror.ErrorHandler(err, apperror.ErrCodeDBQuery, "failed to check email existence")
	}
	if taken {
		// Return the specific application error
		return apperror.ErrEmailAlreadyExist.WithMessage("email already taken")
	}
	return nil
}

func (s *userService) checkUsernameNotTaken(ctx context.Context, username string) error {
	taken, err := s.userRepo.ExistByUserName(ctx, username)
	if err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeDBQuery, "failed to check username existence")
	}
	if taken {
		return apperror.ErrUsernameAlreadyExists.WithMessage("username already taken")
	}
	return nil
}

func (s *userService) checkPhoneNumberNotTaken(ctx context.Context, phoneNumber string) error {
	taken, err := s.userRepo.ExistByPhoneNumber(ctx, phoneNumber)
	if err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeDBQuery, "failed to check phone number existence")
	}
	if taken {
		return apperror.ErrPhoneNumberAlreadyExists.WithMessage("phone number already taken")
	}
	return nil
}

func (s *userService) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	return s.userRepo.FindByEmail(ctx, email)
}
func (s *userService) FindByUserName(ctx context.Context, username string) (*model.User, error) {
	return s.userRepo.FindByUserName(ctx, username)
}
func (s *userService) FindByPhoneNumber(ctx context.Context, phoneNumber string) (*model.User, error) {
	return s.userRepo.FindByPhoneNumber(ctx, phoneNumber)
}
