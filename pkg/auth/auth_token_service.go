package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/vars7899/iots/internal/cache"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/auth/token"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
)

type authTokenService struct {
	tokenService    token.TokenService
	jtiStoreService cache.JTIStore
	logger          *zap.Logger
}

func NewAuthTokenManger(ts token.TokenService, jss cache.JTIStore, l *zap.Logger) AuthTokenService {
	return &authTokenService{tokenService: ts, jtiStoreService: jss, logger: logger.Named(l, "AuthTokenManager")}
}

func (s *authTokenService) IssueAuthTokenSet(ctx context.Context, userID uuid.UUID, roles []string) (*token.AuthTokenSet, error) {
	set, err := s.tokenService.GenerateAuthTokenSet(userID, roles)
	if err != nil {
		s.logger.Error("failed to generate authentication tokens", zap.String("userID", userID.String()), zap.Error(err))
		return nil, err
	}
	if err := s.jtiStoreService.RecordJTI(ctx, set.RefreshTokenJTI, set.RefreshExpiresAt); err != nil {
		s.logger.Error("failed to record authentication tokens", zap.String("userID", userID.String()), zap.Error(err))
		return nil, err
	}
	return set, nil
}
func (s *authTokenService) ParseRefreshToken(tokenStr string) (*token.RefreshTokenClaims, error) {
	claims, err := s.tokenService.ParseRefreshToken(tokenStr)
	if err != nil {
		s.logger.Error("Refresh token parse ", zap.Error(err))
		return nil, apperror.ErrorHandler(err, apperror.ErrCodeInvalidToken, "failed to parse refresh token").Wrap(err)
	}
	s.logger.Debug("Refresh token parse success", zap.String("token", tokenStr))
	return claims, nil
}
func (s *authTokenService) ParseAccessToken(tokenStr string) (*token.AccessTokenClaims, error) {
	claims, err := s.tokenService.ParseAccessToken(tokenStr)
	if err != nil {
		s.logger.Error("Access token parse ", zap.Error(err))
		return nil, apperror.ErrorHandler(err, apperror.ErrCodeInvalidToken, "failed to parse access token").Wrap(err)
	}
	s.logger.Debug("Access token parse success", zap.String("token", tokenStr))
	return claims, nil
}
func (s *authTokenService) ValidateRefreshToken(refreshToken string) (*token.RefreshTokenClaims, error) {
	claims, err := s.tokenService.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}
	return claims, nil
}
func (s *authTokenService) RevokeJTI(ctx context.Context, jti string, expiresAt time.Time) error {
	if jti == "" {
		return nil
	}
	if err := s.jtiStoreService.RevokeJTI(ctx, jti, expiresAt); err != nil {
		s.logger.Error("JTI revoke failure", zap.String("JTI", jti))
		return apperror.ErrorHandler(err, apperror.ErrCodeDBInsert, "failed to revoke jti")
	}
	s.logger.Debug("JTI revoke success", zap.String("JTI", jti))
	return nil
}

func (s *authTokenService) IsJTIRevoked(ctx context.Context, jti string) (bool, error) {
	if jti == "" {
		return false, nil
	}
	isRevoked, err := s.jtiStoreService.IsJTIRevoked(ctx, jti)
	if err != nil {
		s.logger.Error("JTI check failure", zap.String("JTI", jti))
		return false, apperror.ErrorHandler(err, apperror.ErrCodeDBInsert, "failed to  jti")
	}
	s.logger.Debug("JTI check success", zap.String("JTI", jti))
	return isRevoked, nil
}
