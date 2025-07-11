package cli

import (
	"context"
	"fmt"
	"time-tracker/internal/api"
	"time-tracker/internal/domain"
	"time-tracker/internal/errors"
)

// StopCommand handles the stop command
type StopCommand struct {
	api api.API
}

// NewStopCommand creates a new stop command handler
func NewStopCommand(app *App) *StopCommand {
	return &StopCommand{api: app.api}
}

// Execute runs the stop command
func (c *StopCommand) Execute(ctx context.Context, args []string) error {
	if len(args) != 0 {
		return errors.NewInvalidInputError("command", "stop", "usage: tt stop")
	}
	return c.stopRunningTasks(ctx)
}

// stopRunningTasks marks all running tasks as complete
func (c *StopCommand) stopRunningTasks(ctx context.Context) error {
	// Search for tasks with no end time
	opts := domain.SearchOptions{}
	entries, err := c.api.SearchTimeEntries(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to search for running tasks: %w", err)
	}

	now := timeNow()
	for _, entry := range entries {
		if entry.EndTime == nil {
			entry.EndTime = &now
			if err := c.api.UpdateTimeEntry(ctx, entry.ID, entry.StartTime, entry.EndTime, entry.TaskID); err != nil {
				return fmt.Errorf("failed to update task %d: %w", entry.ID, err)
			}
		}
	}

	fmt.Println("All running tasks have been stopped")
	return nil
}