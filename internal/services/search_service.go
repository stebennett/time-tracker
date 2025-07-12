package services

import (
	"context"
	"sort"
	"strings"
	"time"
	"time-tracker/internal/domain"
	"time-tracker/internal/repository/sqlite"
)

// searchServiceImpl implements the SearchService interface
type searchServiceImpl struct {
	repo        sqlite.Repository
	timeService TimeService
	taskService TaskService
	mapper      *domain.Mapper
}

// NewSearchService creates a new SearchService instance
func NewSearchService(repo sqlite.Repository, timeService TimeService, taskService TaskService) SearchService {
	return &searchServiceImpl{
		repo:        repo,
		timeService: timeService,
		taskService: taskService,
		mapper:      domain.NewMapper(),
	}
}

// filterRunningEntries filters entries to only include running ones
func (s *searchServiceImpl) filterRunningEntries(entries []*sqlite.TimeEntry) []*sqlite.TimeEntry {
	runningEntries := make([]*sqlite.TimeEntry, 0)
	for _, entry := range entries {
		if entry.EndTime == nil {
			runningEntries = append(runningEntries, entry)
		}
	}
	return runningEntries
}

// matchesTextFilter checks if a task name matches the text filter
func (s *searchServiceImpl) matchesTextFilter(taskName, textFilter string) bool {
	if textFilter == "" {
		return true
	}
	return strings.Contains(strings.ToLower(taskName), strings.ToLower(textFilter))
}

// buildSearchOptions builds repository search options from criteria
func (s *searchServiceImpl) buildSearchOptions(criteria SearchCriteria) sqlite.SearchOptions {
	searchOpts := sqlite.SearchOptions{}
	
	if criteria.TimeRange != nil {
		searchOpts.StartTime = &criteria.TimeRange.Start
		searchOpts.EndTime = &criteria.TimeRange.End
	}
	
	if criteria.TaskID != nil {
		searchOpts.TaskID = criteria.TaskID
	}
	
	return searchOpts
}

// SearchTasks searches for tasks based on criteria and returns task activities
func (s *searchServiceImpl) SearchTasks(ctx context.Context, criteria SearchCriteria) ([]*TaskActivity, error) {
	// Get all tasks
	dbTasks, err := s.repo.ListTasks(ctx)
	if err != nil {
		return nil, err
	}

	// Build task activities
	activities := make([]*TaskActivity, 0)
	
	for _, dbTask := range dbTasks {
		// Filter by specific task ID if specified
		if criteria.TaskID != nil && dbTask.ID != *criteria.TaskID {
			continue
		}
		
		// Filter by text if specified
		if !s.matchesTextFilter(dbTask.TaskName, criteria.TextFilter) {
			continue
		}
		
		// Get time entries for this task
		taskCriteria := criteria
		taskCriteria.TaskID = &dbTask.ID // Override to get entries for this specific task
		searchOpts := s.buildSearchOptions(taskCriteria)
		
		entries, err := s.repo.SearchTimeEntries(ctx, searchOpts)
		if err != nil {
			return nil, err
		}
		
		// Filter for running only if specified
		if criteria.RunningOnly {
			entries = s.filterRunningEntries(entries)
		}
		
		// Skip tasks with no matching entries
		if len(entries) == 0 {
			continue
		}
		
		// Calculate task activity
		activity := s.buildTaskActivity(dbTask, entries)
		activities = append(activities, activity)
	}

	return activities, nil
}

