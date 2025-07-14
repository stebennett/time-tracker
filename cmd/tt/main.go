package main

import (
	"fmt"
	"os"

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

	// Create Cobra root command with configuration
	rootCmd := cli.NewRootCommand(cfg)

	// Execute the root command (PreRun will handle flag processing)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
