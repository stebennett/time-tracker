package cli

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"time-tracker/internal/api"
	"time-tracker/internal/config"
	"time-tracker/internal/errors"
	"time-tracker/internal/repository/sqlite"
)

// timeNow is a variable that can be replaced in tests
var timeNow = time.Now

// App represents the main CLI application
type App struct {
	businessAPI api.BusinessAPI // BusinessAPI for all commands
	config      *config.Config
	registry    *CommandRegistry
}


// NewApp creates a new CLI application instance with BusinessAPI
func NewApp(businessAPI api.BusinessAPI) *App {
	app := &App{
		businessAPI: businessAPI,
		config:      nil, // Will be set by caller
	}
	app.registry = NewCommandRegistry(app)
	return app
}

// NewAppWithConfig creates a new CLI application instance with BusinessAPI and configuration
func NewAppWithConfig(businessAPI api.BusinessAPI, cfg *config.Config) *App {
	app := &App{
		businessAPI: businessAPI,
		config:      cfg,
	}
	app.registry = NewCommandRegistry(app)
	return app
}

// NewAppWithDefaultRepository creates a new CLI application instance with the default SQLite repository
// This is used for production and creates only the BusinessAPI
func NewAppWithDefaultRepository() (*App, error) {
	// Load configuration
	loader := config.NewLoader()
	cfg, err := loader.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get database path from configuration
	dbPath := cfg.GetDatabasePath()

	// Initialize SQLite repository with configuration
	repo, err := sqlite.NewWithConfig(dbPath, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Create BusinessAPI instance
	businessAPI := api.NewBusinessAPI(repo)

	app := &App{
		businessAPI: businessAPI,
		config:      cfg,
	}
	app.registry = NewCommandRegistry(app)
	return app, nil
}

// Run executes the CLI application with the given arguments
func (a *App) Run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.NewInvalidInputError("command", "", a.registry.GetUsage())
	}

	commandName := args[0]
	commandArgs := args[1:]

	return a.registry.Execute(ctx, commandName, commandArgs)
}

// parseTimeShorthand parses time shorthand like "30m", "2h", "1d", etc.
func parseTimeShorthand(shorthand string) (time.Duration, error) {
	re := regexp.MustCompile(`^(\d+)(m|h|d|w|mo|y)$`)
	matches := re.FindStringSubmatch(shorthand)
	if matches == nil {
		return 0, errors.NewInvalidInputError("time_format", shorthand, "invalid time format")
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, errors.NewInvalidInputError("time_number", shorthand, "invalid number in time format")
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
		return 0, errors.NewInvalidInputError("time_unit", unit, "invalid time unit")
	}

	return duration, nil
}
