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

func TestTaskService_CreateTask(t *testing.T) {
	tests := []struct {
		name           string
		taskName       string
		errorAssertion func(t *testing.T, err error)
	}{
		{
			name:     "should create task with valid name",
			taskName: "Test Task",
		},
		{
			name:     "should create task with minimum length name",
			taskName: "T",
		},
		{
			name:     "should return validation error for empty name",
			taskName: "",
			errorAssertion: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "name")
			},
		},
		{
			name:     "should return validation error for whitespace-only name",
			taskName: "   ",
			errorAssertion: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "name")
			},
		},
		{
			name:     "should return validation error for very long name",
			taskName: string(make([]byte, 300)), // 300 characters, over typical 255 limit
			errorAssertion: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "name")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			service := setupTaskService(t)
			ctx := context.Background()

			// Act
			result, err := service.CreateTask(ctx, tt.taskName)

			// Assert
			if tt.errorAssertion != nil {
				tt.errorAssertion(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Greater(t, result.ID, int64(0))
				assert.Equal(t, tt.taskName, result.TaskName)
			}
		})
	}
}

func TestTaskService_GetTask(t *testing.T) {
	tests := []struct {
		name           string
		taskID         int64
		setupTasks     []*domain.Task
		errorAssertion func(t *testing.T, err error)
	}{
		{
			name:   "should return existing task",
			taskID: 1,
			setupTasks: []*domain.Task{
				{TaskName: "Test Task"},
			},
		},
		{
			name:   "should return not found error for non-existent task",
			taskID: 999,
			errorAssertion: func(t *testing.T, err error) {
				assert.Error(t, err)
				var appErr *errors.AppError
				assert.ErrorAs(t, err, &appErr)
				assert.True(t, appErr.IsType(errors.ErrorTypeNotFound))
			},
		},
		{
			name:   "should return validation error for invalid ID",
			taskID: 0,
			errorAssertion: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "id")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			service, repo := setupTaskServiceWithData(t, tt.setupTasks, nil)
			defer repo.Close()
			ctx := context.Background()

			// Use actual ID from created task if we have setup tasks
			actualTaskID := tt.taskID
			if len(tt.setupTasks) > 0 && tt.taskID == 1 {
				actualTaskID = tt.setupTasks[0].ID
			}

			// Act
			result, err := service.GetTask(ctx, actualTaskID)

			// Assert
			if tt.errorAssertion != nil {
				tt.errorAssertion(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, actualTaskID, result.ID)
				assert.Equal(t, tt.setupTasks[0].TaskName, result.TaskName)
			}
		})
	}
}

func TestTaskService_UpdateTask(t *testing.T) {
	tests := []struct {
		name           string
		taskID         int64
		newName        string
		setupTasks     []*domain.Task
		errorAssertion func(t *testing.T, err error)
	}{
		{
			name:    "should update existing task",
			taskID:  1,
			newName: "Updated Task",
			setupTasks: []*domain.Task{
				{TaskName: "Original Task"},
			},
		},
		{
			name:    "should return not found error for non-existent task",
			taskID:  999,
			newName: "Updated Task",
			errorAssertion: func(t *testing.T, err error) {
				assert.Error(t, err)
				var appErr *errors.AppError
				assert.ErrorAs(t, err, &appErr)
				assert.True(t, appErr.IsType(errors.ErrorTypeNotFound))
			},
		},
		{
			name:    "should return validation error for empty name",
			taskID:  1,
			newName: "",
			setupTasks: []*domain.Task{
				{TaskName: "Original Task"},
			},
			errorAssertion: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "name")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			service, repo := setupTaskServiceWithData(t, tt.setupTasks, nil)
			defer repo.Close()
			ctx := context.Background()

			// Use actual ID from created task if we have setup tasks
			actualTaskID := tt.taskID
			if len(tt.setupTasks) > 0 && tt.taskID == 1 {
				actualTaskID = tt.setupTasks[0].ID
			}

			// Act
			result, err := service.UpdateTask(ctx, actualTaskID, tt.newName)

			// Assert
			if tt.errorAssertion != nil {
				tt.errorAssertion(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, actualTaskID, result.ID)
				assert.Equal(t, tt.newName, result.TaskName)
			}
		})
	}
}

