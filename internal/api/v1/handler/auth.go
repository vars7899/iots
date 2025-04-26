package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/config"
	"github.com/vars7899/iots/internal/api/v1/dto"
	"github.com/vars7899/iots/internal/cache"
	"github.com/vars7899/iots/internal/domain/model"
	"github.com/vars7899/iots/internal/middleware"
	"github.com/vars7899/iots/internal/service"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/auth/token"
	"github.com/vars7899/iots/pkg/contextkey"
	"github.com/vars7899/iots/pkg/di"
	"github.com/vars7899/iots/pkg/logger"
	"github.com/vars7899/iots/pkg/response"
	"github.com/vars7899/iots/pkg/utils"
	"go.uber.org/zap"
)

type AuthHandler struct {
	userService  *service.UserService
	tokenService token.TokenService
	jtiService   cache.JTIStore
	log          *zap.Logger
}

func NewAuthHandler(deps *di.Provider, baseLogger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		userService:  deps.Services.UserService,
		tokenService: deps.Helpers.TokenService,
		jtiService:   deps.Helpers.JTIService,
		log:          logger.Named(baseLogger, "AuthHandler"),
	}
}

func (h *AuthHandler) RegisterRoutes(e *echo.Group) {
	e.POST("/register", h.Register)
	e.POST("/login", h.Login)
	e.POST("/refresh", h.Refresh)
	e.POST("/logout", h.Logout, middleware.JWT_JTI_Middleware(h.tokenService, h.jtiService, h.log))
	// e.POST("/request-password-reset")
}

func (h *AuthHandler) Login(c echo.Context) error {
	var dto dto.LoginUserRequestDTO
	reqPath := utils.GetRequestUrlPath(c)
	reqCtx := c.Request().Context()

	// Bind body request and validate fields
	if err := utils.BindAndValidate(c, &dto); err != nil {
		return err
	}

	if dto.Email == "" && dto.UserName == "" && dto.PhoneNumber == "" {
		return apperror.ErrBadRequest.WithMessage("provide at least one login identifier").WithPath(reqPath)
	}

	userExist, err := h.userService.FindByLoginIdentifier(reqCtx, service.LoginIdentifier{
		Email:       dto.Email,
		Username:    dto.UserName,
		PhoneNumber: dto.PhoneNumber,
	})
	if err != nil {
		return apperror.ErrUnauthorized.WithMessage("invalid credentials").WithPath(reqPath).Wrap(err)
	}

	// Verify password
	if err := userExist.ComparePassword(dto.Password); err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeInvalidCredentials, "failed to login").WithPath(reqPath).Wrap(err)
	}

	set, err := h.generateToken(reqCtx, userExist)
	if err != nil {
		return apperror.ErrInternal.WithMessage("failed to generate required token(s)").WithPath(reqPath).Wrap(err)
	}

	h.bindAccessTokenToResponseHeader(c, set.AccessToken)
	h.bindRefreshTokenCookie(c, set.RefreshToken, h.tokenService.GetRefreshTTL())

	// update last login
	if err := h.userService.SetLastLogin(c.Request().Context(), userExist.ID); err != nil {
		h.log.Error("failed to set last login for user", zap.String("userID", userExist.ID.String()), zap.Error(err))
	}

	if config.InProd() {
		return response.JSON(c, http.StatusOK, echo.Map{
			"message": "user logged in successfully",
			"user":    userExist.PublicView(),
		})
	}

	return response.JSON(c, http.StatusOK, echo.Map{
		"message":                 "user logged in successfully",
		"user":                    userExist.PublicView(),
		"authorization_token_set": set,
	})
}

