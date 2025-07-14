package api

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

func TestGetTask(t *testing.T) {
	tests := []struct {
		name           string
		taskID         int64
		setupTasks     []*domain.Task // Tasks that should exist in the system
		expectedTask   *domain.Task
		expectedError  error
		errorAssertion func(t *testing.T, err error)
	}{
		{
			name:   "should return existing task when valid ID is provided",
			taskID: 0, // Will be set by setupTestBusinessAPI
			setupTasks: []*domain.Task{
				{TaskName: "Test Task"},
				{TaskName: "Another Task"},
			},
			expectedTask: &domain.Task{TaskName: "Test Task"},
		},
		{
			name:       "should return not found error when task does not exist",
			taskID:     999,
			setupTasks: []*domain.Task{},
			errorAssertion: func(t *testing.T, err error) {
				assert.Error(t, err)
				var appErr *errors.AppError
				assert.ErrorAs(t, err, &appErr)
				assert.True(t, appErr.IsType(errors.ErrorTypeNotFound))
				assert.Contains(t, err.Error(), "task")
			},
		},
		{
			name:   "should return validation error when invalid task ID is provided",
			taskID: 0,
			errorAssertion: func(t *testing.T, err error) {
				assert.Error(t, err)
				var appErr *errors.AppError
				assert.ErrorAs(t, err, &appErr)
				assert.True(t, appErr.IsType(errors.ErrorTypeValidation))
				assert.Contains(t, err.Error(), "task ID")
			},
		},
		{
			name:   "should return validation error when negative task ID is provided",
			taskID: -1,
			errorAssertion: func(t *testing.T, err error) {
				assert.Error(t, err)
				var appErr *errors.AppError
				assert.ErrorAs(t, err, &appErr)
				assert.True(t, appErr.IsType(errors.ErrorTypeValidation))
				assert.Contains(t, err.Error(), "task ID")
			},
		},
		{
			name:   "should return correct task when multiple tasks exist",
			taskID: 0, // Will be set to second task ID
			setupTasks: []*domain.Task{
				{TaskName: "First Task"},
				{TaskName: "Second Task"},
				{TaskName: "Third Task"},
			},
			expectedTask: &domain.Task{TaskName: "Second Task"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			api := setupTestBusinessAPI(t, tt.setupTasks, nil)
			ctx := context.Background()

			// Determine the actual task ID to use
			taskID := tt.taskID
			if taskID == 0 && len(tt.setupTasks) > 0 {
				// Use the ID of the first task for "should return existing task" test
				if tt.name == "should return existing task when valid ID is provided" {
					taskID = tt.setupTasks[0].ID
				}
				// Use the ID of the second task for "multiple tasks" test
				if tt.name == "should return correct task when multiple tasks exist" && len(tt.setupTasks) > 1 {
					taskID = tt.setupTasks[1].ID
				}
			}

			// Act
			result, err := api.GetTask(ctx, taskID)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, result)
			} else if tt.errorAssertion != nil {
				tt.errorAssertion(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedTask.TaskName, result.TaskName)
				assert.Greater(t, result.ID, int64(0)) // Should have a valid database ID
			}
		})
	}
}

// setupTestBusinessAPI creates a test API instance with mock data
// This is a behavioral test helper that sets up the API in a specific state
// without relying on implementation details
func setupTestBusinessAPI(t *testing.T, tasks []*domain.Task, timeEntries []*domain.TimeEntry) BusinessAPI {
	// Create real implementation using in-memory database
	repo, err := sqlite.New(":memory:")
	require.NoError(t, err)
	
	businessAPI := NewBusinessAPI(repo)
	
	// Set up test data directly through repository
	ctx := context.Background()
	for _, task := range tasks {
		dbTask := &sqlite.Task{TaskName: task.TaskName}
		err := repo.CreateTask(ctx, dbTask)
		require.NoError(t, err)
		// Update the task ID to match what was created
		task.ID = dbTask.ID
	}
	
	for _, entry := range timeEntries {
		// Map TaskID: 0 to the first task, TaskID: 1 to the second task, etc.
		actualTaskID := entry.TaskID
		if entry.TaskID == 0 && len(tasks) > 0 {
			actualTaskID = tasks[0].ID
		} else if entry.TaskID == 1 && len(tasks) > 1 {
			actualTaskID = tasks[1].ID
		} else if entry.TaskID == 2 && len(tasks) > 2 {
			actualTaskID = tasks[2].ID
		} else if entry.TaskID == 3 && len(tasks) > 3 {
			actualTaskID = tasks[3].ID
		}
		
		dbEntry := &sqlite.TimeEntry{
			TaskID:    actualTaskID,
			StartTime: entry.StartTime,
			EndTime:   entry.EndTime,
		}
		err := repo.CreateTimeEntry(ctx, dbEntry)
		require.NoError(t, err)
		// Update the entry to reflect the actual task ID
		entry.TaskID = actualTaskID
	}
	
	return businessAPI
}

