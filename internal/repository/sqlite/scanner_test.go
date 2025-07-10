package sqlite

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestScanner implements the Scanner interface for testing
type TestScanner struct {
	data []interface{}
	err  error
}

func (ts *TestScanner) Scan(dest ...interface{}) error {
	if ts.err != nil {
		return ts.err
	}
	
	if len(dest) != len(ts.data) {
		return errors.New("mismatch in number of destinations")
	}
	
	for i, d := range dest {
		switch v := d.(type) {
		case *int64:
			*v = ts.data[i].(int64)
		case *time.Time:
			*v = ts.data[i].(time.Time)
		case *sql.NullTime:
			*v = ts.data[i].(sql.NullTime)
		case *string:
			*v = ts.data[i].(string)
		}
	}
	
	return nil
}

func TestScanTimeEntry(t *testing.T) {
	tests := []struct {
		name        string
		scanner     *TestScanner
		expected    *TimeEntry
		expectError bool
	}{
		{
			name: "Valid time entry with end time",
			scanner: &TestScanner{
				data: []interface{}{
					int64(1),
					time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
					sql.NullTime{Time: time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC), Valid: true},
					int64(100),
				},
			},
			expected: &TimeEntry{
				ID:        1,
				TaskID:    100,
				StartTime: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
				EndTime:   func() *time.Time { t := time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC); return &t }(),
			},
			expectError: false,
		},
		{
			name: "Valid time entry without end time (running task)",
			scanner: &TestScanner{
				data: []interface{}{
					int64(2),
					time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC),
					sql.NullTime{Valid: false},
					int64(200),
				},
			},
			expected: &TimeEntry{
				ID:        2,
				TaskID:    200,
				StartTime: time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC),
				EndTime:   nil,
			},
			expectError: false,
		},
		{
			name: "Scanner error",
			scanner: &TestScanner{
				err: sql.ErrNoRows,
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ScanTimeEntry(tt.scanner)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected.ID, result.ID)
				assert.Equal(t, tt.expected.TaskID, result.TaskID)
				assert.True(t, tt.expected.StartTime.Equal(result.StartTime))
				if tt.expected.EndTime == nil {
					assert.Nil(t, result.EndTime)
				} else {
					assert.NotNil(t, result.EndTime)
					assert.True(t, tt.expected.EndTime.Equal(*result.EndTime))
				}
			}
		})
	}
}

func TestScanTask(t *testing.T) {
	tests := []struct {
		name        string
		scanner     *TestScanner
		expected    *Task
		expectError bool
	}{
		{
			name: "Valid task",
			scanner: &TestScanner{
				data: []interface{}{
					int64(1),
					"Test Task",
				},
			},
			expected: &Task{
				ID:       1,
				TaskName: "Test Task",
			},
			expectError: false,
		},
		{
			name: "Empty task name",
			scanner: &TestScanner{
				data: []interface{}{
					int64(2),
					"",
				},
			},
			expected: &Task{
				ID:       2,
				TaskName: "",
			},
			expectError: false,
		},
		{
			name: "Scanner error",
			scanner: &TestScanner{
				err: sql.ErrNoRows,
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ScanTask(tt.scanner)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected.ID, result.ID)
				assert.Equal(t, tt.expected.TaskName, result.TaskName)
			}
		})
	}
}

// TestRows implements the Rows interface for testing
type TestRows struct {
	rows      [][]interface{}
	currentRow int
	err       error
}

func (tr *TestRows) Next() bool {
	if tr.err != nil {
		return false
	}
	if tr.currentRow >= len(tr.rows) {
		return false
	}
	tr.currentRow++
	return tr.currentRow <= len(tr.rows)
}

func (tr *TestRows) Scan(dest ...interface{}) error {
	if tr.err != nil {
		return tr.err
	}
	
	if tr.currentRow == 0 || tr.currentRow > len(tr.rows) {
		return errors.New("no current row")
	}
	
	rowData := tr.rows[tr.currentRow-1]
	
	if len(dest) != len(rowData) {
		return errors.New("mismatch in number of destinations")
	}
	
	for i, d := range dest {
		switch v := d.(type) {
		case *int64:
			*v = rowData[i].(int64)
		case *time.Time:
			*v = rowData[i].(time.Time)
		case *sql.NullTime:
			*v = rowData[i].(sql.NullTime)
		case *string:
			*v = rowData[i].(string)
		}
	}
	
	return nil
}

func (tr *TestRows) Err() error {
	return tr.err
}

func TestScanTimeEntries(t *testing.T) {
	tests := []struct {
		name        string
		rows        *TestRows
		expected    []*TimeEntry
		expectError bool
	}{
		{
			name: "Multiple time entries",
			rows: &TestRows{
				rows: [][]interface{}{
					{
						int64(1),
						time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
						sql.NullTime{Time: time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC), Valid: true},
						int64(100),
					},
					{
						int64(2),
						time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
						sql.NullTime{Valid: false},
						int64(200),
					},
				},
			},
			expected: []*TimeEntry{
				{
					ID:        1,
					TaskID:    100,
					StartTime: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
					EndTime:   func() *time.Time { t := time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC); return &t }(),
				},
				{
					ID:        2,
					TaskID:    200,
					StartTime: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
					EndTime:   nil,
				},
			},
			expectError: false,
		},
		{
			name: "Empty result set",
			rows: &TestRows{
				rows: [][]interface{}{},
			},
			expected:    []*TimeEntry{},
			expectError: false,
		},
		{
			name: "Scan error",
			rows: &TestRows{
				rows: [][]interface{}{
					{int64(1), time.Now(), sql.NullTime{}, int64(100)},
				},
				err: sql.ErrConnDone,
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ScanTimeEntries(tt.rows)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, len(tt.expected))
				for i, expected := range tt.expected {
					assert.Equal(t, expected.ID, result[i].ID)
					assert.Equal(t, expected.TaskID, result[i].TaskID)
					assert.True(t, expected.StartTime.Equal(result[i].StartTime))
					if expected.EndTime == nil {
						assert.Nil(t, result[i].EndTime)
					} else {
						assert.NotNil(t, result[i].EndTime)
						assert.True(t, expected.EndTime.Equal(*result[i].EndTime))
					}
				}
			}
		})
	}
}

func TestScanTasks(t *testing.T) {
	tests := []struct {
		name        string
		rows        *TestRows
		expected    []*Task
		expectError bool
	}{
		{
			name: "Multiple tasks",
			rows: &TestRows{
				rows: [][]interface{}{
					{int64(1), "Task 1"},
					{int64(2), "Task 2"},
				},
			},
			expected: []*Task{
				{ID: 1, TaskName: "Task 1"},
				{ID: 2, TaskName: "Task 2"},
			},
			expectError: false,
		},
		{
			name: "Empty result set",
			rows: &TestRows{
				rows: [][]interface{}{},
			},
			expected:    []*Task{},
			expectError: false,
		},
		{
			name: "Scan error",
			rows: &TestRows{
				rows: [][]interface{}{
					{int64(1), "Task 1"},
				},
				err: sql.ErrConnDone,
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ScanTasks(tt.rows)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, len(tt.expected))
				for i, expected := range tt.expected {
					assert.Equal(t, expected.ID, result[i].ID)
					assert.Equal(t, expected.TaskName, result[i].TaskName)
				}
			}
		})
	}
}