func (h *AuthHandler) Register(c echo.Context) error {
	var dto dto.RegisterUserRequestDTO
	reqPath := utils.GetRequestUrlPath(c)
	reqCtx := c.Request().Context()

	// Bind body request and validate fields
	if err := utils.BindAndValidate(c, &dto); err != nil {
		return err
	}

	createdUser, err := h.userService.CreateUser(reqCtx, dto.AsModel())
	if err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeInternal, "something went wrong while registering user").WithPath(reqPath)
	}

	set, err := h.generateToken(reqCtx, createdUser)
	if err != nil {
		return apperror.ErrInternal.WithMessage("failed to generate required token(s)").WithPath(reqPath).Wrap(err)
	}

	h.bindAccessTokenToResponseHeader(c, set.AccessToken)
	h.bindRefreshTokenCookie(c, set.RefreshToken, h.tokenService.GetRefreshTTL())

	if config.InProd() {
		return response.JSON(c, http.StatusOK, echo.Map{
			"message": "user registered successfully",
			"user":    createdUser.PublicView(),
		})
	}

	return response.JSON(c, http.StatusCreated, echo.Map{
		"message":                 "user registered successfully",
		"user":                    createdUser.PublicView(),
		"authorization_token_set": set,
	})
}

func (h *AuthHandler) Logout(c echo.Context) error {
	claims, err := middleware.GetAccessTokenClaims(c)
	reqPath := utils.GetRequestUrlPath(c)
	reqCtx := c.Request().Context()

	if err != nil {
		h.log.Warn("failed to get access token claims", zap.Error(err))
		return apperror.ErrUnauthorized.WithMessage("invalid/missing authorization token").WithPath(reqPath).Wrap(err)
	}

	// Clear refresh token cookie
	c.SetCookie(&http.Cookie{
		Name:     string(contextkey.Auth_refreshToken),
		Value:    "",
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		Secure:   config.InProd(),
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		h.log.Error("invalid user ID in claims", zap.String("userID", claims.Subject), zap.Error(err))
		return apperror.ErrInvalidToken.WithMessage("invalid/missing authorization token").WithPath(reqPath).Wrap(err)
	}

	// Revoke access token JTI
	if err := h.jtiService.RevokeJTI(reqCtx, claims.ID, claims.ExpiresAt.Time); err != nil {
		h.log.Error("failed to revoke access token", zap.String("jti", claims.ID), zap.Error(err))
	}

	// Get refresh token from cookie and revoke it too if present
	refreshCookie, err := c.Cookie(string(contextkey.Auth_refreshToken))
	if err == nil && refreshCookie.Value != "" {
		refreshClaims, err := h.tokenService.ParseRefreshToken(refreshCookie.Value)
		if err == nil {
			if err := h.jtiService.RevokeJTI(reqCtx, refreshClaims.ID, refreshClaims.ExpiresAt.Time); err != nil {
				h.log.Error("failed to revoke refresh token", zap.String("jti", refreshClaims.ID), zap.Error(err))
			}
		}
	}

	if err := h.userService.SetLastLogin(reqCtx, userID); err != nil {
		h.log.Error("failed to set last login for user", zap.String("userID", userID.String()), zap.Error(err))
	}

	return response.JSON(c, http.StatusOK, echo.Map{
		"message": "user logged out successfully",
	})
}

func (h *AuthHandler) Refresh(c echo.Context) error {
	reqPath := utils.GetRequestUrlPath(c)
	refreshToken, err := c.Cookie(string(contextkey.Auth_refreshToken))
	reqCtx := c.Request().Context()

	if err != nil || refreshToken.Value == "" {
		return apperror.ErrUnauthorized.WithMessage("missing/invalid refresh token").WithPath(reqPath).Wrap(err)
	}

	claims, err := h.tokenService.ParseRefreshToken(refreshToken.Value)
	if err != nil {
		return apperror.ErrUnauthorized.WithMessage("missing/invalid refresh token, failed to parse token").WithPath(reqPath).Wrap(err)
	}

	// Check if the refresh token is expired.
	if time.Now().After(claims.ExpiresAt.Time) {
		return apperror.ErrUnauthorized.WithMessage("refresh token expired").WithPath(reqPath)
	}

	// Check if JTI is revoked
	revoked, err := h.jtiService.IsJTIRevoked(reqCtx, claims.ID)
	if err != nil {
		return apperror.ErrInternal.WithMessage("failed to verify token jti").WithPath(reqPath).Wrap(err)
	}
	if revoked {
		h.log.Warn("refresh attempt with used/revoked token", zap.String("jti", claims.ID))
		return apperror.ErrUnauthorized.WithMessage("refresh token reuse detected").WithPath(reqPath)
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return apperror.ErrUnauthorized.WithMessage("invalid user credentials").WithPath(reqPath).Wrap(err)
	}

	userData, err := h.userService.GetUserByID(reqCtx, userID)
	if err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeUnauthorized, "missing/invalid user credentials").WithPath(reqPath)
	}

	set, err := h.generateToken(reqCtx, userData)
	if err != nil {
		return apperror.ErrInternal.WithMessage("failed to generate required token(s)").WithPath(reqPath).Wrap(err)
	}

	// Mark the old refresh token JTI as revoked
	if err := h.jtiService.RevokeJTI(reqCtx, claims.ID, claims.ExpiresAt.Time); err != nil {
		h.log.Warn("failed to revoke old refresh token jti", zap.String("jti", claims.ID), zap.Error(err))
	}

	h.bindAccessTokenToResponseHeader(c, set.AccessToken)
	h.bindRefreshTokenCookie(c, set.RefreshToken, h.tokenService.GetRefreshTTL())

	if config.InProd() {
		return response.JSON(c, http.StatusOK, echo.Map{
			"message": "token refreshed successfully",
			"user":    userData.PublicView(),
		})
	}

	return response.JSON(c, http.StatusOK, echo.Map{
		"message":                  "token refreshed successfully",
		"user":                     userData.PublicView(),
		"authentication_token_set": set,
	})
}

