package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/internal/cache/redis"
	"github.com/vars7899/iots/pkg/apperror"
	"go.uber.org/zap"
)

// JTIValidationMiddleware checks if a token's JTI has been revoked
func JTIValidationMiddleware(jtiStore redis.JTIStore, logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get claims from context (set by JwtTokenMiddleware)
			claims, err := GetAccessTokenClaims(c)
			if err != nil {
				return apperror.ErrUnauthorized.WithMessage("invalid token").Wrap(err)
			}

			// Extract JTI from the token claims
			jti := claims.ID
			if jti == "" {
				logger.Warn("Access token missing JTI")
				return apperror.ErrUnauthorized.WithMessage("invalid token format")
			}

			// Check if token is revoked
			ctx := c.Request().Context()
			revoked, err := jtiStore.IsJTIRevoked(ctx, jti)
			if err != nil {
				logger.Error("Failed to check JTI status", zap.String("jti", jti), zap.Error(err))
				return apperror.ErrInternal.WithMessage("authorization verification failed")
			}

			if revoked {
				logger.Debug("Attempt to use revoked token", zap.String("jti", jti))
				return apperror.ErrUnauthorized.WithMessage("token has been revoked")
			}

			return next(c)
		}
	}
}
