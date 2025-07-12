package api

import (
	"context"
	"fmt"
	"time"
	"time-tracker/internal/domain"
	"time-tracker/internal/errors"
	"time-tracker/internal/repository/sqlite"
	"time-tracker/internal/validation"
)

// SortOrder defines how task results should be sorted
type SortOrder string

const (
	SortByRecentFirst SortOrder = "recent_first" // Most recently worked (default)
	SortByOldestFirst SortOrder = "oldest_first" // Least recently worked (good for cleanup)
	SortByName        SortOrder = "name"         // Alphabetical by task name
	SortByDuration    SortOrder = "duration"     // By total time spent
)

// Business domain types
type TaskSession struct {
	Task      *domain.Task      `json:"task"`
	TimeEntry *domain.TimeEntry `json:"time_entry"`
	Duration  string            `json:"duration"` // Human-readable for running tasks
}

type TaskActivity struct {
	Task         *domain.Task `json:"task"`
	LastWorked   time.Time    `json:"last_worked"`
	TotalTime    string       `json:"total_time"`    // Human-readable total duration
	SessionCount int          `json:"session_count"`
	IsRunning    bool         `json:"is_running"`
}

type TaskSummary struct {
	Task         *domain.Task        `json:"task"`
	TimeEntries  []*domain.TimeEntry `json:"time_entries"`
	TotalTime    string              `json:"total_time"`
	SessionCount int                 `json:"session_count"`
	RunningCount int                 `json:"running_count"`
	FirstEntry   time.Time           `json:"first_entry"`
	LastEntry    time.Time           `json:"last_entry"`
	IsRunning    bool                `json:"is_running"`
}

type DashboardData struct {
	RunningTask *TaskSession    `json:"running_task"`
	RecentTasks []*TaskActivity `json:"recent_tasks"`
	TodayStats  *DayStatistics  `json:"today_stats"`
}

type DayStatistics struct {
	TotalTime      string `json:"total_time"`
	TaskCount      int    `json:"task_count"`
	SessionCount   int    `json:"session_count"`
	CompletedCount int    `json:"completed_count"`
}

type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type TimeEntryWithTask struct {
	TimeEntry *domain.TimeEntry `json:"time_entry"`
	Task      *domain.Task      `json:"task"`
	Duration  string            `json:"duration"`
}

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
	repo               sqlite.Repository
	mapper             *domain.Mapper
	taskValidator      *validation.TaskValidator
	timeEntryValidator *validation.TimeEntryValidator
}

// NewBusinessAPI creates a new BusinessAPI instance
func NewBusinessAPI(repo sqlite.Repository) BusinessAPI {
	return &businessAPIImpl{
		repo:               repo,
		mapper:             domain.NewMapper(),
		taskValidator:      validation.NewTaskValidator(),
		timeEntryValidator: validation.NewTimeEntryValidator(),
	}
}

// ========== Task Management Workflows ==========

func (b *businessAPIImpl) StartNewTask(ctx context.Context, taskName string) (*TaskSession, error) {
	// 1. Validate task name
	if err := b.taskValidator.ValidateTaskForCreation(taskName); err != nil {
		return nil, errors.NewValidationError("invalid task name", err)
	}
	
	// Get cleaned task name
	cleanedName, err := b.taskValidator.GetValidTaskName(taskName)
	if err != nil {
		return nil, errors.NewValidationError("invalid task name", err)
	}

	// 2. Stop all running tasks first
	_, err = b.StopAllRunningTasks(ctx)
	if err != nil {
		return nil, err
	}

	// 3. Create new task
	dbTask := &sqlite.Task{TaskName: cleanedName}
	err = b.repo.CreateTask(ctx, dbTask)
	if err != nil {
		return nil, err
	}

	// 4. Create time entry for new task
	now := time.Now()
	dbEntry := &sqlite.TimeEntry{
		TaskID:    dbTask.ID,
		StartTime: now,
		EndTime:   nil, // Running task
	}
	err = b.repo.CreateTimeEntry(ctx, dbEntry)
	if err != nil {
		return nil, err
	}

	// 5. Convert to domain models and build TaskSession
	domainTask := b.mapper.Task.FromDatabase(*dbTask)
	domainEntry := b.mapper.TimeEntry.FromDatabase(*dbEntry)

	return &TaskSession{
		Task:      &domainTask,
		TimeEntry: &domainEntry,
		Duration:  "running", // TODO: Calculate actual duration
	}, nil
}

