package services

import (
	"context"
	"testing"
	"time"
	"time-tracker/internal/domain"
	"time-tracker/internal/repository/sqlite"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchService_SearchTasks(t *testing.T) {
	tests := []struct {
		name           string
		criteria       SearchCriteria
		setupTasks     []*domain.Task
		setupEntries   []*domain.TimeEntry
		expectedCount  int
		errorAssertion func(t *testing.T, err error)
	}{
		{
			name:     "should return all tasks when no criteria specified",
			criteria: SearchCriteria{},
			setupTasks: []*domain.Task{
				{TaskName: "Task 1"},
				{TaskName: "Task 2"},
			},
			setupEntries: []*domain.TimeEntry{
				{TaskID: 1, StartTime: time.Now().Add(-1 * time.Hour), EndTime: timePtr(time.Now())},
				{TaskID: 2, StartTime: time.Now().Add(-30 * time.Minute), EndTime: nil},
			},
			expectedCount: 2,
		},
		{
			name: "should filter tasks by text filter",
			criteria: SearchCriteria{
				TextFilter: "Design",
			},
			setupTasks: []*domain.Task{
				{TaskName: "Design Frontend"},
				{TaskName: "Backend API"},
				{TaskName: "Design Database"},
			},
			setupEntries: []*domain.TimeEntry{
				{TaskID: 1, StartTime: time.Now().Add(-2 * time.Hour), EndTime: timePtr(time.Now().Add(-1 * time.Hour))},
				{TaskID: 2, StartTime: time.Now().Add(-1 * time.Hour), EndTime: timePtr(time.Now())},
				{TaskID: 3, StartTime: time.Now().Add(-30 * time.Minute), EndTime: nil},
			},
			expectedCount: 2, // Only tasks containing "Design"
		},
		{
			name: "should filter tasks by time range",
			criteria: SearchCriteria{
				TimeRange: &TimeRange{
					Start: time.Now().Add(-1 * time.Hour),
					End:   time.Now(),
				},
			},
			setupTasks: []*domain.Task{
				{TaskName: "Recent Task"},
				{TaskName: "Old Task"},
			},
			setupEntries: []*domain.TimeEntry{
				{TaskID: 1, StartTime: time.Now().Add(-30 * time.Minute), EndTime: nil}, // Recent
				{TaskID: 2, StartTime: time.Now().Add(-5 * time.Hour), EndTime: timePtr(time.Now().Add(-4 * time.Hour))}, // Old
			},
			expectedCount: 1, // Only recent task
		},
		{
			name: "should filter by specific task ID",
			criteria: SearchCriteria{
				TaskID: func() *int64 { id := int64(1); return &id }(),
			},
			setupTasks: []*domain.Task{
				{TaskName: "Target Task"},
				{TaskName: "Other Task"},
			},
			setupEntries: []*domain.TimeEntry{
				{TaskID: 1, StartTime: time.Now().Add(-1 * time.Hour), EndTime: timePtr(time.Now())},
				{TaskID: 2, StartTime: time.Now().Add(-30 * time.Minute), EndTime: nil},
			},
			expectedCount: 1,
		},
		{
			name: "should filter for running tasks only",
			criteria: SearchCriteria{
				RunningOnly: true,
			},
			setupTasks: []*domain.Task{
				{TaskName: "Running Task"},
				{TaskName: "Completed Task"},
			},
			setupEntries: []*domain.TimeEntry{
				{TaskID: 1, StartTime: time.Now().Add(-30 * time.Minute), EndTime: nil}, // Running
				{TaskID: 2, StartTime: time.Now().Add(-2 * time.Hour), EndTime: timePtr(time.Now().Add(-1 * time.Hour))}, // Completed
			},
			expectedCount: 1, // Only running task
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			service, repo := setupSearchServiceWithData(t, tt.setupTasks, tt.setupEntries)
			defer repo.Close()
			ctx := context.Background()

			// Update criteria with actual task IDs
			actualCriteria := tt.criteria
			if actualCriteria.TaskID != nil && *actualCriteria.TaskID == 1 && len(tt.setupTasks) > 0 {
				actualCriteria.TaskID = &tt.setupTasks[0].ID
			}

			// Act
			result, err := service.SearchTasks(ctx, actualCriteria)

			// Assert
			if tt.errorAssertion != nil {
				tt.errorAssertion(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Len(t, result, tt.expectedCount)
				
				// Verify task activities have required fields
				for _, activity := range result {
					assert.NotNil(t, activity.Task)
					assert.NotEmpty(t, activity.TotalTime)
					assert.GreaterOrEqual(t, activity.SessionCount, 0)
				}
			}
		})
	}
}

func TestSearchService_SearchTimeEntries(t *testing.T) {
	tests := []struct {
		name           string
		criteria       SearchCriteria
		setupTasks     []*domain.Task
		setupEntries   []*domain.TimeEntry
		expectedCount  int
		errorAssertion func(t *testing.T, err error)
	}{
		{
			name:     "should return all time entries when no criteria specified",
			criteria: SearchCriteria{},
			setupTasks: []*domain.Task{
				{TaskName: "Task 1"},
				{TaskName: "Task 2"},
			},
			setupEntries: []*domain.TimeEntry{
				{TaskID: 1, StartTime: time.Now().Add(-2 * time.Hour), EndTime: timePtr(time.Now().Add(-1 * time.Hour))},
				{TaskID: 2, StartTime: time.Now().Add(-30 * time.Minute), EndTime: nil},
			},
			expectedCount: 2,
		},
		{
			name: "should filter time entries by time range",
			criteria: SearchCriteria{
				TimeRange: &TimeRange{
					Start: time.Now().Add(-1 * time.Hour),
					End:   time.Now(),
				},
			},
			setupTasks: []*domain.Task{
				{TaskName: "Task 1"},
				{TaskName: "Task 2"},
			},
			setupEntries: []*domain.TimeEntry{
				{TaskID: 1, StartTime: time.Now().Add(-30 * time.Minute), EndTime: nil}, // Recent
				{TaskID: 2, StartTime: time.Now().Add(-5 * time.Hour), EndTime: timePtr(time.Now().Add(-4 * time.Hour))}, // Old
			},
			expectedCount: 1, // Only recent entry
		},
		{
			name: "should filter time entries by running only",
			criteria: SearchCriteria{
				RunningOnly: true,
			},
			setupTasks: []*domain.Task{
				{TaskName: "Task 1"},
				{TaskName: "Task 2"},
			},
			setupEntries: []*domain.TimeEntry{
				{TaskID: 1, StartTime: time.Now().Add(-30 * time.Minute), EndTime: nil}, // Running
				{TaskID: 2, StartTime: time.Now().Add(-2 * time.Hour), EndTime: timePtr(time.Now().Add(-1 * time.Hour))}, // Completed
			},
			expectedCount: 1, // Only running entry
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			service, repo := setupSearchServiceWithData(t, tt.setupTasks, tt.setupEntries)
			defer repo.Close()
			ctx := context.Background()

			// Act
			result, err := service.SearchTimeEntries(ctx, tt.criteria)

			// Assert
			if tt.errorAssertion != nil {
				tt.errorAssertion(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Len(t, result, tt.expectedCount)
				
				// Verify time entries with tasks have required fields
				for _, entryWithTask := range result {
					assert.NotNil(t, entryWithTask.TimeEntry)
					assert.NotNil(t, entryWithTask.Task)
					assert.NotEmpty(t, entryWithTask.Duration)
				}
			}
		})
	}
}

func TestSearchService_FilterTasksByTime(t *testing.T) {
	tests := []struct {
		name           string
		tasks          []*TaskActivity
		timeRange      *TimeRange
		expectedCount  int
	}{
		{
			name: "should return all tasks when no time range specified",
			tasks: []*TaskActivity{
				{Task: &domain.Task{ID: 1, TaskName: "Task 1"}, LastWorked: time.Now().Add(-30 * time.Minute)},
				{Task: &domain.Task{ID: 2, TaskName: "Task 2"}, LastWorked: time.Now().Add(-2 * time.Hour)},
			},
			timeRange:     nil,
			expectedCount: 2,
		},
		{
			name: "should filter tasks by last worked time",
			tasks: []*TaskActivity{
				{Task: &domain.Task{ID: 1, TaskName: "Recent Task"}, LastWorked: time.Now().Add(-30 * time.Minute)},
				{Task: &domain.Task{ID: 2, TaskName: "Old Task"}, LastWorked: time.Now().Add(-5 * time.Hour)},
			},
			timeRange: &TimeRange{
				Start: time.Now().Add(-1 * time.Hour),
				End:   time.Now(),
			},
			expectedCount: 1, // Only recent task
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			service := setupSearchService(t)

			// Act
			result := service.FilterTasksByTime(tt.tasks, tt.timeRange)

			// Assert
			assert.Len(t, result, tt.expectedCount)
		})
	}
}

