# Dependency Injection in Time Tracker

This document explains how dependency injection has been implemented in the Time Tracker application and how to use it for different environments.

## Overview

The Time Tracker application now uses dependency injection to inject the repository into the App. This provides several benefits:

- **Testability**: Easy to inject mock repositories for testing
- **Flexibility**: Can use different repository types for different environments
- **Separation of Concerns**: App logic is separated from data access logic
- **Environment Independence**: No hard dependencies on environment variables or production paths

## Architecture

### Repository Interface

The application uses a `Repository` interface defined in `internal/repository/sqlite/repository.go`:

```go
type Repository interface {
    // Create operations
    CreateTimeEntry(entry *TimeEntry) error
    CreateTask(task *Task) error
    
    // Read operations
    GetTimeEntry(id int64) (*TimeEntry, error)
    ListTimeEntries() ([]*TimeEntry, error)
    SearchTimeEntries(opts SearchOptions) ([]*TimeEntry, error)
    GetTask(id int64) (*Task, error)
    ListTasks() ([]*Task, error)
    
    // Update operations
    UpdateTimeEntry(entry *TimeEntry) error
    UpdateTask(task *Task) error
    
    // Delete operations
    DeleteTimeEntry(id int64) error
    DeleteTask(id int64) error
    
    // Utility
    Close() error
}
```

### App Constructor

The `App` now has two constructors:

1. **`NewApp(repo Repository)`**: Creates an App with an injected repository (for dependency injection)
2. **`NewAppWithDefaultRepository()`**: Creates an App with the default SQLite repository (for backward compatibility)

### Configuration Package

The configuration logic is now in `internal/config/repository.go` and provides:

- **Environment Detection**: Determines the current environment from `TT_ENV` environment variable
- **Repository Factory**: Creates appropriate repository instances based on environment
- **Environment Types**: Development, Testing, and Production configurations

## Environment Configuration

The application supports different environments through the `TT_ENV` environment variable:

### Development Environment

```bash
export TT_ENV=development
```

- Uses a local SQLite database file (`tt.db`) in the project directory
- Ideal for development work
- Database file is created in the current working directory

### Testing Environment

```bash
export TT_ENV=testing
```

- Uses an in-memory SQLite database (`:memory:`)
- Perfect for unit tests and integration tests
- No persistent data storage

### Production Environment

```bash
export TT_ENV=production
# or omit TT_ENV (defaults to production)
```

- Uses the default SQLite database location (`~/.tt/tt.db`)
- Respects the `TT_DB_DIR` and `TT_DB_FILENAME` environment variables for custom database paths
- Creates the `.tt` directory if it doesn't exist

## Usage Examples

### Basic Usage

```go
import "time-tracker/internal/config"

// Create a repository factory
factory := config.NewRepositoryFactory(config.Production)

// Create repository
repo, err := factory.CreateRepository()
if err != nil {
    // Handle error
}
defer repo.Close()

// Create app with injected repository
app := cli.NewApp(repo)

// Use the app
err = app.Run(args)
```

### Environment-Based Configuration

```go
import "time-tracker/internal/config"

// Get current environment
env := config.GetEnvironment()

// Create factory for current environment
factory := config.NewRepositoryFactory(env)

// Create repository
repo, err := factory.CreateRepository()
if err != nil {
    // Handle error
}
defer repo.Close()

// Create app with injected repository
app := cli.NewApp(repo)
```

### Testing with Mock Repository

```go
// Create a mock repository for testing
mockRepo := NewMockRepository()

// Create app with injected mock repository
app := cli.NewApp(mockRepo)

// Test the app
err := app.createNewTask("Test Task")
```

### Custom Repository Implementation

You can create your own repository implementation by implementing the `Repository` interface:

```go
type CustomRepository struct {
    // Your custom fields
}

func (r *CustomRepository) CreateTimeEntry(entry *sqlite.TimeEntry) error {
    // Your implementation
}

// Implement all other methods...

// Use it with the app
app := cli.NewApp(customRepo)
```

## Benefits

### 1. Improved Testability

Before dependency injection:
```go
// Hard to test - app creates its own repository
app, err := cli.NewApp() // Creates SQLite repository internally
```

After dependency injection:
```go
// Easy to test - inject mock repository
mockRepo := NewMockRepository()
app := cli.NewApp(mockRepo) // Uses injected mock
```

### 2. Environment Flexibility

```go
// Different environments use different repositories
switch env {
case Development:
    return sqlite.New("tt.db") // Local file
case Testing:
    return sqlite.New(":memory:") // In-memory
case Production:
    return sqlite.New("~/.tt/tt.db") // Production path
}
```

### 3. Separation of Concerns

- **App**: Handles CLI logic and user interaction
- **Repository**: Handles data access and persistence
- **Config**: Handles environment detection and repository creation
- **Factory**: Handles repository creation based on environment

## Migration Guide

### For Existing Code

If you have existing code that uses the old `NewApp()` constructor:

```go
// Old way
app, err := cli.NewApp()

// New way
app, err := cli.NewAppWithDefaultRepository()
```

### For New Code

Use the dependency injection approach:

```go
// Create repository using config package
env := config.GetEnvironment()
factory := config.NewRepositoryFactory(env)
repo, err := factory.CreateRepository()
if err != nil {
    // Handle error
}
defer repo.Close()

// Create app with injected repository
app := cli.NewApp(repo)
```

## Testing

The dependency injection system makes testing much easier. See `internal/cli/app_test.go` for examples of how to test the application with mock repositories.

### Running Tests

```bash
# Run all tests
go test ./...

# Run specific test
go test ./internal/cli -v
go test ./internal/config -v

# Run tests with coverage
go test ./... -cover
```

## Future Enhancements

The dependency injection system can be extended to support:

1. **Different Database Types**: PostgreSQL, MySQL, etc.
2. **Cloud Storage**: AWS S3, Google Cloud Storage, etc.
3. **Caching Layers**: Redis, Memcached, etc.
4. **Multiple Repositories**: Read replicas, write-ahead logs, etc.

To add a new repository type, simply implement the `Repository` interface and update the factory in `internal/config/repository.go` to support it.

If you want to start a new task using the CLI, use:

    tt start "Task name"

instead of relying on default or implicit task creation. 