func TestStartNewTask(t *testing.T) {
	tests := []struct {
		name              string
		taskName          string
		existingTasks     []*domain.Task
		existingEntries   []*domain.TimeEntry
		expectedTaskName  string
		expectRunningTask bool
		errorAssertion    func(t *testing.T, err error)
	}{
		{
			name:              "should create new task and start tracking when no running tasks exist",
			taskName:          "New Feature Development",
			existingTasks:     []*domain.Task{},
			existingEntries:   []*domain.TimeEntry{},
			expectedTaskName:  "New Feature Development",
			expectRunningTask: true,
		},
		{
			name:     "should create new task and stop existing running task",
			taskName: "Bug Fix",
			existingTasks: []*domain.Task{
				{TaskName: "Previous Task"},
			},
			existingEntries: []*domain.TimeEntry{
				{TaskID: 0, StartTime: timeNow().Add(-1 * time.Hour), EndTime: nil}, // Running task
			},
			expectedTaskName:  "Bug Fix",
			expectRunningTask: true,
		},
		{
			name:     "should create task when completed tasks exist",
			taskName: "Documentation Update",
			existingTasks: []*domain.Task{
				{TaskName: "Completed Task"},
			},
			existingEntries: []*domain.TimeEntry{
				{TaskID: 0, StartTime: timeNow().Add(-2 * time.Hour), EndTime: timePtr(timeNow().Add(-1 * time.Hour))}, // Completed
			},
			expectedTaskName:  "Documentation Update",
			expectRunningTask: true,
		},
		{
			name:     "should return validation error for empty task name",
			taskName: "",
			errorAssertion: func(t *testing.T, err error) {
				assert.Error(t, err)
				var appErr *errors.AppError
				assert.ErrorAs(t, err, &appErr)
				assert.True(t, appErr.IsType(errors.ErrorTypeValidation))
				assert.Contains(t, err.Error(), "task")
			},
		},
		{
			name:     "should return validation error for whitespace-only task name",
			taskName: "   ",
			errorAssertion: func(t *testing.T, err error) {
				assert.Error(t, err)
				var appErr *errors.AppError
				assert.ErrorAs(t, err, &appErr)
				assert.True(t, appErr.IsType(errors.ErrorTypeValidation))
			},
		},
		{
			name:     "should return validation error for too long task name",
			taskName: string(make([]byte, 300)), // Assuming max is 255
			errorAssertion: func(t *testing.T, err error) {
				assert.Error(t, err)
				var appErr *errors.AppError
				assert.ErrorAs(t, err, &appErr)
				assert.True(t, appErr.IsType(errors.ErrorTypeValidation))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			api := setupTestBusinessAPI(t, tt.existingTasks, tt.existingEntries)
			ctx := context.Background()

			// Fix task IDs in existing entries to match created tasks
			if len(tt.existingTasks) > 0 && len(tt.existingEntries) > 0 {
				for i, entry := range tt.existingEntries {
					if entry.TaskID == 0 { // Use first task
						entry.TaskID = tt.existingTasks[0].ID
						// Create the entry in repository
						repo, _ := setupTestRepo(t)
						dbEntry := &sqlite.TimeEntry{
							TaskID:    entry.TaskID,
							StartTime: entry.StartTime,
							EndTime:   entry.EndTime,
						}
						repo.CreateTimeEntry(ctx, dbEntry)
						tt.existingEntries[i].ID = dbEntry.ID
					}
				}
			}

			// Act
			result, err := api.StartNewTask(ctx, tt.taskName)

			// Assert
			if tt.errorAssertion != nil {
				tt.errorAssertion(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				
				// Verify task was created correctly
				assert.Equal(t, tt.expectedTaskName, result.Task.TaskName)
				assert.Greater(t, result.Task.ID, int64(0))
				
				// Verify time entry was created correctly
				assert.Equal(t, result.Task.ID, result.TimeEntry.TaskID)
				assert.Greater(t, result.TimeEntry.ID, int64(0))
				assert.WithinDuration(t, timeNow(), result.TimeEntry.StartTime, 5*time.Second)
				
				// Verify task is running
				if tt.expectRunningTask {
					assert.Nil(t, result.TimeEntry.EndTime)
					assert.Contains(t, result.Duration, "running") // Should show running duration
				}
				
				// Verify no other tasks are running
				currentSession, err := api.GetCurrentSession(ctx)
				require.NoError(t, err)
				assert.Equal(t, result.Task.ID, currentSession.Task.ID)
			}
		})
	}
}

// Helper functions for tests
func timeNow() time.Time {
	return time.Now().Truncate(time.Second) // Truncate for easier comparison
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func setupTestRepo(t *testing.T) (sqlite.Repository, func()) {
	repo, err := sqlite.New(":memory:")
	require.NoError(t, err)
	cleanup := func() { repo.Close() }
	return repo, cleanup
}

func TestStopAllRunningTasks(t *testing.T) {
	tests := []struct {
		name              string
		existingTasks     []*domain.Task
		existingEntries   []*domain.TimeEntry
		expectedStopped   int
		expectNoRunning   bool
	}{
		{
			name:              "should return empty list when no running tasks exist",
			existingTasks:     []*domain.Task{},
			existingEntries:   []*domain.TimeEntry{},
			expectedStopped:   0,
			expectNoRunning:   true,
		},
		{
			name: "should stop single running task",
			existingTasks: []*domain.Task{
				{TaskName: "Running Task"},
			},
			existingEntries: []*domain.TimeEntry{
				{TaskID: 0, StartTime: timeNow().Add(-1 * time.Hour), EndTime: nil}, // Running
			},
			expectedStopped: 1,
			expectNoRunning: true,
		},
		{
			name: "should stop multiple running tasks",
			existingTasks: []*domain.Task{
				{TaskName: "Task 1"},
				{TaskName: "Task 2"},
			},
			existingEntries: []*domain.TimeEntry{
				{TaskID: 0, StartTime: timeNow().Add(-2 * time.Hour), EndTime: nil}, // Running task 1
				{TaskID: 1, StartTime: timeNow().Add(-1 * time.Hour), EndTime: nil}, // Running task 2
			},
			expectedStopped: 2,
			expectNoRunning: true,
		},
		{
			name: "should not affect completed tasks",
			existingTasks: []*domain.Task{
				{TaskName: "Completed Task"},
				{TaskName: "Running Task"},
			},
			existingEntries: []*domain.TimeEntry{
				{TaskID: 0, StartTime: timeNow().Add(-3 * time.Hour), EndTime: timePtr(timeNow().Add(-2 * time.Hour))}, // Completed
				{TaskID: 1, StartTime: timeNow().Add(-1 * time.Hour), EndTime: nil}, // Running
			},
			expectedStopped: 1,
			expectNoRunning: true,
		},
		{
			name: "should handle mix of completed and running tasks",
			existingTasks: []*domain.Task{
				{TaskName: "Completed Task 1"},
				{TaskName: "Running Task 1"},
				{TaskName: "Completed Task 2"},
				{TaskName: "Running Task 2"},
			},
			existingEntries: []*domain.TimeEntry{
				{TaskID: 0, StartTime: timeNow().Add(-4 * time.Hour), EndTime: timePtr(timeNow().Add(-3 * time.Hour))}, // Completed
				{TaskID: 1, StartTime: timeNow().Add(-2 * time.Hour), EndTime: nil}, // Running
				{TaskID: 2, StartTime: timeNow().Add(-90 * time.Minute), EndTime: timePtr(timeNow().Add(-80 * time.Minute))}, // Completed
				{TaskID: 3, StartTime: timeNow().Add(-30 * time.Minute), EndTime: nil}, // Running
			},
			expectedStopped: 2,
			expectNoRunning: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			api := setupTestBusinessAPI(t, tt.existingTasks, tt.existingEntries)
			ctx := context.Background()

			// Act
			stoppedEntries, err := api.StopAllRunningTasks(ctx)

			// Assert
			require.NoError(t, err)
			assert.Len(t, stoppedEntries, tt.expectedStopped)

			// Verify all returned entries have EndTime set
			for _, entry := range stoppedEntries {
				assert.NotNil(t, entry.EndTime, "Stopped entry should have EndTime set")
				assert.WithinDuration(t, timeNow(), *entry.EndTime, 5*time.Second)
			}

			// Verify no running tasks remain
			if tt.expectNoRunning {
				currentSession, err := api.GetCurrentSession(ctx)
				assert.Error(t, err, "Should not have any running tasks")
				assert.Nil(t, currentSession)
				
				var appErr *errors.AppError
				assert.ErrorAs(t, err, &appErr)
				assert.True(t, appErr.IsType(errors.ErrorTypeNotFound))
			}
		})
	}
}