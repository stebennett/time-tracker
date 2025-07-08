package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTask(t *testing.T) {
	tests := []struct {
		name     string
		taskName string
		expected Task
	}{
		{
			name:     "creates task with name",
			taskName: "Test Task",
			expected: Task{TaskName: "Test Task"},
		},
		{
			name:     "creates task with empty name",
			taskName: "",
			expected: Task{TaskName: ""},
		},
		{
			name:     "creates task with special characters",
			taskName: "Task-with_special@chars!",
			expected: Task{TaskName: "Task-with_special@chars!"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewTask(tt.taskName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTask_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		task     Task
		expected bool
	}{
		{
			name:     "valid task with name",
			task:     Task{ID: 1, TaskName: "Valid Task"},
			expected: true,
		},
		{
			name:     "invalid task with empty name",
			task:     Task{ID: 1, TaskName: ""},
			expected: false,
		},
		{
			name:     "valid task with zero ID",
			task:     Task{ID: 0, TaskName: "Valid Task"},
			expected: true,
		},
		{
			name:     "valid task with whitespace",
			task:     Task{ID: 1, TaskName: "   "},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.task.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTask_String(t *testing.T) {
	tests := []struct {
		name     string
		task     Task
		expected string
	}{
		{
			name:     "returns task name",
			task:     Task{ID: 1, TaskName: "My Task"},
			expected: "My Task",
		},
		{
			name:     "returns empty string for empty task name",
			task:     Task{ID: 1, TaskName: ""},
			expected: "",
		},
		{
			name:     "returns task name with special characters",
			task:     Task{ID: 1, TaskName: "Task-with_special@chars!"},
			expected: "Task-with_special@chars!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.task.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}