// SearchTimeEntries searches for time entries based on criteria
func (s *searchServiceImpl) SearchTimeEntries(ctx context.Context, criteria SearchCriteria) ([]*TimeEntryWithTask, error) {
	var entries []*sqlite.TimeEntry
	var err error
	
	// If no criteria specified, get all time entries
	if criteria.TimeRange == nil && criteria.TaskID == nil && !criteria.RunningOnly {
		entries, err = s.repo.ListTimeEntries(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		// Build search options from criteria
		searchOpts := s.buildSearchOptions(criteria)
		
		// Get time entries
		entries, err = s.repo.SearchTimeEntries(ctx, searchOpts)
		if err != nil {
			return nil, err
		}
	}
	
	// Filter for running only if specified
	if criteria.RunningOnly {
		entries = s.filterRunningEntries(entries)
	}
	
	// Convert to TimeEntryWithTask
	result := make([]*TimeEntryWithTask, 0, len(entries))
	
	for _, entry := range entries {
		// Get the task for this entry
		dbTask, err := s.repo.GetTask(ctx, entry.TaskID)
		if err != nil {
			return nil, err
		}
		
		// Filter by text if specified
		if !s.matchesTextFilter(dbTask.TaskName, criteria.TextFilter) {
			continue
		}
		
		// Convert to domain models
		domainEntry := s.mapper.TimeEntry.FromDatabase(*entry)
		domainTask := s.mapper.Task.FromDatabase(*dbTask)
		duration := s.timeService.CalculateDuration(domainEntry.StartTime, domainEntry.EndTime)
		
		entryWithTask := &TimeEntryWithTask{
			TimeEntry: &domainEntry,
			Task:      &domainTask,
			Duration:  duration,
		}
		
		result = append(result, entryWithTask)
	}

	return result, nil
}

// FilterTasksByTime filters tasks by their last worked time
func (s *searchServiceImpl) FilterTasksByTime(tasks []*TaskActivity, timeRange *TimeRange) []*TaskActivity {
	if timeRange == nil {
		return tasks
	}
	
	filtered := make([]*TaskActivity, 0, len(tasks))
	
	for _, task := range tasks {
		if task.LastWorked.After(timeRange.Start) && task.LastWorked.Before(timeRange.End) {
			filtered = append(filtered, task)
		}
	}
	
	return filtered
}

// SortTasks sorts tasks according to the specified order
func (s *searchServiceImpl) SortTasks(tasks []*TaskActivity, order SortOrder) []*TaskActivity {
	// Make a copy to avoid modifying the original
	sorted := make([]*TaskActivity, len(tasks))
	copy(sorted, tasks)
	
	switch order {
	case SortByRecentFirst:
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].LastWorked.After(sorted[j].LastWorked)
		})
	case SortByOldestFirst:
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].LastWorked.Before(sorted[j].LastWorked)
		})
	case SortByName:
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Task.TaskName < sorted[j].Task.TaskName
		})
	case SortByDuration:
		// For duration sorting, we'd need to parse the duration strings
		// For now, keep original order
		break
	}
	
	return sorted
}

// SortTimeEntries sorts time entries according to the specified order
func (s *searchServiceImpl) SortTimeEntries(entries []*TimeEntryWithTask, order SortOrder) []*TimeEntryWithTask {
	// Make a copy to avoid modifying the original
	sorted := make([]*TimeEntryWithTask, len(entries))
	copy(sorted, entries)
	
	switch order {
	case SortByRecentFirst:
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].TimeEntry.StartTime.After(sorted[j].TimeEntry.StartTime)
		})
	case SortByOldestFirst:
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].TimeEntry.StartTime.Before(sorted[j].TimeEntry.StartTime)
		})
	case SortByName:
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Task.TaskName < sorted[j].Task.TaskName
		})
	case SortByDuration:
		// For duration sorting, we'd need to parse the duration strings
		// For now, keep original order
		break
	}
	
	return sorted
}

// GetRecentTasks returns the most recently worked tasks within the time range
func (s *searchServiceImpl) GetRecentTasks(ctx context.Context, timeRange *TimeRange, limit int) ([]*TaskActivity, error) {
	// Search for all tasks
	criteria := SearchCriteria{
		TimeRange: timeRange,
	}
	
	activities, err := s.SearchTasks(ctx, criteria)
	if err != nil {
		return nil, err
	}
	
	// Sort by most recent first
	sorted := s.SortTasks(activities, SortByRecentFirst)
	
	// Apply limit
	if limit > 0 && len(sorted) > limit {
		sorted = sorted[:limit]
	}
	
	return sorted, nil
}

// FindTasksWithActivity returns tasks that have activity within the time range
func (s *searchServiceImpl) FindTasksWithActivity(ctx context.Context, timeRange *TimeRange) ([]*TaskActivity, error) {
	// Search for all tasks
	criteria := SearchCriteria{
		TimeRange: timeRange,
	}
	
	activities, err := s.SearchTasks(ctx, criteria)
	if err != nil {
		return nil, err
	}
	
	// Filter for tasks with actual activity (session count > 0)
	filtered := make([]*TaskActivity, 0)
	for _, activity := range activities {
		if activity.SessionCount > 0 {
			filtered = append(filtered, activity)
		}
	}
	
	return filtered, nil
}

// buildTaskActivity creates a TaskActivity from a task and its time entries
func (s *searchServiceImpl) buildTaskActivity(dbTask *sqlite.Task, entries []*sqlite.TimeEntry) *TaskActivity {
	domainTask := s.mapper.Task.FromDatabase(*dbTask)
	
	// Calculate statistics
	var lastWorked time.Time
	var totalDuration time.Duration
	sessionCount := len(entries)
	isRunning := false
	
	for _, entry := range entries {
		// Update last worked time
		if entry.StartTime.After(lastWorked) {
			lastWorked = entry.StartTime
		}
		
		// Calculate duration
		if entry.EndTime != nil {
			totalDuration += entry.EndTime.Sub(entry.StartTime)
		} else {
			// Running entry
			isRunning = true
			totalDuration += time.Since(entry.StartTime)
		}
	}
	
	totalTime := s.timeService.FormatDuration(totalDuration)
	
	return &TaskActivity{
		Task:         &domainTask,
		LastWorked:   lastWorked,
		TotalTime:    totalTime,
		SessionCount: sessionCount,
		IsRunning:    isRunning,
	}
}