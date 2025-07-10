package sqlite

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatTimeForDB(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "Valid time",
			input:    time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC),
			expected: "2024-01-15T10:30:45Z",
		},
		{
			name:     "Zero time",
			input:    time.Time{},
			expected: "0001-01-01T00:00:00Z",
		},
		{
			name:     "Time with timezone",
			input:    time.Date(2024, 6, 15, 14, 30, 0, 0, time.FixedZone("EST", -5*3600)),
			expected: "2024-06-15T14:30:00-05:00",
		},
		{
			name:     "Time with nanoseconds",
			input:    time.Date(2024, 3, 10, 9, 15, 30, 123456789, time.UTC),
			expected: "2024-03-10T09:15:30Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTimeForDB(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatTimePtrForDB(t *testing.T) {
	tests := []struct {
		name     string
		input    *time.Time
		expected interface{}
	}{
		{
			name:     "Nil pointer",
			input:    nil,
			expected: nil,
		},
		{
			name:     "Valid time pointer",
			input:    &time.Time{},
			expected: "0001-01-01T00:00:00Z",
		},
		{
			name: "Non-zero time pointer",
			input: func() *time.Time {
				t := time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)
				return &t
			}(),
			expected: "2024-01-15T10:30:45Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTimePtrForDB(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseTimeFromDB(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    time.Time
		expectError bool
	}{
		{
			name:        "Valid RFC3339 time",
			input:       "2024-01-15T10:30:45Z",
			expected:    time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC),
			expectError: false,
		},
		{
			name:        "Valid RFC3339 time with timezone",
			input:       "2024-06-15T14:30:00-05:00",
			expected:    time.Date(2024, 6, 15, 14, 30, 0, 0, time.FixedZone("", -5*3600)),
			expectError: false,
		},
		{
			name:        "Valid RFC3339 time with nanoseconds",
			input:       "2024-03-10T09:15:30.123456789Z",
			expected:    time.Date(2024, 3, 10, 9, 15, 30, 123456789, time.UTC),
			expectError: false,
		},
		{
			name:        "Invalid time format",
			input:       "2024-01-15 10:30:45",
			expected:    time.Time{},
			expectError: true,
		},
		{
			name:        "Empty string",
			input:       "",
			expected:    time.Time{},
			expectError: true,
		},
		{
			name:        "Invalid date",
			input:       "2024-13-45T10:30:45Z",
			expected:    time.Time{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTimeFromDB(tt.input)
			if tt.expectError {
				assert.Error(t, err)
				assert.True(t, result.IsZero())
			} else {
				assert.NoError(t, err)
				assert.True(t, tt.expected.Equal(result))
			}
		})
	}
}

func TestFormatTimeForDB_RoundTrip(t *testing.T) {
	// Test that formatting and parsing are consistent
	// Note: RFC3339 format truncates nanoseconds to seconds, so we test without nanoseconds
	originalTime := time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)
	
	formatted := FormatTimeForDB(originalTime)
	parsed, err := ParseTimeFromDB(formatted)
	
	assert.NoError(t, err)
	assert.True(t, originalTime.Equal(parsed))
}

func TestFormatTimePtrForDB_Integration(t *testing.T) {
	// Test that the function works correctly with actual database-like scenarios
	tests := []struct {
		name     string
		input    *time.Time
		nilCheck bool
	}{
		{
			name:     "Running task (nil end time)",
			input:    nil,
			nilCheck: true,
		},
		{
			name: "Completed task (with end time)",
			input: func() *time.Time {
				t := time.Date(2024, 1, 15, 17, 0, 0, 0, time.UTC)
				return &t
			}(),
			nilCheck: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTimePtrForDB(tt.input)
			if tt.nilCheck {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.IsType(t, "", result)
			}
		})
	}
}