package cli

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"
	"time-tracker/internal/api"
	"time-tracker/internal/domain"
	"time-tracker/internal/errors"
)

// ResumeCommand handles the resume command
type ResumeCommand struct {
	api api.API
}

// NewResumeCommand creates a new resume command handler
func NewResumeCommand(app *App) *ResumeCommand {
	return &ResumeCommand{api: app.api}
}

// Execute runs the resume command
func (c *ResumeCommand) Execute(args []string) error {
	return c.resumeTask(args)
}

// resumeTask implements the resume scenario
func (c *ResumeCommand) resumeTask(args []string) error {
	// Determine time range (default: today)
	var startTime time.Time
	now := timeNow()
	if len(args) > 0 {
		dur, err := parseTimeShorthand(args[0])
		if err != nil {
			return errors.NewInvalidInputError("time_shorthand", args[0], "invalid time shorthand")
		}
		startTime = now.Add(-dur)
	} else {
		y, m, d := now.Date()
		startTime = time.Date(y, m, d, 0, 0, 0, 0, now.Location())
	}

	// Find all time entries in the period, most recent first
	opts := domain.SearchOptions{StartTime: &startTime, EndTime: &now}
	entries, err := c.api.SearchTimeEntries(opts)
	if err != nil {
		return fmt.Errorf("failed to search time entries: %w", err)
	}
	if len(entries) == 0 {
		fmt.Println("No tasks found in the selected period.")
		return nil
	}

	// Group by task, show most recent entry for each task
	taskMap := make(map[int64]*domain.TimeEntry)
	for i := len(entries) - 1; i >= 0; i-- { // reverse for most recent
		entry := entries[i]
		if _, ok := taskMap[entry.TaskID]; !ok {
			taskMap[entry.TaskID] = entry
		}
	}

	// Build a slice for display
	var taskIDs []int64
	for id := range taskMap {
		taskIDs = append(taskIDs, id)
	}
	// Sort by most recent start time
	sort.Slice(taskIDs, func(i, j int) bool {
		return taskMap[taskIDs[i]].StartTime.After(taskMap[taskIDs[j]].StartTime)
	})

	fmt.Println("Select a task to resume:")
	for i, id := range taskIDs {
		task, _ := c.api.GetTask(id)
		last := taskMap[id].StartTime.Format("2006-01-02 15:04:05")
		fmt.Printf("%d. %s (last worked: %s)\n", i+1, task.TaskName, last)
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
	if err != nil || idx < 1 || idx > len(taskIDs) {
		return errors.NewInvalidInputError("selection", input, "invalid selection")
	}
	selectedTaskID := taskIDs[idx-1]

	// Stop any running tasks
	if err := c.stopRunningTasks(); err != nil {
		return err
	}

	// Create a new time entry for the selected task
	_, err = c.api.CreateTimeEntry(selectedTaskID, timeNow(), nil)
	if err != nil {
		return fmt.Errorf("failed to resume task: %w", err)
	}
	task, _ := c.api.GetTask(selectedTaskID)
	fmt.Printf("Resumed task: %s\n", task.TaskName)
	return nil
}

// stopRunningTasks marks all running tasks as complete
func (c *ResumeCommand) stopRunningTasks() error {
	// Search for tasks with no end time
	opts := domain.SearchOptions{}
	entries, err := c.api.SearchTimeEntries(opts)
	if err != nil {
		return fmt.Errorf("failed to search for running tasks: %w", err)
	}

	now := timeNow()
	for _, entry := range entries {
		if entry.EndTime == nil {
			entry.EndTime = &now
			if err := c.api.UpdateTimeEntry(entry.ID, entry.StartTime, entry.EndTime, entry.TaskID); err != nil {
				return fmt.Errorf("failed to update task %d: %w", entry.ID, err)
			}
		}
	}

	fmt.Println("All running tasks have been stopped")
	return nil
}