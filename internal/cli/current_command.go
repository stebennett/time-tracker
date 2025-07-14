package cli

import (
	"context"
	"errors"
	"fmt"
	"time-tracker/internal/api"
	appErrors "time-tracker/internal/errors"
)

// CurrentCommand handles the current command
type CurrentCommand struct {
	businessAPI api.BusinessAPI
}

// NewCurrentCommand creates a new current command handler
func NewCurrentCommand(app *App) *CurrentCommand {
	return &CurrentCommand{businessAPI: app.businessAPI}
}

// Execute runs the current command
func (c *CurrentCommand) Execute(ctx context.Context, args []string) error {
	return c.showCurrentTask(ctx)
}

// showCurrentTask displays the currently running task
func (c *CurrentCommand) showCurrentTask(ctx context.Context) error {
	// Use BusinessAPI's GetCurrentSession
	session, err := c.businessAPI.GetCurrentSession(ctx)
	if err != nil {
		// Check if it's a "not found" error
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) && appErr.IsType(appErrors.ErrorTypeNotFound) {
			fmt.Println("No task is currently running")
			return nil
		}
		return fmt.Errorf("failed to get current session: %w", err)
	}

	fmt.Printf("Current task: %s (%s)\n", session.Task.TaskName, session.Duration)
	return nil
}