package v1

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
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
}

type TokenResponse struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	TokenType    string        `json:"TokenType"`
	ExpiresIn    time.Duration `json:"expires_in"`
}

func NewAuthHandler(dep APIDependencies) *AuthHandler {
	_logger := logger.Lgr.Named("AuthHandler")
	return &AuthHandler{UserService: dep.UserService, TokenService: dep.TokenService, log: _logger}
}

func (h *AuthHandler) RegisterRoutes(e *echo.Group) {
	e.POST("/register", h.Register)
	e.POST("/login", h.Login)
	e.POST("/logout", h.Logout, middleware.JwtTokenMiddleware(h.TokenService))
	e.POST("/refresh", h.Refresh)
}

func (h *AuthHandler) Login(c echo.Context) error {
	var _req dto.LoginUserRequestDTO
	if err := c.Bind(&_req); err != nil {
		return response.ErrBadRequest.WithDetails(echo.Map{
			"error": "invalid request body format",
		})
	}
	if err := _req.Validate(); err != nil {
		return response.ErrBadRequest.WithDetails(echo.Map{
			"error": "validation failed",
		}).Wrap(err)
	}

	// Lookup user using service (supports multiple identifiers)
	_user, err := h.UserService.FindByLoginIdentifier(c.Request().Context(), service.LoginIdentifier{
		Email:       _req.Email,
		Username:    _req.UserName,
		PhoneNumber: _req.PhoneNumber,
	})
	if err != nil {
		return response.ErrUnauthorized.WithDetails(echo.Map{
			"error": "invalid user credentials, invalid user identifier or password",
		}).Wrap(err)
	}

	// Verify password
	if err := _user.ComparePassword(_req.Password); err != nil {
		return response.ErrUnauthorized.WithDetails(echo.Map{
			"error": "invalid user credentials, invalid user identifier or password",
		})
	}

	_accessToken, _refreshToken, err := h.generateToken(_user)
	if err != nil {
		return response.ErrInternal.WithDetails(echo.Map{
			"error": err,
		})
	}

	// bind access token to response header
	h.bindAccessTokenToResponseHeader(c, *_accessToken)
	// bind refresh token to cookie
	h.bindRefreshTokenCookie(c, *_refreshToken, h.TokenService.GetRefreshTTL())
	// update last login
	if err := h.UserService.SetLastLogin(c.Request().Context(), _user.ID); err != nil {
		h.log.Error("failed to set last login for user", zap.String("userID", _user.ID.String()), zap.Error(err))
		return response.ErrInternal.WithDetails(echo.Map{
			"error": err,
		})
	}

	return response.JSON(c, http.StatusOK, echo.Map{
		"message": "login successful",
		"user": echo.Map{
			"id":           _user.ID,
			"username":     _user.Username,
			"email":        _user.Email,
			"phone_number": _user.PhoneNumber,
		},
		"token": &TokenResponse{
			AccessToken:  *_accessToken,
			RefreshToken: *_refreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    h.TokenService.GetAccessTTL(),
		},
	})
}

func (h *AuthHandler) Register(c echo.Context) error {
	var _req dto.RegisterUserRequestDTO
	if err := c.Bind(&_req); err != nil {
		return response.ErrBadRequest.WithDetails(echo.Map{
			"error": "invalid request body format",
		})
	}
	if err := _req.Validate(); err != nil {
		return response.ErrBadRequest.WithDetails(echo.Map{
			"error": err,
		})
	}

	u := &user.User{
		Username:    _req.UserName,
		Email:       _req.Email,
		PhoneNumber: _req.PhoneNumber,
		Password:    _req.Password,
	}

	_createdUser, err := h.UserService.CreateUser(c.Request().Context(), u)
	if err != nil {
		return response.ErrInternal.WithDetails(echo.Map{
			"error": err,
		})
	}

	_accessToken, _refreshToken, err := h.generateToken(_createdUser)
	if err != nil {
		return response.ErrInternal.WithDetails(echo.Map{
			"error": err,
		})
	}

	// bind access token to response header
	h.bindAccessTokenToResponseHeader(c, *_accessToken)
	// bind refresh token to cookie
	h.bindRefreshTokenCookie(c, *_refreshToken, h.TokenService.GetRefreshTTL())

	return response.JSON(c, http.StatusCreated, echo.Map{
		"message": "user registered successfully",
		"token": &TokenResponse{
			AccessToken:  *_accessToken,
			RefreshToken: *_refreshToken,
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
			"error": "invalid/missing authorization token",
		}).Wrap(err)
	}

	c.SetCookie(&http.Cookie{
		Name:     "auth_refresh_token",
		Value:    "",
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
		Secure:   false, // todo: turn to true under production
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	_userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		h.log.Error("invalid user ID in claims", zap.String("userID", claims.UserID), zap.Error(err))
		return response.ErrInternal.WithDetails(echo.Map{
			"error": err,
		})
	}
	if err := h.UserService.SetLastLogin(c.Request().Context(), _userID); err != nil {
		h.log.Error("failed to set last login for user", zap.String("userID", _userID.String()), zap.Error(err))
		return response.ErrInternal.WithDetails(echo.Map{
			"error": err,
		})
	}
	h.log.Info("user logged out successfully", zap.String("userID", _userID.String()))
	return response.JSON(c, http.StatusOK, echo.Map{
		"message": "logged out successfully",
	})
}

func (h *AuthHandler) Refresh(c echo.Context) error {
	return response.JSON(c, http.StatusOK, echo.Map{
		"token": "to be implemented",
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
		Secure:   false, // todo: turn to true under production
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})
}

func (h *AuthHandler) generateToken(u *user.User) (*string, *string, error) {
	var _accessToken string
	var _refreshToken string
	var err error
	// Map roles to []string
	_roles := make([]string, len(u.Roles))
	for i, r := range u.Roles {
		_roles[i] = r.Name
	}
	if _accessToken, err = h.TokenService.GenerateAccessToken(u.ID, _roles); err != nil {
		return nil, nil, err
	}
	if _refreshToken, err = h.TokenService.GenerateRefreshToken(u.ID); err != nil {
		return nil, nil, err
	}
	return &_accessToken, &_refreshToken, nil
}
