package repository

import "errors"

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
