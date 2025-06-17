package main

import (
	"fmt"
	"os"

	"time-tracker/internal/cli"
	"time-tracker/internal/repository/sqlite"
)

func main() {
	// Create repository factory based on environment
	env := getEnvironment()
	factory := NewRepositoryFactory(env)
	
	// Create repository with dependency injection
	repo, err := factory.CreateRepository()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating repository: %v\n", err)
		os.Exit(1)
	}
	defer repo.Close()

	// Create app with injected repository
	app := cli.NewApp(repo)

	if err := app.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// createRepository creates and returns a repository instance
// This function can be easily modified for different environments (dev, test, prod)
func createRepository() (sqlite.Repository, error) {
	// For now, we'll use the default SQLite repository
	// In the future, this could be extended to support different repository types
	// or different configurations based on environment variables
	
	// Get database path
	dbPath, err := cli.GetDatabasePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get database path: %w", err)
	}

	// Initialize SQLite repository
	repo, err := sqlite.New(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return repo, nil
} 