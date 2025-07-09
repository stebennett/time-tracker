package errors

import (
	"errors"
	"testing"
)

func TestErrorType_String(t *testing.T) {
	tests := []struct {
		name     string
		errorType ErrorType
		expected string
	}{
		{"Validation", ErrorTypeValidation, "validation"},
		{"NotFound", ErrorTypeNotFound, "not_found"},
		{"Database", ErrorTypeDatabase, "database"},
		{"InvalidInput", ErrorTypeInvalidInput, "invalid_input"},
		{"Timeout", ErrorTypeTimeout, "timeout"},
		{"Permission", ErrorTypePermission, "permission"},
		{"Unknown", ErrorType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.errorType.String()
			if result != tt.expected {
				t.Errorf("ErrorType.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		appError *AppError
		expected string
	}{
		{
			name: "Error without cause",
			appError: &AppError{
				Type:    ErrorTypeValidation,
				Message: "invalid input",
			},
			expected: "validation: invalid input",
		},
		{
			name: "Error with cause",
			appError: &AppError{
				Type:    ErrorTypeDatabase,
				Message: "connection failed",
				Cause:   errors.New("timeout"),
			},
			expected: "database: connection failed (caused by: timeout)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.appError.Error()
			if result != tt.expected {
				t.Errorf("AppError.Error() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	cause := errors.New("original error")
	appError := &AppError{
		Type:    ErrorTypeDatabase,
		Message: "wrapped error",
		Cause:   cause,
	}

	if appError.Unwrap() != cause {
		t.Errorf("AppError.Unwrap() = %v, want %v", appError.Unwrap(), cause)
	}
}

func TestAppError_Is(t *testing.T) {
	appError1 := &AppError{
		Type: ErrorTypeValidation,
		Code: "VALIDATION_FAILED",
	}
	appError2 := &AppError{
		Type: ErrorTypeValidation,
		Code: "VALIDATION_FAILED",
	}
	appError3 := &AppError{
		Type: ErrorTypeDatabase,
		Code: "DATABASE_ERROR",
	}
	regularError := errors.New("regular error")

	tests := []struct {
		name     string
		err      *AppError
		target   error
		expected bool
	}{
		{"Same type and code", appError1, appError2, true},
		{"Different type", appError1, appError3, false},
		{"Regular error", appError1, regularError, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Is(tt.target)
			if result != tt.expected {
				t.Errorf("AppError.Is() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAppError_IsType(t *testing.T) {
	appError := &AppError{
		Type:    ErrorTypeValidation,
		Message: "test error",
	}

	if !appError.IsType(ErrorTypeValidation) {
		t.Errorf("AppError.IsType() = false, want true for matching type")
	}

	if appError.IsType(ErrorTypeDatabase) {
		t.Errorf("AppError.IsType() = true, want false for different type")
	}
}

func TestAppError_WithContext(t *testing.T) {
	appError := &AppError{
		Type:    ErrorTypeValidation,
		Message: "test error",
	}

	result := appError.WithContext("field", "username")
	
	if result != appError {
		t.Errorf("WithContext should return the same instance")
	}

	if appError.Context == nil {
		t.Errorf("Context should be initialized")
	}

	if appError.Context["field"] != "username" {
		t.Errorf("Context should contain the added key-value pair")
	}
}

func TestAppError_GetContext(t *testing.T) {
	appError := &AppError{
		Type:    ErrorTypeValidation,
		Message: "test error",
		Context: map[string]interface{}{
			"field": "username",
		},
	}

	value, exists := appError.GetContext("field")
	if !exists {
		t.Errorf("GetContext should return true for existing key")
	}
	if value != "username" {
		t.Errorf("GetContext should return correct value")
	}

	_, exists = appError.GetContext("nonexistent")
	if exists {
		t.Errorf("GetContext should return false for non-existing key")
	}

	// Test with nil context
	appError.Context = nil
	_, exists = appError.GetContext("field")
	if exists {
		t.Errorf("GetContext should return false when context is nil")
	}
}