# Debug Logging

The time-tracker application includes a debug logging system that allows you to see detailed information about internal operations when needed.

## Enabling Debug Mode

To enable debug logging, set the `TT_DEBUG` environment variable to any value:

```bash
# On Unix-like systems (Linux, macOS)
export TT_DEBUG=1

# On Windows PowerShell
$env:TT_DEBUG="1"

# On Windows Command Prompt
set TT_DEBUG=1
```

## What Debug Logging Shows

When debug mode is enabled, you'll see additional information about:

- **Database migrations**: Which migrations are being applied and their progress
- **Database backups**: When backups are created and cleaned up
- **Migration errors**: Detailed warnings about parsing issues during data migrations
- **Test debugging**: Additional output during test runs to help debug test failures

## Example Output

With debug mode enabled, you might see output like:

```
Database backup created: /path/to/db.backup.20250624_135644
Applying migration version 1 (type: sql)
Applying migration version 2 (type: sql)
Applying migration version 3 (type: go)
Migration complete: processed 4 rows, updated 4 values, encountered 0 errors
Removed backup file: /path/to/db.backup.20250624_135644
```

Without debug mode, these messages are suppressed and the application runs silently for these operations.

## Usage in Development

Debug logging is particularly useful during development when you need to:

- Troubleshoot migration issues
- Understand what's happening during database operations
- Debug test failures
- Monitor backup and restore operations

## Usage in Production

In production environments, debug logging should typically be disabled to avoid cluttering logs and potentially exposing sensitive information. Only enable it temporarily when troubleshooting specific issues.

## Implementation Details

The debug logging system is implemented in the `internal/logging` package and provides:

- `DebugEnabled()`: Check if debug mode is enabled
- `Debugf(format, args...)`: Print formatted debug messages
- `Debugln(args...)`: Print debug messages with newline

All debug logging functions automatically check the `TT_DEBUG` environment variable and only output messages when it's set. 