package config

import (
	"strconv"
	"time"
)

// Loader handles loading configuration from multiple sources
type Loader struct {
	config *Config
}

// NewLoader creates a new configuration loader
func NewLoader() *Loader {
	return &Loader{
		config: NewConfig(),
	}
}

// Load loads configuration using the cascading strategy:
// 1. Start with defaults
// 2. Override with environment variables
// 3. Override with command line flags (handled by cobra)
func (l *Loader) Load() (*Config, error) {
	// Step 1: Start with defaults (already done in NewConfig)
	
	// Step 2: Load from environment variables
	if err := l.config.LoadFromEnvironment(); err != nil {
		return nil, err
	}
	
	// Step 3: Validate the configuration
	if err := l.config.Validate(); err != nil {
		return nil, err
	}
	
	return l.config, nil
}

// LoadWithOverrides loads configuration and applies command line overrides
func (l *Loader) LoadWithOverrides(overrides *ConfigOverrides) (*Config, error) {
	// Load base configuration
	config, err := l.Load()
	if err != nil {
		return nil, err
	}
	
	// Apply command line overrides
	if overrides != nil {
		l.applyOverrides(config, overrides)
	}
	
	// Re-validate after applying overrides
	if err := config.Validate(); err != nil {
		return nil, err
	}
	
	return config, nil
}

// ConfigOverrides holds command line flag overrides
type ConfigOverrides struct {
	// Database overrides
	DBDir            *string
	DBFilename       *string
	DBQueryTimeout   *time.Duration
	DBWriteTimeout   *time.Duration
	DBDirPermissions *uint32

	// Time overrides
	TimeFormat *string

	// Validation overrides
	TaskNameMinLength *int
	TaskNameMaxLength *int
	MaxDuration       *time.Duration

	// Display overrides
	SummaryWidth  *int
	RunningStatus *string
	DateOnly      *bool

	// Application overrides
	Timeout *time.Duration
	Verbose *bool

	// Commands overrides
	ListDefaultFormat   *string
	OutputDefaultFormat *string
}

// applyOverrides applies command line overrides to the configuration
func (l *Loader) applyOverrides(config *Config, overrides *ConfigOverrides) {
	// Database overrides
	if overrides.DBDir != nil {
		config.Database.Dir = *overrides.DBDir
	}
	if overrides.DBFilename != nil {
		config.Database.Filename = *overrides.DBFilename
	}
	if overrides.DBQueryTimeout != nil {
		config.Database.QueryTimeout = *overrides.DBQueryTimeout
	}
	if overrides.DBWriteTimeout != nil {
		config.Database.WriteTimeout = *overrides.DBWriteTimeout
	}
	if overrides.DBDirPermissions != nil {
		config.Database.DirPermissions = *overrides.DBDirPermissions
	}

	// Time overrides
	if overrides.TimeFormat != nil {
		config.Time.DisplayFormat = *overrides.TimeFormat
	}

	// Validation overrides
	if overrides.TaskNameMinLength != nil {
		config.Validation.TaskNameMinLength = *overrides.TaskNameMinLength
	}
	if overrides.TaskNameMaxLength != nil {
		config.Validation.TaskNameMaxLength = *overrides.TaskNameMaxLength
	}
	if overrides.MaxDuration != nil {
		config.Validation.MaxDuration = *overrides.MaxDuration
	}

	// Display overrides
	if overrides.SummaryWidth != nil {
		config.Display.SummaryWidth = *overrides.SummaryWidth
	}
	if overrides.RunningStatus != nil {
		config.Display.RunningStatus = *overrides.RunningStatus
	}
	if overrides.DateOnly != nil {
		config.Display.DateOnly = *overrides.DateOnly
	}

	// Application overrides
	if overrides.Timeout != nil {
		config.Application.Timeout = *overrides.Timeout
	}
	if overrides.Verbose != nil {
		config.Application.Verbose = *overrides.Verbose
	}

	// Commands overrides
	if overrides.ListDefaultFormat != nil {
		config.Commands.ListDefaultFormat = *overrides.ListDefaultFormat
	}
	if overrides.OutputDefaultFormat != nil {
		config.Commands.OutputDefaultFormat = *overrides.OutputDefaultFormat
	}
}


// ParseDurationWithFallback parses a duration string with a fallback value
func ParseDurationWithFallback(s string, fallback time.Duration) time.Duration {
	if d, err := time.ParseDuration(s); err == nil {
		return d
	}
	return fallback
}

// ParseIntWithFallback parses an integer string with a fallback value
func ParseIntWithFallback(s string, fallback int) int {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return fallback
}

// ParseBoolWithFallback parses a boolean string with a fallback value
func ParseBoolWithFallback(s string, fallback bool) bool {
	if b, err := strconv.ParseBool(s); err == nil {
		return b
	}
	return fallback
}

// ParseUint32WithFallback parses a uint32 string with a fallback value
func ParseUint32WithFallback(s string, base int, fallback uint32) uint32 {
	if u, err := strconv.ParseUint(s, base, 32); err == nil {
		return uint32(u)
	}
	return fallback
}