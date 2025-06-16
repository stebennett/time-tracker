package cli

import (
	"fmt"
	"strings"
)

// App represents the main CLI application
type App struct {
	// Add configuration fields here as needed
}

// NewApp creates a new CLI application instance
func NewApp() *App {
	return &App{}
}

// Run executes the CLI application with the given arguments
func (a *App) Run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: tt \"your text here\"")
	}

	// Join all arguments with spaces
	text := strings.Join(args, " ")
	
	// Print the text
	fmt.Println(text)
	return nil
} 