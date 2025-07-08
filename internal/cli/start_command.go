package cli

import (
	"fmt"
	"strings"
	"time-tracker/internal/api"
	"time-tracker/internal/repository/sqlite"
)

// StartCommand handles the start command
type StartCommand struct {
	api api.API
}

// NewStartCommand creates a new start command handler
func NewStartCommand(app *App) *StartCommand {
	return &StartCommand{api: app.api}
}

// Execute runs the start command
func (c *StartCommand) Execute(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: tt start \"your text here\"")
	}
	text := strings.Join(args, " ")
	return c.createNewTask(text)
}

// createNewTask creates a new task
func (c *StartCommand) createNewTask(taskName string) error {
	// First, stop any running tasks
	if err := c.stopRunningTasks(); err != nil {
		return err
	}

	// Always create a new task
	task, err := c.api.CreateTask(taskName)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	// Create new time entry
	now := timeNow()
	_, err = c.api.CreateTimeEntry(task.ID, now, nil)
	if err != nil {
		return fmt.Errorf("failed to create new task: %w", err)
	}

	fmt.Printf("Started new task: %s\n", taskName)
	return nil
}

// stopRunningTasks marks all running tasks as complete
func (c *StartCommand) stopRunningTasks() error {
	// Search for tasks with no end time
	opts := sqlite.SearchOptions{}
	entries, err := c.api.SearchTimeEntries(opts)
	if err != nil {
		return fmt.Errorf("failed to search for running tasks: %w", err)
	}

	now := timeNow()
	for _, entry := range entries {
		if entry.EndTime == nil {
			entry.EndTime = &now
			if err := c.api.UpdateTimeEntry(entry.ID, entry.StartTime, entry.EndTime, entry.TaskID); err != nil {
				return fmt.Errorf("failed to update task %d: %w", entry.ID, err)
			}
		}
	}

	fmt.Println("All running tasks have been stopped")
	return nil
}