package apperror

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
)

type ErrorCode string

const (
	StatusBadRequest          = http.StatusBadRequest
	StatusUnauthorized        = http.StatusUnauthorized
	StatusForbidden           = http.StatusForbidden
	StatusNotFound            = http.StatusNotFound
	StatusConflict            = http.StatusConflict
	StatusUnprocessableEntity = http.StatusUnprocessableEntity
	StatusInternalServerError = http.StatusInternalServerError
	StatusServiceUnavailable  = http.StatusServiceUnavailable
)

const (
	// General errors (1xxx)
	ErrCodeInternal   ErrorCode = "ERR-1000"
	ErrCodeBadRequest ErrorCode = "ERR-1001"
	ErrCodeNotFound   ErrorCode = "ERR-1002"
	ErrCodeConflict   ErrorCode = "ERR-1003"
	ErrCodeForbidden  ErrorCode = "ERR-1004"

	// Auth errors (2xxx)
	ErrCodeUnauthorized   ErrorCode = "ERR-2000"
	ErrCodeInvalidToken   ErrorCode = "ERR-2001"
	ErrCodeExpiredToken   ErrorCode = "ERR-2002"
	ErrCodeMalformedToken ErrorCode = "ERR-2003"
	ErrCodeInvalidUUID    ErrorCode = "ERR-2004"
	ErrCodeRefreshFailed  ErrorCode = "ERR-2005"
	ErrCodeMissingAuth    ErrorCode = "ERR-2006"

	// Validation errors (3xxx)
	ErrCodeValidation ErrorCode = "ERR-3000"

	// User errors (4xxx)
	ErrCodeUserExists   ErrorCode = "ERR-4000"
	ErrCodeUserNotFound ErrorCode = "ERR-4001"
	ErrCodeInvalidCreds ErrorCode = "ERR-4002"
	ErrCodeUpdateLogin  ErrorCode = "ERR-4003"

	// Database errors (5xxx)
	ErrCodeDBQuery         ErrorCode = "ERR-5000"
	ErrCodeDBInsert        ErrorCode = "ERR-5001"
	ErrCodeDBUpdate        ErrorCode = "ERR-5002"
	ErrCodeDBDelete        ErrorCode = "ERR-5003"
	ErrCodeDBConnect       ErrorCode = "ERR-5004"
	ErrCodeDuplicateKey    ErrorCode = "ERR-5005"
	ErrCodeDBForeignKey    ErrorCode = "ERR-5006"
	ErrCodeInvalidData     ErrorCode = "ERR-5007"
	ErrCodeDBResourceLimit ErrorCode = "ERR-5008"
	ErrCodeDBDeadlock      ErrorCode = "ERR-5009"
	ErrCodeDBConflict      ErrorCode = "ERR-5010"
	ErrCodeDBTimeout       ErrorCode = "ERR-5011"
	ErrCodeDBMissing       ErrorCode = "ERR-5012" // Database not found or uninitialized
	ErrCodeDBMigration     ErrorCode = "ERR-5013" // Migration failed or pending
	ErrCodeDBTxnFailed     ErrorCode = "ERR-5014" // Transaction commit/rollback failed
	ErrCodeDBPing          ErrorCode = "ERR-5015"

	// Timeout errors (6xxx)
	ErrCodeTimeout ErrorCode = "ERR-6000"

	// Context errors (7xxx)
	ErrCodeContextCancelled ErrorCode = "ERR-7000"
	ErrCodeContextTimeout   ErrorCode = "ERR-7001"

	// Configuration & environment errors (8xxx)
	ErrCodeMissingEnv    ErrorCode = "ERR-8000" // Required env variable not set
	ErrCodeInvalidEnv    ErrorCode = "ERR-8001" // Env value is invalid
	ErrCodeConfigLoad    ErrorCode = "ERR-8002" // Failed to load config file
	ErrCodeConfigParse   ErrorCode = "ERR-8003" // Invalid config structure
	ErrCodeMissingSecret ErrorCode = "ERR-8004" // Secret key/token missing
	ErrCodeInvalidSecret ErrorCode = "ERR-8005" // Secret value malformed or unauthorized
	ErrCodeMissingConfig ErrorCode = "ERR-8006" // Required config key missing
	ErrCodeInvalidConfig ErrorCode = "ERR-8007" // Config value invalid

	// Dependency & initialization errors (9xxx)
	ErrCodeMissingDependency ErrorCode = "ERR-9000"
)