func (b *businessAPIImpl) ResumeTask(ctx context.Context, taskID int64) (*TaskSession, error) {
	// 1. Validate task ID
	if err := b.taskValidator.ValidateTaskID(taskID); err != nil {
		return nil, errors.NewValidationError("invalid task ID", err)
	}

	// 2. Check if task exists
	dbTask, err := b.repo.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// 3. Stop all running tasks
	_, err = b.StopAllRunningTasks(ctx)
	if err != nil {
		return nil, err
	}

	// 4. Create new time entry for existing task
	now := time.Now()
	dbEntry := &sqlite.TimeEntry{
		TaskID:    taskID,
		StartTime: now,
		EndTime:   nil, // Running task
	}
	err = b.repo.CreateTimeEntry(ctx, dbEntry)
	if err != nil {
		return nil, err
	}

	// 5. Convert to domain models and build TaskSession
	domainTask := b.mapper.Task.FromDatabase(*dbTask)
	domainEntry := b.mapper.TimeEntry.FromDatabase(*dbEntry)

	return &TaskSession{
		Task:      &domainTask,
		TimeEntry: &domainEntry,
		Duration:  "running",
	}, nil
}

func (b *businessAPIImpl) StopAllRunningTasks(ctx context.Context) ([]*domain.TimeEntry, error) {
	// 1. Find all running tasks (entries with nil EndTime)
	searchOpts := sqlite.SearchOptions{} // Empty search returns running tasks only
	runningEntries, err := b.repo.SearchTimeEntries(ctx, searchOpts)
	if err != nil {
		return nil, err
	}

	// 2. Stop each running entry
	now := time.Now()
	stoppedEntries := make([]*domain.TimeEntry, 0, len(runningEntries))
	
	for _, entry := range runningEntries {
		if entry.EndTime == nil { // Confirm it's running
			entry.EndTime = &now
			err := b.repo.UpdateTimeEntry(ctx, entry)
			if err != nil {
				return nil, err
			}
			
			domainEntry := b.mapper.TimeEntry.FromDatabase(*entry)
			stoppedEntries = append(stoppedEntries, &domainEntry)
		}
	}

	return stoppedEntries, nil
}

func (b *businessAPIImpl) DeleteTaskWithEntries(ctx context.Context, taskID int64) error {
	// 1. Validate task ID
	if err := b.taskValidator.ValidateTaskID(taskID); err != nil {
		return errors.NewValidationError("invalid task ID", err)
	}

	// 2. Check if task exists
	_, err := b.repo.GetTask(ctx, taskID)
	if err != nil {
		return err // Return not found or other repository error
	}

	// 3. Delete all time entries for this task
	searchOpts := sqlite.SearchOptions{TaskID: &taskID}
	timeEntries, err := b.repo.SearchTimeEntries(ctx, searchOpts)
	if err != nil {
		return err
	}

	for _, entry := range timeEntries {
		err := b.repo.DeleteTimeEntry(ctx, entry.ID)
		if err != nil {
			return err
		}
	}

	// 4. Delete the task itself
	err = b.repo.DeleteTask(ctx, taskID)
	if err != nil {
		return err
	}

	return nil
}

func (b *businessAPIImpl) UpdateTaskName(ctx context.Context, taskID int64, newName string) (*domain.Task, error) {
	// 1. Validate task ID
	if err := b.taskValidator.ValidateTaskID(taskID); err != nil {
		return nil, errors.NewValidationError("invalid task ID", err)
	}

	// 2. Validate new task name
	if err := b.taskValidator.ValidateTaskForUpdate(taskID, newName); err != nil {
		return nil, errors.NewValidationError("invalid task name", err)
	}

	// Get cleaned name
	cleanedName, err := b.taskValidator.GetValidTaskName(newName)
	if err != nil {
		return nil, errors.NewValidationError("invalid task name", err)
	}

	// 3. Get existing task to ensure it exists
	dbTask, err := b.repo.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// 4. Update task name
	dbTask.TaskName = cleanedName
	err = b.repo.UpdateTask(ctx, dbTask)
	if err != nil {
		return nil, err
	}

	// 5. Convert to domain model
	domainTask := b.mapper.Task.FromDatabase(*dbTask)
	return &domainTask, nil
}

// ========== Query Operations ==========

