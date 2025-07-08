package domain

import (
	"time-tracker/internal/repository/sqlite"
)

// TaskMapper handles conversion between domain and database Task models.
type TaskMapper struct{}

// NewTaskMapper creates a new TaskMapper instance.
func NewTaskMapper() *TaskMapper {
	return &TaskMapper{}
}

// ToDatabase converts a domain Task to a database Task.
func (m *TaskMapper) ToDatabase(domainTask Task) sqlite.Task {
	return sqlite.Task{
		ID:       domainTask.ID,
		TaskName: domainTask.TaskName,
	}
}

// FromDatabase converts a database Task to a domain Task.
func (m *TaskMapper) FromDatabase(dbTask sqlite.Task) Task {
	return Task{
		ID:       dbTask.ID,
		TaskName: dbTask.TaskName,
	}
}

// ToDatabaseSlice converts a slice of domain Tasks to database Tasks.
func (m *TaskMapper) ToDatabaseSlice(domainTasks []Task) []sqlite.Task {
	dbTasks := make([]sqlite.Task, len(domainTasks))
	for i, task := range domainTasks {
		dbTasks[i] = m.ToDatabase(task)
	}
	return dbTasks
}

// FromDatabaseSlice converts a slice of database Tasks to domain Tasks.
func (m *TaskMapper) FromDatabaseSlice(dbTasks []sqlite.Task) []Task {
	domainTasks := make([]Task, len(dbTasks))
	for i, task := range dbTasks {
		domainTasks[i] = m.FromDatabase(task)
	}
	return domainTasks
}

// TimeEntryMapper handles conversion between domain and database TimeEntry models.
type TimeEntryMapper struct{}

// NewTimeEntryMapper creates a new TimeEntryMapper instance.
func NewTimeEntryMapper() *TimeEntryMapper {
	return &TimeEntryMapper{}
}

// ToDatabase converts a domain TimeEntry to a database TimeEntry.
func (m *TimeEntryMapper) ToDatabase(domainEntry TimeEntry) sqlite.TimeEntry {
	return sqlite.TimeEntry{
		ID:        domainEntry.ID,
		TaskID:    domainEntry.TaskID,
		StartTime: domainEntry.StartTime,
		EndTime:   domainEntry.EndTime,
	}
}

// FromDatabase converts a database TimeEntry to a domain TimeEntry.
func (m *TimeEntryMapper) FromDatabase(dbEntry sqlite.TimeEntry) TimeEntry {
	return TimeEntry{
		ID:        dbEntry.ID,
		TaskID:    dbEntry.TaskID,
		StartTime: dbEntry.StartTime,
		EndTime:   dbEntry.EndTime,
	}
}

// ToDatabaseSlice converts a slice of domain TimeEntries to database TimeEntries.
func (m *TimeEntryMapper) ToDatabaseSlice(domainEntries []TimeEntry) []sqlite.TimeEntry {
	dbEntries := make([]sqlite.TimeEntry, len(domainEntries))
	for i, entry := range domainEntries {
		dbEntries[i] = m.ToDatabase(entry)
	}
	return dbEntries
}

// FromDatabaseSlice converts a slice of database TimeEntries to domain TimeEntries.
func (m *TimeEntryMapper) FromDatabaseSlice(dbEntries []sqlite.TimeEntry) []TimeEntry {
	domainEntries := make([]TimeEntry, len(dbEntries))
	for i, entry := range dbEntries {
		domainEntries[i] = m.FromDatabase(entry)
	}
	return domainEntries
}

// SearchOptionsMapper handles conversion between domain and database SearchOptions.
type SearchOptionsMapper struct{}

// NewSearchOptionsMapper creates a new SearchOptionsMapper instance.
func NewSearchOptionsMapper() *SearchOptionsMapper {
	return &SearchOptionsMapper{}
}

// ToDatabase converts domain SearchOptions to database SearchOptions.
func (m *SearchOptionsMapper) ToDatabase(domainOpts SearchOptions) sqlite.SearchOptions {
	return sqlite.SearchOptions{
		StartTime: domainOpts.StartTime,
		EndTime:   domainOpts.EndTime,
		TaskID:    domainOpts.TaskID,
		TaskName:  domainOpts.TaskName,
	}
}

// FromDatabase converts database SearchOptions to domain SearchOptions.
func (m *SearchOptionsMapper) FromDatabase(dbOpts sqlite.SearchOptions) SearchOptions {
	return SearchOptions{
		StartTime: dbOpts.StartTime,
		EndTime:   dbOpts.EndTime,
		TaskID:    dbOpts.TaskID,
		TaskName:  dbOpts.TaskName,
	}
}

// Mapper provides a unified interface for all mapping operations.
type Mapper struct {
	Task          *TaskMapper
	TimeEntry     *TimeEntryMapper
	SearchOptions *SearchOptionsMapper
}

// NewMapper creates a new Mapper instance with all sub-mappers.
func NewMapper() *Mapper {
	return &Mapper{
		Task:          NewTaskMapper(),
		TimeEntry:     NewTimeEntryMapper(),
		SearchOptions: NewSearchOptionsMapper(),
	}
}