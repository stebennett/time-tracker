package config

import (
	"fmt"

	"time-tracker/internal/repository/sqlite"
)

// CreateRepository creates a repository instance using the configuration system
func CreateRepository(config *Config) (sqlite.Repository, error) {
	// Get database path from configuration
	dbPath := config.GetDatabasePath()

	// Initialize SQLite repository with configuration
	repo, err := sqlite.NewWithConfig(dbPath, config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return repo, nil
}

// CreateTestRepository creates an in-memory repository for testing
func CreateTestRepository() (sqlite.Repository, error) {
	// For testing, use an in-memory database
	dbPath := ":memory:"
	
	// Initialize SQLite repository without configuration
	repo, err := sqlite.New(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize test database: %w", err)
	}

	return repo, nil
} 