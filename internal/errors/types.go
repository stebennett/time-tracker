package errors

import (
	"fmt"
)

// ErrorType represents the category of error
type ErrorType int

const (
	ErrorTypeValidation ErrorType = iota
	ErrorTypeNotFound
	ErrorTypeDatabase
	ErrorTypeInvalidInput
	ErrorTypeTimeout
	ErrorTypePermission
)

// String returns the string representation of the error type
func (et ErrorType) String() string {
	switch et {
	case ErrorTypeValidation:
		return "validation"
	case ErrorTypeNotFound:
		return "not_found"
	case ErrorTypeDatabase:
		return "database"
	case ErrorTypeInvalidInput:
		return "invalid_input"
	case ErrorTypeTimeout:
		return "timeout"
	case ErrorTypePermission:
		return "permission"
	default:
		return "unknown"
	}
}

// AppError represents a structured application error
type AppError struct {
	Type    ErrorType
	Message string
	Code    string
	Cause   error
	Context map[string]interface{}
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type.String(), e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type.String(), e.Message)
}

// Unwrap returns the underlying error for error unwrapping
func (e *AppError) Unwrap() error {
	return e.Cause
}

// Is checks if this error matches the target error type
func (e *AppError) Is(target error) bool {
	if appErr, ok := target.(*AppError); ok {
		return e.Type == appErr.Type && e.Code == appErr.Code
	}
	return false
}

// IsType checks if this error is of the specified type
func (e *AppError) IsType(errorType ErrorType) bool {
	return e.Type == errorType
}

// WithContext adds context information to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// GetContext retrieves context information from the error
func (e *AppError) GetContext(key string) (interface{}, bool) {
	if e.Context == nil {
		return nil, false
	}
	value, exists := e.Context[key]
	return value, exists
}