func (b *businessAPIImpl) GetCurrentSession(ctx context.Context) (*TaskSession, error) {
	// 1. Find currently running task entry
	searchOpts := sqlite.SearchOptions{} // Empty search returns running tasks only
	runningEntries, err := b.repo.SearchTimeEntries(ctx, searchOpts)
	if err != nil {
		return nil, err
	}

	// 2. Find the first running entry
	var runningEntry *sqlite.TimeEntry
	for _, entry := range runningEntries {
		if entry.EndTime == nil {
			runningEntry = entry
			break
		}
	}

	if runningEntry == nil {
		return nil, errors.NewNotFoundError("running task", "")
	}

	// 3. Get the associated task
	dbTask, err := b.repo.GetTask(ctx, runningEntry.TaskID)
	if err != nil {
		return nil, err
	}

	// 4. Convert to domain models and build TaskSession
	domainTask := b.mapper.Task.FromDatabase(*dbTask)
	domainEntry := b.mapper.TimeEntry.FromDatabase(*runningEntry)

	// 5. Calculate duration
	duration := fmt.Sprintf("running for %v", time.Since(runningEntry.StartTime).Truncate(time.Second))

	return &TaskSession{
		Task:      &domainTask,
		TimeEntry: &domainEntry,
		Duration:  duration,
	}, nil
}

func (b *businessAPIImpl) GetTask(ctx context.Context, id int64) (*domain.Task, error) {
	// Validate input using business rules
	if err := b.taskValidator.ValidateTaskID(id); err != nil {
		return nil, errors.NewValidationError("invalid task ID", err)
	}

	// Get task from repository
	dbTask, err := b.repo.GetTask(ctx, id)
	if err != nil {
		// Repository already returns proper AppError types, just pass through
		return nil, err
	}

	// Convert to domain model
	domainTask := b.mapper.Task.FromDatabase(*dbTask)
	return &domainTask, nil
}

func (b *businessAPIImpl) GetTaskSummary(ctx context.Context, taskID int64) (*TaskSummary, error) {
	// 1. Validate task ID
	if err := b.taskValidator.ValidateTaskID(taskID); err != nil {
		return nil, errors.NewValidationError("invalid task ID", err)
	}
	
	// 2. Get task details
	dbTask, err := b.repo.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}
	
	// 3. Get all time entries for this task
	searchOpts := sqlite.SearchOptions{TaskID: &taskID}
	entries, err := b.repo.SearchTimeEntries(ctx, searchOpts)
	if err != nil {
		return nil, err
	}
	
	// 4. Calculate summary statistics
	domainTask := b.mapper.Task.FromDatabase(*dbTask)
	domainEntries := make([]*domain.TimeEntry, len(entries))
	sessionCount := len(entries)
	runningCount := 0
	var firstEntry, lastEntry time.Time
	
	for i, entry := range entries {
		domainEntries[i] = &domain.TimeEntry{}
		*domainEntries[i] = b.mapper.TimeEntry.FromDatabase(*entry)
		
		if entry.EndTime == nil {
			runningCount++
		}
		
		if firstEntry.IsZero() || entry.StartTime.Before(firstEntry) {
			firstEntry = entry.StartTime
		}
		if lastEntry.IsZero() || entry.StartTime.After(lastEntry) {
			lastEntry = entry.StartTime
		}
	}
	
	return &TaskSummary{
		Task:         &domainTask,
		TimeEntries:  domainEntries,
		TotalTime:    "0h 0m", // TODO: Calculate actual total
		SessionCount: sessionCount,
		RunningCount: runningCount,
		FirstEntry:   firstEntry,
		LastEntry:    lastEntry,
		IsRunning:    runningCount > 0,
	}, nil
}

// ========== Search and Discovery Operations ==========

func (b *businessAPIImpl) ParseTimeRange(ctx context.Context, timeStr string) (*TimeRange, error) {
	// TODO: Extract time parsing logic from CLI
	// For now, implement basic time parsing
	if timeStr == "" {
		return nil, errors.NewValidationError("time range cannot be empty", nil)
	}

	// This is a simplified implementation - in real implementation we'd extract
	// the parseTimeShorthand logic from CLI
	now := time.Now()
	var duration time.Duration

	switch timeStr {
	case "30m":
		duration = 30 * time.Minute
	case "1h":
		duration = 1 * time.Hour
	case "2h":
		duration = 2 * time.Hour
	case "1d":
		duration = 24 * time.Hour
	case "1w":
		duration = 7 * 24 * time.Hour
	default:
		return nil, errors.NewValidationError("invalid time format", nil)
	}

	start := now.Add(-duration)
	return &TimeRange{
		Start: start,
		End:   now,
	}, nil
}

