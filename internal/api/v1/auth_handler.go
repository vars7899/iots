package v1

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/internal/api/v1/dto"
	"github.com/vars7899/iots/internal/domain/user"
	"github.com/vars7899/iots/internal/service"
	"github.com/vars7899/iots/pkg/auth/token"
	"github.com/vars7899/iots/pkg/response"
)

type AuthHandler struct {
	UserService  *service.UserService
	TokenService token.TokenService
}

func NewAuthHandler(dep APIDependencies) *AuthHandler {
	return &AuthHandler{UserService: dep.UserService, TokenService: dep.TokenService}
}

func (h *AuthHandler) RegisterRoutes(e *echo.Group) {
	e.POST("/register", h.Register)
	e.POST("/login", h.Login)
	e.POST("/logout", h.Logout)
	e.POST("/refresh", h.Refresh)
}

func (h *AuthHandler) Login(c echo.Context) error {
	return response.Error(c, http.StatusBadRequest, "invalid request body")
	// var _req dto.LoginUserRequestDTO
	// if err := c.Bind(&_req); err != nil {
	// 	return response.Error(c, http.StatusBadRequest, "invalid request body")
	// }
	// if err := _req.Validate(); err != nil {
	// 	return response.Error(c, http.StatusBadRequest, err.Error())
	// }

	// _identifier := ""
	// if _req.Email != "" {
	// 	_identifier = _req.Email
	// } else if _req.UserName != "" {
	// 	_identifier = _req.UserName
	// } else if _req.PhoneNumber != "" {
	// 	_identifier = _req.PhoneNumber
	// } else {
	// 	return response.Error(c, http.StatusBadRequest, "missing user identifier data, provide either email or username or phone number")
	// }

	// // Generate token (example placeholder)
	// token := "dummy-jwt-token"

	// return c.JSON(http.StatusOK, echo.Map{
	// 	"token": token,
	// 	"user":  _identifier, // be careful to exclude sensitive fields
	// })
}

func (h *AuthHandler) Register(c echo.Context) error {

	var _req dto.RegisterUserRequestDTO
	if err := c.Bind(&_req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body")
	}
	if err := _req.Validate(); err != nil {
		return response.Error(c, http.StatusBadRequest, err.Error())
	}

	u := &user.User{
		Username:    _req.UserName,
		Email:       _req.Email,
		PhoneNumber: _req.PhoneNumber,
		Password:    _req.Password,
	}

	_createdUser, err := h.UserService.CreateUser(c.Request().Context(), u)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, err.Error())
	}

	// Map roles to []string
	_roles := make([]string, len(_createdUser.Roles))
	for i, r := range _createdUser.Roles {
		_roles[i] = r.Name
	}

	_accessToken, err := h.TokenService.GenerateAccessToken(_createdUser.ID, _roles)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
	}
	_refreshToken, err := h.TokenService.GenerateRefreshToken(_createdUser.ID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
	}

	return response.JSON(c, http.StatusCreated, echo.Map{
		"message":       "user registered successfully",
		"access_token":  _accessToken,
		"refresh_token": _refreshToken,
		"token_type":    "Bearer",
		"expires_in":    h.TokenService.GetAccessTTL().Seconds(),
	})
}

func (h *AuthHandler) Logout(c echo.Context) error {
	return nil
}

func (h *AuthHandler) Refresh(c echo.Context) error {
	return response.JSON(c, http.StatusOK, echo.Map{
		"token": "to be implemented",
	})
}
