package cli

import (
	"errors"
	"testing"
	apperrors "time-tracker/internal/errors"
	"time-tracker/internal/validation"
)

func TestErrorHandler_Handle(t *testing.T) {
	eh := NewErrorHandler()

	tests := []struct {
		name      string
		operation string
		err       error
		expected  string
	}{
		{
			name:      "Validation error",
			operation: "create user",
			err:       apperrors.NewValidationError("invalid input", nil),
			expected:  "failed to create user: invalid input",
		},
		{
			name:      "Not found error",
			operation: "get user",
			err:       apperrors.NewNotFoundError("user", "123"),
			expected:  "failed to get user: user not found: 123",
		},
		{
			name:      "Database error",
			operation: "save user",
			err:       apperrors.NewDatabaseError("insert", errors.New("timeout")),
			expected:  "failed to save user: A database error occurred. Please try again.",
		},
		{
			name:      "Regular error",
			operation: "process",
			err:       errors.New("regular error"),
			expected:  "failed to process: regular error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := eh.Handle(tt.operation, tt.err)
			if result.Error() != tt.expected {
				t.Errorf("ErrorHandler.Handle() = %v, want %v", result.Error(), tt.expected)
			}
		})
	}
}

func TestErrorHandler_HandleSimple(t *testing.T) {
	eh := NewErrorHandler()

	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "Validation error",
			err:      apperrors.NewValidationError("invalid input", nil),
			expected: "invalid input",
		},
		{
			name:     "Not found error",
			err:      apperrors.NewNotFoundError("user", "123"),
			expected: "user not found: 123",
		},
		{
			name:     "Database error",
			err:      apperrors.NewDatabaseError("insert", errors.New("timeout")),
			expected: "A database error occurred. Please try again.",
		},
		{
			name:     "Regular error",
			err:      errors.New("regular error"),
			expected: "regular error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := eh.HandleSimple(tt.err)
			if result.Error() != tt.expected {
				t.Errorf("ErrorHandler.HandleSimple() = %v, want %v", result.Error(), tt.expected)
			}
		})
	}
}

func TestErrorHandler_IsValidationError(t *testing.T) {
	eh := NewErrorHandler()

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "AppError validation",
			err:      apperrors.NewValidationError("invalid input", nil),
			expected: true,
		},
		{
			name: "Legacy validation error",
			err: &validation.ValidationError{
				Errors: []validation.FieldError{
					{Field: "test", Message: "invalid"},
				},
			},
			expected: true,
		},
		{
			name:     "Database error",
			err:      apperrors.NewDatabaseError("insert", nil),
			expected: false,
		},
		{
			name:     "Regular error",
			err:      errors.New("regular error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := eh.IsValidationError(tt.err)
			if result != tt.expected {
				t.Errorf("ErrorHandler.IsValidationError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestErrorHandler_IsNotFoundError(t *testing.T) {
	eh := NewErrorHandler()

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Not found error",
			err:      apperrors.NewNotFoundError("user", "123"),
			expected: true,
		},
		{
			name:     "Validation error",
			err:      apperrors.NewValidationError("invalid input", nil),
			expected: false,
		},
		{
			name:     "Regular error",
			err:      errors.New("regular error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := eh.IsNotFoundError(tt.err)
			if result != tt.expected {
				t.Errorf("ErrorHandler.IsNotFoundError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestErrorHandler_IsDatabaseError(t *testing.T) {
	eh := NewErrorHandler()

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Database error",
			err:      apperrors.NewDatabaseError("insert", nil),
			expected: true,
		},
		{
			name:     "Validation error",
			err:      apperrors.NewValidationError("invalid input", nil),
			expected: false,
		},
		{
			name:     "Regular error",
			err:      errors.New("regular error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := eh.IsDatabaseError(tt.err)
			if result != tt.expected {
				t.Errorf("ErrorHandler.IsDatabaseError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestErrorHandler_GetErrorCode(t *testing.T) {
	eh := NewErrorHandler()

	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "App error",
			err:      apperrors.NewValidationError("invalid input", nil),
			expected: "VALIDATION_FAILED",
		},
		{
			name:     "Regular error",
			err:      errors.New("regular error"),
			expected: "UNKNOWN_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := eh.GetErrorCode(tt.err)
			if result != tt.expected {
				t.Errorf("ErrorHandler.GetErrorCode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestErrorHandler_HandleValidationError(t *testing.T) {
	eh := NewErrorHandler()

	// Create a mock validation error
	validationErr := &validation.ValidationError{
		Errors: []validation.FieldError{
			{Field: "test", Message: "test validation error"},
		},
	}
	
	result := eh.Handle("test operation", validationErr)
	expected := "failed to test operation: test validation error"
	
	if result.Error() != expected {
		t.Errorf("ErrorHandler.Handle() with validation error = %v, want %v", result.Error(), expected)
	}
}

func TestErrorHandler_HandleNilError(t *testing.T) {
	eh := NewErrorHandler()

	// Test with nil error should not panic
	result := eh.Handle("test operation", nil)
	if result == nil {
		t.Errorf("ErrorHandler.Handle() with nil error should not return nil")
	}
}