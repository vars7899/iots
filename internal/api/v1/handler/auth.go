package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/config"
	"github.com/vars7899/iots/internal/api/v1/dto"
	"github.com/vars7899/iots/internal/middleware"
	"github.com/vars7899/iots/internal/service"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/auth"
	"github.com/vars7899/iots/pkg/contextkey"
	"github.com/vars7899/iots/pkg/di"
	"github.com/vars7899/iots/pkg/logger"
	"github.com/vars7899/iots/pkg/response"
	"github.com/vars7899/iots/pkg/utils"
	"go.uber.org/zap"
)

type AuthHandler struct {
	authService          service.AuthService
	authTokenService     auth.AuthTokenService
	accessControlService auth.AccessControlService
	config               *config.AppConfig
	logger               *zap.Logger
}

func NewAuthHandler(container *di.AppContainer, baseLogger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authService:          container.Services.AuthService,
		authTokenService:     container.CoreServices.AuthTokenService,
		accessControlService: container.CoreServices.AccessControlService,
		config:               container.Config,
		logger:               logger.Named(baseLogger, "AuthHandler"),
	}
}

func (h *AuthHandler) SetupRoutes(e *echo.Group) {
	e.POST("/register", h.Register)
	e.POST("/login", h.Login)
	e.POST("/token/refresh", h.Refresh)
	e.POST("/request-password-reset", h.RequestPasswordReset)
	e.POST("/password-reset", h.PasswordReset)

	// protected routes
	e.POST("/logout", h.Logout, middleware.NewJWTMiddleware(h.authTokenService, h.logger), middleware.NewJTIMiddleware(h.authTokenService, h.logger))
}

func (h *AuthHandler) Login(c echo.Context) error {
	ctx := c.Request().Context()
	var dto dto.LoginUserRequestDTO

	if err := utils.BindAndValidate(c, &dto); err != nil {
		return err
	}

	userData, tokenData, err := h.authService.LoginUser(ctx, dto.AsModel())
	if err != nil {
		return err
	}

	h.bindAccessTokenToHeader(c, tokenData.AccessToken)
	h.bindRefreshTokenToCookie(c, tokenData.RefreshToken)

	responsePayload := echo.Map{
		"message": "user logged in successfully",
		"user":    userData.PublicView(),
	}
	if !config.InProd() {
		responsePayload["authorization_token_set"] = tokenData
	}
	return response.JSON(c, http.StatusOK, responsePayload)
}

func (h *AuthHandler) Register(c echo.Context) error {
	ctx := c.Request().Context()
	var dto dto.RegisterUserRequestDTO

	if err := utils.BindAndValidate(c, &dto); err != nil {
		return err
	}

	user, token, err := h.authService.RegisterUser(ctx, dto.AsModel())
	if err != nil {
		return err
	}

	h.bindAccessTokenToHeader(c, token.AccessToken)
	h.bindRefreshTokenToCookie(c, token.RefreshToken)

	responsePayload := echo.Map{
		"message": "user logged in successfully",
		"user":    user.PublicView(),
	}
	if !config.InProd() {
		responsePayload["authorization_token_set"] = token
	}
	return response.JSON(c, http.StatusOK, responsePayload)
}

func (h *AuthHandler) Logout(c echo.Context) error {
	ctx := c.Request().Context()

	claims, err := middleware.GetAccessTokenClaims(c)
	if err != nil {
		return err
	}
	userID, err := middleware.GetAccessUserIDClaims(c)
	if err != nil {
		return err
	}

	refreshTokenValue := ""
	refreshCookie, err := c.Cookie(string(contextkey.AuthRefreshToken))
	if err == nil && refreshCookie.Value != "" {
		refreshTokenValue = refreshCookie.Value
	}

	if err := h.authService.LogoutUser(ctx, userID, claims, refreshTokenValue); err != nil {
		return err
	}

	h.expireRefreshTokenCookie(c)
	return response.JSON(c, http.StatusOK, echo.Map{
		"message": "user logged out successfully",
	})
}

func (h *AuthHandler) Refresh(c echo.Context) error {
	reqPath := utils.GetRequestUrlPath(c)
	ctx := c.Request().Context()

	// lookup for auth refresh token in cookie
	refreshToken, err := c.Cookie(string(contextkey.AuthRefreshToken))
	if err != nil || refreshToken.Value == "" {
		return apperror.ErrUnauthorized.WithMessage("missing/invalid refresh token").WithPath(reqPath).Wrap(err)
	}

	// validate & generate user auth tokens
	user, token, err := h.authService.RefreshAuthTokens(ctx, refreshToken.Value)
	if err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeUnauthorized, "failed to refresh tokens").WithPath(reqPath)
	}

	// send proper response
	h.bindAccessTokenToHeader(c, token.AccessToken)
	h.bindRefreshTokenToCookie(c, token.RefreshToken)

	responsePayload := echo.Map{
		"message": "user logged in successfully",
		"user":    user.PublicView(),
	}
	if !config.InProd() {
		responsePayload["authorization_token_set"] = token
	}
	return response.JSON(c, http.StatusOK, responsePayload)
}

func (h *AuthHandler) RequestPasswordReset(c echo.Context) error {
	var dto dto.RequestPasswordResetDTO
	ctx := c.Request().Context()

	if err := utils.BindAndValidate(c, &dto); err != nil {
		return err
	}

	responsePayload := echo.Map{
		"message": "If your email is registered with us, you will receive password reset instructions",
	}
	resetToken, resetLink, err := h.authService.RequestPasswordReset(ctx, dto.Email)

	if err != nil {
		return response.JSON(c, http.StatusOK, responsePayload)
	}

	if !config.InProd() {
		responsePayload["reset_password_token"] = resetToken
		responsePayload["reset_link"] = resetLink
	}
	return response.JSON(c, http.StatusOK, responsePayload)
}

func (h *AuthHandler) PasswordReset(c echo.Context) error {
	var dto dto.ResetPasswordDTO
	ctx := c.Request().Context()

	if err := utils.BindAndValidate(c, &dto); err != nil {
		h.logger.Warn("invalid reset password request", zap.Error(err))
		return err
	}

	if err := h.authService.ResetPassword(ctx, dto.Token, dto.NewPassword); err != nil {
		return err
	}

	return response.JSON(c, http.StatusOK, echo.Map{
		"message": "Password reset successful.",
	})
}

func (h *AuthHandler) bindAccessTokenToHeader(c echo.Context, t string) {
	c.Response().Header().Set(echo.HeaderAuthorization, "Bearer "+t)
}

func (h *AuthHandler) bindRefreshTokenToCookie(c echo.Context, tokenValue string) {
	ttl := h.config.Jwt.RefreshTokenTTL
	maxAgeSeconds := int(ttl.Seconds())

	if maxAgeSeconds < 0 {
		maxAgeSeconds = -1
	}

	c.SetCookie(&http.Cookie{
		Name:     string(contextkey.AuthRefreshToken),
		Value:    tokenValue,
		HttpOnly: true,
		Expires:  time.Now().Add(ttl),
		MaxAge:   maxAgeSeconds,
		Secure:   config.InProd(),
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
}

func (h *AuthHandler) expireRefreshTokenCookie(c echo.Context) {
	c.SetCookie(&http.Cookie{
		Name:     string(contextkey.AuthRefreshToken),
		Value:    "",
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		Secure:   config.InProd(),
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
}
