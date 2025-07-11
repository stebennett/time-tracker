package sqlite

import (
	"context"
	"database/sql"

	"time-tracker/internal/errors"
)

// HandleDatabaseError converts database errors to structured app errors
func HandleDatabaseError(operation string, err error) error {
	return errors.NewDatabaseError(operation, err)
}

// HandleNoRowsError handles sql.ErrNoRows errors consistently
func HandleNoRowsError(err error, entityType string, id string) error {
	if err == sql.ErrNoRows {
		return errors.NewNotFoundError(entityType, id)
	}
	return err
}

// ValidateRowsAffected checks if a database operation affected the expected number of rows
func ValidateRowsAffected(result sql.Result, entityType string, id string) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return HandleDatabaseError("get rows affected", err)
	}
	if rows == 0 {
		return errors.NewNotFoundError(entityType, id)
	}
	return nil
}

// ExecuteWithLastInsertID executes a query and returns the last insert ID
func ExecuteWithLastInsertID(ctx context.Context, db *sql.DB, query string, args ...interface{}) (int64, error) {
	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, HandleDatabaseError("execute query", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, HandleDatabaseError("get last insert ID", err)
	}

	return id, nil
}

// ExecuteWithRowsAffected executes a query and validates that rows were affected
func ExecuteWithRowsAffected(ctx context.Context, db *sql.DB, query string, entityType string, id string, args ...interface{}) error {
	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return HandleDatabaseError("execute query", err)
	}

	return ValidateRowsAffected(result, entityType, id)
}

// QuerySingle executes a query that returns a single row and scans it
func QuerySingle[T any](ctx context.Context, db *sql.DB, query string, scanFunc func(Scanner) (*T, error), entityType string, id string, args ...interface{}) (*T, error) {
	row := db.QueryRowContext(ctx, query, args...)
	result, err := scanFunc(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError(entityType, id)
		}
		return nil, HandleDatabaseError("scan "+entityType, err)
	}
	return result, nil
}

// QueryMultiple executes a query that returns multiple rows and scans them
func QueryMultiple[T any](ctx context.Context, db *sql.DB, query string, scanFunc func(Rows) ([]*T, error), entityType string, args ...interface{}) ([]*T, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, HandleDatabaseError("query "+entityType, err)
	}
	defer rows.Close()

	results, err := scanFunc(rows)
	if err != nil {
		return nil, HandleDatabaseError("scan "+entityType, err)
	}

	return results, nil
}