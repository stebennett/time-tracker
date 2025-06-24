package api

import (
	"errors"
	"time"
	"time-tracker/internal/repository/sqlite"
)

// API defines the interface for all task and time entry operations.
type API interface {
	// Task operations
	CreateTask(name string) (*sqlite.Task, error)
	GetTask(id int64) (*sqlite.Task, error)
	ListTasks() ([]*sqlite.Task, error)
	UpdateTask(id int64, name string) error
	DeleteTask(id int64) error

	// Time entry operations
	CreateTimeEntry(taskID int64, startTime time.Time, endTime *time.Time) (*sqlite.TimeEntry, error)
	GetTimeEntry(id int64) (*sqlite.TimeEntry, error)
	ListTimeEntries() ([]*sqlite.TimeEntry, error)
	UpdateTimeEntry(id int64, startTime time.Time, endTime *time.Time, taskID int64) error
	DeleteTimeEntry(id int64) error

	// Business logic implementations
	StartTask(taskID int64) (*sqlite.TimeEntry, error)
	StopTask(entryID int64) error
	ResumeTask(taskID int64) (*sqlite.TimeEntry, error)
	GetCurrentlyRunningTask() (*sqlite.TimeEntry, error)
	ListTodayTasks() ([]*sqlite.Task, error)
}

type apiImpl struct {
	repo sqlite.Repository
}

// New creates a new API instance.
func New(repo sqlite.Repository) API {
	return &apiImpl{repo: repo}
}

// Task CRUD implementations
func (a *apiImpl) CreateTask(name string) (*sqlite.Task, error) {
	task := &sqlite.Task{TaskName: name}
	err := a.repo.CreateTask(task)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (a *apiImpl) GetTask(id int64) (*sqlite.Task, error) {
	return a.repo.GetTask(id)
}

func (a *apiImpl) ListTasks() ([]*sqlite.Task, error) {
	return a.repo.ListTasks()
}

func (a *apiImpl) UpdateTask(id int64, name string) error {
	task, err := a.repo.GetTask(id)
	if err != nil {
		return err
	}
	task.TaskName = name
	return a.repo.UpdateTask(task)
}

func (a *apiImpl) DeleteTask(id int64) error {
	return a.repo.DeleteTask(id)
}

// TimeEntry CRUD implementations
func (a *apiImpl) CreateTimeEntry(taskID int64, startTime time.Time, endTime *time.Time) (*sqlite.TimeEntry, error) {
	entry := &sqlite.TimeEntry{TaskID: taskID, StartTime: startTime, EndTime: endTime}
	err := a.repo.CreateTimeEntry(entry)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func (a *apiImpl) GetTimeEntry(id int64) (*sqlite.TimeEntry, error) {
	return a.repo.GetTimeEntry(id)
}

func (a *apiImpl) ListTimeEntries() ([]*sqlite.TimeEntry, error) {
	return a.repo.ListTimeEntries()
}

func (a *apiImpl) UpdateTimeEntry(id int64, startTime time.Time, endTime *time.Time, taskID int64) error {
	entry, err := a.repo.GetTimeEntry(id)
	if err != nil {
		return err
	}
	entry.StartTime = startTime
	entry.EndTime = endTime
	entry.TaskID = taskID
	return a.repo.UpdateTimeEntry(entry)
}

func (a *apiImpl) DeleteTimeEntry(id int64) error {
	return a.repo.DeleteTimeEntry(id)
}

// Business logic implementations

// StartTask stops any running tasks and starts a new one for the given taskID.
func (a *apiImpl) StartTask(taskID int64) (*sqlite.TimeEntry, error) {
	// Stop all running tasks
	running, _ := a.GetCurrentlyRunningTask()
	if running != nil {
		err := a.StopTask(running.ID)
		if err != nil {
			return nil, err
		}
	}
	entry := &sqlite.TimeEntry{
		TaskID:    taskID,
		StartTime: time.Now(),
	}
	err := a.repo.CreateTimeEntry(entry)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

// StopTask sets the EndTime for the given entryID to now.
func (a *apiImpl) StopTask(entryID int64) error {
	entry, err := a.repo.GetTimeEntry(entryID)
	if err != nil {
		return err
	}
	if entry.EndTime != nil {
		return errors.New("task already stopped")
	}
	now := time.Now()
	entry.EndTime = &now
	return a.repo.UpdateTimeEntry(entry)
}

// ResumeTask stops any running tasks and starts a new entry for the given taskID.
func (a *apiImpl) ResumeTask(taskID int64) (*sqlite.TimeEntry, error) {
	running, _ := a.GetCurrentlyRunningTask()
	if running != nil {
		err := a.StopTask(running.ID)
		if err != nil {
			return nil, err
		}
	}
	entry := &sqlite.TimeEntry{
		TaskID:    taskID,
		StartTime: time.Now(),
	}
	err := a.repo.CreateTimeEntry(entry)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

// GetCurrentlyRunningTask returns the currently running time entry, or error if none.
func (a *apiImpl) GetCurrentlyRunningTask() (*sqlite.TimeEntry, error) {
	entries, err := a.repo.SearchTimeEntries(sqlite.SearchOptions{})
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.EndTime == nil {
			return entry, nil
		}
	}
	return nil, errors.New("no running task")
}

// ListTodayTasks returns all tasks with time entries for today.
func (a *apiImpl) ListTodayTasks() ([]*sqlite.Task, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	entries, err := a.repo.SearchTimeEntries(sqlite.SearchOptions{
		StartTime: &startOfDay,
		EndTime:   &endOfDay,
	})
	if err != nil {
		return nil, err
	}
	taskMap := make(map[int64]*sqlite.Task)
	for _, entry := range entries {
		task, err := a.repo.GetTask(entry.TaskID)
		if err == nil {
			taskMap[task.ID] = task
		}
	}
	tasks := make([]*sqlite.Task, 0, len(taskMap))
	for _, t := range taskMap {
		tasks = append(tasks, t)
	}
	return tasks, nil
}
