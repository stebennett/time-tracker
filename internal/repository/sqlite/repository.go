package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"time-tracker/internal/errors"
	"time-tracker/internal/repository/sqlite/migrations"

	_ "modernc.org/sqlite"
)

// Database operation timeout constants
// Note: These are now configurable via Config, these are fallback defaults
const (
	// DefaultDatabaseQueryTimeout is the default maximum time allowed for database queries
	DefaultDatabaseQueryTimeout = 10 * time.Second
	// DefaultDatabaseWriteTimeout is the default maximum time allowed for database writes
	DefaultDatabaseWriteTimeout = 5 * time.Second
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
	CreateTimeEntry(ctx context.Context, entry *TimeEntry) error
	CreateTask(ctx context.Context, task *Task) error

	// Read operations
	GetTimeEntry(ctx context.Context, id int64) (*TimeEntry, error)
	ListTimeEntries(ctx context.Context) ([]*TimeEntry, error)
	SearchTimeEntries(ctx context.Context, opts SearchOptions) ([]*TimeEntry, error)
	GetTask(ctx context.Context, id int64) (*Task, error)
	ListTasks(ctx context.Context) ([]*Task, error)

	// Update operations
	UpdateTimeEntry(ctx context.Context, entry *TimeEntry) error
	UpdateTask(ctx context.Context, task *Task) error

	// Delete operations
	DeleteTimeEntry(ctx context.Context, id int64) error
	DeleteTask(ctx context.Context, id int64) error

	// Utility
	Close() error
}

// DatabaseConfig interface for repository configuration
type DatabaseConfig interface {
	GetQueryTimeout() time.Duration
	GetWriteTimeout() time.Duration
}

// SQLiteRepository implements the Repository interface
type SQLiteRepository struct {
	db     *sql.DB
	config DatabaseConfig
}

// withTimeout creates a context with timeout for database operations
func (r *SQLiteRepository) withTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}

// withQueryTimeout creates a context with query timeout
func (r *SQLiteRepository) withQueryTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	timeout := DefaultDatabaseQueryTimeout
	if r.config != nil {
		timeout = r.config.GetQueryTimeout()
	}
	return context.WithTimeout(ctx, timeout)
}

// withWriteTimeout creates a context with write timeout
func (r *SQLiteRepository) withWriteTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	timeout := DefaultDatabaseWriteTimeout
	if r.config != nil {
		timeout = r.config.GetWriteTimeout()
	}
	return context.WithTimeout(ctx, timeout)
}

// New creates a new SQLite repository instance
func New(dbPath string) (*SQLiteRepository, error) {
	return NewWithConfig(dbPath, nil)
}

// NewWithConfig creates a new SQLite repository instance with configuration
func NewWithConfig(dbPath string, config DatabaseConfig) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, errors.NewDatabaseError("open database", err)
	}

	// Run migrations
	if err := migrations.RunMigrations(db); err != nil {
		db.Close()
		return nil, errors.NewDatabaseError("run migrations", err)
	}

	return &SQLiteRepository{db: db, config: config}, nil
}

// Close closes the database connection
func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}

// CreateTimeEntry creates a new time entry
func (r *SQLiteRepository) CreateTimeEntry(ctx context.Context, entry *TimeEntry) error {
	// Add timeout for write operations
	timeoutCtx, cancel := r.withWriteTimeout(ctx)
	defer cancel()
	
	query := `
	INSERT INTO time_entries (start_time, end_time, task_id)
	VALUES (?, ?, ?)`

	id, err := ExecuteWithLastInsertID(timeoutCtx, r.db, query, FormatTimeForDB(entry.StartTime), FormatTimePtrForDB(entry.EndTime), entry.TaskID)
	if err != nil {
		return err
	}

	entry.ID = id
	return nil
}

// GetTimeEntry retrieves a time entry by ID
func (r *SQLiteRepository) GetTimeEntry(ctx context.Context, id int64) (*TimeEntry, error) {
	// Add timeout for read operations
	timeoutCtx, cancel := r.withQueryTimeout(ctx)
	defer cancel()
	
	query := `
	SELECT id, start_time, end_time, task_id
	FROM time_entries
	WHERE id = ?`

	return QuerySingle(timeoutCtx, r.db, query, ScanTimeEntry, "time entry", fmt.Sprintf("%d", id), id)
}

// ListTimeEntries retrieves all time entries
func (r *SQLiteRepository) ListTimeEntries(ctx context.Context) ([]*TimeEntry, error) {
	query := `
	SELECT id, start_time, end_time, task_id
	FROM time_entries
	ORDER BY start_time ASC`

	return QueryMultiple(ctx, r.db, query, ScanTimeEntries, "time entries")
}

// UpdateTimeEntry updates an existing time entry
func (r *SQLiteRepository) UpdateTimeEntry(ctx context.Context, entry *TimeEntry) error {
	query := `
	UPDATE time_entries
	SET start_time = ?, end_time = ?, task_id = ?
	WHERE id = ?`

	return ExecuteWithRowsAffected(ctx, r.db, query, "time entry", fmt.Sprintf("%d", entry.ID), FormatTimeForDB(entry.StartTime), FormatTimePtrForDB(entry.EndTime), entry.TaskID, entry.ID)
}

// DeleteTimeEntry deletes a time entry by ID
func (r *SQLiteRepository) DeleteTimeEntry(ctx context.Context, id int64) error {
	query := `DELETE FROM time_entries WHERE id = ?`
	return ExecuteWithRowsAffected(ctx, r.db, query, "time entry", fmt.Sprintf("%d", id), id)
}

// CreateTask creates a new task
func (r *SQLiteRepository) CreateTask(ctx context.Context, task *Task) error {
	query := `INSERT INTO tasks (task_name) VALUES (?)`
	id, err := ExecuteWithLastInsertID(ctx, r.db, query, task.TaskName)
	if err != nil {
		return err
	}
	task.ID = id
	return nil
}

// GetTask retrieves a task by ID
func (r *SQLiteRepository) GetTask(ctx context.Context, id int64) (*Task, error) {
	query := `SELECT id, task_name FROM tasks WHERE id = ?`
	return QuerySingle(ctx, r.db, query, ScanTask, "task", fmt.Sprintf("%d", id), id)
}

// ListTasks retrieves all tasks
func (r *SQLiteRepository) ListTasks(ctx context.Context) ([]*Task, error) {
	query := `SELECT id, task_name FROM tasks ORDER BY task_name ASC`
	return QueryMultiple(ctx, r.db, query, ScanTasks, "tasks")
}

// UpdateTask updates an existing task
func (r *SQLiteRepository) UpdateTask(ctx context.Context, task *Task) error {
	query := `UPDATE tasks SET task_name = ? WHERE id = ?`
	return ExecuteWithRowsAffected(ctx, r.db, query, "task", fmt.Sprintf("%d", task.ID), task.TaskName, task.ID)
}

// DeleteTask deletes a task by ID
func (r *SQLiteRepository) DeleteTask(ctx context.Context, id int64) error {
	query := `DELETE FROM tasks WHERE id = ?`
	return ExecuteWithRowsAffected(ctx, r.db, query, "task", fmt.Sprintf("%d", id), id)
}

// SearchTimeEntries searches for time entries based on the provided options
func (r *SQLiteRepository) SearchTimeEntries(ctx context.Context, opts SearchOptions) ([]*TimeEntry, error) {
	// Add timeout for potentially long-running search operations
	timeoutCtx, cancel := r.withQueryTimeout(ctx)
	defer cancel()
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
	return QueryMultiple(timeoutCtx, r.db, query, ScanTimeEntries, "time entries", args...)
}
