package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vars7899/iots/internal/api/v1/dto"
	"github.com/vars7899/iots/internal/domain"
	"github.com/vars7899/iots/internal/domain/model"
	"github.com/vars7899/iots/internal/repository"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserRepositoryPostgres struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewUserRepositoryPostgres(db *gorm.DB, baseLogger *zap.Logger) repository.UserRepository {
	return &UserRepositoryPostgres{db: db, logger: logger.Named(baseLogger, "UserRepositoryPostgres")}
}

func (r *UserRepositoryPostgres) GetByID(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("id = ?", userID).Preload("Roles").Preload("Roles.Permissions").First(&user).Error; err != nil {
		return nil, apperror.MapDBError(err, domain.EntityUser)
	}
	return &user, nil
}

func (r *UserRepositoryPostgres) Create(ctx context.Context, userData *model.User) (*model.User, error) {
	if err := r.db.WithContext(ctx).Model(&model.User{}).Clauses(clause.Returning{}).Create(&userData).Error; err != nil {
		return nil, apperror.MapDBError(err, domain.EntitySensor)
	}
	return userData, nil
}

func (r *UserRepositoryPostgres) Update(ctx context.Context, userData *model.User) (*model.User, error) {
	var updatedUser model.User

	tx := r.db.WithContext(ctx).Model(&model.User{}).Clauses(clause.Returning{}).Where("id = ?", userData.ID).Updates(userData).Scan(&updatedUser)
	if tx.Error != nil {
		return nil, apperror.MapDBError(tx.Error, domain.EntityUser)
	}
	if tx.RowsAffected == 0 {
		return nil, apperror.ErrNotFound.WithMessagef("error encountered while performing %s update operation, please retry", domain.EntityUser)
	}

	return &updatedUser, nil
}

func (r *UserRepositoryPostgres) HardDelete(ctx context.Context, userID uuid.UUID) error {
	tx := r.db.WithContext(ctx).Unscoped().Where("id = ?", userID).Delete(&model.User{})
	if tx.Error != nil {
		return apperror.MapDBError(tx.Error, domain.EntityUser)
	}
	if tx.RowsAffected == 0 {
		return apperror.ErrNotFound.WithMessagef("cannot hard delete %s: no matching record found", domain.EntityUser)
	}
	return nil
}

func (r *UserRepositoryPostgres) Delete(ctx context.Context, userID uuid.UUID) error {
	tx := r.db.WithContext(ctx).Where("id = ?", userID).Delete(&model.User{})
	if tx.Error != nil {
		return apperror.MapDBError(tx.Error, domain.EntityUser)
	}
	if tx.RowsAffected == 0 {
		return apperror.ErrNotFound.WithMessagef("cannot delete %s: no matching record found", domain.EntityUser)
	}
	return nil
}

func (r *UserRepositoryPostgres) Restore(ctx context.Context, userID uuid.UUID) error {
	tx := r.db.WithContext(ctx).Model(&model.User{}).Unscoped().Where("id = ?", userID).Update("deleted_at", nil)

	if tx.Error != nil {
		return apperror.MapDBError(tx.Error, domain.EntityUser)
	}
	if tx.RowsAffected == 0 {
		return apperror.ErrNotFound.WithMessagef("cannot restore %s: no matching record found", domain.EntityUser)
	}
	return nil
}

func (r *UserRepositoryPostgres) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var userExist model.User
	if err := r.db.WithContext(ctx).Model(&model.User{}).Preload("Roles").Preload("Roles.Permissions").Where("email = ?", email).First(&userExist).Error; err != nil {
		return nil, apperror.MapDBError(err, domain.EntityUser)
	}
	return &userExist, nil
}

func (r *UserRepositoryPostgres) FindByUserName(ctx context.Context, userName string) (*model.User, error) {
	var userExist model.User
	if err := r.db.WithContext(ctx).Model(&model.User{}).Preload("Roles").Preload("Roles.Permissions").Where("username = ?", userName).First(&userExist).Error; err != nil {
		return nil, apperror.MapDBError(err, domain.EntityUser)
	}
	return &userExist, nil
}

