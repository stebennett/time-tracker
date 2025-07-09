package validation

import (
	"time-tracker/internal/domain"
)

// TaskValidator provides validation for Task-related operations
type TaskValidator struct {
	validator *Validator
}

// NewTaskValidator creates a new task validator
func NewTaskValidator() *TaskValidator {
	return &TaskValidator{
		validator: NewValidator(),
	}
}

// ValidateTaskName validates a task name for creation or update
func (tv *TaskValidator) ValidateTaskName(name string) error {
	validationError := NewValidationError()
	
	// Trim whitespace
	trimmedName := tv.validator.TrimAndValidateString(name)
	
	// Check if name is empty
	if !tv.validator.IsNonEmptyString(trimmedName) {
		validationError.AddRequiredError("task_name")
		return validationError
	}
	
	// Check length constraints (1-255 characters)
	if !tv.validator.IsValidStringLength(trimmedName, 1, 255) {
		validationError.AddInvalidLengthError("task_name", trimmedName, 1, 255)
	}
	
	// Check for valid characters
	if !tv.validator.IsValidTaskName(trimmedName) {
		validationError.AddInvalidCharacterError("task_name", trimmedName)
	}
	
	if validationError.HasErrors() {
		return validationError
	}
	
	return nil
}

// ValidateTaskForCreation validates a task for creation
func (tv *TaskValidator) ValidateTaskForCreation(name string) error {
	return tv.ValidateTaskName(name)
}

// ValidateTaskForUpdate validates a task for update
func (tv *TaskValidator) ValidateTaskForUpdate(id int64, name string) error {
	validationError := NewValidationError()
	
	// Validate task ID
	if !tv.validator.IsValidTaskID(id) {
		validationError.AddInvalidValueError("task_id", id, "must be a positive integer")
	}
	
	// Validate task name
	if nameErr := tv.ValidateTaskName(name); nameErr != nil {
		if nameValidationErr, ok := nameErr.(*ValidationError); ok {
			validationError.Errors = append(validationError.Errors, nameValidationErr.Errors...)
		}
	}
	
	if validationError.HasErrors() {
		return validationError
	}
	
	return nil
}

// ValidateTask validates a domain.Task object
func (tv *TaskValidator) ValidateTask(task domain.Task) error {
	validationError := NewValidationError()
	
	// Validate task name
	if nameErr := tv.ValidateTaskName(task.TaskName); nameErr != nil {
		if nameValidationErr, ok := nameErr.(*ValidationError); ok {
			validationError.Errors = append(validationError.Errors, nameValidationErr.Errors...)
		}
	}
	
	// If task has an ID, validate it
	if task.ID != 0 && !tv.validator.IsValidTaskID(task.ID) {
		validationError.AddInvalidValueError("task_id", task.ID, "must be a positive integer")
	}
	
	if validationError.HasErrors() {
		return validationError
	}
	
	return nil
}

// ValidateTaskID validates a task ID
func (tv *TaskValidator) ValidateTaskID(id int64) error {
	if !tv.validator.IsValidTaskID(id) {
		validationError := NewValidationError()
		validationError.AddInvalidValueError("task_id", id, "must be a positive integer")
		return validationError
	}
	return nil
}

// GetValidTaskName returns a cleaned task name if valid
func (tv *TaskValidator) GetValidTaskName(name string) (string, error) {
	if err := tv.ValidateTaskName(name); err != nil {
		return "", err
	}
	return tv.validator.TrimAndValidateString(name), nil
}