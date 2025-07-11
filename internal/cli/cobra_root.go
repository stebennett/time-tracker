package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"time-tracker/internal/api"
	"time-tracker/internal/config"
)

// RootCommand represents the base command when called without any subcommands
type RootCommand struct {
	cmd    *cobra.Command
	api    api.API
	config *config.Config
}

// NewRootCommand creates the root cobra command with global flags
func NewRootCommand(apiInstance api.API, cfg *config.Config) *RootCommand {
	root := &RootCommand{
		api:    apiInstance,
		config: cfg,
	}

	root.cmd = &cobra.Command{
		Use:   "tt",
		Short: "A command-line time tracking application",
		Long: `Time Tracker (tt) is a command-line application for tracking time spent on tasks.

FEATURES:
  • Start and stop time tracking for named tasks
  • List and filter time entries by time range or task name  
  • Export data to CSV format
  • Resume previous tasks from interactive menus
  • Generate detailed summaries and delete tasks
  • Fully configurable via environment variables and command-line flags

EXAMPLES:
  tt start "Working on feature X"          # Start tracking a new task
  tt list 2h                               # List tasks from last 2 hours
  tt list 1d "meeting"                     # List tasks from last day containing "meeting"
  tt current                               # Show currently running task
  tt stop                                  # Stop all running tasks
  tt resume                                # Resume a previous task (interactive)
  tt summary 1w                            # Summary of tasks from last week
  tt output format=csv > tasks.csv         # Export to CSV file

CONFIGURATION:
  Configuration follows this priority order: command-line flags > environment variables > defaults
  
  Database Configuration:
    TT_DB_DIR                              Database directory (default: ~/.tt)
    TT_DB_FILENAME                         Database filename (default: tt.db)
    TT_DB_QUERY_TIMEOUT                    Query timeout (default: 10s)
    TT_DB_WRITE_TIMEOUT                    Write timeout (default: 5s)
  
  Display Configuration:
    TT_TIME_DISPLAY_FORMAT                 Time format (default: 2006-01-02 15:04:05)
    TT_DISPLAY_RUNNING_STATUS              Running status text (default: running)
    TT_DISPLAY_SUMMARY_WIDTH               Summary display width (default: 75)
    TT_DISPLAY_DATE_ONLY                   Show date only (default: false)
  
  Validation Configuration:
    TT_VALIDATION_TASK_NAME_MIN            Min task name length (default: 1)
    TT_VALIDATION_TASK_NAME_MAX            Max task name length (default: 255)
    TT_VALIDATION_MAX_DURATION             Max time entry duration (default: 24h)
  
  Application Configuration:
    TT_APP_TIMEOUT                         Application timeout (default: 60s)
    TT_APP_VERBOSE                         Enable verbose output (default: false)
  
  Command Configuration:
    TT_LIST_DEFAULT_FORMAT                 Default list format (default: table)
    TT_OUTPUT_DEFAULT_FORMAT               Default output format (default: csv)

TIME FORMATS:
  Use these shorthand formats for time filtering:
    30m, 2h, 1d, 2w, 3mo, 1y              # Minutes, hours, days, weeks, months, years

GETTING HELP:
  tt [command] --help                      # Get help for any specific command
  tt completion bash                       # Generate bash completion script
  tt completion zsh                        # Generate zsh completion script`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Apply configuration overrides from flags before any command runs
			return root.getConfigFromFlags()
		},
	}

	// Add global flags for configuration overrides
	root.addGlobalFlags()
	
	// Add all subcommands
	root.addSubcommands()

	return root
}

// Execute runs the root command
func (r *RootCommand) Execute() error {
	return r.cmd.Execute()
}

