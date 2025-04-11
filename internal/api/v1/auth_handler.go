package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/config"
	"github.com/vars7899/iots/internal/api/v1/dto"
	"github.com/vars7899/iots/internal/domain/user"
	"github.com/vars7899/iots/internal/middleware"
	"github.com/vars7899/iots/internal/service"
	"github.com/vars7899/iots/pkg/auth/token"
	"github.com/vars7899/iots/pkg/logger"
	"github.com/vars7899/iots/pkg/response"
	"go.uber.org/zap"
)

type AuthHandler struct {
	UserService  *service.UserService
	TokenService token.TokenService
	log          *zap.Logger
	RequestTTL   time.Duration
}

type TokenResponse struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	TokenType    string        `json:"token_type"`
	ExpiresIn    time.Duration `json:"expires_in"`
}

func NewAuthHandler(dep APIDependencies) *AuthHandler {
	_logger := logger.Lgr.Named("AuthHandler")
	return &AuthHandler{UserService: dep.UserService, TokenService: dep.TokenService, log: _logger, RequestTTL: 5 * time.Second}
}

func (h *AuthHandler) RegisterRoutes(e *echo.Group) {
	e.POST("/register", h.Register)
	e.POST("/login", h.Login)
	e.POST("/logout", h.Logout, middleware.JwtTokenMiddleware(h.TokenService))
	e.POST("/refresh", h.Refresh)
}

func (h *AuthHandler) Login(c echo.Context) error {
	var userLoginDto dto.LoginUserRequestDTO
	if err := c.Bind(&userLoginDto); err != nil {
		return response.ErrBadRequest.WithDetails(echo.Map{
			"error": "invalid request body",
		}).Wrap(err)
	}
	if err := userLoginDto.Validate(); err != nil {
		return response.ErrBadRequest.WithDetails(echo.Map{
			"error":   "validation error",
			"details": err.Error(),
		}).Wrap(err)
	}

	userData, err := h.UserService.FindByLoginIdentifier(c.Request().Context(), service.LoginIdentifier{
		Email:       userLoginDto.Email,
		Username:    userLoginDto.UserName,
		PhoneNumber: userLoginDto.PhoneNumber,
	})
	if err != nil {
		return response.ErrUnauthorized.WithDetails(echo.Map{
			"error": "invalid credentials",
		}).Wrap(err)
	}

	// Verify password
	if err := userData.ComparePassword(userLoginDto.Password); err != nil {
		return response.ErrUnauthorized.WithDetails(echo.Map{
			"error": "invalid credentials",
		}).Wrap(err)
	}

	genAccessToken, genRefreshToken, err := h.generateToken(userData)
	if err != nil {
		return response.ErrInternal.WithDetails(echo.Map{
			"error":  "something went wrong",
			"reason": "failed to generate token",
		}).Wrap(err)
	}

	h.bindAccessTokenToResponseHeader(c, *genAccessToken)
	h.bindRefreshTokenCookie(c, *genRefreshToken, h.TokenService.GetRefreshTTL())

	// update last login
	if err := h.UserService.SetLastLogin(c.Request().Context(), userData.ID); err != nil {
		h.log.Error("failed to set last login for user", zap.String("userID", userData.ID.String()), zap.Error(err))
		return response.ErrInternal.WithDetails(echo.Map{
			"error":  "something went wrong",
			"reason": "failed to update last login",
		}).Wrap(err)
	}

	return response.JSON(c, http.StatusOK, echo.Map{
		"message": "user logged in successfully",
		"user": echo.Map{
			"id":           userData.ID,
			"username":     userData.Username,
			"email":        userData.Email,
			"phone_number": userData.PhoneNumber,
		},
		"token": &TokenResponse{
			AccessToken:  *genAccessToken,
			RefreshToken: *genRefreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    h.TokenService.GetAccessTTL(),
		},
	})
}

func (h *AuthHandler) Register(c echo.Context) error {
	var userRegisterDto dto.RegisterUserRequestDTO
	if err := c.Bind(&userRegisterDto); err != nil {
		return response.ErrBadRequest.WithDetails(echo.Map{
			"error": "invalid request body",
		}).Wrap(err)
	}
	if err := userRegisterDto.Validate(); err != nil {
		return response.ErrBadRequest.WithDetails(echo.Map{
			"error": "validation error",
		}).Wrap(err)
	}

	newUserObj := &user.User{
		Username:    userRegisterDto.UserName,
		Email:       userRegisterDto.Email,
		PhoneNumber: userRegisterDto.PhoneNumber,
		Password:    userRegisterDto.Password,
	}

	createdUser, err := h.UserService.CreateUser(c.Request().Context(), newUserObj)
	if err != nil {
		return response.ErrInternal.WithDetails(echo.Map{
			"error":  "something went wrong",
			"reason": "failed to create user",
		}).Wrap(err)
	}

	genAccessToken, genRefreshToken, err := h.generateToken(createdUser)
	if err != nil {
		return response.ErrInternal.WithDetails(echo.Map{
			"error":  "something went wrong",
			"reason": "failed to generate token",
		}).Wrap(err)
	}

	h.bindAccessTokenToResponseHeader(c, *genAccessToken)
	h.bindRefreshTokenCookie(c, *genRefreshToken, h.TokenService.GetRefreshTTL())

	return response.JSON(c, http.StatusCreated, echo.Map{
		"message": "user registered successfully",
		"token": &TokenResponse{
			AccessToken:  *genAccessToken,
			RefreshToken: *genRefreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    h.TokenService.GetAccessTTL(),
		},
	})
}

