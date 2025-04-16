package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vars7899/iots/internal/domain/user"
	"github.com/vars7899/iots/internal/repository"
	"github.com/vars7899/iots/pkg/logger"
	"github.com/vars7899/iots/pkg/validatorutils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserRepositoryPostgres struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewUserRepositoryPostgres(db *gorm.DB) repository.UserRepository {
	_logger := logger.Lgr.Named("UserRepositoryPostgres")
	return &UserRepositoryPostgres{db: db, log: _logger}
}

func (r *UserRepositoryPostgres) GetByID(ctx context.Context, userID uuid.UUID) (*user.User, error) {
	var user user.User
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		r.log.Error("failed to get user by ID", zap.String("user_id", userID.String()), zap.Error(err))
		return nil, repository.ErrInternal
	}
	return &user, nil
}

func (r *UserRepositoryPostgres) Create(ctx context.Context, u *user.User) (*user.User, error) {
	if err := r.db.WithContext(ctx).Model(&user.User{}).Create(u).Error; err != nil {
		if pgErr := validatorutils.IsPgDuplicateKeyError(err); pgErr != nil {
			return nil, repository.ErrDuplicateKey
		}
		r.log.Error("failed to create new user", zap.Any("user", u), zap.Error(err))
		return nil, repository.ErrInternal
	}
	return r.FindByEmail(ctx, u.Email)
}

func (r *UserRepositoryPostgres) Update(ctx context.Context, userID uuid.UUID, u *user.User) (*user.User, error) {
	result := r.db.WithContext(ctx).Model(&user.User{}).Where("id = ?", userID).Select("*").Updates(u)

	if result.RowsAffected == 0 {
		return nil, repository.ErrNotFound
	}
	if result.Error != nil {
		r.log.Error("failed to update user", zap.String("user_id", userID.String()), zap.Any("user_update", u), zap.Error(result.Error))
		return nil, repository.ErrInternal
	}

	var updatedUser user.User
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&updatedUser).Error; err != nil {
		r.log.Error("failed to get updated user", zap.String("user_id", userID.String()), zap.Error(err))
		return nil, repository.ErrInternal
	}
	return &updatedUser, nil
}

func (r *UserRepositoryPostgres) Delete(ctx context.Context, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).Unscoped().Where("id = ?", userID).Delete(&user.User{})

	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	if result.Error != nil {
		r.log.Error("failed to delete user", zap.String("user_id", userID.String()), zap.Error(result.Error))
		return repository.ErrInternal
	}
	return nil
}

func (r *UserRepositoryPostgres) SoftDelete(ctx context.Context, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", userID).Delete(&user.User{})

	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	if result.Error != nil {
		r.log.Error("failed to delete user", zap.String("user_id", userID.String()), zap.Error(result.Error))
		return repository.ErrInternal
	}
	return nil
}

func (r *UserRepositoryPostgres) Restore(ctx context.Context, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&user.User{}).Unscoped().Where("id = ?", userID).Update("deleted_at", nil)

	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	if result.Error != nil {
		r.log.Error("failed to restore user", zap.String("user_id", userID.String()), zap.Error(result.Error))
		return repository.ErrInternal
	}
	return nil
}

func (r *UserRepositoryPostgres) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	var u user.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		r.log.Error("failed to find user by email", zap.String("user_email", email), zap.Error(err))
		return nil, repository.ErrInternal
	}
	return &u, nil
}

func (r *UserRepositoryPostgres) FindByUserName(ctx context.Context, userName string) (*user.User, error) {
	var u user.User
	if err := r.db.WithContext(ctx).Where("username = ?", userName).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		r.log.Error("failed to find user by username", zap.String("user_username", userName), zap.Error(err))
		return nil, repository.ErrInternal
	}
	return &u, nil
}

func (r *UserRepositoryPostgres) FindByPhoneNumber(ctx context.Context, phoneNumber string) (*user.User, error) {
	var u user.User
	if err := r.db.WithContext(ctx).Where("phone_number = ?", phoneNumber).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		r.log.Error("failed to find user by username", zap.String("user_phone_number", phoneNumber), zap.Error(err))
		return nil, repository.ErrInternal
	}
	return &u, nil
}

func (r *UserRepositoryPostgres) FindByRoles(ctx context.Context, userID uuid.UUID) (*user.User, error) {
	var u user.User
	if err := r.db.WithContext(ctx).Preload("Roles.Permissions").First(&u, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		r.log.Error("failed to find user by roles", zap.String("user_id", userID.String()), zap.Error(err))
		return nil, repository.ErrInternal
	}
	return &u, nil
}

