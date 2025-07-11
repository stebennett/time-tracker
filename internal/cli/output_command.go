package cli

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"time-tracker/internal/api"
	"time-tracker/internal/errors"
)

// OutputCommand handles the output command
type OutputCommand struct {
	api api.API
}

// NewOutputCommand creates a new output command handler
func NewOutputCommand(app *App) *OutputCommand {
	return &OutputCommand{api: app.api}
}

// Execute runs the output command
func (c *OutputCommand) Execute(ctx context.Context, args []string) error {
	return c.outputTasks(ctx, args)
}

// outputTasks outputs tasks in the specified format
func (c *OutputCommand) outputTasks(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.NewInvalidInputError("command", "output", "usage: tt output format=csv")
	}

	// Parse format option
	format := args[0]
	if !strings.HasPrefix(format, "format=") {
		return errors.NewInvalidInputError("format", format, "invalid format option")
	}

	format = strings.TrimPrefix(format, "format=")
	switch format {
	case "csv":
		return c.outputCSV(ctx)
	default:
		return errors.NewInvalidInputError("format", format, "unsupported format")
	}
}

// outputCSV outputs all tasks in CSV format
func (c *OutputCommand) outputCSV(ctx context.Context) error {
	// Get all tasks
	entries, err := c.api.ListTimeEntries(ctx)
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	// Create CSV writer
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write header
	header := []string{"ID", "Start Time", "End Time", "Duration (hours)", "Task Name"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write entries
	for _, entry := range entries {
		// Format start time
		startTime := entry.StartTime.Format(time.RFC3339)

		// Format end time
		var endTime string
		var duration float64
		if entry.EndTime != nil {
			endTime = entry.EndTime.Format(time.RFC3339)
			duration = entry.EndTime.Sub(entry.StartTime).Hours()
		}

		task, err := c.api.GetTask(ctx, entry.TaskID)
		if err != nil {
			return fmt.Errorf("failed to get task for entry %d: %w", entry.ID, err)
		}

		// Write row
		row := []string{
			strconv.FormatInt(entry.ID, 10),
			startTime,
			endTime,
			fmt.Sprintf("%.2f", duration),
			task.TaskName,
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}