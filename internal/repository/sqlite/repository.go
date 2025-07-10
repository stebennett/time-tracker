package sqlite

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"time-tracker/internal/errors"
	"time-tracker/internal/repository/sqlite/migrations"

	_ "modernc.org/sqlite"
)


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
		return nil, errors.NewDatabaseError("open database", err)
	}

	// Run migrations
	if err := migrations.RunMigrations(db); err != nil {
		db.Close()
		return nil, errors.NewDatabaseError("run migrations", err)
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

	id, err := ExecuteWithLastInsertID(r.db, query, FormatTimeForDB(entry.StartTime), FormatTimePtrForDB(entry.EndTime), entry.TaskID)
	if err != nil {
		return err
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

	return QuerySingle(r.db, query, ScanTimeEntry, "time entry", fmt.Sprintf("%d", id), id)
}

// ListTimeEntries retrieves all time entries
func (r *SQLiteRepository) ListTimeEntries() ([]*TimeEntry, error) {
	query := `
	SELECT id, start_time, end_time, task_id
	FROM time_entries
	ORDER BY start_time ASC`

	return QueryMultiple(r.db, query, ScanTimeEntries, "time entries")
}

// UpdateTimeEntry updates an existing time entry
func (r *SQLiteRepository) UpdateTimeEntry(entry *TimeEntry) error {
	query := `
	UPDATE time_entries
	SET start_time = ?, end_time = ?, task_id = ?
	WHERE id = ?`

	return ExecuteWithRowsAffected(r.db, query, "time entry", fmt.Sprintf("%d", entry.ID), FormatTimeForDB(entry.StartTime), FormatTimePtrForDB(entry.EndTime), entry.TaskID, entry.ID)
}

// DeleteTimeEntry deletes a time entry by ID
func (r *SQLiteRepository) DeleteTimeEntry(id int64) error {
	query := `DELETE FROM time_entries WHERE id = ?`
	return ExecuteWithRowsAffected(r.db, query, "time entry", fmt.Sprintf("%d", id), id)
}

// CreateTask creates a new task
func (r *SQLiteRepository) CreateTask(task *Task) error {
	query := `INSERT INTO tasks (task_name) VALUES (?)`
	id, err := ExecuteWithLastInsertID(r.db, query, task.TaskName)
	if err != nil {
		return err
	}
	task.ID = id
	return nil
}

// GetTask retrieves a task by ID
func (r *SQLiteRepository) GetTask(id int64) (*Task, error) {
	query := `SELECT id, task_name FROM tasks WHERE id = ?`
	return QuerySingle(r.db, query, ScanTask, "task", fmt.Sprintf("%d", id), id)
}

// ListTasks retrieves all tasks
func (r *SQLiteRepository) ListTasks() ([]*Task, error) {
	query := `SELECT id, task_name FROM tasks ORDER BY task_name ASC`
	return QueryMultiple(r.db, query, ScanTasks, "tasks")
}

// UpdateTask updates an existing task
func (r *SQLiteRepository) UpdateTask(task *Task) error {
	query := `UPDATE tasks SET task_name = ? WHERE id = ?`
	return ExecuteWithRowsAffected(r.db, query, "task", fmt.Sprintf("%d", task.ID), task.TaskName, task.ID)
}

// DeleteTask deletes a task by ID
func (r *SQLiteRepository) DeleteTask(id int64) error {
	query := `DELETE FROM tasks WHERE id = ?`
	return ExecuteWithRowsAffected(r.db, query, "task", fmt.Sprintf("%d", id), id)
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
			args = append(args, FormatTimePtrForDB(opts.StartTime))
		}
		if opts.StartTime != nil && opts.EndTime != nil {
			timeCondition += " AND "
		}
		if opts.EndTime != nil {
			timeCondition += "start_time <= ?"
			args = append(args, FormatTimePtrForDB(opts.EndTime))
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
	return QueryMultiple(r.db, query, ScanTimeEntries, "time entries", args...)
}
