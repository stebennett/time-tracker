package sqlite

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"time-tracker/internal/repository/sqlite/migrations"

	_ "modernc.org/sqlite"
)

// formatTimeForDB formats a time.Time value as RFC3339 string for consistent database storage
func formatTimeForDB(t time.Time) string {
	return t.Format(time.RFC3339)
}

// formatTimePtrForDB formats a *time.Time value as RFC3339 string, returning nil if the pointer is nil
func formatTimePtrForDB(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return formatTimeForDB(*t)
}

// SearchOptions contains all possible search parameters
type SearchOptions struct {
	StartTime *time.Time
	EndTime   *time.Time
	TaskID    *int64
	TaskName  *string
}

// Repository defines the interface for database operations
type Repository interface {
	// Create operations
	CreateTimeEntry(entry *TimeEntry) error
	CreateTask(task *Task) error

	// Read operations
	GetTimeEntry(id int64) (*TimeEntry, error)
	ListTimeEntries() ([]*TimeEntry, error)
	SearchTimeEntries(opts SearchOptions) ([]*TimeEntry, error)
	GetTask(id int64) (*Task, error)
	ListTasks() ([]*Task, error)

	// Update operations
	UpdateTimeEntry(entry *TimeEntry) error
	UpdateTask(task *Task) error

	// Delete operations
	DeleteTimeEntry(id int64) error
	DeleteTask(id int64) error

	// Utility
	Close() error
}

// SQLiteRepository implements the Repository interface
type SQLiteRepository struct {
	db *sql.DB
}

// New creates a new SQLite repository instance
func New(dbPath string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Run migrations
	if err := migrations.RunMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
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
	INSERT INTO time_entries (start_time, end_time, task_id)
	VALUES (?, ?, ?)`

	result, err := r.db.Exec(query, formatTimeForDB(entry.StartTime), formatTimePtrForDB(entry.EndTime), entry.TaskID)
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
	SELECT id, start_time, end_time, task_id
	FROM time_entries
	WHERE id = ?`

	entry := &TimeEntry{}
	var endTime sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&entry.ID,
		&entry.StartTime,
		&endTime,
		&entry.TaskID,
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
	SELECT id, start_time, end_time, task_id
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
			&entry.TaskID,
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
	SET start_time = ?, end_time = ?, task_id = ?
	WHERE id = ?`

	result, err := r.db.Exec(query, formatTimeForDB(entry.StartTime), formatTimePtrForDB(entry.EndTime), entry.TaskID, entry.ID)
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

// CreateTask creates a new task
func (r *SQLiteRepository) CreateTask(task *Task) error {
	query := `INSERT INTO tasks (task_name) VALUES (?)`
	result, err := r.db.Exec(query, task.TaskName)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}
	task.ID, err = result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}
	return nil
}

// GetTask retrieves a task by ID
func (r *SQLiteRepository) GetTask(id int64) (*Task, error) {
	query := `SELECT id, task_name FROM tasks WHERE id = ?`
	task := &Task{}
	err := r.db.QueryRow(query, id).Scan(&task.ID, &task.TaskName)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found: %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return task, nil
}

// ListTasks retrieves all tasks
func (r *SQLiteRepository) ListTasks() ([]*Task, error) {
	query := `SELECT id, task_name FROM tasks ORDER BY task_name ASC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		task := &Task{}
		err := rows.Scan(&task.ID, &task.TaskName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tasks: %w", err)
	}
	return tasks, nil
}

// UpdateTask updates an existing task
func (r *SQLiteRepository) UpdateTask(task *Task) error {
	query := `UPDATE tasks SET task_name = ? WHERE id = ?`
	result, err := r.db.Exec(query, task.TaskName, task.ID)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("task not found: %d", task.ID)
	}
	return nil
}

// DeleteTask deletes a task by ID
func (r *SQLiteRepository) DeleteTask(id int64) error {
	query := `DELETE FROM tasks WHERE id = ?`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("task not found: %d", id)
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
			timeCondition += "start_time >= ?"
			args = append(args, formatTimePtrForDB(opts.StartTime))
		}
		if opts.StartTime != nil && opts.EndTime != nil {
			timeCondition += " AND "
		}
		if opts.EndTime != nil {
			timeCondition += "start_time <= ?"
			args = append(args, formatTimePtrForDB(opts.EndTime))
		}
		timeCondition += ")"
		conditions = append(conditions, timeCondition)
	} else if opts.TaskID == nil && opts.TaskName == nil {
		// Only filter for running tasks if no search criteria are provided
		conditions = append(conditions, "end_time IS NULL")
	}

	// Build task_id condition
	if opts.TaskID != nil {
		conditions = append(conditions, "task_id = ?")
		args = append(args, *opts.TaskID)
	}

	// Build task name condition (join with tasks)
	joinTasks := false
	if opts.TaskName != nil && *opts.TaskName != "" {
		joinTasks = true
		conditions = append(conditions, "tasks.task_name LIKE ?")
		args = append(args, "%"+*opts.TaskName+"%")
	}

	// Build the final query
	query := `
	SELECT time_entries.id, start_time, end_time, task_id
	FROM time_entries`
	if joinTasks {
		query += " JOIN tasks ON time_entries.task_id = tasks.id"
	}
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
			&entry.TaskID,
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
