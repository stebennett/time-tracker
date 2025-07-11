package config

import (
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Config holds all configuration options for the time tracker application
type Config struct {
	Database    DatabaseConfig
	Time        TimeConfig
	Validation  ValidationConfig
	Display     DisplayConfig
	Application ApplicationConfig
	Commands    CommandsConfig
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Dir            string        `env:"TT_DB_DIR"`
	Filename       string        `env:"TT_DB_FILENAME"`
	QueryTimeout   time.Duration `env:"TT_DB_QUERY_TIMEOUT"`
	WriteTimeout   time.Duration `env:"TT_DB_WRITE_TIMEOUT"`
	DirPermissions uint32        `env:"TT_DB_DIR_PERMISSIONS"`
}

// TimeConfig holds time formatting configuration
type TimeConfig struct {
	DisplayFormat string `env:"TT_TIME_DISPLAY_FORMAT"`
}

// ValidationConfig holds validation rules configuration
type ValidationConfig struct {
	TaskNameMinLength int           `env:"TT_VALIDATION_TASK_NAME_MIN"`
	TaskNameMaxLength int           `env:"TT_VALIDATION_TASK_NAME_MAX"`
	MaxDuration       time.Duration `env:"TT_VALIDATION_MAX_DURATION"`
}

// DisplayConfig holds display formatting configuration
type DisplayConfig struct {
	SummaryWidth    int    `env:"TT_DISPLAY_SUMMARY_WIDTH"`
	RunningStatus   string `env:"TT_DISPLAY_RUNNING_STATUS"`
	DateOnly        bool   `env:"TT_DISPLAY_DATE_ONLY"`
}

// ApplicationConfig holds application-level configuration
type ApplicationConfig struct {
	Timeout time.Duration `env:"TT_APP_TIMEOUT"`
	Verbose bool          `env:"TT_APP_VERBOSE"`
}

// CommandsConfig holds command-specific defaults
type CommandsConfig struct {
	ListDefaultFormat   string `env:"TT_LIST_DEFAULT_FORMAT"`
	OutputDefaultFormat string `env:"TT_OUTPUT_DEFAULT_FORMAT"`
}

// NewConfig creates a new configuration with sensible defaults
func NewConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	defaultDBDir := filepath.Join(homeDir, ".tt")
	
	return &Config{
		Database: DatabaseConfig{
			Dir:            defaultDBDir,
			Filename:       "tt.db",
			QueryTimeout:   10 * time.Second,
			WriteTimeout:   5 * time.Second,
			DirPermissions: 0755,
		},
		Time: TimeConfig{
			DisplayFormat: "2006-01-02 15:04:05",
		},
		Validation: ValidationConfig{
			TaskNameMinLength: 1,
			TaskNameMaxLength: 255,
			MaxDuration:       24 * time.Hour,
		},
		Display: DisplayConfig{
			SummaryWidth:  75,
			RunningStatus: "running",
			DateOnly:      false,
		},
		Application: ApplicationConfig{
			Timeout: 60 * time.Second,
			Verbose: false,
		},
		Commands: CommandsConfig{
			ListDefaultFormat:   "table",
			OutputDefaultFormat: "csv",
		},
	}
}

// GetDatabasePath returns the full path to the database file
func (c *Config) GetDatabasePath() string {
	return filepath.Join(c.Database.Dir, c.Database.Filename)
}

// GetQueryTimeout returns the database query timeout
func (c *Config) GetQueryTimeout() time.Duration {
	return c.Database.QueryTimeout
}

// GetWriteTimeout returns the database write timeout
func (c *Config) GetWriteTimeout() time.Duration {
	return c.Database.WriteTimeout
}

