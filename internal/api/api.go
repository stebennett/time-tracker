package api

import (
	"time"
	"time-tracker/internal/domain"
	"time-tracker/internal/errors"
	"time-tracker/internal/repository/sqlite"
	"time-tracker/internal/validation"
)

// API defines the interface for all task and time entry operations.
type API interface {
	// Task operations
	CreateTask(name string) (*domain.Task, error)
	GetTask(id int64) (*domain.Task, error)
	ListTasks() ([]*domain.Task, error)
	UpdateTask(id int64, name string) error
	DeleteTask(id int64) error

	// Time entry operations
	CreateTimeEntry(taskID int64, startTime time.Time, endTime *time.Time) (*domain.TimeEntry, error)
	GetTimeEntry(id int64) (*domain.TimeEntry, error)
	ListTimeEntries() ([]*domain.TimeEntry, error)
	SearchTimeEntries(opts domain.SearchOptions) ([]*domain.TimeEntry, error)
	UpdateTimeEntry(id int64, startTime time.Time, endTime *time.Time, taskID int64) error
	DeleteTimeEntry(id int64) error

	// Business logic implementations
	StartTask(taskID int64) (*domain.TimeEntry, error)
	StopTask(entryID int64) error
	ResumeTask(taskID int64) (*domain.TimeEntry, error)
	GetCurrentlyRunningTask() (*domain.TimeEntry, error)
	ListTodayTasks() ([]*domain.Task, error)
}

type apiImpl struct {
	repo            sqlite.Repository
	mapper          *domain.Mapper
	taskValidator   *validation.TaskValidator
	timeEntryValidator *validation.TimeEntryValidator
}

// New creates a new API instance.
func New(repo sqlite.Repository) API {
	return &apiImpl{
		repo:            repo,
		mapper:          domain.NewMapper(),
		taskValidator:   validation.NewTaskValidator(),
		timeEntryValidator: validation.NewTimeEntryValidator(),
	}
}

// Task CRUD implementations
func (a *apiImpl) CreateTask(name string) (*domain.Task, error) {
	// Validate input
	if err := a.taskValidator.ValidateTaskForCreation(name); err != nil {
		return nil, err
	}
	
	// Get cleaned name
	cleanedName, err := a.taskValidator.GetValidTaskName(name)
	if err != nil {
		return nil, err
	}
	
	dbTask := &sqlite.Task{TaskName: cleanedName}
	err = a.repo.CreateTask(dbTask)
	if err != nil {
		return nil, err
	}
	domainTask := a.mapper.Task.FromDatabase(*dbTask)
	return &domainTask, nil
}

func (a *apiImpl) GetTask(id int64) (*domain.Task, error) {
	// Validate input
	if err := a.taskValidator.ValidateTaskID(id); err != nil {
		return nil, err
	}
	
	dbTask, err := a.repo.GetTask(id)
	if err != nil {
		return nil, err
	}
	domainTask := a.mapper.Task.FromDatabase(*dbTask)
	return &domainTask, nil
}

func (a *apiImpl) ListTasks() ([]*domain.Task, error) {
	dbTasks, err := a.repo.ListTasks()
	if err != nil {
		return nil, err
	}
	domainTasks := make([]*domain.Task, len(dbTasks))
	for i, dbTask := range dbTasks {
		domainTask := a.mapper.Task.FromDatabase(*dbTask)
		domainTasks[i] = &domainTask
	}
	return domainTasks, nil
}

func (a *apiImpl) UpdateTask(id int64, name string) error {
	// Validate input
	if err := a.taskValidator.ValidateTaskForUpdate(id, name); err != nil {
		return err
	}
	
	// Get cleaned name
	cleanedName, err := a.taskValidator.GetValidTaskName(name)
	if err != nil {
		return err
	}
	
	dbTask, err := a.repo.GetTask(id)
	if err != nil {
		return err
	}
	dbTask.TaskName = cleanedName
	return a.repo.UpdateTask(dbTask)
}

func (a *apiImpl) DeleteTask(id int64) error {
	// Validate input
	if err := a.taskValidator.ValidateTaskID(id); err != nil {
		return err
	}
	
	return a.repo.DeleteTask(id)
}

// TimeEntry CRUD implementations
func (a *apiImpl) CreateTimeEntry(taskID int64, startTime time.Time, endTime *time.Time) (*domain.TimeEntry, error) {
	// Validate input
	if err := a.timeEntryValidator.ValidateTimeEntryForCreation(taskID, startTime, endTime); err != nil {
		return nil, err
	}
	
	dbEntry := &sqlite.TimeEntry{TaskID: taskID, StartTime: startTime, EndTime: endTime}
	err := a.repo.CreateTimeEntry(dbEntry)
	if err != nil {
		return nil, err
	}
	domainEntry := a.mapper.TimeEntry.FromDatabase(*dbEntry)
	return &domainEntry, nil
}

func (a *apiImpl) GetTimeEntry(id int64) (*domain.TimeEntry, error) {
	// Validate input
	if err := a.timeEntryValidator.ValidateTimeEntryID(id); err != nil {
		return nil, err
	}
	
	dbEntry, err := a.repo.GetTimeEntry(id)
	if err != nil {
		return nil, err
	}
	domainEntry := a.mapper.TimeEntry.FromDatabase(*dbEntry)
	return &domainEntry, nil
}

