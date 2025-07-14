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

// SummaryCommand handles the summary command
type SummaryCommand struct {
	businessAPI api.BusinessAPI
}

// NewSummaryCommand creates a new summary command handler
func NewSummaryCommand(app *App) *SummaryCommand {
	return &SummaryCommand{businessAPI: app.businessAPI}
}

// Execute runs the summary command
func (c *SummaryCommand) Execute(ctx context.Context, args []string) error {
	return c.summaryTask(ctx, args)
}

// summaryTask implements the summary command
func (c *SummaryCommand) summaryTask(ctx context.Context, args []string) error {
	// Determine time range and search text
	var timeRange string
	var textFilter string

	if len(args) > 0 {
		// Check if first argument is a time shorthand
		if _, err := parseTimeShorthand(args[0]); err == nil {
			// Time shorthand found
			timeRange = args[0]

			// If there are more arguments, use them as search text
			if len(args) > 1 {
				textFilter = strings.Join(args[1:], " ")
			}
		} else {
			// No time shorthand, treat all arguments as search text
			textFilter = strings.Join(args, " ")
		}
	}

	// Search for tasks using BusinessAPI
	tasks, err := c.businessAPI.SearchTasks(ctx, timeRange, textFilter, api.SortByName)
	if err != nil {
		return fmt.Errorf("failed to search tasks: %w", err)
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks found matching the criteria.")
		return nil
	}

	// If only one task, show its summary directly
	if len(tasks) == 1 {
		return c.showTaskSummary(ctx, tasks[0].Task.ID)
	}

	// Multiple tasks found, let user choose
	fmt.Println("Select a task to summarize:")
	for i, task := range tasks {
		fmt.Printf("%d. %s\n", i+1, task.Task.TaskName)
	}
	fmt.Print("Enter number to summarize, or 'q' to quit: ")

	// Read user input
	var input string
	fmt.Fscanln(os.Stdin, &input)
	if input == "q" || input == "Q" {
		fmt.Println("Summary cancelled.")
		return nil
	}
	idx, err := strconv.Atoi(input)
	if err != nil || idx < 1 || idx > len(tasks) {
		return errors.NewInvalidInputError("selection", input, "invalid selection")
	}
	selectedTask := tasks[idx-1]

	return c.showTaskSummary(ctx, selectedTask.Task.ID)
}

// showTaskSummary displays a detailed summary for a specific task
func (c *SummaryCommand) showTaskSummary(ctx context.Context, taskID int64) error {
	// Get task summary using BusinessAPI
	summary, err := c.businessAPI.GetTaskSummary(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task summary: %w", err)
	}

	// Print summary header
	fmt.Printf("\nSummary for: %s\n", summary.Task.TaskName)
	fmt.Println(strings.Repeat("=", len(summary.Task.TaskName)+12))

	// Print table header
	fmt.Printf("%-20s %-20s %-15s %s\n", "Start Time", "End Time", "Duration", "Status")
	fmt.Println(strings.Repeat("-", 75))

	// Print each session
	for _, entry := range summary.TimeEntries {
		startStr := entry.StartTime.Format("2006-01-02 15:04:05")
		var endStr, durationStr, status string

		if entry.EndTime != nil {
			endStr = entry.EndTime.Format("2006-01-02 15:04:05")
			duration := entry.EndTime.Sub(entry.StartTime)
			hours := int(duration.Hours())
			minutes := int(duration.Minutes()) % 60
			durationStr = fmt.Sprintf("%dh %dm", hours, minutes)
			status = "Completed"
		} else {
			endStr = "running"
			duration := timeNow().Sub(entry.StartTime)
			hours := int(duration.Hours())
			minutes := int(duration.Minutes()) % 60
			durationStr = fmt.Sprintf("%dh %dm", hours, minutes)
			status = "Running"
		}

		fmt.Printf("%-20s %-20s %-15s %s\n", startStr, endStr, durationStr, status)
	}

	// Print summary footer
	fmt.Println(strings.Repeat("-", 75))

	// Format time range
	earliestStr := summary.FirstEntry.Format("2006-01-02 15:04:05")
	latestStr := summary.LastEntry.Format("2006-01-02 15:04:05")

	fmt.Printf("Total Sessions: %d", summary.SessionCount)
	if summary.RunningCount > 0 {
		fmt.Printf(" (%d running)", summary.RunningCount)
	}
	fmt.Printf("\n")
	fmt.Printf("Time Range: %s to %s\n", earliestStr, latestStr)
	fmt.Printf("Total Time: %s\n", summary.TotalTime)

	return nil
}