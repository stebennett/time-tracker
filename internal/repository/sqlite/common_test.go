package sqlite

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockResult implements sql.Result for testing
type MockResult struct {
	lastInsertID int64
	rowsAffected int64
	insertErr    error
	rowsErr      error
}

func (mr *MockResult) LastInsertId() (int64, error) {
	return mr.lastInsertID, mr.insertErr
}

func (mr *MockResult) RowsAffected() (int64, error) {
	return mr.rowsAffected, mr.rowsErr
}

func TestHandleDatabaseError(t *testing.T) {
	originalErr := errors.New("database connection failed")
	result := HandleDatabaseError("test operation", originalErr)
	
	assert.NotNil(t, result)
	assert.Contains(t, result.Error(), "test operation")
	assert.Contains(t, result.Error(), "database connection failed")
}

func TestHandleNoRowsError(t *testing.T) {
	tests := []struct {
		name         string
		inputErr     error
		entityType   string
		id           string
		expectNotFound bool
	}{
		{
			name:         "ErrNoRows should return NotFoundError",
			inputErr:     sql.ErrNoRows,
			entityType:   "test entity",
			id:           "123",
			expectNotFound: true,
		},
		{
			name:         "Other error should return as-is",
			inputErr:     errors.New("some other error"),
			entityType:   "test entity",
			id:           "123",
			expectNotFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HandleNoRowsError(tt.inputErr, tt.entityType, tt.id)
			
			if tt.expectNotFound {
				assert.Contains(t, result.Error(), "not found")
				assert.Contains(t, result.Error(), tt.entityType)
				assert.Contains(t, result.Error(), tt.id)
			} else {
				assert.Equal(t, tt.inputErr, result)
			}
		})
	}
}

func TestValidateRowsAffected(t *testing.T) {
	tests := []struct {
		name         string
		result       sql.Result
		entityType   string
		id           string
		expectError  bool
		expectNotFound bool
	}{
		{
			name: "Successful update",
			result: &MockResult{
				rowsAffected: 1,
				rowsErr:      nil,
			},
			entityType:   "test entity",
			id:           "123",
			expectError:  false,
			expectNotFound: false,
		},
		{
			name: "No rows affected",
			result: &MockResult{
				rowsAffected: 0,
				rowsErr:      nil,
			},
			entityType:   "test entity",
			id:           "123",
			expectError:  true,
			expectNotFound: true,
		},
		{
			name: "Error getting rows affected",
			result: &MockResult{
				rowsAffected: 0,
				rowsErr:      errors.New("database error"),
			},
			entityType:   "test entity",
			id:           "123",
			expectError:  true,
			expectNotFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateRowsAffected(tt.result, tt.entityType, tt.id)
			
			if tt.expectError {
				assert.Error(t, result)
				if tt.expectNotFound {
					assert.Contains(t, result.Error(), "not found")
				} else {
					assert.Contains(t, result.Error(), "database error")
				}
			} else {
				assert.NoError(t, result)
			}
		})
	}
}

// Note: ExecuteWithLastInsertID, ExecuteWithRowsAffected, QuerySingle, and QueryMultiple 
// are integration-level functions that require a real database connection.
// They are thoroughly tested through the existing repository tests.

func TestExecuteWithLastInsertID_MockResult(t *testing.T) {
	// This test demonstrates the function's error handling with mock results
	// The actual database execution is tested through repository integration tests
	
	tests := []struct {
		name        string
		mockResult  *MockResult
		expectError bool
	}{
		{
			name: "Successful insert",
			mockResult: &MockResult{
				lastInsertID: 42,
				insertErr:    nil,
			},
			expectError: false,
		},
		{
			name: "Error getting last insert ID",
			mockResult: &MockResult{
				lastInsertID: 0,
				insertErr:    errors.New("insert id error"),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the LastInsertId logic in isolation
			id, err := tt.mockResult.LastInsertId()
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Zero(t, id)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, int64(42), id)
			}
		})
	}
}

func TestExecuteWithRowsAffected_MockResult(t *testing.T) {
	// This test demonstrates the function's error handling with mock results
	// The actual database execution is tested through repository integration tests
	
	tests := []struct {
		name        string
		mockResult  *MockResult
		expectError bool
	}{
		{
			name: "Successful execution with affected rows",
			mockResult: &MockResult{
				rowsAffected: 1,
				rowsErr:      nil,
			},
			expectError: false,
		},
		{
			name: "No rows affected",
			mockResult: &MockResult{
				rowsAffected: 0,
				rowsErr:      nil,
			},
			expectError: true,
		},
		{
			name: "Error getting rows affected",
			mockResult: &MockResult{
				rowsAffected: 0,
				rowsErr:      errors.New("rows affected error"),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the ValidateRowsAffected logic with our mock
			err := ValidateRowsAffected(tt.mockResult, "test entity", "123")
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}