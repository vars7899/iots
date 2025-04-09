package service

import (
	"context"

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

func (s *UserService) CreateUser(ctx context.Context, u *user.User) error {
	createUserError := func(err error) error {
		return s.wrapError("create user", err)
	}

	// check if the email is already taken
	isAlreadyTaken, err := s.repo.ExistByEmail(ctx, u.Email)
	if err != nil {
		return createUserError(err)
	}
	if isAlreadyTaken {
		return createUserError(user.ErrEmailAlreadyExists)
	}
	if err := s.repo.Create(ctx, u); err != nil {
		return createUserError(err)
	}
	return nil
}

func (s *UserService) GetUser(ctx context.Context, userID uuid.UUID) (*user.User, error) {
	userData, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, s.wrapError("get user", err)
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

func (s *UserService) wrapError(action string, err error) error {
	return errorz.WrapError(s.name, action, err)
}
