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

func TestReportingService_GetTaskSummary(t *testing.T) {
	tests := []struct {
		name           string
		taskID         int64
		setupTasks     []*domain.Task
		setupEntries   []*domain.TimeEntry
		expectedSummary func(t *testing.T, summary *TaskSummary)
		errorAssertion func(t *testing.T, err error)
	}{
		{
			name:   "should return summary for task with completed entries",
			taskID: 1,
			setupTasks: []*domain.Task{
				{TaskName: "Test Task"},
			},
			setupEntries: []*domain.TimeEntry{
				{TaskID: 1, StartTime: time.Now().Add(-2 * time.Hour), EndTime: timePtr(time.Now().Add(-1 * time.Hour))},
				{TaskID: 1, StartTime: time.Now().Add(-30 * time.Minute), EndTime: timePtr(time.Now())},
			},
			expectedSummary: func(t *testing.T, summary *TaskSummary) {
				assert.Equal(t, "Test Task", summary.Task.TaskName)
				assert.Len(t, summary.TimeEntries, 2)
				assert.Equal(t, 2, summary.SessionCount)
				assert.Equal(t, 0, summary.RunningCount)
				assert.False(t, summary.IsRunning)
				assert.Contains(t, summary.TotalTime, "h") // Should contain hours
			},
		},
		{
			name:   "should return summary for task with running entry",
			taskID: 1,
			setupTasks: []*domain.Task{
				{TaskName: "Running Task"},
			},
			setupEntries: []*domain.TimeEntry{
				{TaskID: 1, StartTime: time.Now().Add(-1 * time.Hour), EndTime: timePtr(time.Now())},
				{TaskID: 1, StartTime: time.Now().Add(-30 * time.Minute), EndTime: nil}, // Running
			},
			expectedSummary: func(t *testing.T, summary *TaskSummary) {
				assert.Equal(t, "Running Task", summary.Task.TaskName)
				assert.Len(t, summary.TimeEntries, 2)
				assert.Equal(t, 2, summary.SessionCount)
				assert.Equal(t, 1, summary.RunningCount)
				assert.True(t, summary.IsRunning)
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
			service, repo := setupReportingServiceWithData(t, tt.setupTasks, tt.setupEntries)
			defer repo.Close()
			ctx := context.Background()

			// Use actual ID from created task if we have setup tasks
			actualTaskID := tt.taskID
			if len(tt.setupTasks) > 0 && tt.taskID == 1 {
				actualTaskID = tt.setupTasks[0].ID
				// Update entry TaskIDs to match actual task ID
				for _, entry := range tt.setupEntries {
					if entry.TaskID == 1 {
						entry.TaskID = actualTaskID
					}
				}
			}

			// Act
			result, err := service.GetTaskSummary(ctx, actualTaskID)

			// Assert
			if tt.errorAssertion != nil {
				tt.errorAssertion(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				tt.expectedSummary(t, result)
			}
		})
	}
}

func TestReportingService_AnalyzeTaskActivity(t *testing.T) {
	tests := []struct {
		name               string
		entries            []*domain.TimeEntry
		expectedAnalysis   func(t *testing.T, analysis *ActivityAnalysis)
	}{
		{
			name:    "should return zero analysis for empty entries",
			entries: []*domain.TimeEntry{},
			expectedAnalysis: func(t *testing.T, analysis *ActivityAnalysis) {
				assert.Equal(t, time.Duration(0), analysis.TotalDuration)
				assert.Equal(t, time.Duration(0), analysis.AverageDuration)
				assert.Equal(t, time.Duration(0), analysis.LongestSession)
				assert.Equal(t, time.Duration(0), analysis.ShortestSession)
				assert.Equal(t, 0, analysis.SessionCount)
			},
		},
		{
			name: "should analyze completed entries",
			entries: []*domain.TimeEntry{
				{ID: 1, StartTime: time.Now().Add(-3 * time.Hour), EndTime: timePtr(time.Now().Add(-2 * time.Hour))}, // 1h
				{ID: 2, StartTime: time.Now().Add(-90 * time.Minute), EndTime: timePtr(time.Now().Add(-60 * time.Minute))}, // 30m
				{ID: 3, StartTime: time.Now().Add(-45 * time.Minute), EndTime: timePtr(time.Now().Add(-15 * time.Minute))}, // 30m
			},
			expectedAnalysis: func(t *testing.T, analysis *ActivityAnalysis) {
				assert.InDelta(t, 2*time.Hour, analysis.TotalDuration, float64(time.Millisecond))
				assert.InDelta(t, 40*time.Minute, analysis.AverageDuration, float64(time.Millisecond)) // 120min / 3 = 40min
				assert.InDelta(t, 1*time.Hour, analysis.LongestSession, float64(time.Millisecond))
				assert.InDelta(t, 30*time.Minute, analysis.ShortestSession, float64(time.Millisecond))
				assert.Equal(t, 3, analysis.SessionCount)
			},
		},
		{
			name: "should handle single entry",
			entries: []*domain.TimeEntry{
				{ID: 1, StartTime: time.Now().Add(-1 * time.Hour), EndTime: timePtr(time.Now())},
			},
			expectedAnalysis: func(t *testing.T, analysis *ActivityAnalysis) {
				assert.InDelta(t, 1*time.Hour, analysis.TotalDuration, float64(time.Millisecond))
				assert.InDelta(t, 1*time.Hour, analysis.AverageDuration, float64(time.Millisecond))
				assert.InDelta(t, 1*time.Hour, analysis.LongestSession, float64(time.Millisecond))
				assert.InDelta(t, 1*time.Hour, analysis.ShortestSession, float64(time.Millisecond))
				assert.Equal(t, 1, analysis.SessionCount)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			service := setupReportingService(t)

			// Act
			result := service.AnalyzeTaskActivity(tt.entries)

			// Assert
			require.NotNil(t, result)
			tt.expectedAnalysis(t, result)
		})
	}
}

func TestReportingService_CalculateTaskStatistics(t *testing.T) {
	tests := []struct {
		name           string
		taskID         int64
		setupTasks     []*domain.Task
		setupEntries   []*domain.TimeEntry
		errorAssertion func(t *testing.T, err error)
	}{
		{
			name:   "should calculate statistics for existing task",
			taskID: 1,
			setupTasks: []*domain.Task{
				{TaskName: "Statistics Task"},
			},
			setupEntries: []*domain.TimeEntry{
				{TaskID: 1, StartTime: time.Now().Add(-2 * time.Hour), EndTime: timePtr(time.Now().Add(-1 * time.Hour))},
				{TaskID: 1, StartTime: time.Now().Add(-30 * time.Minute), EndTime: nil}, // Running
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
			service, repo := setupReportingServiceWithData(t, tt.setupTasks, tt.setupEntries)
			defer repo.Close()
			ctx := context.Background()

			// Use actual ID from created task if we have setup tasks
			actualTaskID := tt.taskID
			if len(tt.setupTasks) > 0 && tt.taskID == 1 {
				actualTaskID = tt.setupTasks[0].ID
				// Update entry TaskIDs to match actual task ID
				for _, entry := range tt.setupEntries {
					if entry.TaskID == 1 {
						entry.TaskID = actualTaskID
					}
				}
			}

			// Act
			result, err := service.CalculateTaskStatistics(ctx, actualTaskID)

			// Assert
			if tt.errorAssertion != nil {
				tt.errorAssertion(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, actualTaskID, result.Task.ID)
				assert.GreaterOrEqual(t, result.SessionCount, 0)
			}
		})
	}
}

func TestReportingService_GetDashboardData(t *testing.T) {
	tests := []struct {
		name           string
		timeRange      string
		setupTasks     []*domain.Task
		setupEntries   []*domain.TimeEntry
		expectedData   func(t *testing.T, data *DashboardData)
		errorAssertion func(t *testing.T, err error)
	}{
		{
			name:      "should return dashboard data with running task",
			timeRange: "1d",
			setupTasks: []*domain.Task{
				{TaskName: "Running Task"},
				{TaskName: "Completed Task"},
			},
			setupEntries: []*domain.TimeEntry{
				{TaskID: 1, StartTime: time.Now().Add(-30 * time.Minute), EndTime: nil}, // Running
				{TaskID: 2, StartTime: time.Now().Add(-2 * time.Hour), EndTime: timePtr(time.Now().Add(-1 * time.Hour))}, // Completed
			},
			expectedData: func(t *testing.T, data *DashboardData) {
				require.NotNil(t, data.RunningTask)
				assert.Equal(t, "Running Task", data.RunningTask.Task.TaskName)
				assert.Nil(t, data.RunningTask.TimeEntry.EndTime)
				
				assert.GreaterOrEqual(t, len(data.RecentTasks), 0)
				require.NotNil(t, data.TodayStats)
				assert.GreaterOrEqual(t, data.TodayStats.SessionCount, 0)
			},
		},
		{
			name:      "should return dashboard data without running task",
			timeRange: "1d",
			setupTasks: []*domain.Task{
				{TaskName: "Completed Task"},
			},
			setupEntries: []*domain.TimeEntry{
				{TaskID: 1, StartTime: time.Now().Add(-2 * time.Hour), EndTime: timePtr(time.Now().Add(-1 * time.Hour))},
			},
			expectedData: func(t *testing.T, data *DashboardData) {
				assert.Nil(t, data.RunningTask)
				assert.GreaterOrEqual(t, len(data.RecentTasks), 0)
				require.NotNil(t, data.TodayStats)
			},
		},
		{
			name:      "should return error for invalid time range",
			timeRange: "invalid",
			errorAssertion: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			service, repo := setupReportingServiceWithData(t, tt.setupTasks, tt.setupEntries)
			defer repo.Close()
			ctx := context.Background()

			// Act
			result, err := service.GetDashboardData(ctx, tt.timeRange)

			// Assert
			if tt.errorAssertion != nil {
				tt.errorAssertion(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				tt.expectedData(t, result)
			}
		})
	}
}

func TestReportingService_GetDayStatistics(t *testing.T) {
	tests := []struct {
		name           string
		date           time.Time
		setupTasks     []*domain.Task
		setupEntries   []*domain.TimeEntry
		expectedStats  func(t *testing.T, stats *DayStatistics)
	}{
		{
			name: "should calculate statistics for specific day",
			date: time.Now(),
			setupTasks: []*domain.Task{
				{TaskName: "Today Task 1"},
				{TaskName: "Today Task 2"},
			},
			setupEntries: []*domain.TimeEntry{
				{TaskID: 1, StartTime: time.Now().Add(-2 * time.Hour), EndTime: timePtr(time.Now().Add(-1 * time.Hour))}, // Completed today
				{TaskID: 2, StartTime: time.Now().Add(-30 * time.Minute), EndTime: nil}, // Running today
				{TaskID: 1, StartTime: time.Now().Add(-25 * time.Hour), EndTime: timePtr(time.Now().Add(-24 * time.Hour))}, // Yesterday
			},
			expectedStats: func(t *testing.T, stats *DayStatistics) {
				assert.NotEmpty(t, stats.TotalTime)
				assert.GreaterOrEqual(t, stats.TaskCount, 1)
				assert.GreaterOrEqual(t, stats.SessionCount, 2) // 2 sessions today
				assert.GreaterOrEqual(t, stats.CompletedCount, 1) // 1 completed today
			},
		},
		{
			name: "should return zero stats for day with no activity",
			date: time.Now().Add(-48 * time.Hour), // 2 days ago
			setupTasks: []*domain.Task{
				{TaskName: "Today Task"},
			},
			setupEntries: []*domain.TimeEntry{
				{TaskID: 1, StartTime: time.Now().Add(-1 * time.Hour), EndTime: timePtr(time.Now())}, // Today
			},
			expectedStats: func(t *testing.T, stats *DayStatistics) {
				assert.Equal(t, "0m", stats.TotalTime)
				assert.Equal(t, 0, stats.TaskCount)
				assert.Equal(t, 0, stats.SessionCount)
				assert.Equal(t, 0, stats.CompletedCount)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			service, repo := setupReportingServiceWithData(t, tt.setupTasks, tt.setupEntries)
			defer repo.Close()
			ctx := context.Background()

			// Act
			result, err := service.GetDayStatistics(ctx, tt.date)

			// Assert
			require.NoError(t, err)
			require.NotNil(t, result)
			tt.expectedStats(t, result)
		})
	}
}

func TestReportingService_GetTodayStatistics(t *testing.T) {
	tests := []struct {
		name           string
		setupTasks     []*domain.Task
		setupEntries   []*domain.TimeEntry
		expectedStats  func(t *testing.T, stats *DayStatistics)
	}{
		{
			name: "should calculate today's statistics",
			setupTasks: []*domain.Task{
				{TaskName: "Today Task"},
			},
			setupEntries: []*domain.TimeEntry{
				{TaskID: 1, StartTime: time.Now().Add(-1 * time.Hour), EndTime: timePtr(time.Now())},
				{TaskID: 1, StartTime: time.Now().Add(-25 * time.Hour), EndTime: timePtr(time.Now().Add(-24 * time.Hour))}, // Yesterday
			},
			expectedStats: func(t *testing.T, stats *DayStatistics) {
				assert.NotEmpty(t, stats.TotalTime)
				assert.GreaterOrEqual(t, stats.SessionCount, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			service, repo := setupReportingServiceWithData(t, tt.setupTasks, tt.setupEntries)
			defer repo.Close()
			ctx := context.Background()

			// Act
			result, err := service.GetTodayStatistics(ctx)

			// Assert
			require.NoError(t, err)
			require.NotNil(t, result)
			tt.expectedStats(t, result)
		})
	}
}

func TestReportingService_AggregateTaskData(t *testing.T) {
	tests := []struct {
		name             string
		entries          []*domain.TimeEntry
		expectedTaskIDs  []int64
		expectedCounts   map[int64]int
	}{
		{
			name:            "should return empty map for empty entries",
			entries:         []*domain.TimeEntry{},
			expectedTaskIDs: []int64{},
			expectedCounts:  map[int64]int{},
		},
		{
			name: "should aggregate entries by task",
			entries: []*domain.TimeEntry{
				{ID: 1, TaskID: 1, StartTime: time.Now().Add(-2 * time.Hour), EndTime: timePtr(time.Now().Add(-1 * time.Hour))},
				{ID: 2, TaskID: 1, StartTime: time.Now().Add(-30 * time.Minute), EndTime: nil},
				{ID: 3, TaskID: 2, StartTime: time.Now().Add(-1 * time.Hour), EndTime: timePtr(time.Now())},
			},
			expectedTaskIDs: []int64{1, 2},
			expectedCounts:  map[int64]int{1: 2, 2: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			service := setupReportingService(t)

			// Act
			result := service.AggregateTaskData(tt.entries)

			// Assert
			assert.Len(t, result, len(tt.expectedTaskIDs))
			
			for taskID, expectedCount := range tt.expectedCounts {
				require.Contains(t, result, taskID)
				assert.Equal(t, expectedCount, result[taskID].SessionCount)
			}
		})
	}
}

func TestReportingService_CalculateTotalDuration(t *testing.T) {
	tests := []struct {
		name             string
		entries          []*domain.TimeEntry
		expectedDuration time.Duration
	}{
		{
			name:             "should return zero for empty entries",
			entries:          []*domain.TimeEntry{},
			expectedDuration: 0,
		},
		{
			name: "should calculate total duration for completed entries",
			entries: []*domain.TimeEntry{
				{ID: 1, StartTime: time.Now().Add(-2 * time.Hour), EndTime: timePtr(time.Now().Add(-1 * time.Hour))}, // 1h
				{ID: 2, StartTime: time.Now().Add(-90 * time.Minute), EndTime: timePtr(time.Now().Add(-60 * time.Minute))}, // 30m
			},
			expectedDuration: 90 * time.Minute, // 1h + 30m
		},
		{
			name: "should include running entries in total",
			entries: []*domain.TimeEntry{
				{ID: 1, StartTime: time.Now().Add(-1 * time.Hour), EndTime: timePtr(time.Now())}, // 1h
				{ID: 2, StartTime: time.Now().Add(-30 * time.Minute), EndTime: nil}, // 30m running
			},
			expectedDuration: 90 * time.Minute, // 1h + 30m (approximate)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			service := setupReportingService(t)

			// Act
			result := service.CalculateTotalDuration(tt.entries)

			// Assert
			if tt.expectedDuration == 0 {
				assert.Equal(t, tt.expectedDuration, result)
			} else {
				// Allow some tolerance for running entries
				tolerance := 5 * time.Minute
				assert.InDelta(t, tt.expectedDuration, result, float64(tolerance))
			}
		})
	}
}

func TestReportingService_FormatStatistics(t *testing.T) {
	tests := []struct {
		name          string
		analysis      *ActivityAnalysis
		expectedStats func(t *testing.T, stats *DayStatistics)
	}{
		{
			name: "should format activity analysis to day statistics",
			analysis: &ActivityAnalysis{
				TotalDuration:   2 * time.Hour,
				AverageDuration: 40 * time.Minute,
				SessionCount:    3,
			},
			expectedStats: func(t *testing.T, stats *DayStatistics) {
				assert.Equal(t, "2h 0m", stats.TotalTime)
				assert.Equal(t, 0, stats.TaskCount) // Not calculated from analysis
				assert.Equal(t, 3, stats.SessionCount)
				assert.Equal(t, 0, stats.CompletedCount) // Not calculated from analysis
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			service := setupReportingService(t)

			// Act
			result := service.FormatStatistics(tt.analysis)

			// Assert
			require.NotNil(t, result)
			tt.expectedStats(t, result)
		})
	}
}

// Helper functions
func setupReportingService(t *testing.T) ReportingService {
	repo, err := sqlite.New(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { repo.Close() })
	
	timeService := NewTimeService(repo)
	taskService := NewTaskService(repo, timeService)
	searchService := NewSearchService(repo, timeService, taskService)
	return NewReportingService(repo, timeService, taskService, searchService)
}

func setupReportingServiceWithData(t *testing.T, tasks []*domain.Task, entries []*domain.TimeEntry) (ReportingService, sqlite.Repository) {
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
	searchService := NewSearchService(repo, timeService, taskService)
	service := NewReportingService(repo, timeService, taskService, searchService)
	return service, repo
}