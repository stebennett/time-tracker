package migrations

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"
)

func init() {
	RegisterGoMigration(3, Up_000003_update_time_entries_to_rfc3339, Down_000003_update_time_entries_to_rfc3339)
}

// Up_000003_update_time_entries_to_rfc3339 migrates all time entries to RFC3339 format.
// This migration handles various Go time string formats including:
// - Monotonic clock suffixes (m=+0.000000000)
// - Timezone offsets (+0100, -0700)
// - Zone names (BST, PST, etc.)
// - Nanoseconds
func Up_000003_update_time_entries_to_rfc3339(tx *sql.Tx) error {
	// Read all rows into memory first to avoid locking issues
	type entry struct {
		id        int64
		startTime sql.NullString
		endTime   sql.NullString
	}
	var entries []entry

	rows, err := tx.Query("SELECT id, start_time, end_time FROM time_entries")
	if err != nil {
		return fmt.Errorf("failed to query time entries: %w", err)
	}
	for rows.Next() {
		var e entry
		if err := rows.Scan(&e.id, &e.startTime, &e.endTime); err != nil {
			rows.Close()
			return fmt.Errorf("failed to scan row %d: %w", e.id, err)
		}
		entries = append(entries, e)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return fmt.Errorf("error iterating time entries: %w", err)
	}
	rows.Close()

	updates := 0
	errors := 0
	total := 0

	// Prepare statements for better performance
	startStmt, err := tx.Prepare("UPDATE time_entries SET start_time = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare start_time update statement: %w", err)
	}
	defer startStmt.Close()

	endStmt, err := tx.Prepare("UPDATE time_entries SET end_time = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare end_time update statement: %w", err)
	}
	defer endStmt.Close()

	for _, e := range entries {
		total++

		// Process start_time
		if e.startTime.Valid && e.startTime.String != "" {
			newStartTime, err := parseGoTimeToRFC3339(e.startTime.String)
			if err != nil {
				fmt.Printf("Warning: could not parse start_time for id %d: %v\n", e.id, err)
				errors++
			} else {
				_, err := startStmt.Exec(newStartTime, e.id)
				if err != nil {
					return fmt.Errorf("failed to update start_time for id %d: %w", e.id, err)
				}
				updates++
			}
		}

		// Process end_time
		if e.endTime.Valid && e.endTime.String != "" {
			newEndTime, err := parseGoTimeToRFC3339(e.endTime.String)
			if err != nil {
				fmt.Printf("Warning: could not parse end_time for id %d: %v\n", e.id, err)
				errors++
			} else {
				_, err := endStmt.Exec(newEndTime, e.id)
				if err != nil {
					return fmt.Errorf("failed to update end_time for id %d: %w", e.id, err)
				}
				updates++
			}
		}
	}

	fmt.Printf("Migration complete: processed %d rows, updated %d values, encountered %d errors\n", total, updates, errors)
	return nil
}

// Down_000003_update_time_entries_to_rfc3339 reverts RFC3339 format back to basic format.
// This is a simplified rollback that converts RFC3339 back to basic YYYY-MM-DD HH:MM:SS format.
func Down_000003_update_time_entries_to_rfc3339(tx *sql.Tx) error {
	// Convert RFC3339 format back to basic format
	_, err := tx.Exec(`
		UPDATE time_entries 
		SET start_time = substr(start_time, 1, 10) || ' ' || substr(start_time, 12, 8)
		WHERE start_time GLOB '????-??-??T??:??:??*'
	`)
	if err != nil {
		return fmt.Errorf("failed to revert start_time: %w", err)
	}

	_, err = tx.Exec(`
		UPDATE time_entries 
		SET end_time = substr(end_time, 1, 10) || ' ' || substr(end_time, 12, 8)
		WHERE end_time GLOB '????-??-??T??:??:??*' AND end_time IS NOT NULL
	`)
	if err != nil {
		return fmt.Errorf("failed to revert end_time: %w", err)
	}

	return nil
}

// parseGoTimeToRFC3339 parses various Go time formats and converts them to RFC3339.
// It handles monotonic clock suffixes, timezone offsets, zone names, and nanoseconds.
func parseGoTimeToRFC3339(timeStr string) (string, error) {
	// Strip monotonic clock suffix if present
	timeStr = stripMonotonicSuffix(timeStr)

	// Try different Go time layouts
	layouts := []string{
		"2006-01-02 15:04:05.999999999 -0700 MST", // Go default with zone name
		"2006-01-02 15:04:05.999999999 -0700",     // Go default without zone name
		"2006-01-02 15:04:05 -0700 MST",           // Go without nanoseconds, with zone name
		"2006-01-02 15:04:05 -0700",               // Go without nanoseconds, without zone name
		"2006-01-02 15:04:05.999999999",           // Go with nanoseconds, no timezone
		"2006-01-02 15:04:05",                     // Go basic format
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, timeStr); err == nil {
			return t.Format(time.RFC3339), nil
		}
	}

	// If we can't parse it with Go layouts, try parsing as RFC3339
	if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
		return t.Format(time.RFC3339), nil
	}

	return "", fmt.Errorf("could not parse time format: %s", timeStr)
}

// stripMonotonicSuffix removes the monotonic clock suffix from Go time strings.
func stripMonotonicSuffix(timeStr string) string {
	if idx := strings.Index(timeStr, " m="); idx != -1 {
		return timeStr[:idx]
	}
	return timeStr
}

// isRFC3339 checks if a string is already in RFC3339 format.
func isRFC3339(timeStr string) bool {
	// RFC3339 format: 2006-01-02T15:04:05Z07:00
	// Must have 'T' separator, no space, and timezone offset with colon
	rfc3339Regex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?([+-]\d{2}:\d{2}|Z)$`)
	result := rfc3339Regex.MatchString(timeStr)
	fmt.Printf("DEBUG: isRFC3339(%q) = %v\n", timeStr, result)
	return result
}
