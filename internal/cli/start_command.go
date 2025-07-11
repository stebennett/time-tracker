package cli

import (
	"context"
	"fmt"
	"strings"
	"time-tracker/internal/api"
	"time-tracker/internal/domain"
	"time-tracker/internal/errors"
)

// StartCommand handles the start command
type StartCommand struct {
	api          api.API
	errorHandler *ErrorHandler
}

// NewStartCommand creates a new start command handler
func NewStartCommand(app *App) *StartCommand {
	return &StartCommand{
		api:          app.api,
		errorHandler: NewErrorHandler(),
	}
}

// Execute runs the start command
func (c *StartCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return errors.NewInvalidInputError("command", "start", "usage: tt start \"your text here\"")
	}
	text := strings.Join(args, " ")
	return c.createNewTask(ctx, text)
}

// createNewTask creates a new task
func (c *StartCommand) createNewTask(ctx context.Context, taskName string) error {
	// First, stop any running tasks
	if err := c.stopRunningTasks(ctx); err != nil {
		return err
	}

	// Always create a new task
	task, err := c.api.CreateTask(ctx, taskName)
	if err != nil {
		return c.errorHandler.Handle("create task", err)
	}

	// Create new time entry
	now := timeNow()
	_, err = c.api.CreateTimeEntry(ctx, task.ID, now, nil)
	if err != nil {
		return c.errorHandler.Handle("create time entry", err)
	}

	fmt.Printf("Started new task: %s\n", taskName)
	return nil
}


// stopRunningTasks marks all running tasks as complete
func (c *StartCommand) stopRunningTasks(ctx context.Context) error {
	// Search for tasks with no end time
	opts := domain.SearchOptions{}
	entries, err := c.api.SearchTimeEntries(ctx, opts)
	if err != nil {
		return c.errorHandler.Handle("search for running tasks", err)
	}

	now := timeNow()
	for _, entry := range entries {
		if entry.EndTime == nil {
			entry.EndTime = &now
			if err := c.api.UpdateTimeEntry(ctx, entry.ID, entry.StartTime, entry.EndTime, entry.TaskID); err != nil {
				return c.errorHandler.Handle(fmt.Sprintf("update task %d", entry.ID), err)
			}
		}
	}

	fmt.Println("All running tasks have been stopped")
	return nil
}