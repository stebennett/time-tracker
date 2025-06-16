package cli

import (
	"fmt"
	"strings"
	"time"

	"time-tracker/internal/repository/sqlite"
)

// App represents the main CLI application
type App struct {
	repo sqlite.Repository
}

// NewApp creates a new CLI application instance
func NewApp() (*App, error) {
	// Initialize SQLite repository
	repo, err := sqlite.New("time-tracker.db")
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
		return fmt.Errorf("usage: tt \"your text here\" or tt stop")
	}

	// Only stop if exactly one argument and it is "stop"
	if len(args) == 1 && args[0] == "stop" {
		return a.stopRunningTasks()
	}

	// Otherwise, treat as a new task
	text := strings.Join(args, " ")
	return a.createNewTask(text)
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