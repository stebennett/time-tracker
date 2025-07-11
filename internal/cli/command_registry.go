package cli

import (
	"context"
	"time-tracker/internal/errors"
)

// Command represents a CLI command
type Command interface {
	Execute(ctx context.Context, args []string) error
}

// CommandRegistry manages all available commands
type CommandRegistry struct {
	commands map[string]Command
}

// NewCommandRegistry creates a new command registry
func NewCommandRegistry(app *App) *CommandRegistry {
	registry := &CommandRegistry{
		commands: make(map[string]Command),
	}
	
	// Register all commands
	registry.Register("start", NewStartCommand(app))
	registry.Register("stop", NewStopCommand(app))
	registry.Register("list", NewListCommand(app))
	registry.Register("current", NewCurrentCommand(app))
	registry.Register("output", NewOutputCommand(app))
	registry.Register("resume", NewResumeCommand(app))
	registry.Register("summary", NewSummaryCommand(app))
	registry.Register("delete", NewDeleteCommand(app))
	
	return registry
}

// Register adds a command to the registry
func (r *CommandRegistry) Register(name string, command Command) {
	r.commands[name] = command
}

// Execute runs the specified command with the given arguments
func (r *CommandRegistry) Execute(ctx context.Context, commandName string, args []string) error {
	command, exists := r.commands[commandName]
	if !exists {
		return errors.NewInvalidInputError("command", commandName, "unknown command")
	}
	return command.Execute(ctx, args)
}

// GetUsage returns the usage string for the CLI
func (r *CommandRegistry) GetUsage() string {
	return "usage: tt start \"your text here\" or tt stop or tt list [time] [text] or tt current or tt output format=csv or tt summary [time] [text] or tt resume or tt delete"
}