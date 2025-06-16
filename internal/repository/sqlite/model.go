package sqlite

import "time"

// TimeEntry represents a single time tracking entry
type TimeEntry struct {
	ID          int64
	StartTime   time.Time
	EndTime     *time.Time // Using pointer to allow NULL values
	Description string
} 