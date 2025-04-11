package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/vars7899/iots/internal/domain/user"
	"github.com/vars7899/iots/internal/errorz"
	"github.com/vars7899/iots/internal/repository"
)

type UserService struct {
	name string
	repo repository.UserRepository
}

func NewUserService(r repository.UserRepository) *UserService {
	return &UserService{name: "user", repo: r}
}

func (s *UserService) CreateUser(ctx context.Context, u *user.User) (*user.User, error) {
	createUserError := func(err error) error {
		return s.wrapError("create user", err)
	}

	// check if the email is already taken
	isAlreadyTaken, err := s.repo.ExistByEmail(ctx, u.Email)
	if err != nil {
		return nil, createUserError(err)
	}
	if isAlreadyTaken {
		return nil, createUserError(user.ErrEmailAlreadyExists)
	}
	if err := u.SetPassword(u.Password); err != nil {
		return nil, createUserError(err)
	}
	_newUser, err := s.repo.Create(ctx, u)
	if err != nil {
		return nil, createUserError(err)
	}
	return _newUser, nil
}

func (s *UserService) GetUserByID(ctx context.Context, userID uuid.UUID) (*user.User, error) {
	userData, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, s.wrapError("get user by id", err)
	}
	return userData, nil
}

func (s *UserService) UpdateUser(ctx context.Context, userID uuid.UUID, u *user.User) (*user.User, error) {
	userData, err := s.repo.Update(ctx, userID, u)
	if err != nil {
		return nil, s.wrapError("update user", err)
	}
	return userData, nil
}

func (s *UserService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	if err := s.repo.SoftDelete(ctx, userID); err != nil {
		return s.wrapError("delete user", err)
	}
	return nil
}

func (s *UserService) GetUser(ctx context.Context) ([]*user.User, error) {
	userList, err := s.repo.List(ctx, user.UserFilter{})
	if err != nil {
		return nil, s.wrapError("get user", err)
	}
	return userList, nil
}

func (s *UserService) SetLastLogin(ctx context.Context, userID uuid.UUID) error {
	if err := s.repo.SetLastLogin(ctx, userID, time.Now()); err != nil {
		return s.wrapError("set last login", err)
	}
	return nil
}

type LoginIdentifier struct {
	Email       string
	Username    string
	PhoneNumber string
}

// UserService.go
func (s *UserService) FindByLoginIdentifier(ctx context.Context, identifiers LoginIdentifier) (*user.User, error) {
	if identifiers.Email != "" {
		return s.repo.FindByEmail(ctx, identifiers.Email)
	}
	if identifiers.PhoneNumber != "" {
		return s.repo.FindByPhoneNumber(ctx, identifiers.PhoneNumber)
	}
	if identifiers.Username != "" {
		return s.repo.FindByUserName(ctx, identifiers.Username)
	}
	return nil, s.wrapError("find by login identifier", errors.New("no login identifier provided"))
}

func (s *UserService) wrapError(action string, err error) error {
	return errorz.WrapError(s.name, action, err)
}
