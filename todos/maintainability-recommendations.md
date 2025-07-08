# Time Tracker Maintainability Recommendations

This file contains detailed recommendations for improving the maintainability of the time tracker codebase. Each recommendation is structured to be picked up individually in future coding sessions.

## Status Legend
- 🟢 **COMPLETED** - Fully implemented and tested
- 🟡 **IN PROGRESS** - Partially implemented or currently being worked on
- 🔴 **PENDING** - Not started, ready to be picked up

---

## 🔴 HIGH PRIORITY RECOMMENDATIONS

### 1. 🟢 **COMPLETED**: Refactor Monolithic CLI Layer
**Status**: COMPLETED ✅
**Problem**: `internal/cli/app.go` was 800+ lines with too many responsibilities
**Impact**: Hard to test, modify, and understand
**Solution Implemented**: 
- Extracted command handlers into separate files (`start_command.go`, `list_command.go`, etc.)
- Created command registry pattern
- Moved complete business logic to command handlers, including SummaryCommand and DeleteCommand
- **Result**: Reduced from 821 lines to 139 lines (83% reduction)

### 2. 🔴 **Separate Domain Models from Database Models**
**Status**: PENDING
**Problem**: SQLite models (`sqlite.Task`, `sqlite.TimeEntry`) are used throughout API layer
**Impact**: Tight coupling between business logic and database schema, hard to change database implementation
**Files to Create**:
- `internal/domain/task.go` - Domain model for Task
- `internal/domain/time_entry.go` - Domain model for TimeEntry
- `internal/domain/mapper.go` - Conversion between domain and database models
**Files to Modify**:
- `internal/api/api.go` - Update interface to use domain models
- All command handlers - Update to use domain models
- Tests - Update to work with domain models
**Acceptance Criteria**:
- All API interfaces use domain models
- Database models are only used in repository layer
- Mapper functions handle conversion between layers
- All tests pass with no functionality changes

### 3. 🔴 **Add Input Validation Layer**
**Status**: PENDING
**Problem**: No validation at API boundaries, invalid data can reach database
**Impact**: Poor error messages, potential data corruption, hard to debug issues
**Files to Create**:
- `internal/validation/validator.go` - Core validation framework
- `internal/validation/task_validator.go` - Task-specific validation rules
- `internal/validation/time_entry_validator.go` - TimeEntry validation rules
- `internal/validation/errors.go` - Structured validation error types
**Files to Modify**:
- `internal/api/api.go` - Add validation calls to all API methods
- All command handlers - Handle validation errors appropriately
- Tests - Add validation test cases
**Validation Rules to Implement**:
- Task names: non-empty, max length, valid characters
- Time entries: valid time ranges, non-negative durations
- Search parameters: valid time formats, reasonable date ranges
**Acceptance Criteria**:
- All API methods validate input before processing
- Validation errors are user-friendly and actionable
- Invalid data cannot reach the database layer
- Comprehensive test coverage for validation scenarios

### 4. 🔴 **Implement Custom Error Types**
**Status**: PENDING
**Problem**: All errors are generic `error` type, can't distinguish between user errors and system errors
**Impact**: Poor error handling, unclear error messages, hard to handle errors appropriately in UI
**Files to Create**:
- `internal/errors/types.go` - Define error types and constants
- `internal/errors/errors.go` - Error creation and handling utilities
**Error Types to Implement**:
```go
type ErrorType int

const (
    ErrorTypeValidation ErrorType = iota
    ErrorTypeNotFound
    ErrorTypeDatabase
    ErrorTypeInvalidInput
    ErrorTypeTimeout
    ErrorTypePermission
)

type AppError struct {
    Type    ErrorType
    Message string
    Code    string
    Cause   error
    Context map[string]interface{}
}
```
**Files to Modify**:
- `internal/api/api.go` - Return structured errors
- All command handlers - Handle different error types appropriately
- `internal/repository/sqlite/repository.go` - Convert database errors to app errors
**Acceptance Criteria**:
- All functions return structured errors with appropriate types
- Error messages are user-friendly and actionable
- Different error types are handled appropriately (e.g., validation vs system errors)
- Error context includes relevant information for debugging

---

## 🟡 MEDIUM PRIORITY RECOMMENDATIONS

