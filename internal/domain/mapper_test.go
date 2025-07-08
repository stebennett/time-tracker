package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"time-tracker/internal/repository/sqlite"
)

func TestTaskMapper_ToDatabase(t *testing.T) {
	mapper := NewTaskMapper()
	domainTask := Task{
		ID:       1,
		TaskName: "Test Task",
	}

	result := mapper.ToDatabase(domainTask)

	expected := sqlite.Task{
		ID:       1,
		TaskName: "Test Task",
	}
	assert.Equal(t, expected, result)
}

func TestTaskMapper_FromDatabase(t *testing.T) {
	mapper := NewTaskMapper()
	dbTask := sqlite.Task{
		ID:       1,
		TaskName: "Test Task",
	}

	result := mapper.FromDatabase(dbTask)

	expected := Task{
		ID:       1,
		TaskName: "Test Task",
	}
	assert.Equal(t, expected, result)
}

func TestTaskMapper_ToDatabaseSlice(t *testing.T) {
	mapper := NewTaskMapper()
	domainTasks := []Task{
		{ID: 1, TaskName: "Task 1"},
		{ID: 2, TaskName: "Task 2"},
	}

	result := mapper.ToDatabaseSlice(domainTasks)

	expected := []sqlite.Task{
		{ID: 1, TaskName: "Task 1"},
		{ID: 2, TaskName: "Task 2"},
	}
	assert.Equal(t, expected, result)
}

func TestTaskMapper_FromDatabaseSlice(t *testing.T) {
	mapper := NewTaskMapper()
	dbTasks := []sqlite.Task{
		{ID: 1, TaskName: "Task 1"},
		{ID: 2, TaskName: "Task 2"},
	}

	result := mapper.FromDatabaseSlice(dbTasks)

	expected := []Task{
		{ID: 1, TaskName: "Task 1"},
		{ID: 2, TaskName: "Task 2"},
	}
	assert.Equal(t, expected, result)
}

func TestTaskMapper_EmptySlice(t *testing.T) {
	mapper := NewTaskMapper()

	domainResult := mapper.ToDatabaseSlice([]Task{})
	dbResult := mapper.FromDatabaseSlice([]sqlite.Task{})

	assert.Empty(t, domainResult)
	assert.Empty(t, dbResult)
}

func TestTimeEntryMapper_ToDatabase(t *testing.T) {
	mapper := NewTimeEntryMapper()
	endTime := time.Now()
	domainEntry := TimeEntry{
		ID:        1,
		TaskID:    2,
		StartTime: time.Now().Add(-time.Hour),
		EndTime:   &endTime,
	}

	result := mapper.ToDatabase(domainEntry)

	expected := sqlite.TimeEntry{
		ID:        1,
		TaskID:    2,
		StartTime: domainEntry.StartTime,
		EndTime:   &endTime,
	}
	assert.Equal(t, expected, result)
}

func TestTimeEntryMapper_FromDatabase(t *testing.T) {
	mapper := NewTimeEntryMapper()
	endTime := time.Now()
	dbEntry := sqlite.TimeEntry{
		ID:        1,
		TaskID:    2,
		StartTime: time.Now().Add(-time.Hour),
		EndTime:   &endTime,
	}

	result := mapper.FromDatabase(dbEntry)

	expected := TimeEntry{
		ID:        1,
		TaskID:    2,
		StartTime: dbEntry.StartTime,
		EndTime:   &endTime,
	}
	assert.Equal(t, expected, result)
}

func TestTimeEntryMapper_ToDatabaseSlice(t *testing.T) {
	mapper := NewTimeEntryMapper()
	endTime := time.Now()
	domainEntries := []TimeEntry{
		{ID: 1, TaskID: 1, StartTime: time.Now().Add(-time.Hour), EndTime: &endTime},
		{ID: 2, TaskID: 2, StartTime: time.Now().Add(-30 * time.Minute), EndTime: nil},
	}

	result := mapper.ToDatabaseSlice(domainEntries)

	expected := []sqlite.TimeEntry{
		{ID: 1, TaskID: 1, StartTime: domainEntries[0].StartTime, EndTime: &endTime},
		{ID: 2, TaskID: 2, StartTime: domainEntries[1].StartTime, EndTime: nil},
	}
	assert.Equal(t, expected, result)
}

func TestTimeEntryMapper_FromDatabaseSlice(t *testing.T) {
	mapper := NewTimeEntryMapper()
	endTime := time.Now()
	dbEntries := []sqlite.TimeEntry{
		{ID: 1, TaskID: 1, StartTime: time.Now().Add(-time.Hour), EndTime: &endTime},
		{ID: 2, TaskID: 2, StartTime: time.Now().Add(-30 * time.Minute), EndTime: nil},
	}

	result := mapper.FromDatabaseSlice(dbEntries)

	expected := []TimeEntry{
		{ID: 1, TaskID: 1, StartTime: dbEntries[0].StartTime, EndTime: &endTime},
		{ID: 2, TaskID: 2, StartTime: dbEntries[1].StartTime, EndTime: nil},
	}
	assert.Equal(t, expected, result)
}

func TestTimeEntryMapper_EmptySlice(t *testing.T) {
	mapper := NewTimeEntryMapper()

	domainResult := mapper.ToDatabaseSlice([]TimeEntry{})
	dbResult := mapper.FromDatabaseSlice([]sqlite.TimeEntry{})

	assert.Empty(t, domainResult)
	assert.Empty(t, dbResult)
}

func TestTimeEntryMapper_RunningEntry(t *testing.T) {
	mapper := NewTimeEntryMapper()
	domainEntry := TimeEntry{
		ID:        1,
		TaskID:    2,
		StartTime: time.Now().Add(-time.Hour),
		EndTime:   nil,
	}

	dbResult := mapper.ToDatabase(domainEntry)
	domainResult := mapper.FromDatabase(dbResult)

	assert.Equal(t, domainEntry, domainResult)
	assert.Nil(t, dbResult.EndTime)
	assert.Nil(t, domainResult.EndTime)
}

func TestNewMapper(t *testing.T) {
	mapper := NewMapper()

	assert.NotNil(t, mapper)
	assert.NotNil(t, mapper.Task)
	assert.NotNil(t, mapper.TimeEntry)
	assert.IsType(t, &TaskMapper{}, mapper.Task)
	assert.IsType(t, &TimeEntryMapper{}, mapper.TimeEntry)
}

func TestMapper_Integration(t *testing.T) {
	mapper := NewMapper()

	// Test round-trip conversion for Task
	originalTask := Task{ID: 1, TaskName: "Test Task"}
	dbTask := mapper.Task.ToDatabase(originalTask)
	convertedTask := mapper.Task.FromDatabase(dbTask)
	assert.Equal(t, originalTask, convertedTask)

	// Test round-trip conversion for TimeEntry
	endTime := time.Now()
	originalEntry := TimeEntry{
		ID:        1,
		TaskID:    2,
		StartTime: time.Now().Add(-time.Hour),
		EndTime:   &endTime,
	}
	dbEntry := mapper.TimeEntry.ToDatabase(originalEntry)
	convertedEntry := mapper.TimeEntry.FromDatabase(dbEntry)
	assert.Equal(t, originalEntry, convertedEntry)
}