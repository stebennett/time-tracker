package validation

import (
	"fmt"
	"strings"
)

// ValidationErrorType represents the type of validation error
type ValidationErrorType string

const (
	ErrorTypeRequired         ValidationErrorType = "required"
	ErrorTypeInvalidFormat    ValidationErrorType = "invalid_format"
	ErrorTypeInvalidLength    ValidationErrorType = "invalid_length"
	ErrorTypeInvalidValue     ValidationErrorType = "invalid_value"
	ErrorTypeInvalidRange     ValidationErrorType = "invalid_range"
	ErrorTypeInvalidCharacter ValidationErrorType = "invalid_character"
)

// FieldError represents a validation error for a specific field
type FieldError struct {
	Field   string
	Type    ValidationErrorType
	Message string
	Value   interface{}
}

// Error implements the error interface for FieldError
func (fe *FieldError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", fe.Field, fe.Message)
}

// ValidationError represents a collection of validation errors
type ValidationError struct {
	Errors []FieldError
}

// Error implements the error interface for ValidationError
func (ve *ValidationError) Error() string {
	if len(ve.Errors) == 0 {
		return "validation error"
	}
	
	if len(ve.Errors) == 1 {
		return ve.Errors[0].Error()
	}
	
	var messages []string
	for _, err := range ve.Errors {
		messages = append(messages, err.Error())
	}
	
	return fmt.Sprintf("multiple validation errors: %s", strings.Join(messages, "; "))
}

// IsValidationError checks if an error is a ValidationError
func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

// HasErrors returns true if the ValidationError has any errors
func (ve *ValidationError) HasErrors() bool {
	return len(ve.Errors) > 0
}

// AddError adds a new field error to the validation error
func (ve *ValidationError) AddError(field string, errorType ValidationErrorType, message string, value interface{}) {
	ve.Errors = append(ve.Errors, FieldError{
		Field:   field,
		Type:    errorType,
		Message: message,
		Value:   value,
	})
}

// AddRequiredError adds a required field error
func (ve *ValidationError) AddRequiredError(field string) {
	ve.AddError(field, ErrorTypeRequired, fmt.Sprintf("%s is required", field), nil)
}

// AddInvalidFormatError adds an invalid format error
func (ve *ValidationError) AddInvalidFormatError(field string, value interface{}, expectedFormat string) {
	message := fmt.Sprintf("%s has invalid format, expected: %s", field, expectedFormat)
	ve.AddError(field, ErrorTypeInvalidFormat, message, value)
}

// AddInvalidLengthError adds an invalid length error
func (ve *ValidationError) AddInvalidLengthError(field string, value interface{}, min, max int) {
	var message string
	if min > 0 && max > 0 {
		message = fmt.Sprintf("%s must be between %d and %d characters long", field, min, max)
	} else if min > 0 {
		message = fmt.Sprintf("%s must be at least %d characters long", field, min)
	} else if max > 0 {
		message = fmt.Sprintf("%s must be at most %d characters long", field, max)
	} else {
		message = fmt.Sprintf("%s has invalid length", field)
	}
	ve.AddError(field, ErrorTypeInvalidLength, message, value)
}

// AddInvalidValueError adds an invalid value error
func (ve *ValidationError) AddInvalidValueError(field string, value interface{}, reason string) {
	message := fmt.Sprintf("%s has invalid value: %s", field, reason)
	ve.AddError(field, ErrorTypeInvalidValue, message, value)
}

// AddInvalidRangeError adds an invalid range error
func (ve *ValidationError) AddInvalidRangeError(field string, value interface{}, reason string) {
	message := fmt.Sprintf("%s has invalid range: %s", field, reason)
	ve.AddError(field, ErrorTypeInvalidRange, message, value)
}

// AddInvalidCharacterError adds an invalid character error
func (ve *ValidationError) AddInvalidCharacterError(field string, value interface{}) {
	message := fmt.Sprintf("%s contains invalid characters", field)
	ve.AddError(field, ErrorTypeInvalidCharacter, message, value)
}

// NewValidationError creates a new ValidationError
func NewValidationError() *ValidationError {
	return &ValidationError{
		Errors: make([]FieldError, 0),
	}
}

// GetFieldErrors returns all errors for a specific field
func (ve *ValidationError) GetFieldErrors(field string) []FieldError {
	var fieldErrors []FieldError
	for _, err := range ve.Errors {
		if err.Field == field {
			fieldErrors = append(fieldErrors, err)
		}
	}
	return fieldErrors
}

// GetUserFriendlyMessage returns a user-friendly error message
func (ve *ValidationError) GetUserFriendlyMessage() string {
	if len(ve.Errors) == 0 {
		return "Input validation failed"
	}
	
	if len(ve.Errors) == 1 {
		return ve.Errors[0].Message
	}
	
	return fmt.Sprintf("Multiple validation errors occurred:\n%s", 
		strings.Join(ve.getUserFriendlyMessages(), "\n"))
}

// getUserFriendlyMessages returns user-friendly messages for all errors
func (ve *ValidationError) getUserFriendlyMessages() []string {
	var messages []string
	for _, err := range ve.Errors {
		messages = append(messages, fmt.Sprintf("- %s", err.Message))
	}
	return messages
}