package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/vars7899/iots/internal/domain/model"
	"github.com/vars7899/iots/internal/repository"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/logger"
	"github.com/vars7899/iots/pkg/utils"
	"go.uber.org/zap"
)

type resetPasswordTokenService struct {
	rptRepo  repository.ResetPasswordTokenRepository
	userRepo repository.UserRepository
	l        *zap.Logger
}

func NewResetPasswordTokenService(rptRepo repository.ResetPasswordTokenRepository, userRepo repository.UserRepository, baseLogger *zap.Logger) ResetPasswordTokenService {
	return &resetPasswordTokenService{
		rptRepo:  rptRepo,
		userRepo: userRepo,
		l:        logger.Named(baseLogger, "ResetPasswordTokenService"),
	}
}

func (s *resetPasswordTokenService) CreateToken(ctx context.Context, userID uuid.UUID, expiresIn time.Duration) (*model.ResetPasswordToken, error) {
	userExist, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, apperror.ErrorHandler(err, apperror.ErrCodeDBQuery, "failed to validate user ID")
	}
	if userExist == nil {
		return nil, apperror.ErrNotFound.WithMessage("user not found")
	}

	// generate token
	tokenStr, err := utils.GenerateSecureToken(64)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to generate token")
	}

	token := &model.ResetPasswordToken{
		UserID:    userExist.ID,
		Token:     tokenStr,
		ExpiresAt: time.Now().Add(expiresIn),
	}

	// save to db
	createdToken, err := s.rptRepo.Create(ctx, token)
	if err != nil {
		return nil, apperror.ErrorHandler(err, apperror.ErrCodeDBInsert, "failed to create token")
	}

	return createdToken, nil
}

func (s *resetPasswordTokenService) ValidateToken(ctx context.Context, tokenStr string) (*model.ResetPasswordToken, error) {
	token, err := s.rptRepo.FindByToken(ctx, tokenStr)
	if err != nil {
		return nil, err
	}
	if token.IsExpired() {
		return nil, apperror.ErrForbidden.WithMessage("token expired")
	}
	return token, nil
}

func (s *resetPasswordTokenService) DeleteTokensByUserID(ctx context.Context, userID uuid.UUID) error {
	return s.rptRepo.DeleteByUserID(ctx, userID)
}

func (s *resetPasswordTokenService) DeleteExpiredTokens(ctx context.Context) error {
	return s.rptRepo.DeleteExpired(ctx)
}
