package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/vars7899/iots/internal/domain"
	"github.com/vars7899/iots/internal/domain/model"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type passwordResetTokenRepositoryPostgres struct {
	db *gorm.DB
	l  *zap.Logger
}

func NewPasswordResetTokenRepositoryPostgres(db *gorm.DB, baseLogger *zap.Logger) *passwordResetTokenRepositoryPostgres {
	return &passwordResetTokenRepositoryPostgres{
		db: db,
		l:  logger.Named(baseLogger, "TokenRepositoryPostgres"),
	}
}

func (r *passwordResetTokenRepositoryPostgres) Create(ctx context.Context, tokenData *model.PasswordResetToken) (*model.PasswordResetToken, error) {
	if err := r.db.WithContext(ctx).Model(&model.PasswordResetToken{}).Clauses(clause.Returning{}).Create(&tokenData).Error; err != nil {
		return nil, apperror.MapDBError(err, domain.EntityToken)
	}
	return tokenData, nil
}

func (r *passwordResetTokenRepositoryPostgres) FindByToken(ctx context.Context, tokenStr string) (*model.PasswordResetToken, error) {
	var token model.PasswordResetToken
	if err := r.db.WithContext(ctx).Model(&model.PasswordResetToken{}).Where("token = ?", tokenStr).First(&token).Error; err != nil {
		return nil, apperror.MapDBError(err, domain.EntityToken)
	}
	return &token, nil
}

func (r *passwordResetTokenRepositoryPostgres) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	tx := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.PasswordResetToken{})
	if tx.Error != nil {
		return apperror.MapDBError(tx.Error, domain.EntityToken)
	}
	if tx.RowsAffected == 0 {
		return notFoundErr(domain.EntityToken, "delete by user ID")
	}
	return nil
}

func (r *passwordResetTokenRepositoryPostgres) DeleteExpired(ctx context.Context) error {
	tx := r.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&model.PasswordResetToken{})
	if tx.Error != nil {
		return apperror.MapDBError(tx.Error, domain.EntityToken)
	}
	if tx.RowsAffected == 0 {
		return notFoundErr(domain.EntityToken, "delete expired")
	}
	return nil
}
