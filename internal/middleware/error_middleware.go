// pkg/middleware/error_handler.go
package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/pkg/response"
	"go.uber.org/zap"
)

// ErrorHandler returns middleware that handles errors
func ErrorHandler(logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Set logger in context
			c.Set("logger", logger)

			// Generate trace ID if not present
			traceID := c.Request().Header.Get("X-Trace-ID")
			if traceID == "" {
				traceID = uuid.New().String()
				c.Request().Header.Set("X-Trace-ID", traceID)
			}
			c.Set("trace_id", traceID)

			// Set response headers
			c.Response().Header().Set("X-Trace-ID", traceID)

			// Process request
			err := next(c)
			if err != nil {
				return response.Error(c, err)
			}
			return nil
		}
	}
}

// Recovery returns middleware that recovers from panics
func Recovery(logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = response.Errorf(response.ErrCodeInternal, "panic: %v", r)
					}

					// Log the panic
					stackTrace := response.New(response.ErrCodeInternal).Stack
					logger.Error("Panic recovered",
						zap.Error(err),
						zap.String("stack", stackTrace),
						zap.String("path", c.Request().URL.Path),
						zap.String("method", c.Request().Method),
					)

					response.Error(c, response.ErrInternal.Wrap(err))
				}
			}()
			return next(c)
		}
	}
}
