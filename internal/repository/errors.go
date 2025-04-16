package repository

import (
	"errors"

	appError "github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/validatorutils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TODO: Remove these error codes use appError instead
var (
	// Common Errors
	ErrNotFound        = errors.New("record not found")
	ErrDuplicateKey    = errors.New("duplicate key violation")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrInternal        = errors.New("internal error")

	// Connection Errors
	ErrConnectionFailed = errors.New("connection failed")
	ErrTimeout          = errors.New("operation timed out")

	// Data Integrity Errors
	ErrConflict         = errors.New("data conflict")
	ErrValidationFailed = errors.New("validation failed")
	ErrOptimisticLock   = errors.New("optimistic lock failure")
	ErrDataCorruption   = errors.New("data corruption detected")

	// Access Control Errors
	ErrUnauthorized = errors.New("unauthorized access")
	ErrForbidden    = errors.New("forbidden operation")

	// Other Errors
	ErrOutOfRange        = errors.New("index or value out of range")
	ErrNotImplemented    = errors.New("not implemented")
	ErrTransactionFailed = errors.New("transaction failed")
)

func HandleRepoError(operation string, err error, defaultErr *appError.AppError, log *zap.Logger) *appError.AppError {
	log.Error("repository operation failed", zap.String("operation", operation), zap.Error(err))

	if dupErr := validatorutils.IsPgDuplicateKeyError(err); dupErr != nil {
		return appError.ErrDuplicateKey.Wrap(err).WithMessagef("duplicate entry for %s", dupErr.ConstraintName)
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return appError.ErrNotFound.Wrap(err).WithMessage("resource not found")
	}

	return defaultErr.Wrap(err).WithMessagef("failed to %s", operation)
}
