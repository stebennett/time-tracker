package main

import (
	"fmt"
	"os"

	"time-tracker/internal/cli"
	"time-tracker/internal/repository/sqlite"
)

// Environment represents the current environment
type Environment string

const (
	Development Environment = "development"
	Testing     Environment = "testing"
	Production  Environment = "production"
)

// RepositoryFactory creates repository instances based on environment
type RepositoryFactory struct {
	env Environment
}

// NewRepositoryFactory creates a new repository factory for the given environment
func NewRepositoryFactory(env Environment) *RepositoryFactory {
	return &RepositoryFactory{env: env}
}

// CreateRepository creates a repository instance based on the current environment
func (rf *RepositoryFactory) CreateRepository() (sqlite.Repository, error) {
	switch rf.env {
	case Development:
		return rf.createDevelopmentRepository()
	case Testing:
		return rf.createTestingRepository()
	case Production:
		return rf.createProductionRepository()
	default:
		return rf.createProductionRepository() // Default to production
	}
}

// createDevelopmentRepository creates a repository for development
// Uses a local SQLite database in the project directory
func (rf *RepositoryFactory) createDevelopmentRepository() (sqlite.Repository, error) {
	// For development, use a local database file
	dbPath := "tt.db"
	
	// Initialize SQLite repository
	repo, err := sqlite.New(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize development database: %w", err)
	}

	return repo, nil
}

// createTestingRepository creates a repository for testing
// Uses an in-memory SQLite database
func (rf *RepositoryFactory) createTestingRepository() (sqlite.Repository, error) {
	// For testing, use an in-memory database
	dbPath := ":memory:"
	
	// Initialize SQLite repository
	repo, err := sqlite.New(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize testing database: %w", err)
	}

	return repo, nil
}

// createProductionRepository creates a repository for production
// Uses the default SQLite database location
func (rf *RepositoryFactory) createProductionRepository() (sqlite.Repository, error) {
	// Get database path using the existing logic
	dbPath, err := cli.GetDatabasePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get database path: %w", err)
	}

	// Initialize SQLite repository
	repo, err := sqlite.New(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize production database: %w", err)
	}

	return repo, nil
}

// getEnvironment determines the current environment
func getEnvironment() Environment {
	env := os.Getenv("TT_ENV")
	switch env {
	case "development":
		return Development
	case "testing":
		return Testing
	case "production":
		return Production
	default:
		// Default to production for safety
		return Production
	}
} 