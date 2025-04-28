package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/vars7899/iots/internal/domain"
	"github.com/vars7899/iots/internal/domain/model"
	"github.com/vars7899/iots/internal/repository"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ResetPasswordTokenRepositoryPostgres struct {
	db *gorm.DB
	l  *zap.Logger
}

func NewResetPasswordTokenRepositoryPostgres(db *gorm.DB, baseLogger *zap.Logger) repository.ResetPasswordTokenRepository {
	return &ResetPasswordTokenRepositoryPostgres{
		db: db,
		l:  logger.Named(baseLogger, "TokenRepositoryPostgres"),
	}
}

func (r *ResetPasswordTokenRepositoryPostgres) Create(ctx context.Context, tokenData *model.ResetPasswordToken) (*model.ResetPasswordToken, error) {
	if err := r.db.WithContext(ctx).Model(&model.ResetPasswordToken{}).Clauses(clause.Returning{}).Create(&tokenData).Error; err != nil {
		return nil, apperror.MapDBError(err, domain.EntityToken)
	}
	return tokenData, nil
}

func (r *ResetPasswordTokenRepositoryPostgres) FindByToken(ctx context.Context, tokenStr string) (*model.ResetPasswordToken, error) {
	var token model.ResetPasswordToken
	if err := r.db.WithContext(ctx).Model(&model.ResetPasswordToken{}).Where("token = ?", tokenStr).First(&token).Error; err != nil {
		return nil, apperror.MapDBError(err, domain.EntityToken)
	}
	return &token, nil
}

func (r *ResetPasswordTokenRepositoryPostgres) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	tx := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.ResetPasswordToken{})
	if tx.Error != nil {
		return apperror.MapDBError(tx.Error, domain.EntityToken)
	}
	if tx.RowsAffected == 0 {
		return notFoundErr(domain.EntityToken, "delete by user ID")
	}
	return nil
}

func (r *ResetPasswordTokenRepositoryPostgres) DeleteExpired(ctx context.Context) error {
	tx := r.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&model.ResetPasswordToken{})
	if tx.Error != nil {
		return apperror.MapDBError(tx.Error, domain.EntityToken)
	}
	if tx.RowsAffected == 0 {
		return notFoundErr(domain.EntityToken, "delete expired")
	}
	return nil
}
