package errors

import (
	"errors"
	"testing"
)

func TestNewValidationError(t *testing.T) {
	cause := errors.New("field is required")
	err := NewValidationError("validation failed", cause)

	if err.Type != ErrorTypeValidation {
		t.Errorf("NewValidationError type = %v, want %v", err.Type, ErrorTypeValidation)
	}
	if err.Message != "validation failed" {
		t.Errorf("NewValidationError message = %v, want %v", err.Message, "validation failed")
	}
	if err.Code != "VALIDATION_FAILED" {
		t.Errorf("NewValidationError code = %v, want %v", err.Code, "VALIDATION_FAILED")
	}
	if err.Cause != cause {
		t.Errorf("NewValidationError cause = %v, want %v", err.Cause, cause)
	}
}

func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError("user", "123")

	if err.Type != ErrorTypeNotFound {
		t.Errorf("NewNotFoundError type = %v, want %v", err.Type, ErrorTypeNotFound)
	}
	if err.Message != "user not found: 123" {
		t.Errorf("NewNotFoundError message = %v, want %v", err.Message, "user not found: 123")
	}
	if err.Code != "NOT_FOUND" {
		t.Errorf("NewNotFoundError code = %v, want %v", err.Code, "NOT_FOUND")
	}
	
	resource, ok := err.GetContext("resource")
	if !ok || resource != "user" {
		t.Errorf("NewNotFoundError should set resource context")
	}
	
	identifier, ok := err.GetContext("identifier")
	if !ok || identifier != "123" {
		t.Errorf("NewNotFoundError should set identifier context")
	}
}

func TestNewDatabaseError(t *testing.T) {
	cause := errors.New("connection timeout")
	err := NewDatabaseError("create user", cause)

	if err.Type != ErrorTypeDatabase {
		t.Errorf("NewDatabaseError type = %v, want %v", err.Type, ErrorTypeDatabase)
	}
	if err.Message != "database operation failed: create user" {
		t.Errorf("NewDatabaseError message = %v, want %v", err.Message, "database operation failed: create user")
	}
	if err.Code != "DATABASE_ERROR" {
		t.Errorf("NewDatabaseError code = %v, want %v", err.Code, "DATABASE_ERROR")
	}
	if err.Cause != cause {
		t.Errorf("NewDatabaseError cause = %v, want %v", err.Cause, cause)
	}
	
	operation, ok := err.GetContext("operation")
	if !ok || operation != "create user" {
		t.Errorf("NewDatabaseError should set operation context")
	}
}

func TestNewInvalidInputError(t *testing.T) {
	err := NewInvalidInputError("email", "invalid@", "missing domain")

	if err.Type != ErrorTypeInvalidInput {
		t.Errorf("NewInvalidInputError type = %v, want %v", err.Type, ErrorTypeInvalidInput)
	}
	if err.Message != "invalid input for email: missing domain" {
		t.Errorf("NewInvalidInputError message = %v, want %v", err.Message, "invalid input for email: missing domain")
	}
	if err.Code != "INVALID_INPUT" {
		t.Errorf("NewInvalidInputError code = %v, want %v", err.Code, "INVALID_INPUT")
	}
	
	field, ok := err.GetContext("field")
	if !ok || field != "email" {
		t.Errorf("NewInvalidInputError should set field context")
	}
	
	value, ok := err.GetContext("value")
	if !ok || value != "invalid@" {
		t.Errorf("NewInvalidInputError should set value context")
	}
	
	reason, ok := err.GetContext("reason")
	if !ok || reason != "missing domain" {
		t.Errorf("NewInvalidInputError should set reason context")
	}
}

func TestNewTimeoutError(t *testing.T) {
	err := NewTimeoutError("database query", "5s")

	if err.Type != ErrorTypeTimeout {
		t.Errorf("NewTimeoutError type = %v, want %v", err.Type, ErrorTypeTimeout)
	}
	if err.Message != "operation timed out: database query" {
		t.Errorf("NewTimeoutError message = %v, want %v", err.Message, "operation timed out: database query")
	}
	if err.Code != "TIMEOUT" {
		t.Errorf("NewTimeoutError code = %v, want %v", err.Code, "TIMEOUT")
	}
	
	operation, ok := err.GetContext("operation")
	if !ok || operation != "database query" {
		t.Errorf("NewTimeoutError should set operation context")
	}
	
	timeout, ok := err.GetContext("timeout")
	if !ok || timeout != "5s" {
		t.Errorf("NewTimeoutError should set timeout context")
	}
}

func TestNewPermissionError(t *testing.T) {
	err := NewPermissionError("delete", "user")

	if err.Type != ErrorTypePermission {
		t.Errorf("NewPermissionError type = %v, want %v", err.Type, ErrorTypePermission)
	}
	if err.Message != "permission denied for delete on user" {
		t.Errorf("NewPermissionError message = %v, want %v", err.Message, "permission denied for delete on user")
	}
	if err.Code != "PERMISSION_DENIED" {
		t.Errorf("NewPermissionError code = %v, want %v", err.Code, "PERMISSION_DENIED")
	}
	
	operation, ok := err.GetContext("operation")
	if !ok || operation != "delete" {
		t.Errorf("NewPermissionError should set operation context")
	}
	
	resource, ok := err.GetContext("resource")
	if !ok || resource != "user" {
		t.Errorf("NewPermissionError should set resource context")
	}
}

