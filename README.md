# Time Tracker

A command-line time tracking application written in Go.

## Project Structure

```
time-tracker/
├── cmd/            # Application entry points
│   └── tt/        # Main time tracker CLI
├── internal/       # Private application code
│   └── cli/       # CLI implementation
├── pkg/           # Public libraries
├── configs/       # Configuration files
└── docs/          # Documentation
```

## Installation

### Option 1: Using Go Install (Recommended for Developers)
```bash
go install ./cmd/tt
```

### Option 2: Download Binary (Recommended for Users)
1. Download the latest release binary for your operating system from the [releases page](https://github.com/stebennett/time-tracker/releases)
2. Extract the binary to a location of your choice
3. Add the binary location to your system PATH:
   - Windows:
     ```powershell
     # Add to user PATH
     $env:Path += ";C:\path\to\time-tracker"
     # Or add permanently through System Properties > Environment Variables
     ```
   - Linux/macOS:
     ```bash
     # Add to your shell profile (.bashrc, .zshrc, etc.)
     export PATH="$PATH:/path/to/time-tracker"
     ```

### Database Configuration
The time tracker stores its data in a SQLite database. By default, it's located at:
- Windows: `%USERPROFILE%\.tt\tt.db`
- Linux/macOS: `~/.tt/tt.db`

To use a custom database location, set the `TT_DB` environment variable:
```bash
# Windows
set TT_DB=C:\custom\path\to\tt.db

# Linux/macOS
export TT_DB=/custom/path/to/tt.db
```

## Usage

Basic usage:
```bash
# Start a new task
tt "your text here"

# Stop all running tasks
tt stop

# Show current task
tt current

# List tasks
tt list                    # List all tasks
tt list 1h                 # List tasks from last hour
tt list 2d                 # List tasks from last 2 days
tt list "meeting"          # List tasks containing "meeting"
tt list 1w "project"       # List tasks from last week containing "project"

# Show task summary
tt summary                 # Show all tasks to choose from
tt summary "coding"        # Show tasks containing "coding" to choose from
tt summary 2h              # Show tasks worked on in last 2 hours to choose from
tt summary 1d "project"    # Show tasks with "project" in name worked on in last day

# Export tasks
tt output format=csv       # Export all tasks to CSV format
```

Time shorthand formats:
- `nm` = last n minutes (e.g., "30m")
- `nh` = last n hours (e.g., "2h")
- `nd` = last n days (e.g., "1d")
- `nw` = last n weeks (e.g., "2w")
- `nmo` = last n months (e.g., "3mo")
- `ny` = last n years (e.g., "1y")

## Summary Command

The summary command provides detailed information about time entries for a specific task:

- Shows a table of all working sessions with start time, end time, duration, and status
- Displays summary statistics including total sessions, time range, and total time
- Handles running sessions by showing current elapsed time
- Supports task selection when multiple tasks match the criteria

Example summary output:
```
Summary for: coding project
======================
Start Time           End Time             Duration        Status
---------------------------------------------------------------------------
2024-01-01 09:00:00  2024-01-01 11:00:00  2h 0m           Completed
2024-01-01 14:00:00  2024-01-01 16:00:00  2h 0m           Completed
2024-01-01 18:00:00  running              1h 30m           Running
---------------------------------------------------------------------------
Total Sessions: 3 (1 running)
Time Range: 2024-01-01 09:00:00 to 2024-01-01 19:30:00
Total Time: 5h 30m
```

## CSV Export Format

The CSV export includes the following columns:
- ID: Unique identifier for the task
- Start Time: Task start time in RFC3339 format
- End Time: Task end time in RFC3339 format (empty for running tasks)
- Duration (hours): Task duration in hours (empty for running tasks)
- Description: Task description

Example usage:
```bash
# Export to a file
tt output format=csv > tasks.csv

# Export and filter
tt output format=csv | grep "meeting" > meetings.csv
```

## Development

This project is built using Go. To run the project locally:

1. Clone the repository
2. Run `go mod tidy` to install dependencies
3. Run `go run cmd/tt/main.go` to execute the application

## License

MIT