// func (h *AuthHandler) RequestPasswordReset(c echo.Context) error {
// 	var dto dto.RequestPasswordResetDTO
// 	reqPath := utils.GetRequestUrlPath(c)

// 	if err := utils.BindAndValidate(c, &dto); err != nil {
// 		return apperror.ErrorHandler(err, apperror.ErrCodeValidation, "failed to parse and validate request body").WithPath(reqPath)
// 	}

// 	// 1. check if email exists
// 	// 2. if not return error
// 	// 3. if yes generate a request token for the email and send it to the user email as a link

// }

func (h *AuthHandler) bindAccessTokenToResponseHeader(c echo.Context, t string) {
	c.Response().Header().Set(echo.HeaderAuthorization, "Bearer "+t)
}

func (h *AuthHandler) bindRefreshTokenCookie(c echo.Context, t string, ttl time.Duration) {
	c.SetCookie(&http.Cookie{
		Name:     string(contextkey.Auth_refreshToken),
		Value:    t,
		HttpOnly: true,
		Expires:  time.Now().Add(ttl),
		MaxAge:   int(ttl.Seconds()),
		Secure:   config.InProd(),
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
}

func (h *AuthHandler) generateToken(ctx context.Context, u *model.User) (*token.AuthTokenSet, error) {
	roles := make([]string, len(u.Roles))
	for i, r := range u.Roles {
		roles[i] = r.Name
	}

	set, err := h.tokenService.GenerateAuthTokenSet(u.ID, roles)
	if err != nil {
		h.log.Error("failed to generate authentication tokens", zap.String("userID", u.ID.String()), zap.Error(err)) //
		return nil, err
	}

	// store jti
	if err := h.jtiService.RecordJTI(ctx, set.RefreshTokenJTI, set.RefreshExpiresAt); err != nil {
		h.log.Error("failed to record authentication tokens", zap.String("userID", u.ID.String()), zap.Error(err)) //
		return nil, err
	}

	return set, nil
}