func TestTaskService_DeleteTaskWithEntries(t *testing.T) {
	tests := []struct {
		name           string
		taskID         int64
		setupTasks     []*domain.Task
		setupEntries   []*domain.TimeEntry
		errorAssertion func(t *testing.T, err error)
	}{
		{
			name:   "should delete task without time entries",
			taskID: 1,
			setupTasks: []*domain.Task{
				{TaskName: "Task to Delete"},
			},
		},
		{
			name:   "should delete task with time entries",
			taskID: 1,
			setupTasks: []*domain.Task{
				{TaskName: "Task to Delete"},
			},
			setupEntries: []*domain.TimeEntry{
				{TaskID: 1, StartTime: time.Now().Add(-1 * time.Hour), EndTime: timePtr(time.Now())},
			},
		},
		{
			name:   "should return not found error for non-existent task",
			taskID: 999,
			errorAssertion: func(t *testing.T, err error) {
				assert.Error(t, err)
				var appErr *errors.AppError
				assert.ErrorAs(t, err, &appErr)
				assert.True(t, appErr.IsType(errors.ErrorTypeNotFound))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			service, repo := setupTaskServiceWithData(t, tt.setupTasks, tt.setupEntries)
			defer repo.Close()
			ctx := context.Background()

			// Use actual ID from created task if we have setup tasks
			actualTaskID := tt.taskID
			if len(tt.setupTasks) > 0 && tt.taskID == 1 {
				actualTaskID = tt.setupTasks[0].ID
				// Update entry TaskID to match actual task ID
				for _, entry := range tt.setupEntries {
					if entry.TaskID == 1 {
						entry.TaskID = actualTaskID
					}
				}
			}

			// Act
			err := service.DeleteTaskWithEntries(ctx, actualTaskID)

			// Assert
			if tt.errorAssertion != nil {
				tt.errorAssertion(t, err)
			} else {
				require.NoError(t, err)
				
				// Verify task is deleted
				_, err := service.GetTask(ctx, actualTaskID)
				assert.Error(t, err)
				var appErr *errors.AppError
				assert.ErrorAs(t, err, &appErr)
				assert.True(t, appErr.IsType(errors.ErrorTypeNotFound))
			}
		})
	}
}

func TestTaskService_StartNewTask(t *testing.T) {
	tests := []struct {
		name           string
		taskName       string
		setupTasks     []*domain.Task
		errorAssertion func(t *testing.T, err error)
	}{
		{
			name:     "should start new task",
			taskName: "New Task",
		},
		{
			name:     "should start existing task",
			taskName: "Existing Task",
			setupTasks: []*domain.Task{
				{TaskName: "Existing Task"},
			},
		},
		{
			name:     "should return validation error for empty name",
			taskName: "",
			errorAssertion: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "name")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			service, repo := setupTaskServiceWithData(t, tt.setupTasks, nil)
			defer repo.Close()
			ctx := context.Background()

			// Act
			result, err := service.StartNewTask(ctx, tt.taskName)

			// Assert
			if tt.errorAssertion != nil {
				tt.errorAssertion(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.NotNil(t, result.Task)
				require.NotNil(t, result.TimeEntry)
				
				assert.Equal(t, tt.taskName, result.Task.TaskName)
				assert.Equal(t, result.Task.ID, result.TimeEntry.TaskID)
				assert.Nil(t, result.TimeEntry.EndTime)
				assert.NotEmpty(t, result.Duration)
				assert.Contains(t, result.Duration, "running")
			}
		})
	}
}

func TestTaskService_ResumeTask(t *testing.T) {
	tests := []struct {
		name           string
		taskID         int64
		setupTasks     []*domain.Task
		errorAssertion func(t *testing.T, err error)
	}{
		{
			name:   "should resume existing task",
			taskID: 1,
			setupTasks: []*domain.Task{
				{TaskName: "Task to Resume"},
			},
		},
		{
			name:   "should return not found error for non-existent task",
			taskID: 999,
			errorAssertion: func(t *testing.T, err error) {
				assert.Error(t, err)
				var appErr *errors.AppError
				assert.ErrorAs(t, err, &appErr)
				assert.True(t, appErr.IsType(errors.ErrorTypeNotFound))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			service, repo := setupTaskServiceWithData(t, tt.setupTasks, nil)
			defer repo.Close()
			ctx := context.Background()

			// Use actual ID from created task if we have setup tasks
			actualTaskID := tt.taskID
			if len(tt.setupTasks) > 0 && tt.taskID == 1 {
				actualTaskID = tt.setupTasks[0].ID
			}

			// Act
			result, err := service.ResumeTask(ctx, actualTaskID)

			// Assert
			if tt.errorAssertion != nil {
				tt.errorAssertion(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.NotNil(t, result.Task)
				require.NotNil(t, result.TimeEntry)
				
				assert.Equal(t, actualTaskID, result.Task.ID)
				assert.Equal(t, actualTaskID, result.TimeEntry.TaskID)
				assert.Nil(t, result.TimeEntry.EndTime)
				assert.NotEmpty(t, result.Duration)
			}
		})
	}
}

func TestTaskService_GetCurrentSession(t *testing.T) {
	tests := []struct {
		name            string
		setupTasks      []*domain.Task
		setupEntries    []*domain.TimeEntry
		expectedSession bool
		errorAssertion  func(t *testing.T, err error)
	}{
		{
			name:            "should return nil when no running tasks",
			setupTasks:      []*domain.Task{},
			setupEntries:    []*domain.TimeEntry{},
			expectedSession: false,
		},
		{
			name: "should return current session when task is running",
			setupTasks: []*domain.Task{
				{TaskName: "Running Task"},
			},
			setupEntries: []*domain.TimeEntry{
				{TaskID: 1, StartTime: time.Now().Add(-30 * time.Minute), EndTime: nil},
			},
			expectedSession: true,
		},
		{
			name: "should return nil when all tasks are completed",
			setupTasks: []*domain.Task{
				{TaskName: "Completed Task"},
			},
			setupEntries: []*domain.TimeEntry{
				{TaskID: 1, StartTime: time.Now().Add(-1 * time.Hour), EndTime: timePtr(time.Now())},
			},
			expectedSession: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			service, repo := setupTaskServiceWithData(t, tt.setupTasks, tt.setupEntries)
			defer repo.Close()
			ctx := context.Background()

			// Act
			result, err := service.GetCurrentSession(ctx)

			// Assert
			if tt.errorAssertion != nil {
				tt.errorAssertion(t, err)
			} else {
				require.NoError(t, err)
				
				if tt.expectedSession {
					require.NotNil(t, result)
					require.NotNil(t, result.Task)
					require.NotNil(t, result.TimeEntry)
					assert.Nil(t, result.TimeEntry.EndTime)
					assert.NotEmpty(t, result.Duration)
				} else {
					assert.Nil(t, result)
				}
			}
		})
	}
}

