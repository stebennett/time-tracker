package sqlite

import "time"

// Task represents a task
// Add this struct for the new tasks table
//
type Task struct {
	ID       int64
	TaskName string
}

// TimeEntry represents a single time tracking entry
// Update to use TaskID instead of Description
//
type TimeEntry struct {
	ID        int64
	TaskID    int64
	StartTime time.Time
	EndTime   *time.Time // Using pointer to allow NULL values
} 