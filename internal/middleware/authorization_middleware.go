package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/internal/api"
	"github.com/vars7899/iots/internal/service"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/auth/token"
	"github.com/vars7899/iots/pkg/contextkey"
	"github.com/vars7899/iots/pkg/utils"
	"go.uber.org/zap"
)

func CasbinAuthorizationMiddleware(casbinService service.CasbinService, logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claimsInterface := c.Get(string(contextkey.AccessTokenClaimsKey))
			claims, ok := claimsInterface.(*token.AccessTokenClaims)
			if !ok {
				logger.Error("failed to get token claims from request context")
				return apperror.ErrUnauthorized.WithMessage("authorization verification failed")
			}

			userID, err := uuid.Parse(claims.UserID)
			if err != nil {
				logger.Error("failed to parse user id as uuid provided by claims")
				return apperror.ErrInvalidToken.WithMessage("invalid token format")
			}

			// get resource and action from request
			action := utils.GetActionFromHTTPRequest(c)
			resource := utils.GetResourceFromRequest(c, string(api.ApiV1)) // TODO: update this for other api versions, maybe use regex for generic version

			allowed, err := casbinService.CheckPermission(userID, resource, action)
			if err != nil {
				logger.Error("failed to check user permission")
				return apperror.ErrInternal.WithMessage("authorization check failed, please try again later")
			}

			if !allowed {
				logger.Debug("Permission denied", zap.String("user_id", userID.String()), zap.String("resource", resource), zap.String("action", action))
				return apperror.ErrForbidden.WithMessage("insufficient permissions")
			}

			return next(c)
		}
	}
}
