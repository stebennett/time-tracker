package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"time-tracker/internal/api"
	"time-tracker/internal/errors"
)

// DeleteCommand handles the delete command
type DeleteCommand struct {
	businessAPI api.BusinessAPI
}

// NewDeleteCommand creates a new delete command handler
func NewDeleteCommand(app *App) *DeleteCommand {
	return &DeleteCommand{businessAPI: app.businessAPI}
}

// Execute runs the delete command
func (c *DeleteCommand) Execute(ctx context.Context, args []string) error {
	return c.deleteTask(ctx, args)
}

// deleteTask implements the delete command
func (c *DeleteCommand) deleteTask(ctx context.Context, args []string) error {
	// Determine time range and search text
	var timeRange string
	var textFilter string

	if len(args) > 0 {
		if _, err := parseTimeShorthand(args[0]); err == nil {
			// Time shorthand found
			timeRange = args[0]
			if len(args) > 1 {
				textFilter = strings.Join(args[1:], " ")
			}
		} else {
			// Not a valid duration, treat as text filter
			textFilter = strings.Join(args, " ")
			timeRange = "1d" // Default to last 24h
		}
	} else {
		timeRange = "1d" // Default to last 24h
	}

	// Search for tasks using BusinessAPI
	tasks, err := c.businessAPI.SearchTasks(ctx, timeRange, textFilter, api.SortByRecentFirst)
	if err != nil {
		return fmt.Errorf("failed to search tasks: %w", err)
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks found to delete.")
		return nil
	}

	fmt.Println("Select a task to delete:")
	for i, task := range tasks {
		fmt.Printf("%d. %s (last worked: %s)\n", i+1, task.Task.TaskName, task.LastWorked)
	}
	fmt.Print("Enter number to delete, or 'q' to quit: ")

	// Read user input
	var input string
	fmt.Fscanln(os.Stdin, &input)
	if input == "q" || input == "Q" {
		fmt.Println("Delete cancelled.")
		return nil
	}
	idx, err := strconv.Atoi(input)
	if err != nil || idx < 1 || idx > len(tasks) {
		return errors.NewInvalidInputError("selection", input, "invalid selection")
	}
	selectedTask := tasks[idx-1]

	// Delete the task and all its time entries using BusinessAPI
	if err := c.businessAPI.DeleteTaskWithEntries(ctx, selectedTask.Task.ID); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	fmt.Printf("Deleted task: %s\n", selectedTask.Task.TaskName)
	return nil
}