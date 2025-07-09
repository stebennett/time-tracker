package cli

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"time-tracker/internal/api"
	"time-tracker/internal/domain"
	"time-tracker/internal/errors"
)

// SummaryCommand handles the summary command
type SummaryCommand struct {
	api api.API
}

// NewSummaryCommand creates a new summary command handler
func NewSummaryCommand(app *App) *SummaryCommand {
	return &SummaryCommand{api: app.api}
}

// Execute runs the summary command
func (c *SummaryCommand) Execute(args []string) error {
	return c.summaryTask(args)
}

// summaryTask implements the summary command
func (c *SummaryCommand) summaryTask(args []string) error {
	// Determine time range and search text
	var startTime *time.Time
	var searchText string
	now := timeNow()

	if len(args) > 0 {
		// Check if first argument is a time shorthand
		if duration, err := parseTimeShorthand(args[0]); err == nil {
			// Time shorthand found, set time range
			start := now.Add(-duration)
			startTime = &start

			// If there are more arguments, use them as search text
			if len(args) > 1 {
				searchText = strings.Join(args[1:], " ")
			}
		} else {
			// No time shorthand, treat all arguments as search text
			searchText = strings.Join(args, " ")
		}
	}

	// Find tasks that match the criteria
	var matchingTaskIDs []int64

	if startTime != nil {
		// If time filter is specified, find tasks that have ANY entries in the time window
		timeFilterOpts := domain.SearchOptions{
			StartTime: startTime,
			EndTime:   &now,
		}
		if searchText != "" {
			timeFilterOpts.TaskName = &searchText
		}

		timeFilterEntries, err := c.api.SearchTimeEntries(timeFilterOpts)
		if err != nil {
			return fmt.Errorf("failed to search time entries: %w", err)
		}

		// Get unique task IDs from entries in the time window
		taskIDSet := make(map[int64]bool)
		for _, entry := range timeFilterEntries {
			taskIDSet[entry.TaskID] = true
		}

		// Convert to slice
		for taskID := range taskIDSet {
			matchingTaskIDs = append(matchingTaskIDs, taskID)
		}
	} else if searchText != "" {
		// Only text filter, find tasks by name
		textFilterOpts := domain.SearchOptions{
			TaskName: &searchText,
		}

		textFilterEntries, err := c.api.SearchTimeEntries(textFilterOpts)
		if err != nil {
			return fmt.Errorf("failed to search time entries: %w", err)
		}

		// Get unique task IDs
		taskIDSet := make(map[int64]bool)
		for _, entry := range textFilterEntries {
			taskIDSet[entry.TaskID] = true
		}

		// Convert to slice
		for taskID := range taskIDSet {
			matchingTaskIDs = append(matchingTaskIDs, taskID)
		}
	} else {
		// No filters, get all tasks
		allEntries, err := c.api.ListTimeEntries()
		if err != nil {
			return fmt.Errorf("failed to list time entries: %w", err)
		}

		// Get unique task IDs
		taskIDSet := make(map[int64]bool)
		for _, entry := range allEntries {
			taskIDSet[entry.TaskID] = true
		}

		// Convert to slice
		for taskID := range taskIDSet {
			matchingTaskIDs = append(matchingTaskIDs, taskID)
		}
	}

	if len(matchingTaskIDs) == 0 {
		fmt.Println("No tasks found matching the criteria.")
		return nil
	}

	// Sort task IDs by task name for consistent ordering
	sort.Slice(matchingTaskIDs, func(i, j int) bool {
		taskI, _ := c.api.GetTask(matchingTaskIDs[i])
		taskJ, _ := c.api.GetTask(matchingTaskIDs[j])
		return taskI.TaskName < taskJ.TaskName
	})

	// If only one task, show its summary directly
	if len(matchingTaskIDs) == 1 {
		return c.showTaskSummary(matchingTaskIDs[0])
	}

	// Multiple tasks found, let user choose
	fmt.Println("Select a task to summarize:")
	for i, taskID := range matchingTaskIDs {
		task, _ := c.api.GetTask(taskID)
		fmt.Printf("%d. %s\n", i+1, task.TaskName)
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
	if err != nil || idx < 1 || idx > len(matchingTaskIDs) {
		return errors.NewInvalidInputError("selection", input, "invalid selection")
	}
	selectedTaskID := matchingTaskIDs[idx-1]

	return c.showTaskSummary(selectedTaskID)
}

// showTaskSummary displays a detailed summary for a specific task
func (c *SummaryCommand) showTaskSummary(taskID int64) error {
	// Get task details
	task, err := c.api.GetTask(taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Get all time entries for this task
	opts := domain.SearchOptions{}
	entries, err := c.api.SearchTimeEntries(opts)
	if err != nil {
		return fmt.Errorf("failed to get time entries: %w", err)
	}

	if len(entries) == 0 {
		return errors.NewNotFoundError("time entries", fmt.Sprintf("task %d", taskID))
	}

	// Sort entries by start time
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].StartTime.Before(entries[j].StartTime)
	})

	// Calculate summary statistics
	var earliestStart time.Time
	var latestEnd time.Time
	var totalDuration time.Duration
	var runningSessions int

	earliestStart = entries[0].StartTime
	latestEnd = entries[0].StartTime // Initialize with first start time

	for _, entry := range entries {
		// Update earliest start
		if entry.StartTime.Before(earliestStart) {
			earliestStart = entry.StartTime
		}

		// Calculate duration and update latest end
		if entry.EndTime != nil {
			duration := entry.EndTime.Sub(entry.StartTime)
			totalDuration += duration

			if entry.EndTime.After(latestEnd) {
				latestEnd = *entry.EndTime
			}
		} else {
			// Running session
			runningSessions++
			currentDuration := timeNow().Sub(entry.StartTime)
			totalDuration += currentDuration

			// For running sessions, use current time as latest
			if timeNow().After(latestEnd) {
				latestEnd = timeNow()
			}
		}
	}

	// Print summary header
	fmt.Printf("\nSummary for: %s\n", task.TaskName)
	fmt.Println(strings.Repeat("=", len(task.TaskName)+12))

	// Print table header
	fmt.Printf("%-20s %-20s %-15s %s\n", "Start Time", "End Time", "Duration", "Status")
	fmt.Println(strings.Repeat("-", 75))

	// Print each session
	for _, entry := range entries {
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

	// Format total duration
	totalHours := int(totalDuration.Hours())
	totalMinutes := int(totalDuration.Minutes()) % 60

	// Format time range
	earliestStr := earliestStart.Format("2006-01-02 15:04:05")
	latestStr := latestEnd.Format("2006-01-02 15:04:05")

	fmt.Printf("Total Sessions: %d", len(entries))
	if runningSessions > 0 {
		fmt.Printf(" (%d running)", runningSessions)
	}
	fmt.Printf("\n")
	fmt.Printf("Time Range: %s to %s\n", earliestStr, latestStr)
	fmt.Printf("Total Time: %dh %dm\n", totalHours, totalMinutes)

	return nil
}