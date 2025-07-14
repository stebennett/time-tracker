package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time-tracker/internal/api"
	"time-tracker/internal/errors"
)

// ResumeCommand handles the resume command
type ResumeCommand struct {
	businessAPI api.BusinessAPI
}

// NewResumeCommand creates a new resume command handler
func NewResumeCommand(app *App) *ResumeCommand {
	return &ResumeCommand{businessAPI: app.businessAPI}
}

// Execute runs the resume command
func (c *ResumeCommand) Execute(ctx context.Context, args []string) error {
	return c.resumeTask(ctx, args)
}

// resumeTask implements the resume scenario
func (c *ResumeCommand) resumeTask(ctx context.Context, args []string) error {
	// Determine time range (default: today)
	var timeRange string
	if len(args) > 0 {
		// Validate the time shorthand
		if _, err := parseTimeShorthand(args[0]); err != nil {
			return errors.NewInvalidInputError("time_shorthand", args[0], "invalid time shorthand")
		}
		timeRange = args[0]
	} else {
		timeRange = "1d" // Default to today
	}

	// Search for tasks in the period using BusinessAPI
	tasks, err := c.businessAPI.SearchTasks(ctx, timeRange, "", api.SortByRecentFirst)
	if err != nil {
		return fmt.Errorf("failed to search tasks: %w", err)
	}
	if len(tasks) == 0 {
		fmt.Println("No tasks found in the selected period.")
		return nil
	}

	fmt.Println("Select a task to resume:")
	for i, task := range tasks {
		fmt.Printf("%d. %s (last worked: %s)\n", i+1, task.Task.TaskName, task.LastWorked)
	}
	fmt.Print("Enter number to resume, or 'q' to quit: ")

	// Read user input
	var input string
	fmt.Fscanln(os.Stdin, &input)
	if input == "q" || input == "Q" {
		fmt.Println("Resume cancelled.")
		return nil
	}
	idx, err := strconv.Atoi(input)
	if err != nil || idx < 1 || idx > len(tasks) {
		return errors.NewInvalidInputError("selection", input, "invalid selection")
	}
	selectedTask := tasks[idx-1]

	// Resume the selected task using BusinessAPI
	session, err := c.businessAPI.ResumeTask(ctx, selectedTask.Task.ID)
	if err != nil {
		return fmt.Errorf("failed to resume task: %w", err)
	}
	
	fmt.Printf("Resumed task: %s\n", session.Task.TaskName)
	return nil
}

