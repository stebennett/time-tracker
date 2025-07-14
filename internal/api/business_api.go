package api

import (
	"context"
	"time-tracker/internal/domain"
	"time-tracker/internal/errors"
	"time-tracker/internal/repository/sqlite"
	"time-tracker/internal/services"
)

// Re-export types from services for backward compatibility
type SortOrder = services.SortOrder
type TaskSession = services.TaskSession
type TaskActivity = services.TaskActivity
type TaskSummary = services.TaskSummary
type DashboardData = services.DashboardData
type DayStatistics = services.DayStatistics
type TimeRange = services.TimeRange
type TimeEntryWithTask = services.TimeEntryWithTask

// Re-export constants from services
const (
	SortByRecentFirst = services.SortByRecentFirst
	SortByOldestFirst = services.SortByOldestFirst
	SortByName        = services.SortByName
	SortByDuration    = services.SortByDuration
)

// BusinessAPI defines the business-logic-only interface for time tracking operations
type BusinessAPI interface {
	// ========== Task Management Workflows ==========

	// StartNewTask creates a new task and starts tracking time, stopping any running tasks
	StartNewTask(ctx context.Context, taskName string) (*TaskSession, error)

	// ResumeTask starts a new time entry for an existing task, stopping running tasks
	ResumeTask(ctx context.Context, taskID int64) (*TaskSession, error)

	// StopAllRunningTasks stops all currently running time entries
	StopAllRunningTasks(ctx context.Context) ([]*domain.TimeEntry, error)

	// DeleteTaskWithEntries deletes a task and all its time entries (safe cascade delete)
	DeleteTaskWithEntries(ctx context.Context, taskID int64) error

	// UpdateTaskName safely updates a task name with validation
	UpdateTaskName(ctx context.Context, taskID int64, newName string) (*domain.Task, error)

	// ========== Query Operations ==========

	// GetCurrentSession returns the currently running task session, if any
	GetCurrentSession(ctx context.Context) (*TaskSession, error)

	// GetTask returns a single task by ID
	GetTask(ctx context.Context, id int64) (*domain.Task, error)

	// GetTaskSummary returns comprehensive summary for a specific task
	GetTaskSummary(ctx context.Context, taskID int64) (*TaskSummary, error)

	// ========== Search and Discovery Operations ==========

	// ParseTimeRange converts time shorthand ("30m", "2h", "1d") to actual time range
	ParseTimeRange(ctx context.Context, timeStr string) (*TimeRange, error)

	// SearchTasks finds tasks by name and/or time range with rich metadata and configurable sorting
	SearchTasks(ctx context.Context, timeRange string, textFilter string, sortOrder SortOrder) ([]*TaskActivity, error)

	// SearchTimeEntries returns detailed time entries with task information for analysis
	SearchTimeEntries(ctx context.Context, timeRange string, textFilter string) ([]*TimeEntryWithTask, error)

	// ========== Dashboard and Analytics ==========

	// GetDashboardData returns all data needed for a dashboard view
	GetDashboardData(ctx context.Context, timeRange string) (*DashboardData, error)

	// GetTodayStatistics returns summary statistics for today's work
	GetTodayStatistics(ctx context.Context) (*DayStatistics, error)
}

// businessAPIImpl implements the BusinessAPI interface
type businessAPIImpl struct {
	timeService      services.TimeService
	taskService      services.TaskService
	searchService    services.SearchService
	reportingService services.ReportingService
}

// NewBusinessAPI creates a new BusinessAPI instance
func NewBusinessAPI(repo sqlite.Repository) BusinessAPI {
	// Create services
	timeService := services.NewTimeService(repo)
	taskService := services.NewTaskService(repo, timeService)
	searchService := services.NewSearchService(repo, timeService, taskService)
	reportingService := services.NewReportingService(repo, timeService, taskService, searchService)

	return &businessAPIImpl{
		timeService:      timeService,
		taskService:      taskService,
		searchService:    searchService,
		reportingService: reportingService,
	}
}

// ========== Task Management Workflows ==========

func (b *businessAPIImpl) StartNewTask(ctx context.Context, taskName string) (*TaskSession, error) {
	return b.taskService.StartNewTask(ctx, taskName)
}

func (b *businessAPIImpl) ResumeTask(ctx context.Context, taskID int64) (*TaskSession, error) {
	return b.taskService.ResumeTask(ctx, taskID)
}

func (b *businessAPIImpl) StopAllRunningTasks(ctx context.Context) ([]*domain.TimeEntry, error) {
	return b.taskService.StopAllRunningTasks(ctx)
}

func (b *businessAPIImpl) DeleteTaskWithEntries(ctx context.Context, taskID int64) error {
	return b.taskService.DeleteTaskWithEntries(ctx, taskID)
}

func (b *businessAPIImpl) UpdateTaskName(ctx context.Context, taskID int64, newName string) (*domain.Task, error) {
	return b.taskService.UpdateTask(ctx, taskID, newName)
}

// ========== Query Operations ==========

func (b *businessAPIImpl) GetCurrentSession(ctx context.Context) (*TaskSession, error) {
	result, err := b.taskService.GetCurrentSession(ctx)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, errors.NewNotFoundError("running task", "")
	}
	return result, nil
}

func (b *businessAPIImpl) GetTask(ctx context.Context, id int64) (*domain.Task, error) {
	return b.taskService.GetTask(ctx, id)
}

func (b *businessAPIImpl) GetTaskSummary(ctx context.Context, taskID int64) (*TaskSummary, error) {
	return b.reportingService.GetTaskSummary(ctx, taskID)
}

// ========== Search and Discovery Operations ==========

func (b *businessAPIImpl) ParseTimeRange(ctx context.Context, timeStr string) (*TimeRange, error) {
	return b.timeService.ParseTimeRange(timeStr)
}

func (b *businessAPIImpl) SearchTasks(ctx context.Context, timeRange string, textFilter string, sortOrder SortOrder) ([]*TaskActivity, error) {
	// Parse time range only if provided
	var timeRangeObj *services.TimeRange
	var err error
	if timeRange != "" {
		timeRangeObj, err = b.timeService.ParseTimeRange(timeRange)
		if err != nil {
			return nil, err
		}
	}
	
	// Create search criteria
	criteria := services.SearchCriteria{
		TimeRange:  timeRangeObj,
		TextFilter: textFilter,
	}
	
	// Get tasks and then sort them
	tasks, err := b.searchService.SearchTasks(ctx, criteria)
	if err != nil {
		return nil, err
	}
	
	return b.searchService.SortTasks(tasks, services.SortOrder(sortOrder)), nil
}

func (b *businessAPIImpl) SearchTimeEntries(ctx context.Context, timeRange string, textFilter string) ([]*TimeEntryWithTask, error) {
	// Parse time range only if provided
	var timeRangeObj *services.TimeRange
	var err error
	if timeRange != "" {
		timeRangeObj, err = b.timeService.ParseTimeRange(timeRange)
		if err != nil {
			return nil, err
		}
	}
	
	// Create search criteria
	criteria := services.SearchCriteria{
		TimeRange:  timeRangeObj,
		TextFilter: textFilter,
	}
	
	return b.searchService.SearchTimeEntries(ctx, criteria)
}

// ========== Dashboard and Analytics ==========

func (b *businessAPIImpl) GetDashboardData(ctx context.Context, timeRange string) (*DashboardData, error) {
	return b.reportingService.GetDashboardData(ctx, timeRange)
}

func (b *businessAPIImpl) GetTodayStatistics(ctx context.Context) (*DayStatistics, error) {
	return b.reportingService.GetTodayStatistics(ctx)
}