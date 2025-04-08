package response

import "github.com/labstack/echo/v4"

type SuccessResponse[T any] struct {
	Code     int              `json:"code"`
	Data     T                `json:"data"`
	MetaData ResponseMetaData `json:"metadata,omitempty"`
	Success  bool             `json:"success"`
}

type ResponseMetaData map[string]interface{}

func JSON[T any](c echo.Context, code int, data T, args ...ResponseMetaData) error {
	var metadata ResponseMetaData
	if len(args) > 0 {
		metadata = args[0]
	}
	return c.JSON(code, SuccessResponse[T]{
		Code:     code,
		Data:     data,
		MetaData: metadata,
		Success:  true,
	})
}
