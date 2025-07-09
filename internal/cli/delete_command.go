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

// DeleteCommand handles the delete command
type DeleteCommand struct {
	api api.API
}

// NewDeleteCommand creates a new delete command handler
func NewDeleteCommand(app *App) *DeleteCommand {
	return &DeleteCommand{api: app.api}
}

// Execute runs the delete command
func (c *DeleteCommand) Execute(args []string) error {
	return c.deleteTask(args)
}

// deleteTask implements the delete command
func (c *DeleteCommand) deleteTask(args []string) error {
	// Determine time range (default: last 24h or user-supplied duration)
	now := timeNow()
	var startTime time.Time
	var filterText string

	if len(args) > 0 {
		if dur, err := parseTimeShorthand(args[0]); err == nil {
			startTime = now.Add(-dur)
			if len(args) > 1 {
				filterText = strings.Join(args[1:], " ")
			}
		} else {
			// Not a valid duration, treat as text filter
			filterText = strings.Join(args, " ")
			startTime = now.Add(-24 * time.Hour)
		}
	} else {
		startTime = now.Add(-24 * time.Hour)
	}

	// Build search options like listTasks
	opts := domain.SearchOptions{StartTime: &startTime, EndTime: &now}
	if filterText != "" {
		opts.TaskName = &filterText
	}

	entries, err := c.api.SearchTimeEntries(opts)
	if err != nil {
		return fmt.Errorf("failed to search time entries: %w", err)
	}
	if len(entries) == 0 {
		fmt.Println("No tasks found to delete.")
		return nil
	}

	// Group by task, get the last entry for each task
	taskMap := make(map[int64]*domain.TimeEntry)
	for _, entry := range entries {
		if existing, ok := taskMap[entry.TaskID]; !ok || entry.StartTime.After(existing.StartTime) {
			taskMap[entry.TaskID] = entry
		}
	}

	// Build a slice for display and sort by last worked time ascending
	taskIDs := make([]int64, 0, len(taskMap))
	for id := range taskMap {
		taskIDs = append(taskIDs, id)
	}
	sort.Slice(taskIDs, func(i, j int) bool {
		return taskMap[taskIDs[i]].StartTime.Before(taskMap[taskIDs[j]].StartTime)
	})

	if len(taskIDs) == 0 {
		fmt.Println("No tasks found to delete.")
		return nil
	}

	fmt.Println("Select a task to delete:")
	for i, id := range taskIDs {
		task, _ := c.api.GetTask(id)
		last := taskMap[id].StartTime.Format("2006-01-02 15:04:05")
		fmt.Printf("%d. %s (last worked: %s)\n", i+1, task.TaskName, last)
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
	if err != nil || idx < 1 || idx > len(taskIDs) {
		return errors.NewInvalidInputError("selection", input, "invalid selection")
	}
	selectedTaskID := taskIDs[idx-1]
	task, _ := c.api.GetTask(selectedTaskID)

	// Delete all time entries for the selected task only
	entryOpts := domain.SearchOptions{TaskID: &selectedTaskID}
	taskEntries, err := c.api.SearchTimeEntries(entryOpts)
	if err != nil {
		return fmt.Errorf("failed to get time entries for task: %w", err)
	}
	for _, entry := range taskEntries {
		err := c.api.DeleteTimeEntry(entry.ID)
		if err != nil {
			return fmt.Errorf("failed to delete time entry %d: %w", entry.ID, err)
		}
	}

	// Delete the task itself
	if err := c.api.DeleteTask(selectedTaskID); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	fmt.Printf("Deleted task: %s\n", task.TaskName)
	return nil
}