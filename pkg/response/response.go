package response

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/pkg/apperror"
	"go.uber.org/zap"
)

// Response is the base structure for all API responses
type Response struct {
	Success   bool      `json:"success"`
	Timestamp time.Time `json:"timestamp"`
	TraceID   string    `json:"trace_id,omitempty"`
}

// SuccessResponse is the structure for successful responses
type SuccessResponse[T any] struct {
	Response
	Code     int                    `json:"status_code"`
	Data     T                      `json:"data"`
	MetaData map[string]interface{} `json:"metadata,omitempty"`
}

// ErrorResponse is the structure for error responses
type ErrorResponse struct {
	Response
	StatusCode int             `json:"status_code"`
	ErrorCode  string          `json:"error_code"`
	Message    string          `json:"message"`
	Details    json.RawMessage `json:"details,omitempty"`
}

// JSON sends a success response
func JSON[T any](c echo.Context, code int, data T, metadata ...map[string]interface{}) error {
	var meta map[string]interface{}
	if len(metadata) > 0 {
		meta = metadata[0]
	}

	// Get trace ID from context
	traceID := getTraceID(c)

	return c.JSON(code, SuccessResponse[T]{
		Response: Response{
			Success:   true,
			Timestamp: time.Now().UTC(),
			TraceID:   traceID,
		},
		Code:     code,
		Data:     data,
		MetaData: meta,
	})
}

// Error sends an error response
func Error(c echo.Context, err error) error {
	logger := getLogger(c)

	// Convert to AppError if needed
	appError := apperror.FromError(err)
	if appError == nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	// Add request path and trace ID if not already set
	if appError.Path == "" {
		appError = appError.WithPath(c.Request().URL.Path)
	}

	traceID := getTraceID(c)
	if appError.TraceID == "" && traceID != "" {
		appError = appError.WithTraceID(traceID)
	}

	msg, ok := apperror.CodeMessages[appError.Code]
	if !ok {
		msg = "oops!!! something went wrong"
	}
	appError = appError.WithMessage(msg)

	// Log the error with details for debugging
	logFields := []zap.Field{
		zap.String("error_code", string(appError.Code)),
		zap.String("path", appError.Path),
		zap.Int("status", appError.Status()),
	}

	if appError.TraceID != "" {
		logFields = append(logFields, zap.String("trace_id", appError.TraceID))
	}

	if appError.OriginalErr() != nil {
		logFields = append(logFields, zap.Error(appError.OriginalErr()))
	}

	if appError.Stack != "" {
		logFields = append(logFields, zap.String("stack", appError.Stack))
	}

	if appError.Status() >= 500 {
		logger.Error("Internal server error", logFields...)
	} else {
		logger.Info("Client error", logFields...)
	}

	// Don't expose internal details to client for security
	if appError.InternalOnly() {
		return c.JSON(appError.Status(), ErrorResponse{
			Response: Response{
				Success:   false,
				Timestamp: appError.Timestamp,
				TraceID:   appError.TraceID,
			},
			StatusCode: appError.Status(),
			ErrorCode:  string(appError.Code),
			Message:    "An error occurred", // Generic message for internal errors
		})
	}

	return c.JSON(appError.Status(), ErrorResponse{
		Response: Response{
			Success:   false,
			Timestamp: appError.Timestamp,
			TraceID:   appError.TraceID,
		},
		StatusCode: appError.Status(),
		ErrorCode:  string(appError.Code),
		Message:    appError.Message,
		Details:    appError.Details,
	})
}

// Helper functions to retrieve context values
func getTraceID(c echo.Context) string {
	if id := c.Request().Header.Get("X-Trace-ID"); id != "" {
		return id
	}
	if id, ok := c.Get("trace_id").(string); ok {
		return id
	}
	return ""
}

func getLogger(c echo.Context) *zap.Logger {
	if logger, ok := c.Get("logger").(*zap.Logger); ok {
		return logger
	}
	// Fallback to global logger
	return zap.L()
}
