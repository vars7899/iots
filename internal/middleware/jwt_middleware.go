package middleware

import (
	"errors"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/pkg/auth/token"
	"github.com/vars7899/iots/pkg/response"
)

func JwtTokenMiddleware(tokenService token.TokenService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authorizationHeader := c.Request().Header.Get(echo.HeaderAuthorization)
			if authorizationHeader == "" || !strings.HasPrefix(authorizationHeader, "Bearer") {
				return response.ErrUnauthorized.WithDetails(echo.Map{
					"error": "missing/invalid authorization token",
				})
			}
			tokenString := strings.TrimPrefix(authorizationHeader, "Bearer ")

			_claims, err := tokenService.ParseAccessToken(tokenString)
			if err != nil {
				return response.ErrUnauthorized.WithDetails(echo.Map{
					"error": "invalid/expired authorization token",
				})
			}
			c.Set("claims", _claims)
			return next(c)
		}
	}
}

func GetAccessTokenClaims(c echo.Context) (*token.AccessTokenClaims, error) {
	claims, ok := c.Get("claims").(*token.AccessTokenClaims)
	if !ok || claims == nil {
		return nil, errors.New("access token claims not found in context")
	}
	return claims, nil
}
