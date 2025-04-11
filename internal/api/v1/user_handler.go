package v1

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/internal/service"
	"github.com/vars7899/iots/pkg/response"
)

type UserHandler struct {
	UserService *service.UserService
}

func NewUserHandler(dep APIDependencies) *UserHandler {
	return &UserHandler{UserService: dep.UserService}
}

func (h *UserHandler) RegisterRoutes(e *echo.Group) {
	e.GET("", h.GetUsers)
}

// TODO: dismiss this route for testing only
func (h *UserHandler) GetUsers(c echo.Context) error {
	_userList, err := h.UserService.GetUser(c.Request().Context())
	if err != nil {
		return response.ErrBadRequest.WithDetails(echo.Map{
			"error": err,
		})
	}
	return response.JSON(c, http.StatusOK, echo.Map{
		"message": "successfully retrieved user list",
		"users":   _userList,
	})
}