func TestSearchService_SortTasks(t *testing.T) {
	baseTasks := []*TaskActivity{
		{Task: &domain.Task{ID: 1, TaskName: "Zebra Task"}, LastWorked: time.Now().Add(-1 * time.Hour), TotalTime: "2h 30m"},
		{Task: &domain.Task{ID: 2, TaskName: "Alpha Task"}, LastWorked: time.Now().Add(-30 * time.Minute), TotalTime: "1h 15m"},
		{Task: &domain.Task{ID: 3, TaskName: "Beta Task"}, LastWorked: time.Now().Add(-2 * time.Hour), TotalTime: "3h 45m"},
	}

	tests := []struct {
		name      string
		order     SortOrder
		expected  []string // Expected task names in order
	}{
		{
			name:     "should sort by recent first (default)",
			order:    SortByRecentFirst,
			expected: []string{"Alpha Task", "Zebra Task", "Beta Task"},
		},
		{
			name:     "should sort by oldest first",
			order:    SortByOldestFirst,
			expected: []string{"Beta Task", "Zebra Task", "Alpha Task"},
		},
		{
			name:     "should sort by name alphabetically",
			order:    SortByName,
			expected: []string{"Alpha Task", "Beta Task", "Zebra Task"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			service := setupSearchService(t)
			// Make a copy to avoid modifying the original
			tasks := make([]*TaskActivity, len(baseTasks))
			copy(tasks, baseTasks)

			// Act
			result := service.SortTasks(tasks, tt.order)

			// Assert
			require.Len(t, result, len(tt.expected))
			for i, expectedName := range tt.expected {
				assert.Equal(t, expectedName, result[i].Task.TaskName)
			}
		})
	}
}

