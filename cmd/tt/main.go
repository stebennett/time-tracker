package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"time-tracker/internal/api"
	"time-tracker/internal/cli"
	"time-tracker/internal/config"
)

func main() {
	// Create repository factory based on environment
	env := config.GetEnvironment()
	factory := config.NewRepositoryFactory(env)

	// Create repository with dependency injection
	repo, err := factory.CreateRepository()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating repository: %v\n", err)
		os.Exit(1)
	}
	defer repo.Close()

	// Create API instance
	apiInstance := api.New(repo)

	// Create app with injected API
	app := cli.NewApp(apiInstance)

	// Create context with timeout for the application
	// Interactive commands (resume, delete, summary) may need longer timeouts
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	
	if err := app.Run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
