package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vars7899/iots/config"
	"github.com/vars7899/iots/internal/api/v1/dto"
	"github.com/vars7899/iots/internal/domain/model"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/auth"
	"github.com/vars7899/iots/pkg/auth/token"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
)

type authService struct {
	userService          UserService
	roleService          RoleService
	accessControlService auth.AccessControlService
	authTokenService     auth.AuthTokenService
	resetPasswordService ResetPasswordTokenService
	config               *config.AppConfig
	logger               *zap.Logger
}

func NewAuthService(userService UserService, roleService RoleService, accessControlService auth.AccessControlService, authTokenService auth.AuthTokenService, resetPasswordService ResetPasswordTokenService, config *config.AppConfig, baseLogger *zap.Logger) AuthService {
	return &authService{
		userService:          userService,
		roleService:          roleService,
		accessControlService: accessControlService,
		authTokenService:     authTokenService,
		resetPasswordService: resetPasswordService,
		config:               config,
		logger:               logger.Named(baseLogger, "AuthService"),
	}
}

func (s *authService) RegisterUser(ctx context.Context, userData *model.User) (*model.User, *token.AuthTokenSet, error) {
	createdUser, err := s.userService.CreateUser(ctx, userData)
	if err != nil {
		return nil, nil, apperror.ErrorHandler(err, apperror.ErrCodeInternal, "failed to register user")
	}

	tokens, err := s.authTokenService.IssueAuthTokenSet(ctx, createdUser.ID, getUserRolesList(createdUser.Roles)) // Call the new method
	if err != nil {
		return nil, nil, err
	}

	defaultRoleID, err := s.roleService.GetDefaultRoleID(ctx)
	if err != nil {
		s.logger.Error("failed to assign default role", zap.String("userID", createdUser.ID.String()), zap.Error(err))
		rollbackErr := s.userService.HardDeleteUser(ctx, createdUser.ID)
		if rollbackErr != nil {
			s.logger.Error("failed to rollback user creation after role assignment error", zap.String("userID", createdUser.ID.String()), zap.Error(rollbackErr))
		}
		return nil, nil, apperror.ErrInternal.Wrap(err).WithMessage("failed to assign default role to user")
	}

	userWithRoles, err := s.userService.AssignUserRoles(ctx, createdUser.ID, []uuid.UUID{defaultRoleID})
	if err != nil {
		s.logger.Error("Failed to assign default role in repository", zap.String("userID", createdUser.ID.String()), zap.String("roleID", defaultRoleID.String()), zap.Error(err))
		rollbackErr := s.userService.HardDeleteUser(ctx, createdUser.ID) // Example: attempt rollback
		if rollbackErr != nil {
			s.logger.Error("Failed to rollback user creation after role assignment error", zap.String("userID", createdUser.ID.String()), zap.Error(rollbackErr))
		}
		return nil, nil, apperror.ErrorHandler(err, apperror.ErrCodeDBInsert, "failed to assign default role to user")
	}

	if err := s.accessControlService.SyncUserRoles(userWithRoles); err != nil {
		s.logger.Error("Failed to sync user roles with Casbin after creation",
			zap.String("userID", userWithRoles.ID.String()),
			zap.Error(err))
		// Decide how to handle this error. It means the user is created and has roles in DB,
		// but Casbin's view might be out of sync. This is serious for auth.
		// You might want to log, alert, and potentially disable the user account until fixed.
		// For now, we'll log and let the user creation proceed, but be aware of the auth risk.
		// A background worker could periodically sync or you could implement retry logic.
	}

	return userWithRoles, tokens, nil
}

func (s *authService) LoginUser(ctx context.Context, credentials *dto.LoginCredentials) (*model.User, *token.AuthTokenSet, error) {
	userExist, err := s.findUserByLoginIdentifier(ctx, credentials.Email, credentials.Username, credentials.PhoneNumber)
	if err != nil {
		return nil, nil, err
	}

	if err := userExist.ComparePassword(credentials.Password); err != nil {
		return nil, nil, apperror.ErrUnauthorized.WithMessage("invalid credentials").Wrap(err)
	}

	tokens, err := s.authTokenService.IssueAuthTokenSet(ctx, userExist.ID, getUserRolesList(userExist.Roles))
	if err != nil {
		return nil, nil, err
	}

	if err := s.userService.SetLastLogin(ctx, userExist.ID); err != nil {
		s.logger.Error("failed to set last login for user", zap.String("userID", userExist.ID.String()), zap.Error(err))
		// Continue execution
	}

	// 5. Return the authenticated user and the generated tokens
	return userExist, tokens, nil
}

