package utils

import (
	"context"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/contextkey"
)

func ConvertVectorToPointerVector[T any](i []T) []*T {
	o := make([]*T, len(i))
	for idx := range i {
		o[idx] = &i[idx]
	}
	return o
}

// TODO: might remove this later
func CheckContextForError(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return apperror.ErrContextCancelled.WithDetails(err)
	}
	return nil
}

func GetRequestUrlPath(c echo.Context) string {
	return c.Request().URL.Path
}

func BindAndValidate(c echo.Context, dto interface{}) *apperror.AppError {
	reqPath := GetRequestUrlPath(c)
	if err := c.Bind(dto); err != nil {
		return apperror.ErrBadRequest.WithMessage("invalid request body").WithDetails(echo.Map{
			"error": err.Error(),
		}).WithPath(reqPath).Wrap(err)
	}

	if v, ok := dto.(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return apperror.ErrValidation.WithMessage("validation failed").WithDetails(echo.Map{
				"error": err.Error(),
			}).WithPath(reqPath).Wrap(err)
		}
	}
	return nil
}

func GetActionFromHTTPRequest(c echo.Context) string {
	method := c.Request().Method
	switch method {
	case "GET":
		return string(contextkey.ActionRead)
	case "PATCH", "PUT":
		return string(contextkey.ActionUpdate)
	case "DELETE":
		return string(contextkey.ActionDelete)
	case "POST":
		return string(contextkey.ActionCreate)
	default:
		return strings.ToLower(method)
	}
}

func GetResourceFromRequest(c echo.Context, versionPrefix string) string {
	path := c.Request().URL.Path

	if strings.HasPrefix(path, versionPrefix) {
		path = strings.TrimPrefix(path, versionPrefix+"/")
	}

	pathParts := strings.Split(path, "/")
	if len(pathParts) > 0 {
		return pathParts[0]
	}
	return path
}
