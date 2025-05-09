package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/pkg/auth"
	"go.uber.org/zap"
)

type MiddlewareRegistry struct {
	JWT                echo.MiddlewareFunc
	JTI                echo.MiddlewareFunc
	ErrorHandler       echo.MiddlewareFunc
	AccessControl      echo.MiddlewareFunc
	PermissionRequired func(resource string, action string) echo.MiddlewareFunc
}

func NewMiddlewareRegistry(authTokenService auth.AuthTokenService, accessControlService auth.AccessControlService, logger *zap.Logger) *MiddlewareRegistry {

	return &MiddlewareRegistry{
		JWT:                NewJWTMiddleware(authTokenService, logger),
		JTI:                NewJTIMiddleware(authTokenService, logger),
		ErrorHandler:       NewErrorHandlerMiddleware(logger),
		AccessControl:      NewAccessControlMiddleware(accessControlService, logger),
		PermissionRequired: NewPermissionRequiredMiddlewareGenerator(accessControlService, logger),
	}
}
