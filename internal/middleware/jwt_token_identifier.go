package middleware

import (
	"errors"

	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/auth"
	"go.uber.org/zap"
)

func NewJTIMiddleware(authTokenService auth.AuthTokenService, logger *zap.Logger) echo.MiddlewareFunc {
	if authTokenService == nil {
		panic("NewJTIMiddleware: authTokenService is nil")
	}
	if authTokenService == nil {
		panic("NewJTIMiddleware: logger is nil")
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims, err := GetAccessTokenClaims(c)
			if err != nil {
				return apperror.ErrorHandler(err, apperror.ErrCodeUnauthorized, "Invalid token")
			}

			jti := claims.ID
			if jti == "" {
				return apperror.ErrInvalidToken.WithMessage("Invalid token format: JTI claim missing").Wrap(errors.New("JTI claim is empty"))
			}

			revoked, err := authTokenService.IsJTIRevoked(c.Request().Context(), jti)
			if err != nil {
				return apperror.ErrInternal.WithMessage("Authorization verification failed due to internal error").Wrap(err)
			}

			if revoked {
				return apperror.ErrUnauthorized.WithMessage("Access token has been revoked")
			}

			return next(c)
		}
	}
}
