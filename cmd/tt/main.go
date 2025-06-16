package main

import (
	"fmt"
	"os"

	"time-tracker/internal/cli"
)

func main() {
	app, err := cli.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing app: %v\n", err)
		os.Exit(1)
	}

	if err := app.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
} 