func (r *UserRepositoryPostgres) SetLastLogin(ctx context.Context, userID uuid.UUID, timestamp time.Time) error {
	result := r.db.WithContext(ctx).Model(&user.User{}).Where("id = ?", userID).Update("last_login", timestamp)

	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	if result.Error != nil {
		r.log.Error("failed to check user exist by email", zap.String("user_id", userID.String()), zap.Error(result.Error))
		return repository.ErrInternal
	}
	return nil
}

func (r *UserRepositoryPostgres) ExistByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&user.User{}).Where("email = ?", email).Count(&count)
	if result.Error != nil {
		r.log.Error("failed to check user exist by email", zap.String("user_email", email), zap.Error(result.Error))
		return false, repository.ErrInternal
	}

	return count > 0, nil
}

func (r *UserRepositoryPostgres) ExistByUserName(ctx context.Context, username string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Where("username = ?", username).Count(&count)
	if result.Error != nil {
		r.log.Error("failed to check user exist by email", zap.String("user_username", username), zap.Error(result.Error))
		return false, repository.ErrInternal
	}

	return count > 0, nil
}

func (r *UserRepositoryPostgres) AssignRole(ctx context.Context, userID, roleID uuid.UUID) (*user.User, error) {
	_user := &user.User{ID: userID}
	_role := &user.Role{ID: roleID}
	if err := r.db.WithContext(ctx).Model(_user).Association("Roles").Append(_role); err != nil {
		r.log.Error("failed to assign role to user", zap.String("user_id", userID.String()), zap.String("role_id", roleID.String()), zap.Error(err))
		return nil, repository.ErrInternal
	}

	return r.GetByID(ctx, userID)
}

func (r *UserRepositoryPostgres) AssignRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error {
	_user := &user.User{ID: userID}
	var _roles []*user.Role

	for _, rid := range roleIDs {
		_roles = append(_roles, &user.Role{ID: rid})
	}

	err := r.db.WithContext(ctx).Model(_user).Association("Roles").Append(_roles)
	if err != nil {
		r.log.Error("failed to assign roles to user", zap.String("user_id", userID.String()), zap.Any("role_ids", roleIDs), zap.Error(err))
		return repository.ErrInternal
	}

	return nil
}

func (r *UserRepositoryPostgres) RemoveRole(ctx context.Context, userID, roleID uuid.UUID) (*user.User, error) {
	_user := &user.User{ID: userID}
	_role := &user.Role{ID: roleID}
	if err := r.db.WithContext(ctx).Model(_user).Association("Roles").Delete(_role); err != nil {
		r.log.Error("failed to remove role to user", zap.String("user_id", userID.String()), zap.String("role_id", roleID.String()), zap.Error(err))
		return nil, repository.ErrInternal
	}

	return r.GetByID(ctx, userID)
}

func (r *UserRepositoryPostgres) RemoveRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error {
	_user := &user.User{ID: userID}
	var _roles []*user.Role

	for _, rid := range roleIDs {
		_roles = append(_roles, &user.Role{ID: rid})
	}

	err := r.db.WithContext(ctx).Model(_user).Association("Roles").Delete(_roles)
	if err != nil {
		r.log.Error("failed to assign roles to user", zap.String("user_id", userID.String()), zap.Any("role_ids", roleIDs), zap.Error(err))
		return repository.ErrInternal
	}

	return nil
}

// ReplaceRoles replaces all existing roles of a user with the provided roles.
func (r *UserRepositoryPostgres) ReplaceRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error {
	_user := &user.User{ID: userID}
	var _roles []*user.Role
	for _, id := range roleIDs {
		_roles = append(_roles, &user.Role{ID: id})
	}

	err := r.db.WithContext(ctx).Model(_user).Association("Roles").Replace(_roles)
	if err != nil {
		r.log.Error("failed to replace roles for user", zap.String("user_id", userID.String()), zap.Any("role_ids", roleIDs), zap.Error(err))
		return repository.ErrInternal
	}
	return nil
}

// GetRoles retrieves all roles associated with a user.
func (r *UserRepositoryPostgres) GetRoles(ctx context.Context, userID uuid.UUID) ([]user.Role, error) {
	var user user.User
	if err := r.db.WithContext(ctx).Preload("Roles").First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		r.log.Error("failed to get roles for user", zap.String("user_id", userID.String()), zap.Error(err))
		return nil, repository.ErrInternal
	}
	return user.Roles, nil
}

