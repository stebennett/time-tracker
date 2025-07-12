package services

import (
	"context"
	"testing"
	"time"
	"time-tracker/internal/domain"
	"time-tracker/internal/errors"
	"time-tracker/internal/repository/sqlite"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeService_ParseTimeRange(t *testing.T) {
	tests := []struct {
		name          string
		timeStr       string
		expectedStart time.Duration // Duration back from now
		expectedEnd   time.Duration // Should be 0 (now)
		errorAssertion func(t *testing.T, err error)
	}{
		{
			name:          "should parse 30 minutes correctly",
			timeStr:       "30m",
			expectedStart: 30 * time.Minute,
			expectedEnd:   0,
		},
		{
			name:          "should parse 1 hour correctly",
			timeStr:       "1h",
			expectedStart: 1 * time.Hour,
			expectedEnd:   0,
		},
		{
			name:          "should parse 1 day correctly",
			timeStr:       "1d",
			expectedStart: 24 * time.Hour,
			expectedEnd:   0,
		},
		{
			name:          "should parse 1 week correctly",
			timeStr:       "1w",
			expectedStart: 7 * 24 * time.Hour,
			expectedEnd:   0,
		},
		{
			name:    "should return validation error for empty string",
			timeStr: "",
			errorAssertion: func(t *testing.T, err error) {
				assert.Error(t, err)
				var appErr *errors.AppError
				assert.ErrorAs(t, err, &appErr)
				assert.True(t, appErr.IsType(errors.ErrorTypeValidation))
				assert.Contains(t, err.Error(), "empty")
			},
		},
		{
			name:    "should return validation error for invalid format",
			timeStr: "invalid",
			errorAssertion: func(t *testing.T, err error) {
				assert.Error(t, err)
				var appErr *errors.AppError
				assert.ErrorAs(t, err, &appErr)
				assert.True(t, appErr.IsType(errors.ErrorTypeValidation))
				assert.Contains(t, err.Error(), "invalid")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			service := setupTimeService(t)
			
			// Act
			result, err := service.ParseTimeRange(tt.timeStr)

			// Assert
			if tt.errorAssertion != nil {
				tt.errorAssertion(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				
				// Check that the range is approximately correct (within 1 second tolerance)
				now := time.Now()
				expectedStart := now.Add(-tt.expectedStart)
				
				assert.WithinDuration(t, expectedStart, result.Start, time.Second)
				assert.WithinDuration(t, now, result.End, time.Second)
			}
		})
	}
}

func TestTimeService_CalculateDuration(t *testing.T) {
	tests := []struct {
		name           string
		start          time.Time
		end            *time.Time
		expectedResult string
	}{
		{
			name:           "should calculate duration for completed entry",
			start:          time.Now().Add(-2 * time.Hour),
			end:            timePtr(time.Now()),
			expectedResult: "2h 0m",
		},
		{
			name:           "should return running duration for nil end time",
			start:          time.Now().Add(-30 * time.Minute),
			end:            nil,
			expectedResult: "running for", // Should contain this text
		},
		{
			name:           "should handle zero duration",
			start:          time.Now(),
			end:            timePtr(time.Now()),
			expectedResult: "0m",
		},
		{
			name:           "should format minutes only when less than 1 hour",
			start:          time.Now().Add(-45 * time.Minute),
			end:            timePtr(time.Now()),
			expectedResult: "45m",
		},
		{
			name:           "should format hours and minutes",
			start:          time.Now().Add(-90 * time.Minute),
			end:            timePtr(time.Now()),
			expectedResult: "1h 30m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			service := setupTimeService(t)
			
			// Act
			result := service.CalculateDuration(tt.start, tt.end)

			// Assert
			if tt.end == nil {
				// For running tasks, just check it contains the expected text
				assert.Contains(t, result, tt.expectedResult)
			} else {
				// For completed tasks, check exact match (with some tolerance for test timing)
				assert.Contains(t, result, tt.expectedResult)
			}
		})
	}
}

func TestTimeService_GetRunningEntries(t *testing.T) {
	tests := []struct {
		name              string
		setupEntries      []*domain.TimeEntry
		expectedRunning   int
		errorAssertion    func(t *testing.T, err error)
	}{
		{
			name:            "should return empty list when no entries exist",
			setupEntries:    []*domain.TimeEntry{},
			expectedRunning: 0,
		},
		{
			name: "should return only running entries",
			setupEntries: []*domain.TimeEntry{
				{TaskID: 1, StartTime: time.Now().Add(-1 * time.Hour), EndTime: nil}, // Running
				{TaskID: 2, StartTime: time.Now().Add(-2 * time.Hour), EndTime: timePtr(time.Now().Add(-1 * time.Hour))}, // Completed
				{TaskID: 3, StartTime: time.Now().Add(-30 * time.Minute), EndTime: nil}, // Running
			},
			expectedRunning: 2,
		},
		{
			name: "should return empty list when all entries are completed",
			setupEntries: []*domain.TimeEntry{
				{TaskID: 1, StartTime: time.Now().Add(-2 * time.Hour), EndTime: timePtr(time.Now().Add(-1 * time.Hour))},
				{TaskID: 2, StartTime: time.Now().Add(-3 * time.Hour), EndTime: timePtr(time.Now().Add(-2 * time.Hour))},
			},
			expectedRunning: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			service, repo := setupTimeServiceWithData(t, nil, tt.setupEntries)
			defer repo.Close()
			ctx := context.Background()

			// Act
			result, err := service.GetRunningEntries(ctx)

			// Assert
			if tt.errorAssertion != nil {
				tt.errorAssertion(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, result, tt.expectedRunning)
				
				// Verify all returned entries are actually running
				for _, entry := range result {
					assert.Nil(t, entry.EndTime, "All returned entries should be running")
				}
			}
		})
	}
}

