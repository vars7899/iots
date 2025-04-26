package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/vars7899/iots/internal/domain/model"
)

type ResetPasswordTokenRepository interface {
	Create(ctx context.Context, tokenData *model.PasswordResetToken) (*model.PasswordResetToken, error)
	FindByToken(ctx context.Context, tokenStr string) (*model.PasswordResetToken, error)
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}
