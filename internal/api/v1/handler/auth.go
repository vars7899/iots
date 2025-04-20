package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/config"
	"github.com/vars7899/iots/internal/api/v1/dto"
	"github.com/vars7899/iots/internal/domain/model"
	"github.com/vars7899/iots/internal/middleware"
	"github.com/vars7899/iots/internal/service"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/auth/token"
	"github.com/vars7899/iots/pkg/di"
	"github.com/vars7899/iots/pkg/logger"
	"github.com/vars7899/iots/pkg/response"
	"github.com/vars7899/iots/pkg/utils"
	"go.uber.org/zap"
)

type AuthHandler struct {
	userService  *service.UserService
	tokenService token.TokenService
	log          *zap.Logger
}

type TokenResponse struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	TokenType    string        `json:"token_type"`
	ExpiresIn    time.Duration `json:"expires_in"`
}

func NewAuthHandler(deps di.Provider, baseLogger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		userService:  deps.Services.UserService,
		tokenService: deps.Helpers.TokenService,
		log:          logger.Named(baseLogger, "AuthHandler"),
	}
}

func (h *AuthHandler) RegisterRoutes(e *echo.Group) {
	e.POST("/register", h.Register)
	e.POST("/login", h.Login)
	e.POST("/logout", h.Logout, middleware.JwtTokenMiddleware(h.tokenService))
	e.POST("/refresh", h.Refresh)
}

func (h *AuthHandler) Login(c echo.Context) error {
	var dto dto.LoginUserRequestDTO
	reqPath := utils.GetRequestUrlPath(c)

	// Bind body request and validate fields
	if err := utils.BindAndValidate(c, &dto); err != nil {
		return err
	}

	userExist, err := h.userService.FindByLoginIdentifier(c.Request().Context(), service.LoginIdentifier{
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

	genAccessToken, genRefreshToken, err := h.generateToken(userExist)
	if err != nil {
		return apperror.ErrInternal.WithMessage("failed to generate required token(s)").WithPath(reqPath).Wrap(err)
	}

	h.bindAccessTokenToResponseHeader(c, *genAccessToken)
	h.bindRefreshTokenCookie(c, *genRefreshToken, h.tokenService.GetRefreshTTL())

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
		"message": "user logged in successfully",
		"user":    userExist.PublicView(),

		// TODO: Remove token
		"token": &TokenResponse{
			AccessToken:  *genAccessToken,
			RefreshToken: *genRefreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    h.tokenService.GetAccessTTL(),
		},
	})
}

func (h *AuthHandler) Register(c echo.Context) error {
	var dto dto.RegisterUserRequestDTO
	reqPath := utils.GetRequestUrlPath(c)

	// Bind body request and validate fields
	if err := utils.BindAndValidate(c, &dto); err != nil {
		return err
	}

	createdUser, err := h.userService.CreateUser(c.Request().Context(), dto.AsModel())
	if err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeInternal, "something went wrong while registering user").WithPath(reqPath)
	}

	genAccessToken, genRefreshToken, err := h.generateToken(createdUser)
	if err != nil {
		return apperror.ErrInternal.WithMessage("failed to generate required token(s)").WithPath(reqPath).Wrap(err)
	}

	h.bindAccessTokenToResponseHeader(c, *genAccessToken)
	h.bindRefreshTokenCookie(c, *genRefreshToken, h.tokenService.GetRefreshTTL())

	if config.InProd() {
		return response.JSON(c, http.StatusOK, echo.Map{
			"message": "user registered successfully",
			"user":    createdUser.PublicView(),
		})
	}

	return response.JSON(c, http.StatusCreated, echo.Map{
		"message": "user registered successfully",
		"user":    createdUser.PublicView(),
		"token": &TokenResponse{
			AccessToken:  *genAccessToken,
			RefreshToken: *genRefreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    h.tokenService.GetAccessTTL(),
		},
	})
}

