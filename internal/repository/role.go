package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/vars7899/iots/internal/domain/model"
)

type RoleRepository interface {
	GetByID(ctx context.Context, roleID uuid.UUID) (*model.Role, error)
	GetBySlug(ctx context.Context, slug string) (*model.Role, error)
	Create(ctx context.Context, roleData *model.Role) (*model.Role, error)
	Update(ctx context.Context, roleData *model.Role) (*model.Role, error)
	Delete(ctx context.Context, roleID uuid.UUID) error
	List(ctx context.Context) ([]*model.Role, error)
	ListRolesWithPermissions(ctx context.Context) ([]*model.Role, error)
}
