package utils

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/pkg/apperror"
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

func BindAndValidate(c echo.Context, dto interface{}) error {
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
