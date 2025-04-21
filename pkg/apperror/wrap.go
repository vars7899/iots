package apperror

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
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
			return appErr.WithMessagef("%s: %s", message, appErr.Message)
		}
		return appErr
	}

	// Otherwise, create a new AppError with the default code
	newErr := New(defaultCode)
	if message != "" {
		newErr.Message = fmt.Sprintf("%s: %s", message, newErr.Message)
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
		newAppErr = newAppErr.WithMessagef("%s: %s", contextMessage, newAppErr.Message)
	}
	return newAppErr.Wrap(err)
}

// HandleDBError processes database errors with appropriate codes
func MapDBError(err error, entity string) *AppError {
	// fmt.Println("errrrrr--->", err)
	// if err == nil {
	// 	return nil
	// }

	// 1. GORM not found
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound.WithMessagef("no %s found matching the criteria", entity).Wrap(err)
	}
	// 2. Postgres error code
	var pqErr *pgconn.PgError
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23505": //unique violation
			return ErrDuplicateKey.WithMessagef("unique constraint violation on %s", entity).Wrap(pqErr)
		case "23503": // Foreign key violation
			return ErrDBForeignKey.WithMessagef("%s references an invalid or non-existent record", entity).Wrap(pqErr)
		case "23502": // not null violation
			return ErrDBInvalidData.WithMessagef("required field missing for %s", entity).Wrap(pqErr)
		case "23514": // check constraint violation
			return ErrDBInvalidData.WithMessagef("validation failed for %s data", entity).Wrap(pqErr)
		case "22001": // string data too long
			return ErrDBInvalidData.WithMessagef("input for %s exceeds allowed length", entity).Wrap(pqErr)
		case "22003": // numeric value out of range
			return ErrDBInvalidData.WithMessagef("numeric value for %s exceeds allowed range", entity).Wrap(pqErr)
		case "22007", "22008": // datetime format or overflow
			return ErrDBInvalidData.WithMessagef("invalid date/time value for %s", entity).Wrap(pqErr)
		case "22P02": // invalid text representation
			return ErrDBInvalidData.WithMessagef("invalid input syntax for %s", entity).Wrap(pqErr)
		case "08000", "08003", "08006", "08001", "08004", "08007", "08P01": // Connection exceptions (class 08)
			return ErrDBConnect.WithMessagef("database connection error while performing %s-related operation", entity).Wrap(pqErr)
		case "53100", "53200", "53300": // Insufficient resources
			return ErrDBResourceLimit.WithMessagef("database resource limit reached while performing %s-related operation", entity).Wrap(pqErr)
		case "55P03", "57014": // Deadlock or query canceled
			return ErrDBDeadlock.WithMessagef("database conflict while performing %s-related operation, please retry", entity).Wrap(pqErr)
		case "40001": // Serialization failure
			return ErrDBConflict.WithMessagef("transaction conflict while performing %s-related operation, please retry", entity).Wrap(pqErr)
		}
	}
	// 3. String matching
	errMsg := err.Error()
	if strings.Contains(errMsg, "duplicate key") {
		return ErrDuplicateKey.WithMessagef("a %s with the same unique fields already exists", entity).Wrap(err)
	}
	if strings.Contains(errMsg, "context deadline exceeded") || strings.Contains(err.Error(), "context canceled") {
		return ErrDBTimeout.WithMessagef("database operation timeout while performing %s-related operation", entity).Wrap(err)
	}
	if strings.Contains(errMsg, "connect: connection refused") || strings.Contains(errMsg, "EOF") {
		return ErrDBConnect.WithMessagef("failed to connect to the database during %s operation", entity).Wrap(err)
	}
	// 4. Default Fallback
	return ErrDBQuery.WithMessagef("unexpected database error while performing %s-related operation", entity).Wrap(err)
}
