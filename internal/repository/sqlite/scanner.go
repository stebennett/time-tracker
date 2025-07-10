package sqlite

import (
	"database/sql"
)

// Scanner interface defines the common scanning behavior for both sql.Row and sql.Rows
type Scanner interface {
	Scan(dest ...interface{}) error
}

// ScanTimeEntry scans a single time entry from a database row
func ScanTimeEntry(scanner Scanner) (*TimeEntry, error) {
	entry := &TimeEntry{}
	var endTime sql.NullTime

	err := scanner.Scan(
		&entry.ID,
		&entry.StartTime,
		&endTime,
		&entry.TaskID,
	)
	if err != nil {
		return nil, err
	}

	if endTime.Valid {
		entry.EndTime = &endTime.Time
	}

	return entry, nil
}

// Rows interface defines the common behavior for sql.Rows
type Rows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Err() error
}

// ScanTimeEntries scans multiple time entries from database rows
func ScanTimeEntries(rows Rows) ([]*TimeEntry, error) {
	var entries []*TimeEntry
	for rows.Next() {
		entry, err := ScanTimeEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

// ScanTask scans a single task from a database row
func ScanTask(scanner Scanner) (*Task, error) {
	task := &Task{}
	err := scanner.Scan(&task.ID, &task.TaskName)
	if err != nil {
		return nil, err
	}
	return task, nil
}

// ScanTasks scans multiple tasks from database rows
func ScanTasks(rows Rows) ([]*Task, error) {
	var tasks []*Task
	for rows.Next() {
		task, err := ScanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}