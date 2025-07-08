package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTimeEntry(t *testing.T) {
	taskID := int64(1)
	startTime := time.Now()

	result := NewTimeEntry(taskID, startTime)

	assert.Equal(t, taskID, result.TaskID)
	assert.Equal(t, startTime, result.StartTime)
	assert.Nil(t, result.EndTime)
	assert.Equal(t, int64(0), result.ID)
}

func TestTimeEntry_IsRunning(t *testing.T) {
	tests := []struct {
		name     string
		entry    TimeEntry
		expected bool
	}{
		{
			name: "running entry with nil end time",
			entry: TimeEntry{
				ID:        1,
				TaskID:    1,
				StartTime: time.Now(),
				EndTime:   nil,
			},
			expected: true,
		},
		{
			name: "stopped entry with end time",
			entry: TimeEntry{
				ID:        1,
				TaskID:    1,
				StartTime: time.Now().Add(-time.Hour),
				EndTime:   &[]time.Time{time.Now()}[0],
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.entry.IsRunning()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTimeEntry_Stop(t *testing.T) {
	startTime := time.Now().Add(-time.Hour)
	endTime := time.Now()
	entry := TimeEntry{
		ID:        1,
		TaskID:    1,
		StartTime: startTime,
		EndTime:   nil,
	}

	result := entry.Stop(endTime)

	assert.Equal(t, entry.ID, result.ID)
	assert.Equal(t, entry.TaskID, result.TaskID)
	assert.Equal(t, entry.StartTime, result.StartTime)
	assert.NotNil(t, result.EndTime)
	assert.Equal(t, endTime, *result.EndTime)
}

func TestTimeEntry_Duration(t *testing.T) {
	tests := []struct {
		name     string
		entry    TimeEntry
		expected time.Duration
	}{
		{
			name: "stopped entry duration",
			entry: TimeEntry{
				ID:        1,
				TaskID:    1,
				StartTime: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
				EndTime:   &[]time.Time{time.Date(2023, 1, 1, 11, 0, 0, 0, time.UTC)}[0],
			},
			expected: time.Hour,
		},
		{
			name: "30 minute stopped entry",
			entry: TimeEntry{
				ID:        1,
				TaskID:    1,
				StartTime: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
				EndTime:   &[]time.Time{time.Date(2023, 1, 1, 10, 30, 0, 0, time.UTC)}[0],
			},
			expected: 30 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.entry.Duration()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTimeEntry_Duration_Running(t *testing.T) {
	// For running entries, we can't test exact duration due to time.Since()
	// but we can test that it returns a positive duration
	entry := TimeEntry{
		ID:        1,
		TaskID:    1,
		StartTime: time.Now().Add(-time.Hour),
		EndTime:   nil,
	}

	result := entry.Duration()
	assert.True(t, result > 0)
	assert.True(t, result < 2*time.Hour) // Should be less than 2 hours
}

func TestTimeEntry_IsValid(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		entry    TimeEntry
		expected bool
	}{
		{
			name: "valid running entry",
			entry: TimeEntry{
				ID:        1,
				TaskID:    1,
				StartTime: now,
				EndTime:   nil,
			},
			expected: true,
		},
		{
			name: "valid stopped entry",
			entry: TimeEntry{
				ID:        1,
				TaskID:    1,
				StartTime: now.Add(-time.Hour),
				EndTime:   &now,
			},
			expected: true,
		},
		{
			name: "invalid entry with zero task ID",
			entry: TimeEntry{
				ID:        1,
				TaskID:    0,
				StartTime: now,
				EndTime:   nil,
			},
			expected: false,
		},
		{
			name: "invalid entry with negative task ID",
			entry: TimeEntry{
				ID:        1,
				TaskID:    -1,
				StartTime: now,
				EndTime:   nil,
			},
			expected: false,
		},
		{
			name: "invalid entry with zero start time",
			entry: TimeEntry{
				ID:        1,
				TaskID:    1,
				StartTime: time.Time{},
				EndTime:   nil,
			},
			expected: false,
		},
		{
			name: "invalid entry with end time before start time",
			entry: TimeEntry{
				ID:        1,
				TaskID:    1,
				StartTime: now,
				EndTime:   &[]time.Time{now.Add(-time.Hour)}[0],
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.entry.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}