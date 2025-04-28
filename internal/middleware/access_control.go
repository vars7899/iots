package middleware

import (
	"errors"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/internal/api"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/auth"
	"github.com/vars7899/iots/pkg/contextkey"
	"github.com/vars7899/iots/pkg/utils"
	"go.uber.org/zap"
)

func NewAccessControlMiddleware(accessControlService auth.AccessControlService, logger *zap.Logger) echo.MiddlewareFunc {
	if accessControlService == nil {
		panic("NewAccessControlMiddleware: missing dependency accessControlService")
	}
	if logger == nil {
		panic("NewAccessControlMiddleware: missing dependency logger")
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userIDInterface := c.Get(string(contextkey.UserIDKey))
			userID, ok := userIDInterface.(uuid.UUID)
			if !ok || userID == uuid.Nil {
				return apperror.ErrInternal.WithMessage("Authorization user ID missing").Wrap(errors.New("user ID missing or invalid from context"))
			}

			action := utils.GetActionFromHTTPRequest(c)
			resource := utils.GetResourceFromRequest(c, string(api.ApiV1)) // TODO: update this for other api versions, maybe use regex for generic version

			allowed, err := accessControlService.CheckPermission(userID, resource, action)
			if err != nil {
				return apperror.ErrInternal.WithMessage("Authorization check failed due to internal error").Wrap(err)
			}
			if !allowed {
				return apperror.ErrForbidden.WithMessage("Insufficient permissions")
			}

			return next(c)
		}
	}
}
