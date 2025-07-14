package cli

import (
	"context"
	"fmt"
	"strings"
	"time-tracker/internal/api"
	"time-tracker/internal/errors"
)

// StartCommand handles the start command
type StartCommand struct {
	businessAPI  api.BusinessAPI
	errorHandler *ErrorHandler
}

// NewStartCommand creates a new start command handler
func NewStartCommand(app *App) *StartCommand {
	return &StartCommand{
		businessAPI:  app.businessAPI,
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
	// Check if there's a current running task to maintain backward compatibility
	currentSession, err := c.businessAPI.GetCurrentSession(ctx)
	hasRunningTask := err == nil && currentSession != nil

	// Use BusinessAPI's StartNewTask which handles stopping running tasks automatically
	session, err := c.businessAPI.StartNewTask(ctx, taskName)
	if err != nil {
		return c.errorHandler.Handle("start task", err)
	}

	// Show stopping message if there was a running task (for e2e test compatibility)
	if hasRunningTask {
		fmt.Println("All running tasks have been stopped")
	}
	fmt.Printf("Started new task: %s\n", session.Task.TaskName)
	return nil
}