### 5. 🔴 **Reduce Code Duplication**
**Status**: PENDING
**Problem**: Database scanning code repeated across repository methods, time formatting utilities duplicated
**Impact**: Maintenance burden, inconsistent error handling, bug fixes needed in multiple places
**Files to Create**:
- `internal/repository/sqlite/scanner.go` - Generic scanning utilities
- `internal/repository/sqlite/formatters.go` - Time formatting utilities
- `internal/repository/sqlite/common.go` - Common database operations
**Files to Modify**:
- `internal/repository/sqlite/repository.go` - Use common utilities
- All command handlers - Use shared utilities where applicable
**Specific Duplications to Address**:
- Database row scanning patterns
- Time formatting for database storage
- Error handling patterns in repository methods
- Common query building patterns
**Acceptance Criteria**:
- No duplicated database scanning code
- Unified time formatting utilities
- Consistent error handling across repository methods
- Reduced line count in repository.go by at least 20%

### 6. 🔴 **Add Context Support**
**Status**: PENDING
**Problem**: No `context.Context` in interface methods, can't cancel operations or handle timeouts
**Impact**: Poor user experience with long-running operations, no way to cancel operations
**Files to Modify**:
- `internal/api/api.go` - Add context parameter to all methods
- `internal/repository/sqlite/repository.go` - Add context support to database operations
- All command handlers - Pass context through operation chains
- Tests - Update to use context in tests
**Implementation Strategy**:
- Add context.Context as first parameter to all API methods
- Use context.WithTimeout for database operations
- Check context.Done() in long-running operations
- Add context cancellation support in interactive commands (resume, delete, summary)
**Acceptance Criteria**:
- All API methods accept context.Context
- Database operations respect context cancellation
- Long-running operations can be cancelled
- Interactive commands respond to cancellation signals

### 7. 🔴 **Improve Configuration Management**
**Status**: PENDING
**Problem**: Hardcoded values scattered throughout codebase, no comprehensive configuration
**Impact**: Hard to modify behavior without code changes, difficult deployment configuration
**Files to Create**:
- `internal/config/config.go` - Comprehensive configuration struct
- `internal/config/validation.go` - Configuration validation
- `internal/config/loader.go` - Configuration loading from files/environment
- `config.yaml.example` - Example configuration file
**Files to Modify**:
- `internal/cli/app.go` - Use configuration for hardcoded values
- All command handlers - Use configuration for formatting, timeouts, etc.
- `cmd/tt/main.go` - Load and validate configuration at startup
**Configuration Areas**:
- Database connection settings
- Time formatting preferences
- Default time ranges for commands
- Output formatting options
- Validation rules (max task name length, etc.)
**Acceptance Criteria**:
- All hardcoded values moved to configuration
- Configuration validation at startup
- Support for both file-based and environment-based configuration
- Backward compatibility with existing behavior

### 8. 🔴 **Add Structured Logging**
**Status**: PENDING
**Problem**: Debug logging is minimal and inconsistent, hard to troubleshoot issues
**Impact**: Difficult debugging, no audit trail, hard to monitor application health
**Files to Create**:
- `internal/logging/logger.go` - Structured logging interface
- `internal/logging/levels.go` - Log level definitions
- `internal/logging/formatters.go` - Log formatting options
**Files to Modify**:
- `internal/api/api.go` - Add logging to all operations
- All command handlers - Add operation logging
- `internal/repository/sqlite/repository.go` - Add database operation logging
- `cmd/tt/main.go` - Configure logging at startup
**Logging Requirements**:
- Structured logging with JSON format option
- Log levels: DEBUG, INFO, WARN, ERROR
- Request correlation IDs for tracing
- Performance metrics (operation duration)
- Error context and stack traces
**Acceptance Criteria**:
- All operations are logged with appropriate levels
- Logs include correlation IDs for request tracing
- Performance metrics are captured
- Log output is configurable (console, file, JSON)

---

## 🟢 LOW PRIORITY IMPROVEMENTS

### 9. 🔴 **Add Dependency Injection Container**
**Status**: PENDING
**Problem**: Manual dependency wiring in main.go, complex initialization as application grows
**Impact**: Hard to test, complex setup, tight coupling between components
**Files to Create**:
- `internal/container/container.go` - DI container implementation
- `internal/container/providers.go` - Service providers
**Files to Modify**:
- `cmd/tt/main.go` - Use DI container for initialization
- Tests - Use DI container for test setup
**Implementation Options**:
- Lightweight DI container (e.g., `dig` or `wire`)
- Custom service locator pattern
- Factory pattern with interfaces
**Acceptance Criteria**:
- All dependencies are managed by container
- Easy to swap implementations for testing
- Clean initialization in main.go
- Improved testability

