package response

import "github.com/labstack/echo/v4"

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   bool   `json:"error"`
}

func Error(c echo.Context, code int, msg string) error {
	return c.JSON(code, ErrorResponse{
		Code:    code,
		Message: msg,
		Error:   true,
	})
}