func TestTaskService_CreateTaskSession(t *testing.T) {
	tests := []struct {
		name      string
		task      *domain.Task
		timeEntry *domain.TimeEntry
	}{
		{
			name: "should create session for running task",
			task: &domain.Task{ID: 1, TaskName: "Test Task"},
			timeEntry: &domain.TimeEntry{
				ID:        1,
				TaskID:    1,
				StartTime: time.Now().Add(-30 * time.Minute),
				EndTime:   nil,
			},
		},
		{
			name: "should create session for completed task",
			task: &domain.Task{ID: 2, TaskName: "Completed Task"},
			timeEntry: &domain.TimeEntry{
				ID:        2,
				TaskID:    2,
				StartTime: time.Now().Add(-1 * time.Hour),
				EndTime:   timePtr(time.Now()),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			service := setupTaskService(t)

			// Act
			result := service.CreateTaskSession(tt.task, tt.timeEntry)

			// Assert
			require.NotNil(t, result)
			assert.Equal(t, tt.task, result.Task)
			assert.Equal(t, tt.timeEntry, result.TimeEntry)
			assert.NotEmpty(t, result.Duration)
			
			if tt.timeEntry.EndTime == nil {
				assert.Contains(t, result.Duration, "running")
			}
		})
	}
}

func TestTaskService_StopAllRunningTasks(t *testing.T) {
	tests := []struct {
		name             string
		setupTasks       []*domain.Task
		setupEntries     []*domain.TimeEntry
		expectedStopped  int
	}{
		{
			name:            "should return empty list when no running tasks",
			setupTasks:      []*domain.Task{},
			setupEntries:    []*domain.TimeEntry{},
			expectedStopped: 0,
		},
		{
			name: "should stop all running tasks",
			setupTasks: []*domain.Task{
				{TaskName: "Running Task 1"},
				{TaskName: "Running Task 2"},
			},
			setupEntries: []*domain.TimeEntry{
				{TaskID: 1, StartTime: time.Now().Add(-1 * time.Hour), EndTime: nil},
				{TaskID: 2, StartTime: time.Now().Add(-30 * time.Minute), EndTime: nil},
			},
			expectedStopped: 2,
		},
		{
			name: "should not affect completed tasks",
			setupTasks: []*domain.Task{
				{TaskName: "Completed Task"},
				{TaskName: "Running Task"},
			},
			setupEntries: []*domain.TimeEntry{
				{TaskID: 1, StartTime: time.Now().Add(-2 * time.Hour), EndTime: timePtr(time.Now().Add(-1 * time.Hour))},
				{TaskID: 2, StartTime: time.Now().Add(-30 * time.Minute), EndTime: nil},
			},
			expectedStopped: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			service, repo := setupTaskServiceWithData(t, tt.setupTasks, tt.setupEntries)
			defer repo.Close()
			ctx := context.Background()

			// Act
			stoppedEntries, err := service.StopAllRunningTasks(ctx)

			// Assert
			require.NoError(t, err)
			assert.Len(t, stoppedEntries, tt.expectedStopped)
			
			// Verify all returned entries have EndTime set
			for _, entry := range stoppedEntries {
				assert.NotNil(t, entry.EndTime)
				assert.WithinDuration(t, time.Now(), *entry.EndTime, 5*time.Second)
			}
			
			// Verify no running tasks remain
			currentSession, err := service.GetCurrentSession(ctx)
			require.NoError(t, err)
			assert.Nil(t, currentSession)
		})
	}
}

// Helper functions
func setupTaskService(t *testing.T) TaskService {
	repo, err := sqlite.New(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { repo.Close() })
	
	timeService := NewTimeService(repo)
	return NewTaskService(repo, timeService)
}

func setupTaskServiceWithData(t *testing.T, tasks []*domain.Task, entries []*domain.TimeEntry) (TaskService, sqlite.Repository) {
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
	
	timeService := NewTimeService(repo)
	service := NewTaskService(repo, timeService)
	return service, repo
}