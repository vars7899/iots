package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/internal/cache"
	"github.com/vars7899/iots/internal/service"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/auth/token"
	"github.com/vars7899/iots/pkg/contextkey"
	"go.uber.org/zap"
)

func AuthRequired(e *echo.Echo, tokenService token.TokenService, jtiStore cache.JTIStore, casbinService service.CasbinService, logger *zap.Logger) echo.MiddlewareFunc {
	// Chain the middlewares: first JWT authentication, then Casbin authorization
	jwtMiddleware := JWT_JTI_Middleware(tokenService, jtiStore, logger)
	casbinMiddleware := CasbinAuthorizationMiddleware(casbinService, logger)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return jwtMiddleware(casbinMiddleware(next))
	}
}

// CombinedJWTMiddleware combines token parsing and JTI validation in one middleware
func JWT_JTI_Middleware(tokenService token.TokenService, jtiStore cache.JTIStore, logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract and validate token
			authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return apperror.ErrUnauthorized.WithMessage("missing/invalid authorization token")
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			accessTokenClaims, err := tokenService.ParseAccessToken(tokenString)
			if err != nil {
				logger.Debug("Token validation failed", zap.String("token", tokenString), zap.Error(err))
				return apperror.ErrUnauthorized.WithMessage("invalid/expired authorization token").Wrap(err)
			}

			// Extract JTI and validate it's not revoked
			jti := accessTokenClaims.ID
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

			// Store claims in context for the handler
			c.Set(string(contextkey.AccessTokenClaimsKey), accessTokenClaims)
			return next(c)
		}
	}
}
