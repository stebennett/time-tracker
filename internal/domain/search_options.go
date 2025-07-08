package domain

import "time"

// SearchOptions represents search criteria for time entries.
// This is a domain model that mirrors the database search options
// but belongs to the domain layer for proper separation of concerns.
type SearchOptions struct {
	StartTime *time.Time
	EndTime   *time.Time
	TaskID    *int64
	TaskName  *string
}