package cli

import (
	"fmt"
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
	if validationErr, ok := err.(*validation.ValidationError); ok {
		return fmt.Errorf("failed to %s: %s", operation, validationErr.GetUserFriendlyMessage())
	}
	return fmt.Errorf("failed to %s: %w", operation, err)
}

// HandleSimple provides user-friendly error messages without operation context
func (eh *ErrorHandler) HandleSimple(err error) error {
	if validationErr, ok := err.(*validation.ValidationError); ok {
		return fmt.Errorf("%s", validationErr.GetUserFriendlyMessage())
	}
	return err
}

// IsValidationError checks if an error is a validation error
func (eh *ErrorHandler) IsValidationError(err error) bool {
	return validation.IsValidationError(err)
}