func (h *AuthHandler) Logout(c echo.Context) error {
	claims, err := middleware.GetAccessTokenClaims(c)
	if err != nil {
		h.log.Warn("failed to get access token claims", zap.Error(err))
		return response.ErrUnauthorized.WithDetails(echo.Map{
			"error":  "invalid/missing authorization token",
			"reason": "failed to access token claims",
		}).Wrap(err)
	}

	c.SetCookie(&http.Cookie{
		Name:     "auth_refresh_token",
		Value:    "",
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		Secure:   config.IsUnderProduction(),
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		h.log.Error("invalid user ID in claims", zap.String("userID", claims.UserID), zap.Error(err))
		return response.ErrInvalidToken.WithDetails(echo.Map{
			"error":  "invalid/missing authorization token",
			"reason": "failed to parse user id from token",
		}).Wrap(err)
	}
	if err := h.UserService.SetLastLogin(c.Request().Context(), userID); err != nil {
		h.log.Error("failed to set last login for user", zap.String("userID", userID.String()), zap.Error(err))
		return response.ErrInternal.WithDetails(echo.Map{
			"error":  "something went wrong while trying to logout",
			"reason": "failed to update last login",
		}).Wrap(err)
	}
	return response.JSON(c, http.StatusOK, echo.Map{
		"message": "user logged out successfully",
	})
}

func (h *AuthHandler) Refresh(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), h.RequestTTL)
	defer cancel()

	refreshToken, err := c.Cookie("auth_refresh_token")

	if err != nil || refreshToken.Value == "" {
		return response.ErrUnauthorized.WithDetails(echo.Map{
			"error":  "missing/invalid refresh token",
			"reason": "failed to get refresh token",
		})
	}

	claims, err := h.TokenService.ParseRefreshToken(refreshToken.Value)
	if err != nil {
		return response.ErrUnauthorized.WithDetails(echo.Map{
			"error":  "invalid/expired refresh token",
			"reason": "failed to parse refresh token",
		}).Wrap(err)
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return response.ErrUnauthorized.WithDetails(echo.Map{
			"error":  "invalid user credentials",
			"reason": "failed to parse user id from token",
		}).Wrap(err)
	}

	userData, err := h.UserService.GetUserByID(ctx, userID)
	if err != nil {
		return response.ErrUnauthorized.WithDetails(echo.Map{
			"error": "invalid user credentials",
		}).Wrap(err)
	}

	newAccessToken, newRefreshToken, err := h.generateToken(userData)
	if err != nil {
		return response.ErrInternal.WithDetails(echo.Map{
			"error":  "something went wrong",
			"reason": "failed to generate token",
		}).Wrap(err)
	}

	h.bindAccessTokenToResponseHeader(c, *newAccessToken)
	h.bindRefreshTokenCookie(c, *newRefreshToken, h.TokenService.GetRefreshTTL())

	return response.JSON(c, http.StatusOK, echo.Map{
		"message": "token refreshed successfully",
		"token": &TokenResponse{
			AccessToken:  *newAccessToken,
			RefreshToken: *newRefreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    h.TokenService.GetAccessTTL(),
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
		Secure:   config.IsUnderProduction(), // todo: turn to true under production
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
}

func (h *AuthHandler) generateToken(u *user.User) (accessToken *string, refreshToken *string, reason error) {
	var genAccessToken string
	var genRefreshToken string
	var err error
	// Map roles to []string
	_roles := make([]string, len(u.Roles))
	for i, r := range u.Roles {
		_roles[i] = r.Name
	}
	if genAccessToken, err = h.TokenService.GenerateAccessToken(u.ID, _roles); err != nil {
		h.log.Error("failed to generate access token", zap.String("userID", u.ID.String()), zap.Error(err)) //
		return nil, nil, err
	}
	if genRefreshToken, err = h.TokenService.GenerateRefreshToken(u.ID); err != nil {
		h.log.Error("failed to generate refresh token", zap.String("userID", u.ID.String()), zap.Error(err)) //
		return nil, nil, err
	}
	return &genAccessToken, &genRefreshToken, nil
}