func (r *UserRepositoryPostgres) FindByPhoneNumber(ctx context.Context, phoneNumber string) (*model.User, error) {
	var userExist model.User
	if err := r.db.WithContext(ctx).Model(&model.User{}).Preload("Roles").Preload("Roles.Permissions").Where("phone_number = ?", phoneNumber).First(&userExist).Error; err != nil {
		return nil, apperror.MapDBError(err, domain.EntityUser)
	}
	return &userExist, nil
}

func (r *UserRepositoryPostgres) FindByRoles(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	var userExist model.User
	if err := r.db.WithContext(ctx).Model(&model.User{}).Preload("Roles.Permissions").First(&userExist, "id = ?", userID).Error; err != nil {
		return nil, apperror.MapDBError(err, domain.EntityUser)
	}
	return &userExist, nil
}

func (r *UserRepositoryPostgres) SetLastLogin(ctx context.Context, userID uuid.UUID, timestamp time.Time) error {
	tx := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Update("last_login", timestamp)

	if tx.Error != nil {
		return apperror.MapDBError(tx.Error, domain.EntityUser)
	}
	if tx.RowsAffected == 0 {
		return apperror.ErrNotFound.WithMessagef("cannot set last login %s: no matching record found", domain.EntityUser)
	}
	return nil
}

func (r *UserRepositoryPostgres) ExistByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, apperror.MapDBError(err, domain.EntityUser)
	}
	return count > 0, nil
}

func (r *UserRepositoryPostgres) ExistByUserName(ctx context.Context, username string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return false, apperror.MapDBError(err, domain.EntityUser)
	}
	return count > 0, nil
}

func (r *UserRepositoryPostgres) ExistByPhoneNumber(ctx context.Context, phoneNumber string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.User{}).Where("phone_number = ?", phoneNumber).Count(&count).Error; err != nil {
		return false, apperror.MapDBError(err, domain.EntityUser)
	}
	return count > 0, nil
}

func (r *UserRepositoryPostgres) AssignRole(ctx context.Context, userID, roleID uuid.UUID) (*model.User, error) {
	var user model.User

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`
            INSERT INTO user_roles (user_id, role_id) 
            VALUES (?, ?) 
            ON CONFLICT (user_id, role_id) DO NOTHING
        `, userID, roleID).Error; err != nil {
			return err
		}
		return tx.Model(&model.User{}).Where("id = ?", userID).Preload("Roles").Preload("Roles.Permissions").First(&user).Error
	})

	if err != nil {
		return nil, apperror.MapDBError(err, domain.EntityUser)
	}
	return &user, nil
}

func (r *UserRepositoryPostgres) AssignRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) (*model.User, error) {
	var user model.User

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(roleIDs) == 0 {
			return tx.Model(&model.User{}).Where("id = ?", userID).Preload("Roles.Permissions").First(&user).Error
		}

		query := "INSERT INTO user_roles (user_id, role_id) VALUES"
		values := []interface{}{}

		for i, roleID := range roleIDs {
			if i > 0 {
				query += ", "
			}
			query += "(?, ?)"
			values = append(values, userID, roleID)
		}
		query += " ON CONFLICT (user_id, role_id) DO NOTHING"

		if err := tx.Exec(query, values...).Error; err != nil {
			return err
		}

		return tx.Model(&model.User{}).Where("id = ?", userID).Preload("Roles").Preload("Roles.Permissions").First(&user).Error
	})

	if err != nil {
		return nil, apperror.MapDBError(err, domain.EntityUser)
	}

	return &user, nil
}