// CodeMessages maps error codes to default messages (can be overridden in i18n files)
var CodeMessages = map[ErrorCode]string{
	// General errors
	ErrCodeInternal:   "Internal server error",
	ErrCodeBadRequest: "Bad request",
	ErrCodeNotFound:   "Resource not found",
	ErrCodeConflict:   "Resource conflict",
	ErrCodeForbidden:  "Forbidden",

	// Auth errors
	ErrCodeUnauthorized:   "Unauthorized access",
	ErrCodeInvalidToken:   "Invalid or expired access token",
	ErrCodeExpiredToken:   "Token has expired",
	ErrCodeMalformedToken: "Malformed token",
	ErrCodeInvalidUUID:    "Invalid user ID format in token",
	ErrCodeRefreshFailed:  "Could not refresh token",
	ErrCodeMissingAuth:    "Authorization header missing",

	// Validation errors
	ErrCodeValidation: "Validation failed",

	// User errors
	ErrCodeUserExists:   "User already exists",
	ErrCodeUserNotFound: "User not found",
	ErrCodeInvalidCreds: "Invalid credentials",
	ErrCodeUpdateLogin:  "Could not update user session",

	// Database errors
	ErrCodeDBQuery:         "Database query failed",
	ErrCodeDBInsert:        "Failed to insert into database",
	ErrCodeDBUpdate:        "Failed to update database",
	ErrCodeDBDelete:        "Failed to delete from database",
	ErrCodeDBConnect:       "Failed to connect to database",
	ErrCodeDuplicateKey:    "Duplicate key value violates unique constraint",
	ErrCodeDBForeignKey:    "Invalid Foreign key reference",
	ErrCodeInvalidData:     "Invalid or corrupt data",
	ErrCodeDBResourceLimit: "database resource limit reached",
	ErrCodeDBDeadlock:      "database deadlock conflict",
	ErrCodeDBConflict:      "database serialization conflict",
	ErrCodeDBTimeout:       "database timeout",
	ErrCodeDBMissing:       "database not found or uninitialized",
	ErrCodeDBMigration:     "migration failed or pending",
	ErrCodeDBTxnFailed:     "transaction commit/rollback failed",
	ErrCodeDBPing:          "failed to ping database",

	// Timeout errors
	ErrCodeTimeout: "Request deadline exceeded",

	// Context errors
	ErrCodeContextCancelled: "operation ended due to context cancellation",
	ErrCodeContextTimeout:   "operation ended due to context timeout",

	// Configuration & environment errors
	ErrCodeMissingEnv:    "Required environment variable is not set",
	ErrCodeInvalidEnv:    "Invalid value in environment variable",
	ErrCodeConfigLoad:    "Failed to load configuration file",
	ErrCodeConfigParse:   "Configuration parsing failed",
	ErrCodeMissingSecret: "Secret key or token is missing",
	ErrCodeInvalidSecret: "Secret is malformed or unauthorized",
	ErrCodeMissingConfig: "Required configuration key is missing",
	ErrCodeInvalidConfig: "Invalid configuration value",

	// Dependency & initialization errors
	ErrCodeMissingDependency: "Required dependency is missing",
}

// HTTP status mapping for error codes
var CodeStatus = map[ErrorCode]int{
	// General errors
	ErrCodeInternal:   StatusInternalServerError,
	ErrCodeBadRequest: StatusBadRequest,
	ErrCodeNotFound:   StatusNotFound,
	ErrCodeConflict:   StatusConflict,
	ErrCodeForbidden:  StatusForbidden,

	// Auth errors
	ErrCodeUnauthorized:   StatusUnauthorized,
	ErrCodeInvalidToken:   StatusUnauthorized,
	ErrCodeExpiredToken:   StatusUnauthorized,
	ErrCodeMalformedToken: StatusBadRequest,
	ErrCodeInvalidUUID:    StatusBadRequest,
	ErrCodeRefreshFailed:  StatusUnauthorized,
	ErrCodeMissingAuth:    StatusUnauthorized,

	// Validation errors
	ErrCodeValidation: StatusUnprocessableEntity,

	// User errors
	ErrCodeUserExists:   StatusConflict,
	ErrCodeUserNotFound: StatusNotFound,
	ErrCodeInvalidCreds: StatusUnauthorized,
	ErrCodeUpdateLogin:  StatusServiceUnavailable,

	// Database errors
	ErrCodeDBQuery:         StatusInternalServerError,
	ErrCodeDBInsert:        StatusInternalServerError,
	ErrCodeDBUpdate:        StatusInternalServerError,
	ErrCodeDBDelete:        StatusInternalServerError,
	ErrCodeDBConnect:       StatusServiceUnavailable,
	ErrCodeDuplicateKey:    StatusConflict,
	ErrCodeDBForeignKey:    StatusInternalServerError,
	ErrCodeInvalidData:     StatusBadRequest,
	ErrCodeDBResourceLimit: StatusInternalServerError,
	ErrCodeDBDeadlock:      StatusInternalServerError,
	ErrCodeDBConflict:      StatusInternalServerError,
	ErrCodeDBTimeout:       StatusInternalServerError,
	ErrCodeDBMissing:       StatusInternalServerError,
	ErrCodeDBMigration:     StatusInternalServerError,
	ErrCodeDBTxnFailed:     StatusInternalServerError,
	ErrCodeDBPing:          StatusInternalServerError,

	// Context errors
	ErrCodeContextTimeout: StatusInternalServerError,

	// Configuration & environment errors
	ErrCodeMissingEnv:    StatusInternalServerError,
	ErrCodeInvalidEnv:    StatusInternalServerError,
	ErrCodeConfigLoad:    StatusInternalServerError,
	ErrCodeConfigParse:   StatusInternalServerError,
	ErrCodeMissingSecret: StatusInternalServerError,
	ErrCodeInvalidSecret: StatusUnauthorized,
	ErrCodeMissingConfig: StatusInternalServerError,
	ErrCodeInvalidConfig: StatusInternalServerError,

	// Dependency & initialization errors
	ErrCodeMissingDependency: StatusInternalServerError,
}

