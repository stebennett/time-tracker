# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a command-line time tracking application written in Go. It allows users to start, stop, and track tasks with time entries stored in a SQLite database.

## Development Commands

### Building and Running
```bash
# Build the application
go build -o tt cmd/tt/main.go

# Run directly
go run cmd/tt/main.go [commands]

# Install for development
go install ./cmd/tt

# Run tests
go test ./...

# Run specific test
go test ./internal/cli

# Get dependencies
go mod tidy
```

### Testing with Separate Database
**IMPORTANT**: When manually testing the CLI during development, always use a separate test database to avoid interfering with production data:

```bash
# Build test binary
go build -o tt-test cmd/tt/main.go

# Run commands with test database
TT_DB=/tmp/tt_test.db ./tt-test start "Test task"
TT_DB=/tmp/tt_test.db ./tt-test current
TT_DB=/tmp/tt_test.db ./tt-test list
TT_DB=/tmp/tt_test.db ./tt-test stop

# Or export the environment variable for multiple commands
export TT_DB=/tmp/tt_test.db
./tt-test start "Test task"
./tt-test current
./tt-test stop
unset TT_DB
```

### Database Configuration
- Production: `~/.tt/tt.db` (Linux/macOS) or `%USERPROFILE%\.tt\tt.db` (Windows)
- Test: Use `TT_DB` environment variable to specify alternative location
- Unit tests: Use in-memory mock implementations

### Testing
The project uses Go's built-in testing framework with `github.com/stretchr/testify` for assertions. All tests follow the `*_test.go` naming convention.

## Architecture

### Core Components

1. **CLI Layer** (`internal/cli/`): Command-line interface implementation
   - `app.go`: Main application logic, command parsing, and business operations
   - Handles all user commands: start, stop, list, current, output, resume, summary, delete

2. **API Layer** (`internal/api/`): Business logic interface
   - `api.go`: Defines the API interface and implements business logic
   - Acts as a facade over the repository layer
   - Provides methods for task/time entry CRUD and business operations

3. **Repository Layer** (`internal/repository/sqlite/`): Data persistence
   - `repository.go`: Database operations and queries
   - `model.go`: Data structures (Task, TimeEntry)
   - `migrations/`: Database schema migrations
   - Uses SQLite with `modernc.org/sqlite` driver

4. **Configuration** (`internal/config/`): Environment and database configuration
   - Repository factory pattern for dependency injection
   - Environment-based configuration management

### Dependencies and Patterns

- **Dependency Injection**: The application uses constructor injection with interfaces
- **Repository Pattern**: Database operations are abstracted through the Repository interface
- **Factory Pattern**: RepositoryFactory creates appropriate repository instances based on environment
- **Database**: SQLite with migrations handled programmatically

### Key Data Models

- **Task**: Represents a named task (ID, TaskName)
- **TimeEntry**: Represents a time tracking session (ID, TaskID, StartTime, EndTime)
- Running tasks have nil EndTime

## Database Configuration

The application uses SQLite with configurable location:
- Default: `~/.tt/tt.db` (Linux/macOS) or `%USERPROFILE%\.tt\tt.db` (Windows)
- Override with `TT_DB` environment variable

## Coding Standards (from .cursor/rules/)

- Use Go 1.22+ features and idioms
- Follow `gofmt`, `golint`, and `govet` guidelines
- Use explicit error handling (no silent ignores)
- Implement interfaces where appropriate
- Use dependency injection for better testability
- Document all exported functions and types
- Write unit tests for all business logic using table-driven tests
- Use prepared statements for database queries
- Never commit secrets or credentials
- Use environment variables for configuration

## Testing Approach

- Unit tests for all business logic
- Table-driven tests in Go
- Follow AAA (Arrange-Act-Assert) pattern
- Mock external dependencies
- Use `testing` package for Go tests
- Tests located alongside source files with `_test.go` suffix

## CLI Commands

The application provides these commands:
- `tt start "task name"` - Start a new task
- `tt stop` - Stop all running tasks
- `tt list [time] [text]` - List tasks with optional filters
- `tt current` - Show currently running task
- `tt output format=csv` - Export tasks to CSV
- `tt resume` - Resume a previous task
- `tt summary [time] [text]` - Show detailed task summary
- `tt delete` - Delete a task and all its time entries

Time filters support: `30m`, `2h`, `1d`, `2w`, `3mo`, `1y`