package validation

import (
	"strings"
	"testing"
)

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name        string
		errors      []FieldError
		expectError string
	}{
		{"No errors", []FieldError{}, "validation error"},
		{"Single error", []FieldError{{Field: "name", Message: "is required"}}, "validation error for field 'name': is required"},
		{"Multiple errors", []FieldError{
			{Field: "name", Message: "is required"},
			{Field: "age", Message: "must be positive"},
		}, "multiple validation errors"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ve := &ValidationError{Errors: tt.errors}
			result := ve.Error()
			
			if tt.name == "Multiple errors" {
				if !strings.Contains(result, tt.expectError) {
					t.Errorf("ValidationError.Error() = %v, expected to contain %v", result, tt.expectError)
				}
			} else {
				if result != tt.expectError {
					t.Errorf("ValidationError.Error() = %v, expected %v", result, tt.expectError)
				}
			}
		})
	}
}

func TestValidationError_HasErrors(t *testing.T) {
	tests := []struct {
		name     string
		errors   []FieldError
		expected bool
	}{
		{"No errors", []FieldError{}, false},
		{"Has errors", []FieldError{{Field: "name", Message: "is required"}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ve := &ValidationError{Errors: tt.errors}
			result := ve.HasErrors()
			
			if result != tt.expected {
				t.Errorf("ValidationError.HasErrors() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestValidationError_AddError(t *testing.T) {
	ve := NewValidationError()
	
	ve.AddError("name", ErrorTypeRequired, "is required", "")
	
	if len(ve.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(ve.Errors))
	}
	
	if ve.Errors[0].Field != "name" {
		t.Errorf("Expected field 'name', got %s", ve.Errors[0].Field)
	}
	
	if ve.Errors[0].Type != ErrorTypeRequired {
		t.Errorf("Expected error type %v, got %v", ErrorTypeRequired, ve.Errors[0].Type)
	}
}

func TestValidationError_AddRequiredError(t *testing.T) {
	ve := NewValidationError()
	
	ve.AddRequiredError("name")
	
	if len(ve.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(ve.Errors))
	}
	
	if ve.Errors[0].Type != ErrorTypeRequired {
		t.Errorf("Expected error type %v, got %v", ErrorTypeRequired, ve.Errors[0].Type)
	}
	
	if ve.Errors[0].Field != "name" {
		t.Errorf("Expected field 'name', got %s", ve.Errors[0].Field)
	}
}

func TestValidationError_AddInvalidFormatError(t *testing.T) {
	ve := NewValidationError()
	
	ve.AddInvalidFormatError("date", "2023-13-01", "YYYY-MM-DD")
	
	if len(ve.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(ve.Errors))
	}
	
	if ve.Errors[0].Type != ErrorTypeInvalidFormat {
		t.Errorf("Expected error type %v, got %v", ErrorTypeInvalidFormat, ve.Errors[0].Type)
	}
	
	if !strings.Contains(ve.Errors[0].Message, "YYYY-MM-DD") {
		t.Errorf("Expected message to contain expected format, got %s", ve.Errors[0].Message)
	}
}

func TestValidationError_AddInvalidLengthError(t *testing.T) {
	ve := NewValidationError()
	
	ve.AddInvalidLengthError("name", "a", 2, 50)
	
	if len(ve.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(ve.Errors))
	}
	
	if ve.Errors[0].Type != ErrorTypeInvalidLength {
		t.Errorf("Expected error type %v, got %v", ErrorTypeInvalidLength, ve.Errors[0].Type)
	}
	
	if !strings.Contains(ve.Errors[0].Message, "between 2 and 50") {
		t.Errorf("Expected message to contain length range, got %s", ve.Errors[0].Message)
	}
}

func TestValidationError_AddInvalidValueError(t *testing.T) {
	ve := NewValidationError()
	
	ve.AddInvalidValueError("age", -1, "must be positive")
	
	if len(ve.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(ve.Errors))
	}
	
	if ve.Errors[0].Type != ErrorTypeInvalidValue {
		t.Errorf("Expected error type %v, got %v", ErrorTypeInvalidValue, ve.Errors[0].Type)
	}
	
	if !strings.Contains(ve.Errors[0].Message, "must be positive") {
		t.Errorf("Expected message to contain reason, got %s", ve.Errors[0].Message)
	}
}