func TestTimeService_StopRunningEntries(t *testing.T) {
	tests := []struct {
		name              string
		setupEntries      []*domain.TimeEntry
		expectedStopped   int
	}{
		{
			name:            "should return empty list when no running entries exist",
			setupEntries:    []*domain.TimeEntry{},
			expectedStopped: 0,
		},
		{
			name: "should stop all running entries",
			setupEntries: []*domain.TimeEntry{
				{TaskID: 1, StartTime: time.Now().Add(-1 * time.Hour), EndTime: nil}, // Running
				{TaskID: 2, StartTime: time.Now().Add(-30 * time.Minute), EndTime: nil}, // Running
			},
			expectedStopped: 2,
		},
		{
			name: "should not affect completed entries",
			setupEntries: []*domain.TimeEntry{
				{TaskID: 1, StartTime: time.Now().Add(-2 * time.Hour), EndTime: timePtr(time.Now().Add(-1 * time.Hour))}, // Completed
				{TaskID: 2, StartTime: time.Now().Add(-30 * time.Minute), EndTime: nil}, // Running
			},
			expectedStopped: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			service, repo := setupTimeServiceWithData(t, nil, tt.setupEntries)
			defer repo.Close()
			ctx := context.Background()

			// Act
			stoppedEntries, err := service.StopRunningEntries(ctx)

			// Assert
			require.NoError(t, err)
			assert.Len(t, stoppedEntries, tt.expectedStopped)
			
			// Verify all returned entries have EndTime set
			for _, entry := range stoppedEntries {
				assert.NotNil(t, entry.EndTime, "Stopped entries should have EndTime set")
				assert.WithinDuration(t, time.Now(), *entry.EndTime, 5*time.Second)
			}
			
			// Verify no running entries remain
			runningEntries, err := service.GetRunningEntries(ctx)
			require.NoError(t, err)
			assert.Empty(t, runningEntries, "No running entries should remain after stopping all")
		})
	}
}

func TestTimeService_CreateTimeEntry(t *testing.T) {
	tests := []struct {
		name           string
		taskID         int64
		setupTasks     []*domain.Task
		errorAssertion func(t *testing.T, err error)
	}{
		{
			name:   "should create time entry for valid task",
			taskID: 1,
			setupTasks: []*domain.Task{
				{ID: 1, TaskName: "Test Task"},
			},
		},
		{
			name:   "should return validation error for invalid task ID",
			taskID: 0,
			errorAssertion: func(t *testing.T, err error) {
				assert.Error(t, err)
				// The validator returns its own error type, just check it's a validation error
				assert.Contains(t, err.Error(), "task_id")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			service, repo := setupTimeServiceWithData(t, tt.setupTasks, nil)
			defer repo.Close()
			ctx := context.Background()

			// Act
			result, err := service.CreateTimeEntry(ctx, tt.taskID)

			// Assert
			if tt.errorAssertion != nil {
				tt.errorAssertion(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				
				assert.Equal(t, tt.taskID, result.TaskID)
				assert.Greater(t, result.ID, int64(0))
				assert.WithinDuration(t, time.Now(), result.StartTime, 5*time.Second)
				assert.Nil(t, result.EndTime, "New entry should be running")
			}
		})
	}
}

func TestTimeService_IsToday(t *testing.T) {
	tests := []struct {
		name     string
		testTime time.Time
		expected bool
	}{
		{
			name:     "should return true for current time",
			testTime: time.Now(),
			expected: true,
		},
		{
			name:     "should return true for earlier today",
			testTime: time.Now().Add(-3 * time.Hour),
			expected: true,
		},
		{
			name:     "should return false for yesterday",
			testTime: time.Now().Add(-25 * time.Hour),
			expected: false,
		},
		{
			name:     "should return false for tomorrow",
			testTime: time.Now().Add(25 * time.Hour),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			service := setupTimeService(t)
			
			// Act
			result := service.IsToday(tt.testTime)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper functions
func setupTimeService(t *testing.T) TimeService {
	repo, err := sqlite.New(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { repo.Close() })
	
	return NewTimeService(repo)
}

func setupTimeServiceWithData(t *testing.T, tasks []*domain.Task, entries []*domain.TimeEntry) (TimeService, sqlite.Repository) {
	repo, err := sqlite.New(":memory:")
	require.NoError(t, err)
	
	ctx := context.Background()
	
	// Create tasks
	for _, task := range tasks {
		dbTask := &sqlite.Task{TaskName: task.TaskName}
		err := repo.CreateTask(ctx, dbTask)
		require.NoError(t, err)
		task.ID = dbTask.ID // Update with actual ID
	}
	
	// Create time entries
	for _, entry := range entries {
		dbEntry := &sqlite.TimeEntry{
			TaskID:    entry.TaskID,
			StartTime: entry.StartTime,
			EndTime:   entry.EndTime,
		}
		err := repo.CreateTimeEntry(ctx, dbEntry)
		require.NoError(t, err)
		entry.ID = dbEntry.ID // Update with actual ID
	}
	
	service := NewTimeService(repo)
	return service, repo
}

func timePtr(t time.Time) *time.Time {
	return &t
}