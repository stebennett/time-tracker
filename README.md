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

```bash
go install ./cmd/tt
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