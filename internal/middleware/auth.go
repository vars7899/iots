package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/pkg/auth"
	"go.uber.org/zap"
)

func NewAuthMiddleware(authTokenService auth.AuthTokenService, accessControlService auth.AccessControlService, logger *zap.Logger) []echo.MiddlewareFunc {
	if authTokenService == nil {
		panic("NewAuthMiddleware: authTokenService is nil")
	}
	if accessControlService == nil {
		panic("NewAuthMiddleware: accessControlService is nil")
	}
	if logger == nil {
		panic("NewAuthMiddleware: logger is nil")
	}

	jwtMiddleware := NewJWTMiddleware(authTokenService, logger)
	jtiMiddleware := NewJTIMiddleware(authTokenService, logger)
	casbinMiddleware := NewAccessControlMiddleware(accessControlService, logger)

	return []echo.MiddlewareFunc{
		jwtMiddleware,
		jtiMiddleware,
		casbinMiddleware,
	}
}
