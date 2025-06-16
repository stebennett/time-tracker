package cli

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"time-tracker/internal/repository/sqlite"
)

// App represents the main CLI application
type App struct {
	repo sqlite.Repository
}

// getDatabasePath returns the path to the SQLite database file
func getDatabasePath() (string, error) {
	// Check for TT_DB environment variable
	if dbPath := os.Getenv("TT_DB"); dbPath != "" {
		return dbPath, nil
	}

	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Create .tt directory if it doesn't exist
	ttDir := filepath.Join(homeDir, ".tt")
	if err := os.MkdirAll(ttDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create .tt directory: %w", err)
	}

	// Return path to tt.db in .tt directory
	return filepath.Join(ttDir, "tt.db"), nil
}

// NewApp creates a new CLI application instance
func NewApp() (*App, error) {
	// Get database path
	dbPath, err := getDatabasePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get database path: %w", err)
	}

	// Initialize SQLite repository
	repo, err := sqlite.New(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return &App{
		repo: repo,
	}, nil
}

// Run executes the CLI application with the given arguments
func (a *App) Run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: tt \"your text here\" or tt stop or tt list [time] [text] or tt current or tt output format=csv")
	}

	// Handle different commands
	switch args[0] {
	case "stop":
		return a.stopRunningTasks()
	case "list":
		return a.listTasks(args[1:])
	case "current":
		return a.showCurrentTask()
	case "output":
		return a.outputTasks(args[1:])
	default:
		// Otherwise, treat as a new task
		text := strings.Join(args, " ")
		return a.createNewTask(text)
	}
}

// parseTimeShorthand parses time shorthand like "30m", "2h", "1d", etc.
func parseTimeShorthand(shorthand string) (time.Duration, error) {
	re := regexp.MustCompile(`^(\d+)([mhdwmo]|y)$`)
	matches := re.FindStringSubmatch(shorthand)
	if matches == nil {
		return 0, fmt.Errorf("invalid time format: %s", shorthand)
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("invalid number in time format: %s", shorthand)
	}

	unit := matches[2]
	var duration time.Duration

	switch unit {
	case "m":
		duration = time.Duration(value) * time.Minute
	case "h":
		duration = time.Duration(value) * time.Hour
	case "d":
		duration = time.Duration(value) * 24 * time.Hour
	case "w":
		duration = time.Duration(value) * 7 * 24 * time.Hour
	case "mo":
		duration = time.Duration(value) * 30 * 24 * time.Hour
	case "y":
		duration = time.Duration(value) * 365 * 24 * time.Hour
	default:
		return 0, fmt.Errorf("invalid time unit: %s", unit)
	}

	return duration, nil
}

// listTasks handles the list command with various options
func (a *App) listTasks(args []string) error {
	opts := sqlite.SearchOptions{}
	
	// If no arguments, list all tasks
	if len(args) == 0 {
		entries, err := a.repo.ListTimeEntries()
		if err != nil {
			return fmt.Errorf("failed to list tasks: %w", err)
		}
		return a.printEntries(entries)
	}

	// Check if first argument is a time shorthand
	if duration, err := parseTimeShorthand(args[0]); err == nil {
		// Time shorthand found, set time range
		now := time.Now()
		startTime := now.Add(-duration)
		opts.StartTime = &startTime
		opts.EndTime = &now

		// If there are more arguments, use them as search text
		if len(args) > 1 {
			text := strings.Join(args[1:], " ")
			opts.Description = &text
		}
	} else {
		// No time shorthand, treat all arguments as search text
		text := strings.Join(args, " ")
		opts.Description = &text
	}

	// Search for entries
	entries, err := a.repo.SearchTimeEntries(opts)
	if err != nil {
		return fmt.Errorf("failed to search tasks: %w", err)
	}

	return a.printEntries(entries)
}

// printEntries prints the time entries in a formatted way
func (a *App) printEntries(entries []*sqlite.TimeEntry) error {
	if len(entries) == 0 {
		fmt.Println("No tasks found")
		return nil
	}

	for _, entry := range entries {
		// Format start time
		startTime := entry.StartTime.Format("2006-01-02 15:04:05")
		
		// Format end time or "running"
		var endTimeStr string
		if entry.EndTime == nil {
			endTimeStr = "running"
		} else {
			endTimeStr = entry.EndTime.Format("2006-01-02 15:04:05")
		}

		// Calculate duration if task is completed
		var durationStr string
		if entry.EndTime != nil {
			duration := entry.EndTime.Sub(entry.StartTime)
			hours := int(duration.Hours())
			minutes := int(duration.Minutes()) % 60
			durationStr = fmt.Sprintf(" (%dh %dm)", hours, minutes)
		}

		fmt.Printf("%s - %s%s: %s\n", startTime, endTimeStr, durationStr, entry.Description)
	}

	return nil
}

// stopRunningTasks marks all running tasks as complete
func (a *App) stopRunningTasks() error {
	// Search for tasks with no end time
	opts := sqlite.SearchOptions{}
	entries, err := a.repo.SearchTimeEntries(opts)
	if err != nil {
		return fmt.Errorf("failed to search for running tasks: %w", err)
	}

	now := time.Now()
	for _, entry := range entries {
		if entry.EndTime == nil {
			entry.EndTime = &now
			if err := a.repo.UpdateTimeEntry(entry); err != nil {
				return fmt.Errorf("failed to update task %d: %w", entry.ID, err)
			}
		}
	}

	fmt.Println("All running tasks have been stopped")
	return nil
}

// createNewTask creates a new task and stops any running tasks
func (a *App) createNewTask(description string) error {
	// First, stop any running tasks
	if err := a.stopRunningTasks(); err != nil {
		return err
	}

	// Create new task
	now := time.Now()
	entry := &sqlite.TimeEntry{
		StartTime:   now,
		Description: description,
	}

	if err := a.repo.CreateTimeEntry(entry); err != nil {
		return fmt.Errorf("failed to create new task: %w", err)
	}

	fmt.Printf("Started new task: %s\n", description)
	return nil
}

// showCurrentTask displays the currently running task
func (a *App) showCurrentTask() error {
	// Search for tasks with no end time
	opts := sqlite.SearchOptions{}
	entries, err := a.repo.SearchTimeEntries(opts)
	if err != nil {
		return fmt.Errorf("failed to search for running tasks: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("No task is currently running")
		return nil
	}

	// Get the most recent running task
	entry := entries[0]
	duration := time.Since(entry.StartTime)
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60

	fmt.Printf("Current task: %s (running for %dh %dm)\n", 
		entry.Description, hours, minutes)
	return nil
}

// outputTasks outputs tasks in the specified format
func (a *App) outputTasks(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: tt output format=csv")
	}

	// Parse format option
	format := args[0]
	if !strings.HasPrefix(format, "format=") {
		return fmt.Errorf("invalid format option: %s", format)
	}

	format = strings.TrimPrefix(format, "format=")
	switch format {
	case "csv":
		return a.outputCSV()
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// outputCSV outputs all tasks in CSV format
func (a *App) outputCSV() error {
	// Get all tasks
	entries, err := a.repo.ListTimeEntries()
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	// Create CSV writer
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write header
	header := []string{"ID", "Start Time", "End Time", "Duration (hours)", "Description"}
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

		// Write row
		row := []string{
			strconv.FormatInt(entry.ID, 10),
			startTime,
			endTime,
			fmt.Sprintf("%.2f", duration),
			entry.Description,
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
} 