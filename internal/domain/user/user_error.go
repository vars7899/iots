package user

import "errors"

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidUserState   = errors.New("user data violates business rules")

	// Authentication and Authorization Errors
	ErrInvalidCredentials    = errors.New("invalid username or password")
	ErrUnauthorized          = errors.New("unauthorized access")
	ErrForbidden             = errors.New("forbidden operation")
	ErrAccountNotActive      = errors.New("account is not active")
	ErrPasswordMismatch      = errors.New("passwords do not match")
	ErrTokenExpired          = errors.New("token has expired")
	ErrInvalidToken          = errors.New("invalid token")
	ErrPasswordResetRequired = errors.New("password reset required")

	// Validation Errors (beyond basic data types)
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrPhoneNumberInvalid    = errors.New("invalid phone number format")
	ErrPasswordTooWeak       = errors.New("password is too weak")
	ErrInvalidEmailFormat    = errors.New("invalid email format")
	ErrInvalidUsernameFormat = errors.New("invalid username format")
	ErrRoleNotFound          = errors.New("role not found")
	ErrPermissionNotFound    = errors.New("permission not found")
	ErrRoleAlreadyAssigned   = errors.New("role already assigned to user")
	ErrRoleNotAssigned       = errors.New("role not assigned to user")

	// Data Integrity Errors (beyond duplicates)
	ErrDataNotFound            = errors.New("data not found") // More generic not found
	ErrDataConflict            = errors.New("data conflict occurred")
	ErrDataConstraintViolation = errors.New("data constraint violation")
	ErrInvalidOperation        = errors.New("invalid operation on user data")
	ErrSelfActionForbidden     = errors.New("cannot perform this action on yourself") // E.g., demoting own admin status

	// User Management Specific Errors
	ErrUserNotFound       = errors.New("user not found")
	ErrUserDeletionFailed = errors.New("failed to delete user")
	ErrUserUpdateFailed   = errors.New("failed to update user")
	ErrUserCreationFailed = errors.New("failed to create user")
	ErrInvalidRole        = errors.New("invalid role specified")

	// External Service Errors (if your user logic interacts with other services)
	ErrExternalServiceFailure = errors.New("external service call failed")
	ErrCommunicationError     = errors.New("communication error with external service")

	// State Transition Errors (if user state changes have specific rules)
	ErrInvalidStateTransition = errors.New("invalid user state transition")

	// Concurrency Errors
	ErrConcurrentUpdate = errors.New("concurrent update conflict")
)
