package services

import (
	"context"
	"time"
	"time-tracker/internal/domain"
	"time-tracker/internal/repository/sqlite"
)

const (
	// DefaultRecentTasksLimit is the default limit for recent tasks in dashboard
	DefaultRecentTasksLimit = 10
)

// reportingServiceImpl implements the ReportingService interface
type reportingServiceImpl struct {
	repo          sqlite.Repository
	timeService   TimeService
	taskService   TaskService
	searchService SearchService
	mapper        *domain.Mapper
}

// NewReportingService creates a new ReportingService instance
func NewReportingService(repo sqlite.Repository, timeService TimeService, taskService TaskService, searchService SearchService) ReportingService {
	return &reportingServiceImpl{
		repo:          repo,
		timeService:   timeService,
		taskService:   taskService,
		searchService: searchService,
		mapper:        domain.NewMapper(),
	}
}

// GetTaskSummary returns comprehensive summary for a specific task
func (r *reportingServiceImpl) GetTaskSummary(ctx context.Context, id int64) (*TaskSummary, error) {
	// Get the task
	task, err := r.taskService.GetTask(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get all time entries for this task
	searchOpts := sqlite.SearchOptions{
		TaskID: &id,
	}
	
	dbEntries, err := r.repo.SearchTimeEntries(ctx, searchOpts)
	if err != nil {
		return nil, err
	}

	// Convert to domain entries
	timeEntries := make([]*domain.TimeEntry, len(dbEntries))
	for i, dbEntry := range dbEntries {
		domainEntry := r.mapper.TimeEntry.FromDatabase(*dbEntry)
		timeEntries[i] = &domainEntry
	}

	// Calculate summary statistics
	sessionCount := len(timeEntries)
	runningCount := 0
	isRunning := false
	var firstEntry, lastEntry time.Time
	var totalDuration time.Duration

	for i, entry := range timeEntries {
		// Track first and last entry times
		if i == 0 || entry.StartTime.Before(firstEntry) {
			firstEntry = entry.StartTime
		}
		if i == 0 || entry.StartTime.After(lastEntry) {
			lastEntry = entry.StartTime
		}

		// Check if running
		if entry.EndTime == nil {
			runningCount++
			isRunning = true
			totalDuration += time.Since(entry.StartTime)
		} else {
			totalDuration += entry.EndTime.Sub(entry.StartTime)
		}
	}

	totalTime := r.timeService.FormatDuration(totalDuration)

	return &TaskSummary{
		Task:         task,
		TimeEntries:  timeEntries,
		TotalTime:    totalTime,
		SessionCount: sessionCount,
		RunningCount: runningCount,
		FirstEntry:   firstEntry,
		LastEntry:    lastEntry,
		IsRunning:    isRunning,
	}, nil
}

// AnalyzeTaskActivity analyzes time entries and returns detailed activity statistics
func (r *reportingServiceImpl) AnalyzeTaskActivity(entries []*domain.TimeEntry) *ActivityAnalysis {
	if len(entries) == 0 {
		return &ActivityAnalysis{
			TotalDuration:   0,
			AverageDuration: 0,
			LongestSession:  0,
			ShortestSession: 0,
			SessionCount:    0,
			ProductiveHours: []int{},
		}
	}

	var totalDuration time.Duration
	var longestSession time.Duration
	var shortestSession time.Duration
	productiveHoursMap := make(map[int]int)

	for i, entry := range entries {
		var duration time.Duration
		
		if entry.EndTime != nil {
			duration = entry.EndTime.Sub(entry.StartTime)
		} else {
			// Running entry
			duration = time.Since(entry.StartTime)
		}

		totalDuration += duration

		// Track longest and shortest sessions
		if i == 0 || duration > longestSession {
			longestSession = duration
		}
		if i == 0 || duration < shortestSession {
			shortestSession = duration
		}

		// Track productive hours
		hour := entry.StartTime.Hour()
		productiveHoursMap[hour]++
	}

	// Calculate average duration
	averageDuration := totalDuration / time.Duration(len(entries))

	// Convert productive hours map to sorted slice
	productiveHours := make([]int, 0, len(productiveHoursMap))
	for hour := range productiveHoursMap {
		productiveHours = append(productiveHours, hour)
	}

	return &ActivityAnalysis{
		TotalDuration:   totalDuration,
		AverageDuration: averageDuration,
		LongestSession:  longestSession,
		ShortestSession: shortestSession,
		SessionCount:    len(entries),
		ProductiveHours: productiveHours,
	}
}

// CalculateTaskStatistics calculates comprehensive statistics for a task
func (r *reportingServiceImpl) CalculateTaskStatistics(ctx context.Context, id int64) (*TaskSummary, error) {
	// This is essentially the same as GetTaskSummary for now
	return r.GetTaskSummary(ctx, id)
}

// GetDashboardData returns all data needed for a dashboard view
func (r *reportingServiceImpl) GetDashboardData(ctx context.Context, timeRange string) (*DashboardData, error) {
	// Parse time range
	timeRangeObj, err := r.timeService.ParseTimeRange(timeRange)
	if err != nil {
		return nil, err
	}

	// Get current running task
	runningTask, err := r.taskService.GetCurrentSession(ctx)
	if err != nil {
		return nil, err
	}

	// Get recent tasks within time range
	recentTasks, err := r.searchService.GetRecentTasks(ctx, timeRangeObj, DefaultRecentTasksLimit)
	if err != nil {
		return nil, err
	}

	// Get today's statistics
	todayStats, err := r.GetTodayStatistics(ctx)
	if err != nil {
		return nil, err
	}

	return &DashboardData{
		RunningTask: runningTask,
		RecentTasks: recentTasks,
		TodayStats:  todayStats,
	}, nil
}

// GetDayStatistics returns summary statistics for a specific day
func (r *reportingServiceImpl) GetDayStatistics(ctx context.Context, date time.Time) (*DayStatistics, error) {
	// Get date range for the specific day
	dateRange := r.timeService.GetDateRange(date)

	// Search for time entries within the date range
	criteria := SearchCriteria{
		TimeRange: dateRange,
	}

	timeEntries, err := r.searchService.SearchTimeEntries(ctx, criteria)
	if err != nil {
		return nil, err
	}

	// Calculate statistics
	totalDuration := time.Duration(0)
	taskMap := make(map[int64]bool)
	sessionCount := len(timeEntries)
	completedCount := 0

	for _, entryWithTask := range timeEntries {
		// Track unique tasks
		taskMap[entryWithTask.Task.ID] = true

		// Calculate duration
		if entryWithTask.TimeEntry.EndTime != nil {
			totalDuration += entryWithTask.TimeEntry.EndTime.Sub(entryWithTask.TimeEntry.StartTime)
			completedCount++
		} else {
			// Running entry - only count if it started today
			if r.timeService.IsToday(entryWithTask.TimeEntry.StartTime) {
				totalDuration += time.Since(entryWithTask.TimeEntry.StartTime)
			}
		}
	}

	totalTime := r.timeService.FormatDuration(totalDuration)
	taskCount := len(taskMap)

	return &DayStatistics{
		TotalTime:      totalTime,
		TaskCount:      taskCount,
		SessionCount:   sessionCount,
		CompletedCount: completedCount,
	}, nil
}

// GetTodayStatistics returns summary statistics for today
func (r *reportingServiceImpl) GetTodayStatistics(ctx context.Context) (*DayStatistics, error) {
	return r.GetDayStatistics(ctx, time.Now())
}

// AggregateTaskData aggregates time entries by task and returns task activity map
func (r *reportingServiceImpl) AggregateTaskData(entries []*domain.TimeEntry) map[int64]*TaskActivity {
	taskMap := make(map[int64]*TaskActivity)

	for _, entry := range entries {
		taskID := entry.TaskID
		
		// Initialize task activity if not exists
		if _, exists := taskMap[taskID]; !exists {
			taskMap[taskID] = &TaskActivity{
				Task: &domain.Task{ID: taskID}, // Minimal task - would need full task data from elsewhere
				LastWorked:   entry.StartTime,
				TotalTime:    "0m",
				SessionCount: 0,
				IsRunning:    false,
			}
		}

		activity := taskMap[taskID]
		
		// Update session count
		activity.SessionCount++

		// Update last worked time
		if entry.StartTime.After(activity.LastWorked) {
			activity.LastWorked = entry.StartTime
		}

		// Check if running
		if entry.EndTime == nil {
			activity.IsRunning = true
		}
	}

	return taskMap
}

// CalculateTotalDuration calculates total duration across all time entries
func (r *reportingServiceImpl) CalculateTotalDuration(entries []*domain.TimeEntry) time.Duration {
	var totalDuration time.Duration

	for _, entry := range entries {
		if entry.EndTime != nil {
			totalDuration += entry.EndTime.Sub(entry.StartTime)
		} else {
			// Running entry
			totalDuration += time.Since(entry.StartTime)
		}
	}

	return totalDuration
}

// FormatStatistics formats activity analysis into day statistics format
func (r *reportingServiceImpl) FormatStatistics(stats *ActivityAnalysis) *DayStatistics {
	totalTime := r.timeService.FormatDuration(stats.TotalDuration)

	return &DayStatistics{
		TotalTime:      totalTime,
		TaskCount:      0, // Not available from ActivityAnalysis
		SessionCount:   stats.SessionCount,
		CompletedCount: 0, // Not available from ActivityAnalysis
	}
}