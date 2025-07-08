package main

import (
	"fmt"
	"os"

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

	if err := app.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