func (h *AuthHandler) Logout(c echo.Context) error {
	claims, err := middleware.GetAccessTokenClaims(c)
	reqPath := utils.GetRequestUrlPath(c)

	if err != nil {
		h.log.Warn("failed to get access token claims", zap.Error(err))
		return apperror.ErrUnauthorized.WithMessage("invalid/missing authorization token").WithPath(reqPath).Wrap(err)
	}

	c.SetCookie(&http.Cookie{
		Name:     "auth_refresh_token",
		Value:    "",
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		Secure:   config.InProd(),
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		h.log.Error("invalid user ID in claims, failed to parse token", zap.String("userID", claims.UserID), zap.Error(err))
		return apperror.ErrInvalidToken.WithMessage("invalid/missing authorization token").WithPath(reqPath).Wrap(err)
	}
	if err := h.userService.SetLastLogin(c.Request().Context(), userID); err != nil {
		h.log.Error("failed to set last login for user", zap.String("userID", userID.String()), zap.Error(err))
	}
	return response.JSON(c, http.StatusOK, echo.Map{
		"message": "user logged out successfully",
	})
}

func (h *AuthHandler) Refresh(c echo.Context) error {
	reqPath := utils.GetRequestUrlPath(c)

	refreshToken, err := c.Cookie("auth_refresh_token")

	if err != nil || refreshToken.Value == "" {
		return apperror.ErrUnauthorized.WithMessage("missing/invalid refresh token").WithPath(reqPath).Wrap(err)
	}

	claims, err := h.tokenService.ParseRefreshToken(refreshToken.Value)
	if err != nil {
		return apperror.ErrUnauthorized.WithMessage("missing/invalid refresh token, failed to parse token").WithPath(reqPath).Wrap(err)
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		apperror.ErrUnauthorized.WithMessage("invalid user credentials").WithPath(reqPath).Wrap(err)
	}

	userData, err := h.userService.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeUnauthorized, "missing/invalid user credentials").WithPath(reqPath)
	}

	newAccessToken, newRefreshToken, err := h.generateToken(userData)
	if err != nil {
		return apperror.ErrInternal.WithMessage("failed to generate required token(s)").WithPath(reqPath).Wrap(err)
	}

	h.bindAccessTokenToResponseHeader(c, *newAccessToken)
	h.bindRefreshTokenCookie(c, *newRefreshToken, h.tokenService.GetRefreshTTL())

	if config.InProd() {
		return response.JSON(c, http.StatusOK, echo.Map{
			"message": "token refreshed successfully",
			"user":    userData.PublicView(),
		})
	}

	return response.JSON(c, http.StatusOK, echo.Map{
		"message": "token refreshed successfully",
		"user":    userData.PublicView(),
		"token": &TokenResponse{
			AccessToken:  *newAccessToken,
			RefreshToken: *newRefreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    h.tokenService.GetAccessTTL(),
		},
	})
}

func (h *AuthHandler) bindAccessTokenToResponseHeader(c echo.Context, t string) {
	c.Response().Header().Set(echo.HeaderAuthorization, "Bearer "+t)
}

func (h *AuthHandler) bindRefreshTokenCookie(c echo.Context, t string, ttl time.Duration) {

	c.SetCookie(&http.Cookie{
		Name:     "auth_refresh_token",
		Value:    t,
		HttpOnly: true,
		Expires:  time.Now().Add(ttl),
		Secure:   config.InProd(),
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
}

func (h *AuthHandler) generateToken(u *model.User) (accessToken *string, refreshToken *string, reason error) {
	var genAccessToken string
	var genRefreshToken string
	var err error
	// Map roles to []string
	roles := make([]string, len(u.Roles))
	for i, r := range u.Roles {
		roles[i] = r.Name
	}
	if genAccessToken, err = h.tokenService.GenerateAccessToken(u.ID, roles); err != nil {
		h.log.Error("failed to generate access token", zap.String("userID", u.ID.String()), zap.Error(err)) //
		return nil, nil, err
	}
	if genRefreshToken, err = h.tokenService.GenerateRefreshToken(u.ID); err != nil {
		h.log.Error("failed to generate refresh token", zap.String("userID", u.ID.String()), zap.Error(err)) //
		return nil, nil, err
	}
	return &genAccessToken, &genRefreshToken, nil
}
