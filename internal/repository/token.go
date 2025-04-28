package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/vars7899/iots/internal/domain/model"
)

type ResetPasswordTokenRepository interface {
	Create(ctx context.Context, tokenData *model.ResetPasswordToken) (*model.ResetPasswordToken, error)
	FindByToken(ctx context.Context, tokenStr string) (*model.ResetPasswordToken, error)
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}
