package cli

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time-tracker/internal/api"
	"time-tracker/internal/config"
)

// ListCommand handles the list command
type ListCommand struct {
	businessAPI api.BusinessAPI
	config      *config.Config
}

// NewListCommand creates a new list command handler
func NewListCommand(app *App) *ListCommand {
	return &ListCommand{
		businessAPI: app.businessAPI,
		config:      app.config,
	}
}

// Execute runs the list command
func (c *ListCommand) Execute(ctx context.Context, args []string) error {
	return c.listTasks(ctx, args)
}

// listTasks handles the list command with various options
func (c *ListCommand) listTasks(ctx context.Context, args []string) error {
	var timeRange string
	var textFilter string

	// If no arguments, list all tasks
	if len(args) == 0 {
		timeRange = ""
		textFilter = ""
	} else {
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

	// Search for time entries with task information using BusinessAPI
	entries, err := c.businessAPI.SearchTimeEntries(ctx, timeRange, textFilter)
	if err != nil {
		return fmt.Errorf("failed to search tasks: %w", err)
	}

	return c.printTimeEntries(ctx, entries)
}

// printTimeEntries prints one line per time entry in the format:
// startTime - endTime (duration): taskName
// Where endTime is 'running' if the entry is running.
func (c *ListCommand) printTimeEntries(ctx context.Context, entries []*api.TimeEntryWithTask) error {
	if len(entries) == 0 {
		fmt.Println("No tasks found")
		return nil
	}

	// Sort entries: non-running tasks by StartTime ascending, running tasks last
	var runningEntries, finishedEntries []*api.TimeEntryWithTask
	for _, entry := range entries {
		if entry.TimeEntry.EndTime == nil {
			runningEntries = append(runningEntries, entry)
		} else {
			finishedEntries = append(finishedEntries, entry)
		}
	}
	
	sort.Slice(finishedEntries, func(i, j int) bool {
		return finishedEntries[i].TimeEntry.StartTime.Before(finishedEntries[j].TimeEntry.StartTime)
	})
	sort.Slice(runningEntries, func(i, j int) bool {
		return runningEntries[i].TimeEntry.StartTime.Before(runningEntries[j].TimeEntry.StartTime)
	})
	
	sortedEntries := append(finishedEntries, runningEntries...)
	
	for _, entry := range sortedEntries {
		// Use configured time format
		timeFormat := c.getTimeFormat()
		startStr := entry.TimeEntry.StartTime.Format(timeFormat)
		var endStr string
		
		if entry.TimeEntry.EndTime == nil {
			endStr = c.getRunningStatus()
		} else {
			endStr = entry.TimeEntry.EndTime.Format(timeFormat)
		}
		
		// Truncate task name if configured
		taskName := c.truncateTaskName(entry.Task.TaskName)
		fmt.Printf("%s - %s (%s): %s\n", startStr, endStr, entry.Duration, taskName)
	}

	return nil
}

// getTimeFormat returns the configured time format or default
func (c *ListCommand) getTimeFormat() string {
	if c.config != nil && c.config.Time.DisplayFormat != "" {
		return c.config.Time.DisplayFormat
	}
	return "2006-01-02 15:04:05" // Default format
}

// getRunningStatus returns the configured running status text or default
func (c *ListCommand) getRunningStatus() string {
	if c.config != nil && c.config.Display.RunningStatus != "" {
		return c.config.Display.RunningStatus
	}
	return "running" // Default status
}

// truncateTaskName truncates task name if it exceeds configured limit
func (c *ListCommand) truncateTaskName(taskName string) string {
	if c.config != nil && c.config.Display.SummaryWidth > 0 {
		maxLen := c.config.Display.SummaryWidth - 30 // Reserve space for time info
		if maxLen > 0 && len(taskName) > maxLen {
			return taskName[:maxLen-3] + "..."
		}
	}
	return taskName
}