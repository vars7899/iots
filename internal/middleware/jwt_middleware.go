package middleware

import (
	"errors"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/auth/token"
	"github.com/vars7899/iots/pkg/contextkey"
	"go.uber.org/zap"
)

// JwtTokenMiddleware validates JWT tokens and adds claims to context
func JwtTokenMiddleware(tokenService token.TokenService, logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get(echo.HeaderAuthorization)

			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return apperror.ErrUnauthorized.WithMessage("missing/invalid authorization token")
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			accessTokenClaims, err := tokenService.ParseAccessToken(tokenString)
			if err != nil {
				logger.Debug("Token validation failed", zap.Error(err))
				return apperror.ErrUnauthorized.WithMessage("invalid/expired authorization token").Wrap(err)
			}

			// Store claims in context for the handler
			c.Set(string(contextkey.AccessTokenClaimsKey), accessTokenClaims)
			return next(c)
		}
	}
}

func GetAccessTokenClaims(c echo.Context) (*token.AccessTokenClaims, error) {
	claims, ok := c.Get(string(contextkey.AccessTokenClaimsKey)).(*token.AccessTokenClaims)
	if !ok || claims == nil {
		return nil, errors.New("access token claims not found in context")
	}
	return claims, nil
}