// addGlobalFlags adds global configuration flags
func (r *RootCommand) addGlobalFlags() {
	flags := r.cmd.PersistentFlags()

	// Database configuration
	flags.String("db-dir", "", "Database directory (overrides TT_DB_DIR)")
	flags.String("db-filename", "", "Database filename (overrides TT_DB_FILENAME)")
	flags.Duration("db-query-timeout", 0, "Database query timeout (overrides TT_DB_QUERY_TIMEOUT)")
	flags.Duration("db-write-timeout", 0, "Database write timeout (overrides TT_DB_WRITE_TIMEOUT)")

	// Time configuration
	flags.String("time-format", "", "Time display format (overrides TT_TIME_DISPLAY_FORMAT)")

	// Display configuration
	flags.Int("summary-width", 0, "Summary display width (overrides TT_DISPLAY_SUMMARY_WIDTH)")
	flags.String("running-status", "", "Running status text (overrides TT_DISPLAY_RUNNING_STATUS)")
	flags.Bool("date-only", false, "Show date only in displays (overrides TT_DISPLAY_DATE_ONLY)")

	// Validation configuration
	flags.Int("task-name-min-length", 0, "Minimum task name length (overrides TT_VALIDATION_TASK_NAME_MIN)")
	flags.Int("task-name-max-length", 0, "Maximum task name length (overrides TT_VALIDATION_TASK_NAME_MAX)")
	flags.Duration("max-duration", 0, "Maximum time entry duration (overrides TT_VALIDATION_MAX_DURATION)")

	// Application configuration
	flags.Duration("app-timeout", 0, "Application timeout (overrides TT_APP_TIMEOUT)")
	flags.Bool("verbose", false, "Enable verbose output (overrides TT_APP_VERBOSE)")

	// Commands configuration
	flags.String("list-format", "", "Default list format (overrides TT_LIST_DEFAULT_FORMAT)")
	flags.String("output-format", "", "Default output format (overrides TT_OUTPUT_DEFAULT_FORMAT)")
}

// addSubcommands adds all CLI subcommands to the root command
func (r *RootCommand) addSubcommands() {
	// Start command
	startCmd := &cobra.Command{
		Use:   "start [task name]",
		Short: "Start a new task",
		Long:  "Start tracking time for a new task. If a task is already running, it will be stopped first.",
		Args:  cobra.MinimumNArgs(1), // Require at least one argument
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), r.getAppTimeout())
			defer cancel()
			
			startHandler := NewStartCommand(NewAppWithConfig(r.api, r.config))
			return startHandler.Execute(ctx, args)
		},
	}

	// Stop command
	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop all running tasks",
		Long:  "Stop all currently running time tracking tasks.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), r.getAppTimeout())
			defer cancel()
			
			stopHandler := NewStopCommand(NewAppWithConfig(r.api, r.config))
			return stopHandler.Execute(ctx, args)
		},
	}

	// List command
	listCmd := &cobra.Command{
		Use:   "list [time] [text]",
		Short: "List time entries",
		Long: `List time entries with optional filtering.
		
Time filters support: 30m, 2h, 1d, 2w, 3mo, 1y
Text filters search within task names (case-insensitive partial matching)

Examples:
  tt list                    # List all entries
  tt list 1h                 # List entries from last hour
  tt list "project alpha"    # List entries containing "project alpha"
  tt list 2d "meeting"       # List entries from last 2 days containing "meeting"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), r.getAppTimeout())
			defer cancel()
			
			listHandler := NewListCommand(NewAppWithConfig(r.api, r.config))
			return listHandler.Execute(ctx, args)
		},
	}

	// Current command
	currentCmd := &cobra.Command{
		Use:   "current",
		Short: "Show currently running task",
		Long:  "Display information about the currently running task, if any.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), r.getAppTimeout())
			defer cancel()
			
			currentHandler := NewCurrentCommand(NewAppWithConfig(r.api, r.config))
			return currentHandler.Execute(ctx, args)
		},
	}

	// Output command
	outputCmd := &cobra.Command{
		Use:   "output format=csv",
		Short: "Export data in specified format",
		Long: `Export time tracking data in the specified format.
		
Supported formats:
  csv - Comma-separated values format

Example:
  tt output format=csv`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), r.getAppTimeout())
			defer cancel()
			
			outputHandler := NewOutputCommand(NewAppWithConfig(r.api, r.config))
			return outputHandler.Execute(ctx, args)
		},
	}

	// Resume command
	resumeCmd := &cobra.Command{
		Use:   "resume [time]",
		Short: "Resume a previous task",
		Long: `Resume a previous task by selecting from a list of recent tasks.
		
Time filters support: 30m, 2h, 1d, 2w, 3mo, 1y

Examples:
  tt resume      # Resume from today's tasks
  tt resume 3d   # Resume from tasks in the last 3 days`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resume commands may need longer timeout for user interaction
			ctx, cancel := context.WithTimeout(context.Background(), r.getAppTimeout()*2)
			defer cancel()
			
			resumeHandler := NewResumeCommand(NewAppWithConfig(r.api, r.config))
			return resumeHandler.Execute(ctx, args)
		},
	}

	// Summary command
	summaryCmd := &cobra.Command{
		Use:   "summary [time] [text]",
		Short: "Show detailed task summary",
		Long: `Show a detailed summary for selected tasks with time breakdowns.
		
Time filters support: 30m, 2h, 1d, 2w, 3mo, 1y
Text filters search within task names

Examples:
  tt summary           # Summary for all tasks
  tt summary 1w        # Summary for tasks from last week
  tt summary "project" # Summary for tasks containing "project"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Summary commands may need longer timeout for user interaction
			ctx, cancel := context.WithTimeout(context.Background(), r.getAppTimeout()*2)
			defer cancel()
			
			summaryHandler := NewSummaryCommand(NewAppWithConfig(r.api, r.config))
			return summaryHandler.Execute(ctx, args)
		},
	}

	// Delete command
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a task and all its time entries",
		Long: `Delete a task and all its associated time entries.
		
This operation cannot be undone. You will be prompted to select
which task to delete from a list of available tasks.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Delete commands may need longer timeout for user interaction
			ctx, cancel := context.WithTimeout(context.Background(), r.getAppTimeout()*2)
			defer cancel()
			
			deleteHandler := NewDeleteCommand(NewAppWithConfig(r.api, r.config))
			return deleteHandler.Execute(ctx, args)
		},
	}

	// Add all subcommands to root
	r.cmd.AddCommand(
		startCmd,
		stopCmd,
		listCmd,
		currentCmd,
		outputCmd,
		resumeCmd,
		summaryCmd,
		deleteCmd,
	)
}