// LoadFromEnvironment loads configuration from environment variables
func (c *Config) LoadFromEnvironment() error {
	// Database configuration
	if dir := os.Getenv("TT_DB_DIR"); dir != "" {
		c.Database.Dir = dir
	}
	if filename := os.Getenv("TT_DB_FILENAME"); filename != "" {
		c.Database.Filename = filename
	}
	if timeout := os.Getenv("TT_DB_QUERY_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			c.Database.QueryTimeout = d
		}
	}
	if timeout := os.Getenv("TT_DB_WRITE_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			c.Database.WriteTimeout = d
		}
	}
	if perms := os.Getenv("TT_DB_DIR_PERMISSIONS"); perms != "" {
		if p, err := strconv.ParseUint(perms, 8, 32); err == nil {
			c.Database.DirPermissions = uint32(p)
		}
	}

	// Time configuration
	if format := os.Getenv("TT_TIME_DISPLAY_FORMAT"); format != "" {
		c.Time.DisplayFormat = format
	}

	// Validation configuration
	if minLen := os.Getenv("TT_VALIDATION_TASK_NAME_MIN"); minLen != "" {
		if n, err := strconv.Atoi(minLen); err == nil {
			c.Validation.TaskNameMinLength = n
		}
	}
	if maxLen := os.Getenv("TT_VALIDATION_TASK_NAME_MAX"); maxLen != "" {
		if n, err := strconv.Atoi(maxLen); err == nil {
			c.Validation.TaskNameMaxLength = n
		}
	}
	if maxDur := os.Getenv("TT_VALIDATION_MAX_DURATION"); maxDur != "" {
		if d, err := time.ParseDuration(maxDur); err == nil {
			c.Validation.MaxDuration = d
		}
	}

	// Display configuration
	if width := os.Getenv("TT_DISPLAY_SUMMARY_WIDTH"); width != "" {
		if w, err := strconv.Atoi(width); err == nil {
			c.Display.SummaryWidth = w
		}
	}
	if status := os.Getenv("TT_DISPLAY_RUNNING_STATUS"); status != "" {
		c.Display.RunningStatus = status
	}
	if dateOnly := os.Getenv("TT_DISPLAY_DATE_ONLY"); dateOnly != "" {
		if b, err := strconv.ParseBool(dateOnly); err == nil {
			c.Display.DateOnly = b
		}
	}

	// Application configuration
	if timeout := os.Getenv("TT_APP_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			c.Application.Timeout = d
		}
	}
	if verbose := os.Getenv("TT_APP_VERBOSE"); verbose != "" {
		if b, err := strconv.ParseBool(verbose); err == nil {
			c.Application.Verbose = b
		}
	}

	// Commands configuration
	if format := os.Getenv("TT_LIST_DEFAULT_FORMAT"); format != "" {
		c.Commands.ListDefaultFormat = format
	}
	if format := os.Getenv("TT_OUTPUT_DEFAULT_FORMAT"); format != "" {
		c.Commands.OutputDefaultFormat = format
	}

	return nil
}

// Validate validates the configuration and returns any errors
func (c *Config) Validate() error {
	// Validate database configuration
	if c.Database.Dir == "" {
		return &ConfigError{Field: "database.dir", Message: "database directory cannot be empty"}
	}
	if c.Database.Filename == "" {
		return &ConfigError{Field: "database.filename", Message: "database filename cannot be empty"}
	}
	if c.Database.QueryTimeout <= 0 {
		return &ConfigError{Field: "database.query_timeout", Message: "query timeout must be positive"}
	}
	if c.Database.WriteTimeout <= 0 {
		return &ConfigError{Field: "database.write_timeout", Message: "write timeout must be positive"}
	}

	// Validate time configuration
	if c.Time.DisplayFormat == "" {
		return &ConfigError{Field: "time.display_format", Message: "display format cannot be empty"}
	}

	// Validate validation configuration
	if c.Validation.TaskNameMinLength < 1 {
		return &ConfigError{Field: "validation.task_name_min_length", Message: "task name minimum length must be at least 1"}
	}
	if c.Validation.TaskNameMaxLength < c.Validation.TaskNameMinLength {
		return &ConfigError{Field: "validation.task_name_max_length", Message: "task name maximum length must be greater than minimum length"}
	}
	if c.Validation.MaxDuration <= 0 {
		return &ConfigError{Field: "validation.max_duration", Message: "max duration must be positive"}
	}

	// Validate display configuration
	if c.Display.SummaryWidth < 10 {
		return &ConfigError{Field: "display.summary_width", Message: "summary width must be at least 10"}
	}
	if c.Display.RunningStatus == "" {
		return &ConfigError{Field: "display.running_status", Message: "running status text cannot be empty"}
	}

	// Validate application configuration
	if c.Application.Timeout <= 0 {
		return &ConfigError{Field: "application.timeout", Message: "application timeout must be positive"}
	}

	return nil
}

// ConfigError represents a configuration validation error
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return e.Field + ": " + e.Message
}