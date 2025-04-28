package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/vars7899/iots/internal/domain/model"
	"github.com/vars7899/iots/internal/repository"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
)

type roleService struct {
	roleRepo repository.RoleRepository
	logger   *zap.Logger
	// Add configuration for default role slug
	defaultRoleSlug string
}

func NewRoleService(r repository.RoleRepository, defaultRoleSlug string, baseLogger *zap.Logger) RoleService {
	return &roleService{
		roleRepo:        r,
		logger:          logger.Named(baseLogger, "RoleService"),
		defaultRoleSlug: defaultRoleSlug,
	}
}

func (s *roleService) GetDefaultRoleID(ctx context.Context) (uuid.UUID, error) {
	role, err := s.roleRepo.GetBySlug(ctx, s.defaultRoleSlug)
	if err != nil {
		s.logger.Error("Failed to find default role by slug", zap.String("slug", s.defaultRoleSlug), zap.Error(err))
		return uuid.Nil, apperror.ErrorHandler(err, apperror.ErrCodeInternal, fmt.Sprintf("failed to find default role '%s'", s.defaultRoleSlug))
	}
	return role.ID, nil
}

func (s *roleService) GetRoleBySlug(ctx context.Context, slug string) (*model.Role, error) {
	role, err := s.roleRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, apperror.ErrorHandler(err, apperror.ErrCodeDBQuery, fmt.Sprintf("failed to get role by slug '%s'", slug))
	}
	return role, nil
}

func (s *roleService) GetAllRolesWithPermissions(ctx context.Context) ([]*model.Role, error) {
	roles, err := s.roleRepo.ListRolesWithPermissions(ctx)
	if err != nil {
		return nil, apperror.ErrorHandler(err, apperror.ErrCodeDBQuery, "failed to get all roles with permissions")
	}
	return roles, nil
}
