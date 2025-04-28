package middleware

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/auth"
	"github.com/vars7899/iots/pkg/auth/token"
	"github.com/vars7899/iots/pkg/contextkey"
	"go.uber.org/zap"
)

func NewJWTMiddleware(authTokenService auth.AuthTokenService, logger *zap.Logger) echo.MiddlewareFunc {
	if authTokenService == nil {
		panic("NewJWTMiddleware: authTokenService dependency is nil")
	}
	if logger == nil {
		panic("NewJWTMiddleware: logger dependency is nil")
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get(echo.HeaderAuthorization)

			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return apperror.ErrUnauthorized.WithMessage("Missing or invalid 'Authorization: Bearer <token>' header")
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			accessClaims, err := authTokenService.ParseAccessToken(tokenString)
			if err != nil {
				return apperror.ErrUnauthorized.WithMessage("Invalid or expired access token").Wrap(err)
			}

			userID, err := uuid.Parse(accessClaims.Subject)
			if err != nil {
				return apperror.ErrInvalidToken.WithMessage("Invalid token claims: User ID missing or malformed").Wrap(err)
			}

			c.Set(string(contextkey.UserIDKey), userID)
			c.Set(string(contextkey.AccessTokenClaimsKey), accessClaims)
			return next(c)
		}
	}
}

func GetAccessTokenClaims(c echo.Context) (*token.AccessTokenClaims, error) {
	claimsInterface := c.Get(string(contextkey.AccessTokenClaimsKey))
	claims, ok := claimsInterface.(*token.AccessTokenClaims)
	if !ok || claims == nil {
		return nil, apperror.ErrInternal.WithMessage("Authorization context missing").Wrap(errors.New("token claims missing from context"))
	}
	return claims, nil
}

func GetAccessUserIDClaims(c echo.Context) (*uuid.UUID, error) {
	userIDInterface := c.Get(string(contextkey.UserIDKey))
	userID, ok := userIDInterface.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return nil, apperror.ErrInternal.WithMessage("Authorization user ID missing").Wrap(errors.New("user ID missing or invalid from context"))
	}
	return &userID, nil
}
