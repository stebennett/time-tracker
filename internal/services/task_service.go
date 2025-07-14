package services

import (
	"context"
	"strings"
	"time-tracker/internal/domain"
	"time-tracker/internal/errors"
	"time-tracker/internal/repository/sqlite"
	"time-tracker/internal/validation"
)

// taskServiceImpl implements the TaskService interface
type taskServiceImpl struct {
	repo            sqlite.Repository
	timeService     TimeService
	mapper          *domain.Mapper
	taskValidator   *validation.TaskValidator
}

// NewTaskService creates a new TaskService instance
func NewTaskService(repo sqlite.Repository, timeService TimeService) TaskService {
	return &taskServiceImpl{
		repo:          repo,
		timeService:   timeService,
		mapper:        domain.NewMapper(),
		taskValidator: validation.NewTaskValidator(),
	}
}

// validateAndTrimTaskName validates and trims a task name
func (t *taskServiceImpl) validateAndTrimTaskName(name string) (string, error) {
	trimmedName := strings.TrimSpace(name)
	if err := t.taskValidator.ValidateTaskName(trimmedName); err != nil {
		return "", errors.NewValidationError("invalid task name", err)
	}
	return trimmedName, nil
}

// findTaskByName searches for a task by exact name match
func (t *taskServiceImpl) findTaskByName(ctx context.Context, name string) (*domain.Task, error) {
	dbTasks, err := t.repo.ListTasks(ctx)
	if err != nil {
		return nil, err
	}

	// Look for exact match
	for _, dbTask := range dbTasks {
		if dbTask.TaskName == name {
			domainTask := t.mapper.Task.FromDatabase(*dbTask)
			return &domainTask, nil
		}
	}
	
	return nil, nil // Not found
}

// CreateTask creates a new task with the given name
func (t *taskServiceImpl) CreateTask(ctx context.Context, name string) (*domain.Task, error) {
	// Validate task name
	trimmedName, err := t.validateAndTrimTaskName(name)
	if err != nil {
		return nil, err
	}

	// Create database task
	dbTask := &sqlite.Task{
		TaskName: trimmedName,
	}
	
	err = t.repo.CreateTask(ctx, dbTask)
	if err != nil {
		return nil, err
	}

	// Convert to domain model
	domainTask := t.mapper.Task.FromDatabase(*dbTask)
	return &domainTask, nil
}

// GetTask retrieves a task by its ID
func (t *taskServiceImpl) GetTask(ctx context.Context, id int64) (*domain.Task, error) {
	// Validate task ID
	if id <= 0 {
		return nil, errors.NewValidationError("invalid task ID", nil)
	}

	dbTask, err := t.repo.GetTask(ctx, id)
	if err != nil {
		return nil, err
	}

	// Convert to domain model
	domainTask := t.mapper.Task.FromDatabase(*dbTask)
	return &domainTask, nil
}

// UpdateTask updates a task's name
func (t *taskServiceImpl) UpdateTask(ctx context.Context, id int64, name string) (*domain.Task, error) {
	// Validate task ID
	if id <= 0 {
		return nil, errors.NewValidationError("invalid task ID", nil)
	}

	// Validate task name
	trimmedName, err := t.validateAndTrimTaskName(name)
	if err != nil {
		return nil, err
	}

	// Check if task exists
	_, err = t.repo.GetTask(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update task
	dbTask := &sqlite.Task{
		ID:       id,
		TaskName: trimmedName,
	}
	
	err = t.repo.UpdateTask(ctx, dbTask)
	if err != nil {
		return nil, err
	}

	// Convert to domain model
	domainTask := t.mapper.Task.FromDatabase(*dbTask)
	return &domainTask, nil
}

// DeleteTaskWithEntries deletes a task and all its time entries
func (t *taskServiceImpl) DeleteTaskWithEntries(ctx context.Context, id int64) error {
	// Validate task ID
	if id <= 0 {
		return errors.NewValidationError("invalid task ID", nil)
	}

	// Check if task exists
	_, err := t.repo.GetTask(ctx, id)
	if err != nil {
		return err
	}

	// Delete all time entries for this task
	searchOpts := sqlite.SearchOptions{
		TaskID: &id,
	}
	
	entries, err := t.repo.SearchTimeEntries(ctx, searchOpts)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		err = t.repo.DeleteTimeEntry(ctx, entry.ID)
		if err != nil {
			return err
		}
	}

	// Delete the task
	return t.repo.DeleteTask(ctx, id)
}

// StartNewTask creates or finds a task and starts a new time entry for it, stopping any running tasks
func (t *taskServiceImpl) StartNewTask(ctx context.Context, name string) (*TaskSession, error) {
	// Validate task name
	trimmedName, err := t.validateAndTrimTaskName(name)
	if err != nil {
		return nil, err
	}

	// Stop all running tasks first
	_, err = t.StopAllRunningTasks(ctx)
	if err != nil {
		return nil, err
	}

	// Try to find existing task first
	task, err := t.findTaskByName(ctx, trimmedName)
	if err != nil {
		return nil, err
	}

	// Create new task if not found
	if task == nil {
		task, err = t.CreateTask(ctx, trimmedName)
		if err != nil {
			return nil, err
		}
	}

	// Create new time entry
	timeEntry, err := t.timeService.CreateTimeEntry(ctx, task.ID)
	if err != nil {
		return nil, err
	}

	// Create task session
	return t.CreateTaskSession(task, timeEntry), nil
}

// ResumeTask resumes work on an existing task by creating a new time entry, stopping any running tasks
func (t *taskServiceImpl) ResumeTask(ctx context.Context, id int64) (*TaskSession, error) {
	// Validate task ID
	if id <= 0 {
		return nil, errors.NewValidationError("invalid task ID", nil)
	}

	// Get the task
	task, err := t.GetTask(ctx, id)
	if err != nil {
		return nil, err
	}

	// Stop all running tasks first
	_, err = t.StopAllRunningTasks(ctx)
	if err != nil {
		return nil, err
	}

	// Create new time entry
	timeEntry, err := t.timeService.CreateTimeEntry(ctx, task.ID)
	if err != nil {
		return nil, err
	}

	// Create task session
	return t.CreateTaskSession(task, timeEntry), nil
}

// GetCurrentSession returns the currently running task session, if any
func (t *taskServiceImpl) GetCurrentSession(ctx context.Context) (*TaskSession, error) {
	// Get running time entries
	runningEntries, err := t.timeService.GetRunningEntries(ctx)
	if err != nil {
		return nil, err
	}

	if len(runningEntries) == 0 {
		return nil, nil
	}

	// Use the most recent running entry
	entry := runningEntries[0]
	
	// Get the task for this entry
	task, err := t.GetTask(ctx, entry.TaskID)
	if err != nil {
		return nil, err
	}

	// Create task session
	return t.CreateTaskSession(task, entry), nil
}

// CreateTaskSession creates a TaskSession from a task and time entry
func (t *taskServiceImpl) CreateTaskSession(task *domain.Task, entry *domain.TimeEntry) *TaskSession {
	duration := t.timeService.CalculateDuration(entry.StartTime, entry.EndTime)
	
	return &TaskSession{
		Task:      task,
		TimeEntry: entry,
		Duration:  duration,
	}
}

// StopAllRunningTasks stops all currently running tasks
func (t *taskServiceImpl) StopAllRunningTasks(ctx context.Context) ([]*domain.TimeEntry, error) {
	return t.timeService.StopRunningEntries(ctx)
}