func (a *apiImpl) ListTimeEntries() ([]*domain.TimeEntry, error) {
	dbEntries, err := a.repo.ListTimeEntries()
	if err != nil {
		return nil, err
	}
	domainEntries := make([]*domain.TimeEntry, len(dbEntries))
	for i, dbEntry := range dbEntries {
		domainEntry := a.mapper.TimeEntry.FromDatabase(*dbEntry)
		domainEntries[i] = &domainEntry
	}
	return domainEntries, nil
}

func (a *apiImpl) SearchTimeEntries(opts domain.SearchOptions) ([]*domain.TimeEntry, error) {
	// Validate input
	if err := a.timeEntryValidator.ValidateSearchOptions(opts); err != nil {
		return nil, err
	}
	
	dbOpts := a.mapper.SearchOptions.ToDatabase(opts)
	dbEntries, err := a.repo.SearchTimeEntries(dbOpts)
	if err != nil {
		return nil, err
	}
	domainEntries := make([]*domain.TimeEntry, len(dbEntries))
	for i, dbEntry := range dbEntries {
		domainEntry := a.mapper.TimeEntry.FromDatabase(*dbEntry)
		domainEntries[i] = &domainEntry
	}
	return domainEntries, nil
}

func (a *apiImpl) UpdateTimeEntry(id int64, startTime time.Time, endTime *time.Time, taskID int64) error {
	// Validate input
	if err := a.timeEntryValidator.ValidateTimeEntryForUpdate(id, taskID, startTime, endTime); err != nil {
		return err
	}
	
	dbEntry, err := a.repo.GetTimeEntry(id)
	if err != nil {
		return err
	}
	dbEntry.StartTime = startTime
	dbEntry.EndTime = endTime
	dbEntry.TaskID = taskID
	return a.repo.UpdateTimeEntry(dbEntry)
}

func (a *apiImpl) DeleteTimeEntry(id int64) error {
	// Validate input
	if err := a.timeEntryValidator.ValidateTimeEntryID(id); err != nil {
		return err
	}
	
	return a.repo.DeleteTimeEntry(id)
}

// Business logic implementations

// StartTask stops any running tasks and starts a new one for the given taskID.
func (a *apiImpl) StartTask(taskID int64) (*domain.TimeEntry, error) {
	// Validate input
	if err := a.taskValidator.ValidateTaskID(taskID); err != nil {
		return nil, err
	}
	
	// Stop all running tasks
	running, _ := a.GetCurrentlyRunningTask()
	if running != nil {
		err := a.StopTask(running.ID)
		if err != nil {
			return nil, err
		}
	}
	dbEntry := &sqlite.TimeEntry{
		TaskID:    taskID,
		StartTime: time.Now(),
	}
	err := a.repo.CreateTimeEntry(dbEntry)
	if err != nil {
		return nil, err
	}
	domainEntry := a.mapper.TimeEntry.FromDatabase(*dbEntry)
	return &domainEntry, nil
}

// StopTask sets the EndTime for the given entryID to now.
func (a *apiImpl) StopTask(entryID int64) error {
	// Validate input
	if err := a.timeEntryValidator.ValidateTimeEntryID(entryID); err != nil {
		return err
	}
	
	dbEntry, err := a.repo.GetTimeEntry(entryID)
	if err != nil {
		return err
	}
	if dbEntry.EndTime != nil {
		return errors.NewValidationError("task already stopped", nil)
	}
	now := time.Now()
	dbEntry.EndTime = &now
	return a.repo.UpdateTimeEntry(dbEntry)
}

// ResumeTask stops any running tasks and starts a new entry for the given taskID.
func (a *apiImpl) ResumeTask(taskID int64) (*domain.TimeEntry, error) {
	// Validate input
	if err := a.taskValidator.ValidateTaskID(taskID); err != nil {
		return nil, err
	}
	
	running, _ := a.GetCurrentlyRunningTask()
	if running != nil {
		err := a.StopTask(running.ID)
		if err != nil {
			return nil, err
		}
	}
	dbEntry := &sqlite.TimeEntry{
		TaskID:    taskID,
		StartTime: time.Now(),
	}
	err := a.repo.CreateTimeEntry(dbEntry)
	if err != nil {
		return nil, err
	}
	domainEntry := a.mapper.TimeEntry.FromDatabase(*dbEntry)
	return &domainEntry, nil
}

// GetCurrentlyRunningTask returns the currently running time entry, or error if none.
func (a *apiImpl) GetCurrentlyRunningTask() (*domain.TimeEntry, error) {
	dbEntries, err := a.repo.SearchTimeEntries(sqlite.SearchOptions{})
	if err != nil {
		return nil, err
	}
	for _, dbEntry := range dbEntries {
		if dbEntry.EndTime == nil {
			domainEntry := a.mapper.TimeEntry.FromDatabase(*dbEntry)
			return &domainEntry, nil
		}
	}
	return nil, errors.NewNotFoundError("running task", "")
}

// ListTodayTasks returns all tasks with time entries for today.
func (a *apiImpl) ListTodayTasks() ([]*domain.Task, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	dbEntries, err := a.repo.SearchTimeEntries(sqlite.SearchOptions{
		StartTime: &startOfDay,
		EndTime:   &endOfDay,
	})
	if err != nil {
		return nil, err
	}
	taskMap := make(map[int64]*domain.Task)
	for _, dbEntry := range dbEntries {
		dbTask, err := a.repo.GetTask(dbEntry.TaskID)
		if err == nil {
			domainTask := a.mapper.Task.FromDatabase(*dbTask)
			taskMap[domainTask.ID] = &domainTask
		}
	}
	tasks := make([]*domain.Task, 0, len(taskMap))
	for _, t := range taskMap {
		tasks = append(tasks, t)
	}
	return tasks, nil
}
