package cli

import (
	"fmt"
	"time-tracker/internal/api"
	"time-tracker/internal/repository/sqlite"
)

// CurrentCommand handles the current command
type CurrentCommand struct {
	api api.API
}

// NewCurrentCommand creates a new current command handler
func NewCurrentCommand(app *App) *CurrentCommand {
	return &CurrentCommand{api: app.api}
}

// Execute runs the current command
func (c *CurrentCommand) Execute(args []string) error {
	return c.showCurrentTask()
}

// showCurrentTask displays the currently running task
func (c *CurrentCommand) showCurrentTask() error {
	// Search for tasks with no end time
	opts := sqlite.SearchOptions{}
	entries, err := c.api.SearchTimeEntries(opts)
	if err != nil {
		return fmt.Errorf("failed to search for running tasks: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("No task is currently running")
		return nil
	}

	// Get the most recent running task
	entry := entries[0]
	duration := timeNow().Sub(entry.StartTime)
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60

	task, err := c.api.GetTask(entry.TaskID)
	if err != nil {
		return fmt.Errorf("failed to get task for entry %d: %w", entry.ID, err)
	}

	fmt.Printf("Current task: %s (running for %dh %dm)\n",
		task.TaskName, hours, minutes)
	return nil
}