func TestSearchService_SortTimeEntries(t *testing.T) {
	baseEntries := []*TimeEntryWithTask{
		{
			TimeEntry: &domain.TimeEntry{ID: 1, TaskID: 1, StartTime: time.Now().Add(-1 * time.Hour)},
			Task:      &domain.Task{ID: 1, TaskName: "Zebra Task"},
			Duration:  "1h 0m",
		},
		{
			TimeEntry: &domain.TimeEntry{ID: 2, TaskID: 2, StartTime: time.Now().Add(-30 * time.Minute)},
			Task:      &domain.Task{ID: 2, TaskName: "Alpha Task"},
			Duration:  "30m",
		},
		{
			TimeEntry: &domain.TimeEntry{ID: 3, TaskID: 3, StartTime: time.Now().Add(-2 * time.Hour)},
			Task:      &domain.Task{ID: 3, TaskName: "Beta Task"},
			Duration:  "2h 0m",
		},
	}

	tests := []struct {
		name      string
		order     SortOrder
		expected  []string // Expected task names in order
	}{
		{
			name:     "should sort by recent first (default)",
			order:    SortByRecentFirst,
			expected: []string{"Alpha Task", "Zebra Task", "Beta Task"},
		},
		{
			name:     "should sort by oldest first",
			order:    SortByOldestFirst,
			expected: []string{"Beta Task", "Zebra Task", "Alpha Task"},
		},
		{
			name:     "should sort by name alphabetically",
			order:    SortByName,
			expected: []string{"Alpha Task", "Beta Task", "Zebra Task"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			service := setupSearchService(t)
			// Make a copy to avoid modifying the original
			entries := make([]*TimeEntryWithTask, len(baseEntries))
			copy(entries, baseEntries)

			// Act
			result := service.SortTimeEntries(entries, tt.order)

			// Assert
			require.Len(t, result, len(tt.expected))
			for i, expectedName := range tt.expected {
				assert.Equal(t, expectedName, result[i].Task.TaskName)
			}
		})
	}
}

