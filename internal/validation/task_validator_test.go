package validation

import (
	"strings"
	"testing"
	"time-tracker/internal/domain"
)

func TestTaskValidator_ValidateTaskName(t *testing.T) {
	validator := NewTaskValidator()

	tests := []struct {
		name        string
		input       string
		expectError bool
		errorType   ValidationErrorType
	}{
		{"Valid name", "Task 1", false, ""},
		{"Empty name", "", true, ErrorTypeRequired},
		{"Whitespace only", "   ", true, ErrorTypeRequired},
		{"Too long name", strings.Repeat("a", 256), true, ErrorTypeInvalidLength},
		{"Valid long name", strings.Repeat("a", 255), false, ""},
		{"Invalid characters", "Task@#$%", true, ErrorTypeInvalidCharacter},
		{"Valid with punctuation", "Task! (important)", false, ""},
		{"Valid with hyphen", "Task-1", false, ""},
		{"Valid with underscore", "Task_1", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateTaskName(tt.input)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateTaskName(%q) expected error but got nil", tt.input)
					return
				}
				
				validationErr, ok := err.(*ValidationError)
				if !ok {
					t.Errorf("ValidateTaskName(%q) expected ValidationError but got %T", tt.input, err)
					return
				}
				
				if len(validationErr.Errors) == 0 {
					t.Errorf("ValidateTaskName(%q) expected validation errors but got none", tt.input)
					return
				}
				
				if validationErr.Errors[0].Type != tt.errorType {
					t.Errorf("ValidateTaskName(%q) expected error type %v but got %v", tt.input, tt.errorType, validationErr.Errors[0].Type)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateTaskName(%q) expected no error but got %v", tt.input, err)
				}
			}
		})
	}
}

func TestTaskValidator_ValidateTaskForCreation(t *testing.T) {
	validator := NewTaskValidator()

	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{"Valid name", "Task 1", false},
		{"Empty name", "", true},
		{"Invalid characters", "Task@#$%", true},
		{"Valid with spaces", "My Important Task", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateTaskForCreation(tt.input)
			
			if tt.expectError && err == nil {
				t.Errorf("ValidateTaskForCreation(%q) expected error but got nil", tt.input)
			} else if !tt.expectError && err != nil {
				t.Errorf("ValidateTaskForCreation(%q) expected no error but got %v", tt.input, err)
			}
		})
	}
}

func TestTaskValidator_ValidateTaskForUpdate(t *testing.T) {
	validator := NewTaskValidator()

	tests := []struct {
		name        string
		id          int64
		taskName    string
		expectError bool
	}{
		{"Valid update", 1, "Task 1", false},
		{"Invalid ID", 0, "Task 1", true},
		{"Negative ID", -1, "Task 1", true},
		{"Valid ID, invalid name", 1, "", true},
		{"Valid ID, invalid characters", 1, "Task@#$%", true},
		{"Valid ID, valid name", 1, "My Important Task", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateTaskForUpdate(tt.id, tt.taskName)
			
			if tt.expectError && err == nil {
				t.Errorf("ValidateTaskForUpdate(%d, %q) expected error but got nil", tt.id, tt.taskName)
			} else if !tt.expectError && err != nil {
				t.Errorf("ValidateTaskForUpdate(%d, %q) expected no error but got %v", tt.id, tt.taskName, err)
			}
		})
	}
}

func TestTaskValidator_ValidateTask(t *testing.T) {
	validator := NewTaskValidator()

	tests := []struct {
		name        string
		task        domain.Task
		expectError bool
	}{
		{"Valid task", domain.Task{ID: 1, TaskName: "Task 1"}, false},
		{"Valid task without ID", domain.Task{TaskName: "Task 1"}, false},
		{"Invalid task name", domain.Task{ID: 1, TaskName: ""}, true},
		{"Invalid ID", domain.Task{ID: -1, TaskName: "Task 1"}, true},
		{"Invalid characters", domain.Task{ID: 1, TaskName: "Task@#$%"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateTask(tt.task)
			
			if tt.expectError && err == nil {
				t.Errorf("ValidateTask(%+v) expected error but got nil", tt.task)
			} else if !tt.expectError && err != nil {
				t.Errorf("ValidateTask(%+v) expected no error but got %v", tt.task, err)
			}
		})
	}
}

func TestTaskValidator_ValidateTaskID(t *testing.T) {
	validator := NewTaskValidator()

	tests := []struct {
		name        string
		id          int64
		expectError bool
	}{
		{"Valid ID", 1, false},
		{"Zero ID", 0, true},
		{"Negative ID", -1, true},
		{"Large ID", 999999, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateTaskID(tt.id)
			
			if tt.expectError && err == nil {
				t.Errorf("ValidateTaskID(%d) expected error but got nil", tt.id)
			} else if !tt.expectError && err != nil {
				t.Errorf("ValidateTaskID(%d) expected no error but got %v", tt.id, err)
			}
		})
	}
}

func TestTaskValidator_GetValidTaskName(t *testing.T) {
	validator := NewTaskValidator()

	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{"Valid name", "Task 1", "Task 1", false},
		{"Name with spaces", "  Task 1  ", "Task 1", false},
		{"Empty name", "", "", true},
		{"Whitespace only", "   ", "", true},
		{"Invalid characters", "Task@#$%", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.GetValidTaskName(tt.input)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("GetValidTaskName(%q) expected error but got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("GetValidTaskName(%q) expected no error but got %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("GetValidTaskName(%q) = %q, expected %q", tt.input, result, tt.expected)
				}
			}
		})
	}
}