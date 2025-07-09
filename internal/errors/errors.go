package errors

import (
	"errors"
	"fmt"
)

// NewValidationError creates a new validation error
func NewValidationError(message string, cause error) *AppError {
	return &AppError{
		Type:    ErrorTypeValidation,
		Message: message,
		Code:    "VALIDATION_FAILED",
		Cause:   cause,
		Context: make(map[string]interface{}),
	}
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(resource string, identifier string) *AppError {
	return &AppError{
		Type:    ErrorTypeNotFound,
		Message: fmt.Sprintf("%s not found: %s", resource, identifier),
		Code:    "NOT_FOUND",
		Context: map[string]interface{}{
			"resource":   resource,
			"identifier": identifier,
		},
	}
}

// NewDatabaseError creates a new database error
func NewDatabaseError(operation string, cause error) *AppError {
	return &AppError{
		Type:    ErrorTypeDatabase,
		Message: fmt.Sprintf("database operation failed: %s", operation),
		Code:    "DATABASE_ERROR",
		Cause:   cause,
		Context: map[string]interface{}{
			"operation": operation,
		},
	}
}

// NewInvalidInputError creates a new invalid input error
func NewInvalidInputError(field string, value interface{}, reason string) *AppError {
	return &AppError{
		Type:    ErrorTypeInvalidInput,
		Message: fmt.Sprintf("invalid input for %s: %s", field, reason),
		Code:    "INVALID_INPUT",
		Context: map[string]interface{}{
			"field":  field,
			"value":  value,
			"reason": reason,
		},
	}
}

// NewTimeoutError creates a new timeout error
func NewTimeoutError(operation string, timeout interface{}) *AppError {
	return &AppError{
		Type:    ErrorTypeTimeout,
		Message: fmt.Sprintf("operation timed out: %s", operation),
		Code:    "TIMEOUT",
		Context: map[string]interface{}{
			"operation": operation,
			"timeout":   timeout,
		},
	}
}

// NewPermissionError creates a new permission error
func NewPermissionError(operation string, resource string) *AppError {
	return &AppError{
		Type:    ErrorTypePermission,
		Message: fmt.Sprintf("permission denied for %s on %s", operation, resource),
		Code:    "PERMISSION_DENIED",
		Context: map[string]interface{}{
			"operation": operation,
			"resource":  resource,
		},
	}
}

// WrapError wraps an existing error with additional context
func WrapError(err error, errorType ErrorType, message string) *AppError {
	return &AppError{
		Type:    errorType,
		Message: message,
		Code:    errorType.String(),
		Cause:   err,
		Context: make(map[string]interface{}),
	}
}

// IsAppError checks if the error is an AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// AsAppError converts an error to an AppError if possible
func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// IsErrorType checks if the error is of the specified type
func IsErrorType(err error, errorType ErrorType) bool {
	if appErr, ok := AsAppError(err); ok {
		return appErr.IsType(errorType)
	}
	return false
}

// GetUserMessage returns a user-friendly error message
func GetUserMessage(err error) string {
	if appErr, ok := AsAppError(err); ok {
		switch appErr.Type {
		case ErrorTypeValidation:
			return appErr.Message
		case ErrorTypeNotFound:
			return appErr.Message
		case ErrorTypeInvalidInput:
			return appErr.Message
		case ErrorTypeDatabase:
			return "A database error occurred. Please try again."
		case ErrorTypeTimeout:
			return "The operation timed out. Please try again."
		case ErrorTypePermission:
			return appErr.Message
		default:
			return "An unexpected error occurred. Please try again."
		}
	}
	return err.Error()
}

// GetErrorCode returns the error code for the error
func GetErrorCode(err error) string {
	if appErr, ok := AsAppError(err); ok {
		return appErr.Code
	}
	return "UNKNOWN_ERROR"
}

// ShouldLogError determines if an error should be logged based on its type
func ShouldLogError(err error) bool {
	if appErr, ok := AsAppError(err); ok {
		switch appErr.Type {
		case ErrorTypeValidation, ErrorTypeNotFound, ErrorTypeInvalidInput:
			return false // These are user errors, not system errors
		case ErrorTypeDatabase, ErrorTypeTimeout, ErrorTypePermission:
			return true // These are system errors that should be logged
		default:
			return true
		}
	}
	return true // Unknown errors should be logged
}