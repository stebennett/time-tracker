package cli

import (
	"context"
	"fmt"
	"time-tracker/internal/api"
	"time-tracker/internal/errors"
)

// StopCommand handles the stop command
type StopCommand struct {
	businessAPI api.BusinessAPI
}

// NewStopCommand creates a new stop command handler
func NewStopCommand(app *App) *StopCommand {
	return &StopCommand{businessAPI: app.businessAPI}
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
	// Use BusinessAPI's StopAllRunningTasks
	_, err := c.businessAPI.StopAllRunningTasks(ctx)
	if err != nil {
		return fmt.Errorf("failed to stop running tasks: %w", err)
	}

	// Always show the same message for backward compatibility with e2e tests
	fmt.Println("All running tasks have been stopped")
	return nil
}