func TestWrapError(t *testing.T) {
	cause := errors.New("original error")
	err := WrapError(cause, ErrorTypeDatabase, "wrapped message")

	if err.Type != ErrorTypeDatabase {
		t.Errorf("WrapError type = %v, want %v", err.Type, ErrorTypeDatabase)
	}
	if err.Message != "wrapped message" {
		t.Errorf("WrapError message = %v, want %v", err.Message, "wrapped message")
	}
	if err.Code != "database" {
		t.Errorf("WrapError code = %v, want %v", err.Code, "database")
	}
	if err.Cause != cause {
		t.Errorf("WrapError cause = %v, want %v", err.Cause, cause)
	}
}

func TestIsAppError(t *testing.T) {
	appError := &AppError{Type: ErrorTypeValidation}
	regularError := errors.New("regular error")

	if !IsAppError(appError) {
		t.Errorf("IsAppError should return true for AppError")
	}

	if IsAppError(regularError) {
		t.Errorf("IsAppError should return false for regular error")
	}

	if IsAppError(nil) {
		t.Errorf("IsAppError should return false for nil")
	}
}

func TestAsAppError(t *testing.T) {
	appError := &AppError{Type: ErrorTypeValidation}
	regularError := errors.New("regular error")

	result, ok := AsAppError(appError)
	if !ok {
		t.Errorf("AsAppError should return true for AppError")
	}
	if result != appError {
		t.Errorf("AsAppError should return the same AppError instance")
	}

	result, ok = AsAppError(regularError)
	if ok {
		t.Errorf("AsAppError should return false for regular error")
	}
	if result != nil {
		t.Errorf("AsAppError should return nil for regular error")
	}
}

func TestIsErrorType(t *testing.T) {
	appError := &AppError{Type: ErrorTypeValidation}
	regularError := errors.New("regular error")

	if !IsErrorType(appError, ErrorTypeValidation) {
		t.Errorf("IsErrorType should return true for matching type")
	}

	if IsErrorType(appError, ErrorTypeDatabase) {
		t.Errorf("IsErrorType should return false for different type")
	}

	if IsErrorType(regularError, ErrorTypeValidation) {
		t.Errorf("IsErrorType should return false for regular error")
	}
}

func TestGetUserMessage(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "Validation error",
			err:      NewValidationError("invalid input", nil),
			expected: "invalid input",
		},
		{
			name:     "Not found error",
			err:      NewNotFoundError("user", "123"),
			expected: "user not found: 123",
		},
		{
			name:     "Database error",
			err:      NewDatabaseError("query", errors.New("timeout")),
			expected: "A database error occurred. Please try again.",
		},
		{
			name:     "Timeout error",
			err:      NewTimeoutError("query", "5s"),
			expected: "The operation timed out. Please try again.",
		},
		{
			name:     "Permission error",
			err:      NewPermissionError("delete", "user"),
			expected: "permission denied for delete on user",
		},
		{
			name:     "Regular error",
			err:      errors.New("regular error"),
			expected: "regular error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetUserMessage(tt.err)
			if result != tt.expected {
				t.Errorf("GetUserMessage() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetErrorCode(t *testing.T) {
	appError := &AppError{Code: "VALIDATION_FAILED"}
	regularError := errors.New("regular error")

	if GetErrorCode(appError) != "VALIDATION_FAILED" {
		t.Errorf("GetErrorCode should return correct code for AppError")
	}

	if GetErrorCode(regularError) != "UNKNOWN_ERROR" {
		t.Errorf("GetErrorCode should return UNKNOWN_ERROR for regular error")
	}
}

func TestShouldLogError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Validation error",
			err:      NewValidationError("invalid input", nil),
			expected: false,
		},
		{
			name:     "Not found error",
			err:      NewNotFoundError("user", "123"),
			expected: false,
		},
		{
			name:     "Invalid input error",
			err:      NewInvalidInputError("email", "invalid", "format"),
			expected: false,
		},
		{
			name:     "Database error",
			err:      NewDatabaseError("query", errors.New("timeout")),
			expected: true,
		},
		{
			name:     "Timeout error",
			err:      NewTimeoutError("query", "5s"),
			expected: true,
		},
		{
			name:     "Permission error",
			err:      NewPermissionError("delete", "user"),
			expected: true,
		},
		{
			name:     "Regular error",
			err:      errors.New("regular error"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldLogError(tt.err)
			if result != tt.expected {
				t.Errorf("ShouldLogError() = %v, want %v", result, tt.expected)
			}
		})
	}
}