// GetPermissions retrieves all permissions associated with a user by traversing their roles.
func (r *UserRepositoryPostgres) GetPermissions(ctx context.Context, userID uuid.UUID) ([]user.Permission, error) {
	var _user user.User
	if err := r.db.WithContext(ctx).Preload("Roles.Permissions").First(&_user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		r.log.Error("failed to get permissions for user", zap.String("user_id", userID.String()), zap.Error(err))
		return nil, repository.ErrInternal
	}

	// Collect unique permissions
	permissionMap := make(map[uuid.UUID]user.Permission)
	for _, role := range _user.Roles {
		for _, permission := range role.Permissions {
			permissionMap[permission.ID] = permission
		}
	}

	permissions := make([]user.Permission, 0, len(permissionMap))
	for _, perm := range permissionMap {
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

func (r *UserRepositoryPostgres) List(ctx context.Context, f user.UserFilter) ([]*user.User, error) {
	var userList []user.User
	query := r.db.WithContext(ctx).Model(&user.User{}).Preload("Roles")

	query = r.applyUserFilters(query, f)
	query = r.applyListOptions(query, f.Limit, f.Offset, f.SortBy, f.SortOrder, "created_at DESC")

	if err := query.Find(&userList).Error; err != nil {
		r.log.Error("failed to list users", zap.Any("filter", f), zap.Error(err))
		return nil, repository.ErrInternal
	}

	// Convert to []*
	users := make([]*user.User, len(userList))
	for i := range userList {
		users[i] = &userList[i]
	}

	return users, nil
}

// Count retrieves the total number of users based on the provided filter.
func (r *UserRepositoryPostgres) Count(ctx context.Context, f user.UserFilter) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&user.User{})

	query = r.applyUserFilters(query, f)

	if err := query.Count(&count).Error; err != nil {
		r.log.Error("failed to count users", zap.Any("filter", f), zap.Error(err))
		return 0, repository.ErrInternal
	}

	return count, nil
}

func (r *UserRepositoryPostgres) applyUserFilters(q *gorm.DB, f user.UserFilter) *gorm.DB {
	if f.ID != nil {
		q = q.Where("id = ?", f.ID)
	}
	if f.Username != nil {
		q = q.Where("username LIKE ?", fmt.Sprintf("%%%s%%", *f.Username))
	}
	if f.Email != nil {
		q = q.Where("email LIKE ?", fmt.Sprintf("%%%s%%", *f.Email))
	}
	if f.PhoneNumber != nil {
		q = q.Where("phone_number LIKE ?", fmt.Sprintf("%%%s%%", *f.PhoneNumber))
	}
	if f.IsActive != nil {
		q = q.Where("is_active = ?", *f.IsActive)
	}
	if f.CreatedBy != nil {
		q = q.Where("created_by = ?", *f.CreatedBy)
	}
	if f.CreatedAt != nil {
		q = q.Where("created_at >= ?", *f.CreatedAt)
	}
	if f.CreatedAtGTE != nil {
		q = q.Where("created_at >= ?", *f.CreatedAtGTE)
	}
	if f.CreatedAtLTE != nil {
		q = q.Where("created_at <= ?", *f.CreatedAtLTE)
	}
	if f.UpdatedAt != nil {
		q = q.Where("updated_at >= ?", *f.UpdatedAt)
	}
	if f.UpdatedAtGTE != nil {
		q = q.Where("updated_at >= ?", *f.UpdatedAtGTE)
	}
	if f.UpdatedAtLTE != nil {
		q = q.Where("updated_at <= ?", *f.UpdatedAtLTE)
	}
	return q
}

// applyListOptions applies limit, offset, and sorting to a GORM query.
func (r *UserRepositoryPostgres) applyListOptions(query *gorm.DB, limit int, offset int, sortBy string, sortOrder string, defaultSort string) *gorm.DB {
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	if sortBy != "" {
		order := "ASC"
		if sortOrder != "" && (sortOrder == "DESC" || sortOrder == "desc") {
			order = "DESC"
		}
		query = query.Order(fmt.Sprintf("%s %s", sortBy, order))
	} else if defaultSort != "" {
		query = query.Order(defaultSort)
	}
	return query
}
