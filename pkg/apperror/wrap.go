package apperror

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

// ErrorHandler provides a universal way to handle and wrap errors
func ErrorHandler(err error, defaultCode ErrorCode, args ...string) *AppError {
	// If nil error, return nil
	if err == nil {
		return nil
	}

	message := ""
	if len(args) > 0 && args[0] != "" {
		message = args[0]
	}

	// If already an AppError, just update the message if provided
	var appErr *AppError
	if errors.As(err, &appErr) {
		if message != "" {
			return appErr.WithMessage(message)
		}
		return appErr
	}

	// Otherwise, create a new AppError with the default code
	newErr := New(defaultCode)
	if message != "" {
		newErr.Message = message
	}

	// Wrap the original error
	return newErr.Wrap(err)
}

func WrapAppErrWithContext(err error, contextMessage string, fallbackCode ErrorCode) *AppError {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		if contextMessage != "" {
			return appErr.WithMessagef("%s: %s", contextMessage, appErr.Message)
		}
		return appErr
	}

	newAppErr := New(fallbackCode)
	if contextMessage != "" {
		newAppErr = newAppErr.WithMessage(contextMessage)
	}
	return newAppErr.Wrap(err)
}

// HandleDBError processes database errors with appropriate codes
func HandleDBError(err error) *AppError {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return New(ErrCodeNotFound).WithMessagef("resource not found").Wrap(err)
	}
	if strings.Contains(err.Error(), "duplicate key") {
		return New(ErrCodeDuplicateKey).WithMessagef("duplicate record detected").Wrap(err)
	}
	return New(ErrCodeDBQuery).WithMessagef("Database error").Wrap(err)
}
