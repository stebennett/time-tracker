package cli

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
	"time-tracker/internal/api"
	"time-tracker/internal/config"
	"time-tracker/internal/domain"
)

// ListCommand handles the list command
type ListCommand struct {
	api    api.API
	config *config.Config
}

// NewListCommand creates a new list command handler
func NewListCommand(app *App) *ListCommand {
	return &ListCommand{
		api:    app.api,
		config: app.config,
	}
}

// Execute runs the list command
func (c *ListCommand) Execute(ctx context.Context, args []string) error {
	return c.listTasks(ctx, args)
}

// listTasks handles the list command with various options
func (c *ListCommand) listTasks(ctx context.Context, args []string) error {
	opts := domain.SearchOptions{}

	// If no arguments, list all tasks
	if len(args) == 0 {
		entries, err := c.api.ListTimeEntries(ctx)
		if err != nil {
			return fmt.Errorf("failed to list tasks: %w", err)
		}
		return c.printEntries(ctx, entries)
	}

	// Check if first argument is a time shorthand
	if duration, err := parseTimeShorthand(args[0]); err == nil {
		// Time shorthand found, set time range
		now := timeNow()
		startTime := now.Add(-duration)
		opts.StartTime = &startTime
		opts.EndTime = &now

		// If there are more arguments, use them as search text
		if len(args) > 1 {
			text := strings.Join(args[1:], " ")
			opts.TaskName = &text
		}
	} else {
		// No time shorthand, treat all arguments as search text
		text := strings.Join(args, " ")
		opts.TaskName = &text
	}

	// Search for entries
	entries, err := c.api.SearchTimeEntries(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to search tasks: %w", err)
	}

	return c.printEntries(ctx, entries)
}

// printEntries prints one line per task in the format:
// startTime - endTime (duration): taskName
// Where startTime and endTime are from the last time entry for the task, and endTime is 'running' if the entry is running.
func (c *ListCommand) printEntries(ctx context.Context, entries []*domain.TimeEntry) error {
	if len(entries) == 0 {
		fmt.Println("No tasks found")
		return nil
	}

	// Group entries by TaskID and find the last entry for each task
	type lastEntryInfo struct {
		TaskName  string
		StartTime time.Time
		EndTime   *time.Time
		IsRunning bool
	}
	lastEntryMap := make(map[int64]*lastEntryInfo)
	for _, entry := range entries {
		task, err := c.api.GetTask(ctx, entry.TaskID)
		if err != nil {
			return fmt.Errorf("failed to get task for entry %d: %w", entry.ID, err)
		}
		info, ok := lastEntryMap[entry.TaskID]
		if !ok || entry.StartTime.After(info.StartTime) {
			lastEntryMap[entry.TaskID] = &lastEntryInfo{
				TaskName:  task.TaskName,
				StartTime: entry.StartTime,
				EndTime:   entry.EndTime,
				IsRunning: entry.EndTime == nil,
			}
		}
	}

	// Collect and sort: non-running tasks by StartTime ascending, running tasks last
	var runningInfos, finishedInfos []*lastEntryInfo
	for _, info := range lastEntryMap {
		if info.IsRunning {
			runningInfos = append(runningInfos, info)
		} else {
			finishedInfos = append(finishedInfos, info)
		}
	}
	sort.Slice(finishedInfos, func(i, j int) bool {
		return finishedInfos[i].StartTime.Before(finishedInfos[j].StartTime)
	})
	sort.Slice(runningInfos, func(i, j int) bool {
		return runningInfos[i].StartTime.Before(runningInfos[j].StartTime)
	})
	infos := append(finishedInfos, runningInfos...)
	for _, info := range infos {
		// Use configured time format
		timeFormat := c.getTimeFormat()
		startStr := info.StartTime.Format(timeFormat)
		var endStr string
		var duration time.Duration
		if info.IsRunning {
			endStr = c.getRunningStatus()
			duration = timeNow().Sub(info.StartTime)
		} else if info.EndTime != nil {
			endStr = info.EndTime.Format(timeFormat)
			duration = info.EndTime.Sub(info.StartTime)
		}
		hours := int(duration.Hours())
		minutes := int(duration.Minutes()) % 60
		
		// Truncate task name if configured
		taskName := c.truncateTaskName(info.TaskName)
		fmt.Printf("%s - %s (%dh %dm): %s\n", startStr, endStr, hours, minutes, taskName)
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