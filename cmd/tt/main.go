package main

import (
	"fmt"
	"os"

	"time-tracker/internal/api"
	"time-tracker/internal/cli"
	"time-tracker/internal/config"
)

func main() {
	// Initialize configuration system with cascading priority
	loader := config.NewLoader()
	cfg, err := loader.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Create repository factory based on environment
	env := config.GetEnvironment()
	factory := config.NewRepositoryFactory(env)

	// Create repository with dependency injection using configuration
	repo, err := factory.CreateRepository()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating repository: %v\n", err)
		os.Exit(1)
	}
	defer repo.Close()

	// Create API instance
	apiInstance := api.New(repo)

	// Create Cobra root command with configuration
	rootCmd := cli.NewRootCommand(apiInstance, cfg)

	// Apply configuration overrides from command-line flags
	if err := rootCmd.PreRun(); err != nil {
		fmt.Fprintf(os.Stderr, "Error processing command-line flags: %v\n", err)
		os.Exit(1)
	}

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
