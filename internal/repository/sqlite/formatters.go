package sqlite

import (
	"time"
)

// FormatTimeForDB formats a time.Time value as RFC3339 string for consistent database storage
func FormatTimeForDB(t time.Time) string {
	return t.Format(time.RFC3339)
}

// FormatTimePtrForDB formats a *time.Time value as RFC3339 string, returning nil if the pointer is nil
func FormatTimePtrForDB(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return FormatTimeForDB(*t)
}

// ParseTimeFromDB parses an RFC3339 formatted time string from the database
func ParseTimeFromDB(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}