### 10. 🔴 **Enhance Testing Strategy**
**Status**: PENDING
**Problem**: Missing error scenario tests, no benchmarks, limited property-based testing
**Impact**: Potential bugs in error paths, unknown performance characteristics
**Files to Create**:
- `internal/cli/benchmarks_test.go` - Performance benchmarks
- `internal/cli/property_test.go` - Property-based tests
- `internal/cli/error_scenarios_test.go` - Comprehensive error testing
**Files to Modify**:
- All existing test files - Add error scenario coverage
- `internal/repository/sqlite/repository_test.go` - Add benchmark tests
**Testing Enhancements**:
- Property-based testing for time parsing logic
- Comprehensive error scenario tests
- Performance benchmarks for database operations
- Integration tests with real database
- Load testing for concurrent operations
**Acceptance Criteria**:
- 95%+ test coverage including error paths
- Performance benchmarks for critical operations
- Property-based tests for complex logic
- Integration tests with real database

---

## 🔄 ARCHITECTURAL IMPROVEMENTS

### 11. 🔴 **Extract Business Logic to Service Layer**
**Status**: PENDING
**Problem**: Business logic mixed in command handlers and API layer
**Impact**: Hard to test business rules, logic duplication, unclear separation of concerns
**Files to Create**:
- `internal/services/task_service.go` - Task business logic
- `internal/services/time_tracking_service.go` - Time tracking business logic
- `internal/services/reporting_service.go` - Reporting and analysis logic
**Files to Modify**:
- `internal/api/api.go` - Delegate to service layer
- All command handlers - Use service layer for business logic
**Service Responsibilities**:
- Task lifecycle management
- Time tracking rules and validation
- Reporting and analytics
- Data aggregation and calculations
**Acceptance Criteria**:
- Clear separation between presentation, business, and data layers
- Business rules are testable in isolation
- No business logic in command handlers
- Service layer is API-agnostic

### 12. 🔴 **Add Caching Layer**
**Status**: PENDING
**Problem**: No caching for frequently accessed data, repeated database queries
**Impact**: Performance degradation with large datasets, unnecessary database load
**Files to Create**:
- `internal/cache/cache.go` - Caching interface
- `internal/cache/memory.go` - In-memory cache implementation
- `internal/cache/redis.go` - Redis cache implementation (optional)
**Files to Modify**:
- `internal/api/api.go` - Add caching to frequently accessed methods
- `internal/repository/sqlite/repository.go` - Cache query results
**Caching Strategy**:
- Cache task lists and recent time entries
- Cache aggregated reports and summaries
- TTL-based cache invalidation
- Cache warming strategies
**Acceptance Criteria**:
- Significant performance improvement for repeated queries
- Cache hit/miss metrics
- Configurable cache backends
- Proper cache invalidation on data changes

---

## 📊 IMPLEMENTATION PRIORITY MATRIX

| Priority | Effort | Impact | Recommendation |
|----------|---------|---------|----------------|
| High | Medium | High | Domain model separation |
| High | Low | High | Input validation layer |
| High | Low | Medium | Custom error types |
| Medium | Medium | Medium | Code duplication reduction |
| Medium | High | Medium | Context support |
| Medium | Medium | Low | Configuration management |
| Medium | Medium | Low | Structured logging |
| Low | High | Low | Dependency injection |
| Low | Medium | Low | Enhanced testing |
| Low | High | Medium | Service layer extraction |
| Low | High | Medium | Caching layer |

---

## 🎯 QUICK WINS (Can be tackled in single session)

1. **Custom Error Types** - Create error types and update a few key methods
2. **Input Validation** - Add basic validation to API methods
3. **Code Duplication** - Extract common database utilities
4. **Configuration Management** - Move hardcoded values to config struct

## 🏗️ MAJOR REFACTORING (Multi-session work)

1. **Domain Model Separation** - Requires updating all layers
2. **Context Support** - Needs changes throughout the application
3. **Service Layer Extraction** - Significant architectural change
4. **Caching Layer** - New infrastructure component

---

## 📝 NOTES FOR FUTURE SESSIONS

- **Current State**: CLI layer has been successfully refactored with command handlers
- **Next Logical Step**: Domain model separation (most impactful for maintainability)
- **Testing Strategy**: Always create tests before implementing changes
- **Compatibility**: Maintain 100% backward compatibility for user experience
- **Database**: Use `TT_DB=/tmp/tt_test.db` for testing to avoid production interference

Each recommendation includes full context, implementation details, and acceptance criteria to enable picking up any task individually in future sessions.