package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/vars7899/iots/internal/domain"
	"github.com/vars7899/iots/internal/domain/model"
	"github.com/vars7899/iots/internal/repository"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/logger"
	"github.com/vars7899/iots/pkg/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RoleRepositoryPostgres struct {
	db *gorm.DB
	l  *zap.Logger
}

func NewRoleRepositoryPostgres(db *gorm.DB, baseLogger *zap.Logger) repository.RoleRepository {
	return &RoleRepositoryPostgres{
		db: db,
		l:  logger.Named(baseLogger, "RoleRepositoryPostgres"),
	}
}

func (r *RoleRepositoryPostgres) GetByID(ctx context.Context, roleID uuid.UUID) (*model.Role, error) {
	var role model.Role
	if err := r.db.WithContext(ctx).Preload("Permissions").Where("id = ?", roleID).First(&role).Error; err != nil {
		return nil, apperror.MapDBError(err, domain.EntityRole)
	}
	return &role, nil
}

func (r *RoleRepositoryPostgres) GetBySlug(ctx context.Context, slug string) (*model.Role, error) {
	var role model.Role
	if err := r.db.WithContext(ctx).Preload("Permissions").Where("slug = ?", slug).First(&role).Error; err != nil {
		return nil, apperror.MapDBError(err, domain.EntityRole)
	}
	return &role, nil
}

func (r *RoleRepositoryPostgres) Create(ctx context.Context, roleData *model.Role) (*model.Role, error) {
	if err := r.db.WithContext(ctx).Model(&model.Role{}).Clauses(clause.Returning{}).Create(&roleData).Error; err != nil {
		return nil, apperror.MapDBError(err, domain.EntityRole)
	}
	return roleData, nil
}

func (r *RoleRepositoryPostgres) Update(ctx context.Context, roleData *model.Role) (*model.Role, error) {
	var updatedRole model.Role
	tx := r.db.WithContext(ctx).Model(&model.Role{}).Clauses(clause.Returning{}).Where("id = ?", roleData.ID).Updates(roleData).Scan(&roleData)
	if tx.Error != nil {
		return nil, apperror.MapDBError(tx.Error, domain.EntityRole)
	}
	if tx.RowsAffected == 0 {
		return nil, apperror.ErrNotFound.WithMessagef("cannot update %s: no matching record found", domain.EntitySensor)
	}
	return &updatedRole, nil
}

func (r *RoleRepositoryPostgres) Delete(ctx context.Context, roleID uuid.UUID) error {
	var role model.Role
	if err := r.db.WithContext(ctx).Where("id = ?", roleID).First(&role).Error; err != nil {
		return apperror.MapDBError(err, domain.EntityRole)
	}
	if role.IsProtected {
		return apperror.ErrForbidden.WithMessage("missing permissions")
	}

	tx := r.db.WithContext(ctx).Delete(&model.Role{}, roleID)
	if tx.Error != nil {
		return apperror.MapDBError(tx.Error, domain.EntityRole)
	}
	if tx.RowsAffected == 0 {
		return apperror.ErrNotFound.WithMessagef("cannot delete %s: no matching record found after check", domain.EntityRole)
	}
	return nil
}

func (r *RoleRepositoryPostgres) List(ctx context.Context) ([]*model.Role, error) {
	var roles []model.Role
	if err := r.db.WithContext(ctx).Model(&model.Role{}).Find(&roles).Error; err != nil {
		return nil, apperror.MapDBError(err, domain.EntityRole)
	}
	return utils.ConvertVectorToPointerVector(roles), nil
}

func (r *RoleRepositoryPostgres) ListRolesWithPermissions(ctx context.Context) ([]*model.Role, error) {
	var roles []model.Role
	if err := r.db.WithContext(ctx).Preload("Permissions").Find(&roles).Error; err != nil {
		return nil, apperror.MapDBError(err, domain.EntityRole)
	}
	return utils.ConvertVectorToPointerVector(roles), nil
}