func (r *UserRepositoryPostgres) RemoveRole(ctx context.Context, userID, roleID uuid.UUID) (*model.User, error) {
	var user model.User

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM user_roles WHERE user_id = ? AND role_id = ?", userID, roleID).Error; err != nil {
			return err
		}
		return tx.Model(&model.User{}).Where("id = ?", userID).Preload("Roles").First(&user).Error
	})

	if err != nil {
		return nil, apperror.MapDBError(err, domain.EntityUser)
	}
	return &user, nil
}

func (r *UserRepositoryPostgres) RemoveRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) (*model.User, error) {
	var user model.User

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(roleIDs) == 0 {
			return nil // No roles to remove
		}

		query := "DELETE FROM user_roles WHERE user_id = ? AND role_id IN (?)"
		if err := tx.Exec(query, userID, roleIDs).Error; err != nil {
			return err
		}

		// Fetch the updated user with their roles
		return tx.Model(&model.User{}).Where("id = ?", userID).Preload("Roles").First(&user).Error
	})

	if err != nil {
		return nil, apperror.MapDBError(err, domain.EntityUser)
	}
	return &user, nil
}

func (r *UserRepositoryPostgres) ReplaceRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) (*model.User, error) {
	var user model.User

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM user_roles WHERE user_id = ?", userID).Error; err != nil {
			return err
		}

		if len(roleIDs) > 0 {
			query := "INSERT INTO user_roles (user_id, role_id) VALUES "
			values := []interface{}{}

			for i, roleID := range roleIDs {
				if i > 0 {
					query += ", "
				}
				query += "(?, ?)"
				values = append(values, userID, roleID)
			}

			if err := tx.Exec(query, values...).Error; err != nil {
				return err
			}
		}

		return tx.Model(&model.User{}).Where("id = ?", userID).Preload("Roles").First(&user).Error
	})

	if err != nil {
		return nil, apperror.MapDBError(err, domain.EntityUser)
	}

	return &user, nil
}

func (r *UserRepositoryPostgres) GetRoles(ctx context.Context, userID uuid.UUID) ([]model.Role, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Preload("Roles").First(&user, "id = ?", userID).Error; err != nil {
		return nil, apperror.MapDBError(err, domain.EntityUser)
	}
	return user.Roles, nil
}

func (r *UserRepositoryPostgres) GetPermissions(ctx context.Context, userID uuid.UUID) ([]*model.Permission, error) {
	var permissions []*model.Permission

	err := r.db.WithContext(ctx).
		Distinct().
		Select("permissions.*").
		Table("permissions").
		Joins("INNER JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("INNER JOIN user_roles ON role_permissions.role_id = user_roles.role_id").
		Where("user_roles.user_id = ?", userID).
		Find(&permissions).Error

	if err != nil {
		return nil, apperror.MapDBError(err, domain.EntityUser)
	}

	return permissions, nil
}

func (r *UserRepositoryPostgres) List(ctx context.Context, f dto.UserFilter) ([]*model.User, error) {
	var userList []model.User
	query := r.db.WithContext(ctx).Model(&model.User{}).Preload("Roles")

	query = r.applyUserFilters(query, f)
	query = r.applyListOptions(query, f.Limit, f.Offset, f.SortBy, f.SortOrder, "created_at DESC")

	if err := query.Find(&userList).Error; err != nil {
		return nil, apperror.MapDBError(err, domain.EntityUser)
	}

	// Convert to []*
	users := make([]*model.User, len(userList))
	for i := range userList {
		users[i] = &userList[i]
	}

	return users, nil
}

// Count retrieves the total number of users based on the provided filter.
func (r *UserRepositoryPostgres) Count(ctx context.Context, f dto.UserFilter) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.User{})

	query = r.applyUserFilters(query, f)

	if err := query.Count(&count).Error; err != nil {
		return 0, apperror.MapDBError(err, domain.EntityUser)
	}

	return count, nil
}

func (r *UserRepositoryPostgres) applyUserFilters(q *gorm.DB, f dto.UserFilter) *gorm.DB {
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