func TestSearchService_GetRecentTasks(t *testing.T) {
	tests := []struct {
		name           string
		timeRange      *TimeRange
		limit          int
		setupTasks     []*domain.Task
		setupEntries   []*domain.TimeEntry
		expectedCount  int
		errorAssertion func(t *testing.T, err error)
	}{
		{
			name:      "should return recent tasks with limit",
			timeRange: nil, // No time filter
			limit:     2,
			setupTasks: []*domain.Task{
				{TaskName: "Task 1"},
				{TaskName: "Task 2"},
				{TaskName: "Task 3"},
			},
			setupEntries: []*domain.TimeEntry{
				{TaskID: 1, StartTime: time.Now().Add(-3 * time.Hour), EndTime: timePtr(time.Now().Add(-2 * time.Hour))},
				{TaskID: 2, StartTime: time.Now().Add(-1 * time.Hour), EndTime: timePtr(time.Now())},
				{TaskID: 3, StartTime: time.Now().Add(-30 * time.Minute), EndTime: nil},
			},
			expectedCount: 2, // Limit to 2 most recent
		},
		{
			name: "should filter by time range and limit",
			timeRange: &TimeRange{
				Start: time.Now().Add(-2 * time.Hour),
				End:   time.Now(),
			},
			limit: 5,
			setupTasks: []*domain.Task{
				{TaskName: "Recent Task"},
				{TaskName: "Old Task"},
			},
			setupEntries: []*domain.TimeEntry{
				{TaskID: 1, StartTime: time.Now().Add(-30 * time.Minute), EndTime: nil}, // Recent
				{TaskID: 2, StartTime: time.Now().Add(-5 * time.Hour), EndTime: timePtr(time.Now().Add(-4 * time.Hour))}, // Old
			},
			expectedCount: 1, // Only recent task within time range
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			service, repo := setupSearchServiceWithData(t, tt.setupTasks, tt.setupEntries)
			defer repo.Close()
			ctx := context.Background()

			// Act
			result, err := service.GetRecentTasks(ctx, tt.timeRange, tt.limit)

			// Assert
			if tt.errorAssertion != nil {
				tt.errorAssertion(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.LessOrEqual(t, len(result), tt.limit)
				assert.LessOrEqual(t, len(result), tt.expectedCount)
				
				// Verify tasks are sorted by most recent first
				for i := 1; i < len(result); i++ {
					assert.True(t, result[i-1].LastWorked.After(result[i].LastWorked) || result[i-1].LastWorked.Equal(result[i].LastWorked))
				}
			}
		})
	}
}

func TestSearchService_FindTasksWithActivity(t *testing.T) {
	tests := []struct {
		name           string
		timeRange      *TimeRange
		setupTasks     []*domain.Task
		setupEntries   []*domain.TimeEntry
		expectedCount  int
		errorAssertion func(t *testing.T, err error)
	}{
		{
			name:      "should return all tasks with activity when no time range",
			timeRange: nil,
			setupTasks: []*domain.Task{
				{TaskName: "Active Task"},
				{TaskName: "Inactive Task"},
			},
			setupEntries: []*domain.TimeEntry{
				{TaskID: 1, StartTime: time.Now().Add(-1 * time.Hour), EndTime: timePtr(time.Now())},
				// Task 2 has no entries
			},
			expectedCount: 1, // Only tasks with activity
		},
		{
			name: "should filter tasks by activity within time range",
			timeRange: &TimeRange{
				Start: time.Now().Add(-1 * time.Hour),
				End:   time.Now(),
			},
			setupTasks: []*domain.Task{
				{TaskName: "Recent Task"},
				{TaskName: "Old Task"},
			},
			setupEntries: []*domain.TimeEntry{
				{TaskID: 1, StartTime: time.Now().Add(-30 * time.Minute), EndTime: nil}, // Recent
				{TaskID: 2, StartTime: time.Now().Add(-5 * time.Hour), EndTime: timePtr(time.Now().Add(-4 * time.Hour))}, // Old
			},
			expectedCount: 1, // Only tasks with recent activity
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			service, repo := setupSearchServiceWithData(t, tt.setupTasks, tt.setupEntries)
			defer repo.Close()
			ctx := context.Background()

			// Act
			result, err := service.FindTasksWithActivity(ctx, tt.timeRange)

			// Assert
			if tt.errorAssertion != nil {
				tt.errorAssertion(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Len(t, result, tt.expectedCount)
				
				// Verify all returned tasks have activity
				for _, activity := range result {
					assert.Greater(t, activity.SessionCount, 0)
					assert.NotEqual(t, "0m", activity.TotalTime)
				}
			}
		})
	}
}

// Helper functions
func setupSearchService(t *testing.T) SearchService {
	repo, err := sqlite.New(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { repo.Close() })
	
	timeService := NewTimeService(repo)
	taskService := NewTaskService(repo, timeService)
	return NewSearchService(repo, timeService, taskService)
}

func setupSearchServiceWithData(t *testing.T, tasks []*domain.Task, entries []*domain.TimeEntry) (SearchService, sqlite.Repository) {
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
	taskService := NewTaskService(repo, timeService)
	service := NewSearchService(repo, timeService, taskService)
	return service, repo
}