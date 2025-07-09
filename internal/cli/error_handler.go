package cli

import (
	"fmt"
	"time-tracker/internal/errors"
	"time-tracker/internal/validation"
)

// ErrorHandler provides centralized error handling for command handlers
type ErrorHandler struct{}

// NewErrorHandler creates a new error handler
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{}
}

// Handle provides user-friendly error messages for validation and other errors
func (eh *ErrorHandler) Handle(operation string, err error) error {
	// Handle validation errors first (legacy support)
	if validationErr, ok := err.(*validation.ValidationError); ok {
		return fmt.Errorf("failed to %s: %s", operation, validationErr.GetUserFriendlyMessage())
	}
	
	// Handle new AppError types
	if _, ok := errors.AsAppError(err); ok {
		// Return user-friendly message
		userMessage := errors.GetUserMessage(err)
		return fmt.Errorf("failed to %s: %s", operation, userMessage)
	}
	
	// Fallback for unknown errors
	return fmt.Errorf("failed to %s: %w", operation, err)
}

// HandleSimple provides user-friendly error messages without operation context
func (eh *ErrorHandler) HandleSimple(err error) error {
	// Handle validation errors first (legacy support)
	if validationErr, ok := err.(*validation.ValidationError); ok {
		return fmt.Errorf("%s", validationErr.GetUserFriendlyMessage())
	}
	
	// Handle new AppError types
	if _, ok := errors.AsAppError(err); ok {
		// Return user-friendly message
		userMessage := errors.GetUserMessage(err)
		return fmt.Errorf("%s", userMessage)
	}
	
	// Fallback for unknown errors
	return err
}

// IsValidationError checks if an error is a validation error
func (eh *ErrorHandler) IsValidationError(err error) bool {
	if validation.IsValidationError(err) {
		return true
	}
	return errors.IsErrorType(err, errors.ErrorTypeValidation)
}

// IsNotFoundError checks if an error is a not found error
func (eh *ErrorHandler) IsNotFoundError(err error) bool {
	return errors.IsErrorType(err, errors.ErrorTypeNotFound)
}

// IsDatabaseError checks if an error is a database error
func (eh *ErrorHandler) IsDatabaseError(err error) bool {
	return errors.IsErrorType(err, errors.ErrorTypeDatabase)
}

// GetErrorCode returns the error code for structured errors
func (eh *ErrorHandler) GetErrorCode(err error) string {
	return errors.GetErrorCode(err)
}