// AppError represents a standardized error for the application
type AppError struct {
	Code         ErrorCode       `json:"code"`
	Message      string          `json:"message"`
	Details      json.RawMessage `json:"details,omitempty"`
	Timestamp    time.Time       `json:"timestamp"`
	Path         string          `json:"path,omitempty"`
	TraceID      string          `json:"-"`
	Stack        string          `json:"-"`
	status       int             `json:"-"`
	originalErr  error           `json:"-"`
	internalOnly bool            `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.originalErr != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.originalErr.Error())
	}
	return e.Message
}

func (e *AppError) OriginalErr() error {
	return e.originalErr
}
func (e *AppError) InternalOnly() bool {
	return e.internalOnly
}

// Unwrap implements error unwrapping
func (e *AppError) Unwrap() error {
	return e.originalErr
}

// Status returns the HTTP status code
func (e *AppError) Status() int {
	return e.status
}

// WithDetails adds details to the error and returns a copy
func (e *AppError) WithDetails(details interface{}) *AppError {
	clone := e.clone()

	// Convert details to JSON
	if details != nil {
		data, err := json.Marshal(details)
		if err == nil {
			clone.Details = data
		}
	}

	return clone
}

// WithMessage overrides the default message
func (e *AppError) WithMessage(message string) *AppError {
	clone := e.clone()
	clone.Message = message
	return clone
}

func (e *AppError) WithMessagef(format string, args ...interface{}) *AppError {
	clone := e.clone()
	clone.Message = fmt.Sprintf(format, args...)
	return clone
}

// WithPath records the request path
func (e *AppError) WithPath(path string) *AppError {
	clone := e.clone()
	clone.Path = path
	return clone
}

// WithTraceID adds a trace ID for distributed tracing
func (e *AppError) WithTraceID(traceID string) *AppError {
	clone := e.clone()
	clone.TraceID = traceID
	return clone
}

// Wrap adds an underlying error while preserving the original error info
func (e *AppError) Wrap(err error) *AppError {
	if err == nil {
		return e
	}

	clone := e.clone()

	// If we're wrapping another AppError, merge some fields
	var appErr *AppError
	if errors.As(err, &appErr) {
		if clone.Details == nil && appErr.Details != nil {
			clone.Details = appErr.Details
		}
		if clone.TraceID == "" && appErr.TraceID != "" {
			clone.TraceID = appErr.TraceID
		}

		// Keep the most specific error as the cause
		if appErr.originalErr != nil {
			clone.originalErr = appErr.originalErr
		} else {
			clone.originalErr = err
		}
	} else {
		clone.originalErr = err
	}

	// Capture stack trace if not already present
	if clone.Stack == "" {
		clone.Stack = stackTrace(2)
	}

	return clone
}

// AsInternal marks an error as internal only (not exposed to client)
func (e *AppError) AsInternal() *AppError {
	clone := e.clone()
	clone.internalOnly = true
	return clone
}

// clone creates a shallow copy of the error
func (e *AppError) clone() *AppError {
	return &AppError{
		Code:         e.Code,
		Message:      e.Message,
		Details:      e.Details,
		Timestamp:    e.Timestamp,
		Path:         e.Path,
		TraceID:      e.TraceID,
		Stack:        e.Stack,
		status:       e.status,
		originalErr:  e.originalErr,
		internalOnly: e.internalOnly,
	}
}

// New creates a new AppError with the given code
func New(code ErrorCode) *AppError {
	status, exists := CodeStatus[code]
	if !exists {
		status = StatusInternalServerError
	}

	message, exists := CodeMessages[code]
	if !exists {
		message = "An error occurred"
	}

	return &AppError{
		Code:      code,
		Message:   message,
		Timestamp: time.Now().UTC(),
		status:    status,
		Stack:     stackTrace(2),
	}
}

// Errorf creates a new AppError with a formatted message
func Errorf(code ErrorCode, format string, args ...interface{}) *AppError {
	err := New(code)
	err.Message = fmt.Sprintf(format, args...)
	return err
}

// FromError converts a standard error to an AppError
func FromError(err error) *AppError {
	if err == nil {
		return nil
	}

	// If it's already an AppError, return it
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	// Otherwise wrap as internal error
	return New(ErrCodeInternal).Wrap(err)
}

// stackTrace captures the current stack trace
func stackTrace(skip int) string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(skip+1, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var builder strings.Builder
	for {
		frame, more := frames.Next()
		if !strings.Contains(frame.File, "runtime/") {
			fmt.Fprintf(&builder, "%s:%d %s\n", frame.File, frame.Line, frame.Function)
		}
		if !more {
			break
		}
	}
	return builder.String()
}

// Common errors
var (
	// General errors
	ErrInternal   = New(ErrCodeInternal)
	ErrBadRequest = New(ErrCodeBadRequest)
	ErrNotFound   = New(ErrCodeNotFound)
	ErrConflict   = New(ErrCodeConflict)
	ErrForbidden  = New(ErrCodeForbidden)

	// Auth errors
	ErrUnauthorized   = New(ErrCodeUnauthorized)
	ErrInvalidToken   = New(ErrCodeInvalidToken)
	ErrExpiredToken   = New(ErrCodeExpiredToken)
	ErrMalformedToken = New(ErrCodeMalformedToken)
	ErrInvalidUUID    = New(ErrCodeInvalidUUID)
	ErrRefreshFailed  = New(ErrCodeRefreshFailed)
	ErrMissingAuth    = New(ErrCodeMissingAuth)

	// Validation errors
	ErrValidation = New(ErrCodeValidation)

	// User errors
	ErrUserExists   = New(ErrCodeUserExists)
	ErrUserNotFound = New(ErrCodeUserNotFound)
	ErrInvalidCreds = New(ErrCodeInvalidCreds)
	ErrUpdateLogin  = New(ErrCodeUpdateLogin)

	// DB errors
	ErrDBQuery         = New(ErrCodeDBQuery)
	ErrDBInsert        = New(ErrCodeDBInsert)
	ErrDBUpdate        = New(ErrCodeDBUpdate)
	ErrDBDelete        = New(ErrCodeDBDelete)
	ErrDBConnect       = New(ErrCodeDBConnect)
	ErrDuplicateKey    = New(ErrCodeDuplicateKey)
	ErrDBForeignKey    = New(ErrCodeDBForeignKey)
	ErrDBInvalidData   = New(ErrCodeInvalidData)
	ErrDBResourceLimit = New(ErrCodeDBResourceLimit)
	ErrDBDeadlock      = New(ErrCodeDBDeadlock)
	ErrDBConflict      = New(ErrCodeDBConflict)
	ErrDBTimeout       = New(ErrCodeDBTimeout)
	ErrDBMissing       = New(ErrCodeDBMissing)
	ErrDBMigration     = New(ErrCodeDBMigration)
	ErrDBTxnFailed     = New(ErrCodeDBTxnFailed)
	ErrDBPing          = New(ErrCodeDBPing)

	// Timeout errors
	ErrTimeout = New(ErrCodeTimeout)

	// Context errors
	ErrContextCancelled = New(ErrCodeContextCancelled)
	ErrContextTimeout   = New(ErrCodeContextTimeout)

	// Configuration & environment errors
	ErrMissingEnv    = New(ErrCodeMissingEnv)
	ErrInvalidEnv    = New(ErrCodeInvalidEnv)
	ErrConfigLoad    = New(ErrCodeConfigLoad)
	ErrConfigParse   = New(ErrCodeConfigParse)
	ErrMissingSecret = New(ErrCodeMissingSecret)
	ErrInvalidSecret = New(ErrCodeInvalidSecret)
	ErrMissingConfig = New(ErrCodeMissingConfig)
	ErrInvalidConfig = New(ErrCodeInvalidConfig)

	// Dependency & initialization errors
	ErrMissingDependency = New(ErrCodeMissingDependency)
)