func (b *businessAPIImpl) SearchTasks(ctx context.Context, timeRange string, textFilter string, sortOrder SortOrder) ([]*TaskActivity, error) {
	// TODO: Implement comprehensive task search with sorting and time filtering
	// For now, return basic task list
	searchOpts := sqlite.SearchOptions{}
	
	// Add text filter if provided
	if textFilter != "" {
		searchOpts.TaskName = &textFilter
	}
	
	entries, err := b.repo.SearchTimeEntries(ctx, searchOpts)
	if err != nil {
		return nil, err
	}
	
	// Group by task and build TaskActivity objects
	taskMap := make(map[int64]*TaskActivity)
	for _, entry := range entries {
		if _, exists := taskMap[entry.TaskID]; !exists {
			// Get task details
			dbTask, err := b.repo.GetTask(ctx, entry.TaskID)
			if err != nil {
				continue // Skip if task not found
			}
			
			domainTask := b.mapper.Task.FromDatabase(*dbTask)
			taskMap[entry.TaskID] = &TaskActivity{
				Task:        &domainTask,
				LastWorked:  entry.StartTime,
				IsRunning:   entry.EndTime == nil,
				TotalTime:   "0h 0m", // TODO: Calculate actual total
				SessionCount: 1,
			}
		} else {
			taskMap[entry.TaskID].SessionCount++
		}
		
		// Update last worked time if this entry is more recent
		if entry.StartTime.After(taskMap[entry.TaskID].LastWorked) {
			taskMap[entry.TaskID].LastWorked = entry.StartTime
			taskMap[entry.TaskID].IsRunning = entry.EndTime == nil
		}
	}
	
	// Convert map to slice
	result := make([]*TaskActivity, 0, len(taskMap))
	for _, activity := range taskMap {
		result = append(result, activity)
	}
	
	return result, nil
}

func (b *businessAPIImpl) SearchTimeEntries(ctx context.Context, timeRange string, textFilter string) ([]*TimeEntryWithTask, error) {
	// TODO: Implement time range parsing and filtering
	searchOpts := sqlite.SearchOptions{}
	
	// Add text filter if provided
	if textFilter != "" {
		searchOpts.TaskName = &textFilter
	}
	
	entries, err := b.repo.SearchTimeEntries(ctx, searchOpts)
	if err != nil {
		return nil, err
	}
	
	result := make([]*TimeEntryWithTask, 0, len(entries))
	for _, entry := range entries {
		// Get associated task
		dbTask, err := b.repo.GetTask(ctx, entry.TaskID)
		if err != nil {
			continue // Skip if task not found
		}
		
		domainTask := b.mapper.Task.FromDatabase(*dbTask)
		domainEntry := b.mapper.TimeEntry.FromDatabase(*entry)
		
		duration := "completed"
		if entry.EndTime == nil {
			duration = fmt.Sprintf("running for %v", time.Since(entry.StartTime).Truncate(time.Second))
		}
		
		result = append(result, &TimeEntryWithTask{
			TimeEntry: &domainEntry,
			Task:      &domainTask,
			Duration:  duration,
		})
	}
	
	return result, nil
}

// ========== Dashboard and Analytics ==========

func (b *businessAPIImpl) GetDashboardData(ctx context.Context, timeRange string) (*DashboardData, error) {
	// Get current running task
	runningTask, _ := b.GetCurrentSession(ctx) // Ignore error if no running task
	
	// Get recent tasks - simplified implementation
	recentTasks, err := b.SearchTasks(ctx, "", "", SortByRecentFirst)
	if err != nil {
		return nil, err
	}
	
	// Limit to 5 recent tasks
	recentActivities := make([]*TaskActivity, 0, 5)
	for i, task := range recentTasks {
		if i < 5 {
			recentActivities = append(recentActivities, task)
		}
	}
	
	// Get today's statistics
	todayStats, err := b.GetTodayStatistics(ctx)
	if err != nil {
		// Don't fail if stats unavailable
		todayStats = &DayStatistics{}
	}
	
	return &DashboardData{
		RunningTask: runningTask,
		RecentTasks: recentActivities,
		TodayStats:  todayStats,
	}, nil
}

func (b *businessAPIImpl) GetTodayStatistics(ctx context.Context) (*DayStatistics, error) {
	// TODO: Implement proper today filtering
	// For now, return basic stats from all entries
	entries, err := b.repo.SearchTimeEntries(ctx, sqlite.SearchOptions{})
	if err != nil {
		return nil, err
	}
	
	taskSet := make(map[int64]bool)
	completedCount := 0
	
	for _, entry := range entries {
		taskSet[entry.TaskID] = true
		if entry.EndTime != nil {
			completedCount++
		}
	}
	
	return &DayStatistics{
		TotalTime:      "0h 0m", // TODO: Calculate actual total
		TaskCount:      len(taskSet),
		SessionCount:   len(entries),
		CompletedCount: completedCount,
	}, nil
}