// getAppTimeout returns the configured application timeout
func (r *RootCommand) getAppTimeout() time.Duration {
	if r.config != nil {
		return r.config.Application.Timeout
	}
	return 60 * time.Second // Default timeout
}

// getConfigFromFlags updates the configuration with values from command-line flags
func (r *RootCommand) getConfigFromFlags() error {
	if r.config == nil {
		return fmt.Errorf("configuration not initialized")
	}

	flags := r.cmd.PersistentFlags()

	// Database configuration
	if dbDir, _ := flags.GetString("db-dir"); dbDir != "" {
		r.config.Database.Dir = dbDir
	}
	if dbFilename, _ := flags.GetString("db-filename"); dbFilename != "" {
		r.config.Database.Filename = dbFilename
	}
	if queryTimeout, _ := flags.GetDuration("db-query-timeout"); queryTimeout > 0 {
		r.config.Database.QueryTimeout = queryTimeout
	}
	if writeTimeout, _ := flags.GetDuration("db-write-timeout"); writeTimeout > 0 {
		r.config.Database.WriteTimeout = writeTimeout
	}

	// Time configuration
	if timeFormat, _ := flags.GetString("time-format"); timeFormat != "" {
		r.config.Time.DisplayFormat = timeFormat
	}

	// Display configuration
	if summaryWidth, _ := flags.GetInt("summary-width"); summaryWidth > 0 {
		r.config.Display.SummaryWidth = summaryWidth
	}
	if runningStatus, _ := flags.GetString("running-status"); runningStatus != "" {
		r.config.Display.RunningStatus = runningStatus
	}
	if dateOnly, _ := flags.GetBool("date-only"); dateOnly {
		r.config.Display.DateOnly = dateOnly
	}

	// Validation configuration
	if taskNameMinLength, _ := flags.GetInt("task-name-min-length"); taskNameMinLength > 0 {
		r.config.Validation.TaskNameMinLength = taskNameMinLength
	}
	if taskNameMaxLength, _ := flags.GetInt("task-name-max-length"); taskNameMaxLength > 0 {
		r.config.Validation.TaskNameMaxLength = taskNameMaxLength
	}
	if maxDuration, _ := flags.GetDuration("max-duration"); maxDuration > 0 {
		r.config.Validation.MaxDuration = maxDuration
	}

	// Application configuration
	if appTimeout, _ := flags.GetDuration("app-timeout"); appTimeout > 0 {
		r.config.Application.Timeout = appTimeout
	}
	if verbose, _ := flags.GetBool("verbose"); verbose {
		r.config.Application.Verbose = verbose
	}

	// Commands configuration
	if listFormat, _ := flags.GetString("list-format"); listFormat != "" {
		r.config.Commands.ListDefaultFormat = listFormat
	}
	if outputFormat, _ := flags.GetString("output-format"); outputFormat != "" {
		r.config.Commands.OutputDefaultFormat = outputFormat
	}

	return nil
}

// PreRun sets up configuration overrides from flags before running commands
func (r *RootCommand) PreRun() error {
	return r.getConfigFromFlags()
}