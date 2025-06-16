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
tt "your text here"
```

The application will echo back the text you provide.

## Development

This project is built using Go. To run the project locally:

1. Clone the repository
2. Run `go mod tidy` to install dependencies
3. Run `go run cmd/tt/main.go` to execute the application

## License

MIT 