func (s *authService) RefreshAuthTokens(ctx context.Context, refreshToken string) (*model.User, *token.AuthTokenSet, error) {
	claims, err := s.authTokenService.ValidateRefreshToken(refreshToken)
	if err != nil {
		s.logger.Error("invalid user claims", zap.Error(err))
		return nil, nil, apperror.ErrUnauthorized.WithMessage("invalid refresh token").Wrap(err)
	}

	// Check if the refresh token is expired.
	if time.Now().After(claims.ExpiresAt.Time) {
		return nil, nil, apperror.ErrUnauthorized.WithMessage("refresh token expired")
	}

	// Check if JTI is revoked
	revoked, err := s.authTokenService.IsJTIRevoked(ctx, claims.ID)
	if err != nil {
		return nil, nil, apperror.ErrInternal.WithMessage("failed to verify token jti").Wrap(err)
	}
	if revoked {
		s.logger.Warn("refresh attempt with used/revoked token", zap.String("jti", claims.ID))
		return nil, nil, apperror.ErrUnauthorized.WithMessage("refresh token reuse detected")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		s.logger.Error("invalid user ID in claims", zap.Error(err))
		return nil, nil, apperror.ErrInvalidToken.WithMessage("invalid authorization token").Wrap(err)
	}

	user, err := s.userService.GetUserByID(ctx, userID)
	if err != nil {
		return nil, nil, apperror.ErrorHandler(err, apperror.ErrCodeDBQuery, "failed to fetch user during refresh")
	}

	token, err := s.authTokenService.IssueAuthTokenSet(ctx, user.ID, getUserRolesList(user.Roles))
	if err != nil {
		return nil, nil, err
	}

	// Mark the old refresh token JTI as revoked
	if err := s.authTokenService.RevokeJTI(ctx, claims.ID, claims.ExpiresAt.Time); err != nil {
		s.logger.Warn("failed to revoke old refresh token jti", zap.String("jti", claims.ID), zap.Error(err))
	}

	return user, token, nil
}

func (s *authService) LogoutUser(ctx context.Context, userID *uuid.UUID, claims *token.AccessTokenClaims, refreshTokenStr string) error {
	if err := s.authTokenService.RevokeJTI(ctx, claims.ID, claims.ExpiresAt.Time); err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeInternal, "Failed to revoke access token")
	}

	if refreshTokenStr != "" {
		refreshClaims, err := s.authTokenService.ParseRefreshToken(refreshTokenStr)
		if err != nil {
			s.logger.Warn("Failed to parse refresh token during logout, cannot revoke JTI", zap.String("userID", userID.String()), zap.Error(err))
			// continue
		} else if refreshClaims != nil && refreshClaims.ID != "" {
			if err := s.authTokenService.RevokeJTI(ctx, refreshClaims.ID, refreshClaims.ExpiresAt.Time); err != nil {
				s.logger.Error("Failed to revoke refresh token JTI during logout",
					zap.String("userID", userID.String()),
					zap.String("jti", refreshClaims.ID),
					zap.Error(err),
				)
				// continue
			} else {
				s.logger.Info("Refresh token JTI revoked successfully", zap.String("userID", userID.String()), zap.String("jti", refreshClaims.ID))
			}
		} else {
			// Refresh token value was present but empty or claims were nil/missing JTI after parsing
			s.logger.Debug("Refresh token value present but empty or claims incomplete, cannot revoke JTI", zap.String("userID", userID.String()))
		}
	} else {
		s.logger.Debug("No refresh token cookie found for logout", zap.String("userID", userID.String()))
	}

	if err := s.userService.SetLastLogin(ctx, *userID); err != nil {
		s.logger.Warn("failed to set last login while logging out", zap.String("userID", userID.String()), zap.Error(err))
		return err
	}

	return nil
}

func (s *authService) RequestPasswordReset(ctx context.Context, email string) (*model.ResetPasswordToken, *string, error) {
	user, err := s.userService.FindByEmail(ctx, email)
	if err != nil {
		s.logger.Warn("User not found during password reset request", zap.String("email", email), zap.Error(err))
		return nil, nil, err
	}

	if err := s.resetPasswordService.DeleteTokensByUserID(ctx, user.ID); err != nil {
		s.logger.Warn("Failed to invalidate existing user password reset token", zap.String("email", email), zap.Error(err))
		// continue
	}

	resetToken, err := s.resetPasswordService.CreateToken(ctx, user.ID, s.config.Auth.RequestResetPasswordTokenTTL)
	if err != nil {
		return nil, nil, err
	}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s", s.config.Frontend.BaseUrl, resetToken.Token)

	// send email otp later

	return resetToken, &resetLink, nil
}

func (s *authService) ResetPassword(ctx context.Context, resetToken, newRawPassword string) error {
	resetRecord, err := s.resetPasswordService.ValidateToken(ctx, resetToken)
	if err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeInternal, "invalid or expired reset token")
	}

	user, err := s.userService.GetUserByID(ctx, resetRecord.UserID)
	if err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeDBQuery, "failed to find user for password reset")
	}

	if err := user.HashPassword(newRawPassword); err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeInternal, "failed to set new password")
	}

	if _, err := s.userService.SetPassword(ctx, user.ID, user.Password); err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeInternal, "failed to update user password")
	}

	if err := s.resetPasswordService.DeleteTokensByUserID(ctx, user.ID); err != nil {
		s.logger.Warn("failed to delete reset token", zap.String("userID", user.ID.String()))
	}

	return nil
}

func getUserRolesList(roles []model.Role) []string {
	roleList := make([]string, len(roles))
	for i, role := range roles {
		roleList[i] = role.Slug
	}
	return roleList
}

func (s *authService) findUserByLoginIdentifier(ctx context.Context, email, username, phoneNumber string) (*model.User, error) {
	var user *model.User
	var err error

	// Use sequential checks
	if email != "" {
		user, err = s.userService.FindByEmail(ctx, email)
	} else if phoneNumber != "" {
		user, err = s.userService.FindByPhoneNumber(ctx, phoneNumber)
	} else if username != "" {
		user, err = s.userService.FindByUserName(ctx, username)
	} else {
		return nil, apperror.ErrInvalidCredentials.WithMessage("missing login identifier")
	}

	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return nil, apperror.ErrUnauthorized.WithMessage("invalid credentials")
		}
		return nil, apperror.ErrorHandler(err, apperror.ErrCodeDBQuery, "failed to find user by login identifier")
	}

	return user, nil
}