func TestValidationError_AddInvalidRangeError(t *testing.T) {
	ve := NewValidationError()
	
	ve.AddInvalidRangeError("date_range", "2023-01-01 to 2022-12-31", "end must be after start")
	
	if len(ve.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(ve.Errors))
	}
	
	if ve.Errors[0].Type != ErrorTypeInvalidRange {
		t.Errorf("Expected error type %v, got %v", ErrorTypeInvalidRange, ve.Errors[0].Type)
	}
	
	if !strings.Contains(ve.Errors[0].Message, "end must be after start") {
		t.Errorf("Expected message to contain reason, got %s", ve.Errors[0].Message)
	}
}

func TestValidationError_AddInvalidCharacterError(t *testing.T) {
	ve := NewValidationError()
	
	ve.AddInvalidCharacterError("name", "test@#$")
	
	if len(ve.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(ve.Errors))
	}
	
	if ve.Errors[0].Type != ErrorTypeInvalidCharacter {
		t.Errorf("Expected error type %v, got %v", ErrorTypeInvalidCharacter, ve.Errors[0].Type)
	}
	
	if !strings.Contains(ve.Errors[0].Message, "invalid characters") {
		t.Errorf("Expected message to contain 'invalid characters', got %s", ve.Errors[0].Message)
	}
}

func TestValidationError_GetFieldErrors(t *testing.T) {
	ve := NewValidationError()
	
	ve.AddRequiredError("name")
	ve.AddInvalidLengthError("name", "a", 2, 50)
	ve.AddRequiredError("age")
	
	nameErrors := ve.GetFieldErrors("name")
	ageErrors := ve.GetFieldErrors("age")
	missingErrors := ve.GetFieldErrors("missing")
	
	if len(nameErrors) != 2 {
		t.Errorf("Expected 2 errors for 'name', got %d", len(nameErrors))
	}
	
	if len(ageErrors) != 1 {
		t.Errorf("Expected 1 error for 'age', got %d", len(ageErrors))
	}
	
	if len(missingErrors) != 0 {
		t.Errorf("Expected 0 errors for 'missing', got %d", len(missingErrors))
	}
}

func TestValidationError_GetUserFriendlyMessage(t *testing.T) {
	tests := []struct {
		name     string
		errors   []FieldError
		expected string
	}{
		{"No errors", []FieldError{}, "Input validation failed"},
		{"Single error", []FieldError{{Field: "name", Message: "is required"}}, "is required"},
		{"Multiple errors", []FieldError{
			{Field: "name", Message: "is required"},
			{Field: "age", Message: "must be positive"},
		}, "Multiple validation errors occurred"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ve := &ValidationError{Errors: tt.errors}
			result := ve.GetUserFriendlyMessage()
			
			if tt.name == "Multiple errors" {
				if !strings.Contains(result, tt.expected) {
					t.Errorf("GetUserFriendlyMessage() = %v, expected to contain %v", result, tt.expected)
				}
			} else {
				if result != tt.expected {
					t.Errorf("GetUserFriendlyMessage() = %v, expected %v", result, tt.expected)
				}
			}
		})
	}
}

func TestIsValidationError(t *testing.T) {
	ve := NewValidationError()
	ve.AddRequiredError("name")
	
	if !IsValidationError(ve) {
		t.Errorf("IsValidationError() = false, expected true for ValidationError")
	}
	
	regularError := &FieldError{Field: "test", Message: "error"}
	if IsValidationError(regularError) {
		t.Errorf("IsValidationError() = true, expected false for regular error")
	}
}

func TestNewValidationError(t *testing.T) {
	ve := NewValidationError()
	
	if ve == nil {
		t.Error("NewValidationError() returned nil")
	}
	
	if ve.Errors == nil {
		t.Error("NewValidationError() returned ValidationError with nil Errors slice")
	}
	
	if len(ve.Errors) != 0 {
		t.Errorf("NewValidationError() returned ValidationError with %d errors, expected 0", len(ve.Errors))
	}
}