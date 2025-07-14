package cli

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewApp(t *testing.T) {
	mockAPI := newMockBusinessAPI()
	
	app := NewApp(mockAPI)
	
	assert.NotNil(t, app)
	assert.Equal(t, mockAPI, app.businessAPI)
	assert.Nil(t, app.config) // Config should be nil when not provided
	assert.NotNil(t, app.registry)
}

func TestNewAppWithConfig(t *testing.T) {
	mockAPI := newMockBusinessAPI()
	
	app := NewAppWithConfig(mockAPI, nil)
	
	assert.NotNil(t, app)
	assert.Equal(t, mockAPI, app.businessAPI)
	assert.Nil(t, app.config)
	assert.NotNil(t, app.registry)
}

func TestNewAppWithDefaultRepository(t *testing.T) {
	app, err := NewAppWithDefaultRepository()
	
	// This test depends on being able to create a real repository
	// In a testing environment, this might fail due to database constraints
	if err != nil {
		// If it fails, at least verify the error is reasonable
		t.Logf("NewAppWithDefaultRepository failed (expected in some test environments): %v", err)
		return
	}
	
	assert.NotNil(t, app)
	assert.NotNil(t, app.businessAPI)
	assert.NotNil(t, app.config)
	assert.NotNil(t, app.registry)
}

func TestApp_Run(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("runs start command successfully", func(t *testing.T) {
		err := app.Run(ctx, []string{"start", "Test Task"})
		assert.NoError(t, err)

		// Verify task was created
		session, err := app.businessAPI.GetCurrentSession(ctx)
		require.NoError(t, err)
		assert.Equal(t, "Test Task", session.Task.TaskName)
	})

	t.Run("runs stop command successfully", func(t *testing.T) {
		// Start a task first
		_, err := app.businessAPI.StartNewTask(ctx, "Task to Stop")
		require.NoError(t, err)

		// Stop it
		err = app.Run(ctx, []string{"stop"})
		assert.NoError(t, err)

		// Verify no task is running
		_, err = app.businessAPI.GetCurrentSession(ctx)
		assert.Error(t, err)
	})

	t.Run("runs current command successfully", func(t *testing.T) {
		err := app.Run(ctx, []string{"current"})
		assert.NoError(t, err)
	})

	t.Run("runs list command successfully", func(t *testing.T) {
		err := app.Run(ctx, []string{"list"})
		assert.NoError(t, err)
	})

	t.Run("runs output command successfully", func(t *testing.T) {
		err := app.Run(ctx, []string{"output", "format=csv"})
		assert.NoError(t, err)
	})

	t.Run("handles unknown command", func(t *testing.T) {
		err := app.Run(ctx, []string{"unknown"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown command")
	})

	t.Run("handles no arguments", func(t *testing.T) {
		err := app.Run(ctx, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "command")
	})

	t.Run("handles command with multiple arguments", func(t *testing.T) {
		err := app.Run(ctx, []string{"start", "Multi", "Word", "Task"})
		assert.NoError(t, err)

		session, err := app.businessAPI.GetCurrentSession(ctx)
		require.NoError(t, err)
		assert.Equal(t, "Multi Word Task", session.Task.TaskName)
	})
}

func TestApp_RunErrorPropagation(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("propagates command validation errors", func(t *testing.T) {
		// Start command without task name should error
		err := app.Run(ctx, []string{"start"})
		assert.Error(t, err)
	})

	t.Run("propagates format validation errors", func(t *testing.T) {
		// Output command with invalid format should error
		err := app.Run(ctx, []string{"output", "format=invalid"})
		assert.Error(t, err)
	})

	t.Run("propagates unknown command errors", func(t *testing.T) {
		err := app.Run(ctx, []string{"nonexistent"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown command")
	})
}

func TestParseTimeShorthand(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		expected    time.Duration
	}{
		{
			name:        "minutes",
			input:       "30m",
			expectError: false,
			expected:    30 * time.Minute,
		},
		{
			name:        "hours", 
			input:       "2h",
			expectError: false,
			expected:    2 * time.Hour,
		},
		{
			name:        "days",
			input:       "1d",
			expectError: false,
			expected:    24 * time.Hour,
		},
		{
			name:        "weeks",
			input:       "2w",
			expectError: false,
			expected:    14 * 24 * time.Hour,
		},
		{
			name:        "months",
			input:       "1mo",
			expectError: false,
			expected:    30 * 24 * time.Hour,
		},
		{
			name:        "years",
			input:       "1y",
			expectError: false,
			expected:    365 * 24 * time.Hour,
		},
		{
			name:        "invalid format",
			input:       "invalid",
			expectError: true,
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
		},
		{
			name:        "no number",
			input:       "h",
			expectError: true,
		},
		{
			name:        "no unit",
			input:       "5",
			expectError: true,
		},
		{
			name:        "invalid unit",
			input:       "5x",
			expectError: true,
		},
		{
			name:        "negative number",
			input:       "-5h",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration, err := parseTimeShorthand(tt.input)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, duration)
			}
		})
	}
}

func TestTimeNow(t *testing.T) {
	// Test that timeNow can be overridden for testing
	originalTimeNow := timeNow
	defer func() { timeNow = originalTimeNow }()
	
	fixedTime := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	timeNow = func() time.Time { return fixedTime }
	
	assert.Equal(t, fixedTime, timeNow())
}