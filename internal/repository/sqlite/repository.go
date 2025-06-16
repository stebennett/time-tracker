package sqlite

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SearchOptions contains all possible search parameters
type SearchOptions struct {
	StartTime   *time.Time
	EndTime     *time.Time
	Description *string
}

// Repository defines the interface for database operations
type Repository interface {
	// Create operations
	CreateTimeEntry(entry *TimeEntry) error
	
	// Read operations
	GetTimeEntry(id int64) (*TimeEntry, error)
	ListTimeEntries() ([]*TimeEntry, error)
	SearchTimeEntries(opts SearchOptions) ([]*TimeEntry, error)
	
	// Update operations
	UpdateTimeEntry(entry *TimeEntry) error
	
	// Delete operations
	DeleteTimeEntry(id int64) error
	
	// Utility
	Close() error
}

// SQLiteRepository implements the Repository interface
type SQLiteRepository struct {
	db *sql.DB
}

// New creates a new SQLite repository instance
func New(dbPath string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create the table if it doesn't exist
	if err := createTable(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &SQLiteRepository{db: db}, nil
}

// Close closes the database connection
func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}

// CreateTimeEntry creates a new time entry
func (r *SQLiteRepository) CreateTimeEntry(entry *TimeEntry) error {
	query := `
	INSERT INTO time_entries (start_time, end_time, description)
	VALUES (?, ?, ?)`

	result, err := r.db.Exec(query, entry.StartTime, entry.EndTime, entry.Description)
	if err != nil {
		return fmt.Errorf("failed to create time entry: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	entry.ID = id
	return nil
}

// GetTimeEntry retrieves a time entry by ID
func (r *SQLiteRepository) GetTimeEntry(id int64) (*TimeEntry, error) {
	query := `
	SELECT id, start_time, end_time, description
	FROM time_entries
	WHERE id = ?`

	entry := &TimeEntry{}
	var endTime sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&entry.ID,
		&entry.StartTime,
		&endTime,
		&entry.Description,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("time entry not found: %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get time entry: %w", err)
	}

	if endTime.Valid {
		entry.EndTime = &endTime.Time
	}

	return entry, nil
}

// ListTimeEntries retrieves all time entries
func (r *SQLiteRepository) ListTimeEntries() ([]*TimeEntry, error) {
	query := `
	SELECT id, start_time, end_time, description
	FROM time_entries
	ORDER BY start_time ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query time entries: %w", err)
	}
	defer rows.Close()

	var entries []*TimeEntry
	for rows.Next() {
		entry := &TimeEntry{}
		var endTime sql.NullTime

		err := rows.Scan(
			&entry.ID,
			&entry.StartTime,
			&endTime,
			&entry.Description,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan time entry: %w", err)
		}

		if endTime.Valid {
			entry.EndTime = &endTime.Time
		}

		entries = append(entries, entry)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating time entries: %w", err)
	}

	return entries, nil
}

// UpdateTimeEntry updates an existing time entry
func (r *SQLiteRepository) UpdateTimeEntry(entry *TimeEntry) error {
	query := `
	UPDATE time_entries
	SET start_time = ?, end_time = ?, description = ?
	WHERE id = ?`

	result, err := r.db.Exec(query,
		entry.StartTime,
		entry.EndTime,
		entry.Description,
		entry.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update time entry: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("time entry not found: %d", entry.ID)
	}

	return nil
}

// DeleteTimeEntry deletes a time entry by ID
func (r *SQLiteRepository) DeleteTimeEntry(id int64) error {
	query := `DELETE FROM time_entries WHERE id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete time entry: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("time entry not found: %d", id)
	}

	return nil
}

// SearchTimeEntries searches for time entries based on the provided options
func (r *SQLiteRepository) SearchTimeEntries(opts SearchOptions) ([]*TimeEntry, error) {
	var conditions []string
	var args []interface{}

	// Build time range conditions
	if opts.StartTime != nil || opts.EndTime != nil {
		timeCondition := "("
		if opts.StartTime != nil {
			timeCondition += "end_time IS NULL OR end_time >= ?"
			args = append(args, *opts.StartTime)
		}
		if opts.StartTime != nil && opts.EndTime != nil {
			timeCondition += " AND "
		}
		if opts.EndTime != nil {
			timeCondition += "start_time <= ?"
			args = append(args, *opts.EndTime)
		}
		timeCondition += ")"
		conditions = append(conditions, timeCondition)
	} else {
		// If no time range is specified, only return running tasks (end_time IS NULL)
		conditions = append(conditions, "end_time IS NULL")
	}

	// Build description condition
	if opts.Description != nil && *opts.Description != "" {
		conditions = append(conditions, "description LIKE ?")
		args = append(args, "%"+*opts.Description+"%")
	}

	// Build the final query
	query := `
	SELECT id, start_time, end_time, description
	FROM time_entries`

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY start_time ASC"

	// Execute the query
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search time entries: %w", err)
	}
	defer rows.Close()

	var entries []*TimeEntry
	for rows.Next() {
		entry := &TimeEntry{}
		var endTime sql.NullTime

		err := rows.Scan(
			&entry.ID,
			&entry.StartTime,
			&endTime,
			&entry.Description,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan time entry: %w", err)
		}

		if endTime.Valid {
			entry.EndTime = &endTime.Time
		}

		entries = append(entries, entry)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating time entries: %w", err)
	}

	return entries, nil
}

// createTable creates the initial table structure
func createTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS time_entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		start_time DATETIME NOT NULL,
		end_time DATETIME,
		description TEXT
	)`

	_, err := db.Exec(query)
	return err
} 