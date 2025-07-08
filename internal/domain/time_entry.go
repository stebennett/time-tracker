package domain

import (
	"time"
)

// TimeEntry represents a time tracking entry in the domain model.
// This is a pure domain model without database-specific concerns.
type TimeEntry struct {
	ID        int64
	TaskID    int64
	StartTime time.Time
	EndTime   *time.Time
}

// NewTimeEntry creates a new TimeEntry for the given task.
func NewTimeEntry(taskID int64, startTime time.Time) TimeEntry {
	return TimeEntry{
		TaskID:    taskID,
		StartTime: startTime,
	}
}

// IsRunning returns true if the time entry is currently running (no end time).
func (te TimeEntry) IsRunning() bool {
	return te.EndTime == nil
}

// Stop sets the end time for the time entry.
func (te TimeEntry) Stop(endTime time.Time) TimeEntry {
	te.EndTime = &endTime
	return te
}

// Duration returns the duration of the time entry.
// If the entry is still running, it returns the duration up to now.
func (te TimeEntry) Duration() time.Duration {
	if te.EndTime == nil {
		return time.Since(te.StartTime)
	}
	return te.EndTime.Sub(te.StartTime)
}

// IsValid checks if the time entry has valid data.
func (te TimeEntry) IsValid() bool {
	if te.TaskID <= 0 {
		return false
	}
	if te.StartTime.IsZero() {
		return false
	}
	if te.EndTime != nil && te.EndTime.Before(te.StartTime) {
		return false
	}
	return true
}