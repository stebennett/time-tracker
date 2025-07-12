package services

import (
	"context"
	"fmt"
	"time"
	"time-tracker/internal/domain"
	"time-tracker/internal/errors"
	"time-tracker/internal/repository/sqlite"
	"time-tracker/internal/validation"
)

// timeServiceImpl implements the TimeService interface
type timeServiceImpl struct {
	repo               sqlite.Repository
	mapper             *domain.Mapper
	timeEntryValidator *validation.TimeEntryValidator
}

// NewTimeService creates a new TimeService instance
func NewTimeService(repo sqlite.Repository) TimeService {
	return &timeServiceImpl{
		repo:               repo,
		mapper:             domain.NewMapper(),
		timeEntryValidator: validation.NewTimeEntryValidator(),
	}
}

// ParseTimeRange converts time shorthand ("30m", "2h", "1d") to actual time range
func (t *timeServiceImpl) ParseTimeRange(timeStr string) (*TimeRange, error) {
	if timeStr == "" {
		return nil, errors.NewValidationError("time range cannot be empty", nil)
	}

	duration, err := t.parseTimeShorthand(timeStr)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	start := now.Add(-duration)
	
	return &TimeRange{
		Start: start,
		End:   now,
	}, nil
}

// parseTimeShorthand converts shorthand time strings to durations
// TODO: Extract this logic from CLI to make it more comprehensive
func (t *timeServiceImpl) parseTimeShorthand(timeStr string) (time.Duration, error) {
	switch timeStr {
	case "30m":
		return 30 * time.Minute, nil
	case "1h":
		return 1 * time.Hour, nil
	case "2h":
		return 2 * time.Hour, nil
	case "1d":
		return 24 * time.Hour, nil
	case "1w":
		return 7 * 24 * time.Hour, nil
	case "1mo":
		return 30 * 24 * time.Hour, nil
	case "1y":
		return 365 * 24 * time.Hour, nil
	default:
		return 0, errors.NewValidationError("invalid time format", nil)
	}
}

// ValidateTimeEntry validates time entry parameters
func (t *timeServiceImpl) ValidateTimeEntry(taskID int64, start time.Time, end *time.Time) error {
	return t.timeEntryValidator.ValidateTimeEntryForCreation(taskID, start, end)
}

// CalculateDuration calculates human-readable duration between two times
func (t *timeServiceImpl) CalculateDuration(start time.Time, end *time.Time) string {
	if end == nil {
		return t.CalculateRunningDuration(start)
	}
	
	duration := end.Sub(start)
	return t.FormatDuration(duration)
}

// FormatDuration formats a duration into human-readable string
func (t *timeServiceImpl) FormatDuration(duration time.Duration) string {
	if duration < 0 {
		return "0h 0m"
	}
	
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

// CalculateRunningDuration calculates duration for a running task
func (t *timeServiceImpl) CalculateRunningDuration(startTime time.Time) string {
	elapsed := time.Since(startTime)
	return fmt.Sprintf("running for %s", t.FormatDuration(elapsed))
}

// GetRunningEntries returns all currently running time entries
func (t *timeServiceImpl) GetRunningEntries(ctx context.Context) ([]*domain.TimeEntry, error) {
	// Empty search returns running tasks only (as per repository implementation)
	searchOpts := sqlite.SearchOptions{}
	dbEntries, err := t.repo.SearchTimeEntries(ctx, searchOpts)
	if err != nil {
		return nil, err
	}

	// Filter for running entries and convert to domain
	runningEntries := make([]*domain.TimeEntry, 0)
	for _, dbEntry := range dbEntries {
		if dbEntry.EndTime == nil {
			domainEntry := t.mapper.TimeEntry.FromDatabase(*dbEntry)
			runningEntries = append(runningEntries, &domainEntry)
		}
	}

	return runningEntries, nil
}

// StopRunningEntries stops all currently running time entries
func (t *timeServiceImpl) StopRunningEntries(ctx context.Context) ([]*domain.TimeEntry, error) {
	// Get all running entries
	searchOpts := sqlite.SearchOptions{}
	runningEntries, err := t.repo.SearchTimeEntries(ctx, searchOpts)
	if err != nil {
		return nil, err
	}

	// Stop each running entry
	now := time.Now()
	stoppedEntries := make([]*domain.TimeEntry, 0, len(runningEntries))
	
	for _, entry := range runningEntries {
		if entry.EndTime == nil { // Confirm it's running
			entry.EndTime = &now
			err := t.repo.UpdateTimeEntry(ctx, entry)
			if err != nil {
				return nil, err
			}
			
			domainEntry := t.mapper.TimeEntry.FromDatabase(*entry)
			stoppedEntries = append(stoppedEntries, &domainEntry)
		}
	}

	return stoppedEntries, nil
}

// CreateTimeEntry creates a new running time entry for a task
func (t *timeServiceImpl) CreateTimeEntry(ctx context.Context, taskID int64) (*domain.TimeEntry, error) {
	now := time.Now()
	
	// Validate the time entry
	if err := t.ValidateTimeEntry(taskID, now, nil); err != nil {
		return nil, err
	}

	// Create database time entry
	dbEntry := &sqlite.TimeEntry{
		TaskID:    taskID,
		StartTime: now,
		EndTime:   nil, // Running task
	}
	
	err := t.repo.CreateTimeEntry(ctx, dbEntry)
	if err != nil {
		return nil, err
	}

	// Convert to domain model
	domainEntry := t.mapper.TimeEntry.FromDatabase(*dbEntry)
	return &domainEntry, nil
}

// IsToday checks if a given time is within today's date range
func (t *timeServiceImpl) IsToday(timeValue time.Time) bool {
	now := time.Now()
	year1, month1, day1 := timeValue.Date()
	year2, month2, day2 := now.Date()
	return year1 == year2 && month1 == month2 && day1 == day2
}

// GetTodayRange returns the time range for today (start of day to now)
func (t *timeServiceImpl) GetTodayRange() *TimeRange {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return &TimeRange{
		Start: startOfDay,
		End:   now,
	}
}

// GetDateRange returns the time range for a specific date (full day)
func (t *timeServiceImpl) GetDateRange(date time.Time) *TimeRange {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	return &TimeRange{
		Start: startOfDay,
		End:   endOfDay,
	}
}