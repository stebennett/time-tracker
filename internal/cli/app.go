package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"time-tracker/internal/api"
	"time-tracker/internal/repository/sqlite"
)

// timeNow is a variable that can be replaced in tests
var timeNow = time.Now

// App represents the main CLI application
type App struct {
	api      api.API
	registry *CommandRegistry
}

// GetDatabasePath returns the path to the SQLite database file
func GetDatabasePath() (string, error) {
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

// NewApp creates a new CLI application instance with dependency injection
func NewApp(api api.API) *App {
	app := &App{
		api: api,
	}
	app.registry = NewCommandRegistry(app)
	return app
}

// NewAppWithDefaultRepository creates a new CLI application instance with the default SQLite repository
// This maintains backward compatibility and is used for production
func NewAppWithDefaultRepository() (*App, error) {
	// Get database path
	dbPath, err := GetDatabasePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get database path: %w", err)
	}

	// Initialize SQLite repository
	repo, err := sqlite.New(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Create API instance
	apiInstance := api.New(repo)

	app := &App{
		api: apiInstance,
	}
	app.registry = NewCommandRegistry(app)
	return app, nil
}

// Run executes the CLI application with the given arguments
func (a *App) Run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("%s", a.registry.GetUsage())
	}

	commandName := args[0]
	commandArgs := args[1:]

	return a.registry.Execute(commandName, commandArgs)
}

// parseTimeShorthand parses time shorthand like "30m", "2h", "1d", etc.
func parseTimeShorthand(shorthand string) (time.Duration, error) {
	re := regexp.MustCompile(`^(\d+)(m|h|d|w|mo|y)$`)
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
