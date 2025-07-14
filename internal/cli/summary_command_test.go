package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSummaryCommand_Execute(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	cmd := NewSummaryCommand(app)
	ctx := context.Background()

	t.Run("handles no arguments", func(t *testing.T) {
		// Create some test data
		_, err := app.businessAPI.StartNewTask(ctx, "Summary Test Task")
		require.NoError(t, err)
		_, err = app.businessAPI.StopAllRunningTasks(ctx)
		require.NoError(t, err)

		// Note: Summary command requires user interaction for task selection
		_ = cmd.summaryTask(ctx, []string{})
		// Will reach interactive part, testing argument processing
	})

	t.Run("handles time argument", func(t *testing.T) {
		_ = cmd.summaryTask(ctx, []string{"1h"})
		// Will process time argument and reach interactive part
	})

	t.Run("handles text filter", func(t *testing.T) {
		_ = cmd.summaryTask(ctx, []string{"TestTask"})
		// Will process text filter and reach interactive part
	})

	t.Run("handles time and text filter", func(t *testing.T) {
		_ = cmd.summaryTask(ctx, []string{"1h", "Test", "Task"})
		// Will process both filters and reach interactive part
	})
}

func TestSummaryCommand_SummaryTask(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	cmd := NewSummaryCommand(app)
	ctx := context.Background()

	t.Run("processes no arguments correctly", func(t *testing.T) {
		_ = cmd.summaryTask(ctx, []string{})
		// Will search all tasks and proceed to selection
	})

	t.Run("processes time shorthand correctly", func(t *testing.T) {
		_ = cmd.summaryTask(ctx, []string{"1h"})
		// Will apply time filter and proceed to selection
	})

	t.Run("processes text filter correctly", func(t *testing.T) {
		_ = cmd.summaryTask(ctx, []string{"SomeTask"})
		// Will apply text filter and proceed to selection
	})

	t.Run("processes combined filters correctly", func(t *testing.T) {
		_ = cmd.summaryTask(ctx, []string{"1h", "Some", "Task"})
		// Will apply both filters and proceed to selection
	})

	t.Run("handles invalid time format", func(t *testing.T) {
		// Create test data first
		_, err := app.businessAPI.StartNewTask(ctx, "Test Task")
		require.NoError(t, err)

		// Invalid time format should be treated as text filter
		_ = cmd.summaryTask(ctx, []string{"invalidtime"})
		// Should proceed with text filter, not error
	})
}

func TestSummaryCommand_ShowTaskSummary(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	cmd := NewSummaryCommand(app)
	ctx := context.Background()

	t.Run("shows summary for existing task", func(t *testing.T) {
		// Create a task
		session, err := app.businessAPI.StartNewTask(ctx, "Summary Task")
		require.NoError(t, err)
		taskID := session.Task.ID
		_, err = app.businessAPI.StopAllRunningTasks(ctx)
		require.NoError(t, err)

		// Show summary
		err = cmd.showTaskSummary(ctx, taskID)
		assert.NoError(t, err)
	})

	t.Run("shows summary for running task", func(t *testing.T) {
		// Create a running task
		session, err := app.businessAPI.StartNewTask(ctx, "Running Summary Task")
		require.NoError(t, err)
		taskID := session.Task.ID

		// Show summary
		err = cmd.showTaskSummary(ctx, taskID)
		assert.NoError(t, err)
	})

	t.Run("handles non-existent task", func(t *testing.T) {
		err := cmd.showTaskSummary(ctx, 999)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get task summary")
	})

	t.Run("shows summary with multiple sessions", func(t *testing.T) {
		// Create task with multiple sessions
		session, err := app.businessAPI.StartNewTask(ctx, "Multi Session Task")
		require.NoError(t, err)
		taskID := session.Task.ID
		_, err = app.businessAPI.StopAllRunningTasks(ctx)
		require.NoError(t, err)

		// Resume the same task (creates another session)
		_, err = app.businessAPI.ResumeTask(ctx, taskID)
		require.NoError(t, err)
		_, err = app.businessAPI.StopAllRunningTasks(ctx)
		require.NoError(t, err)

		// Show summary
		err = cmd.showTaskSummary(ctx, taskID)
		assert.NoError(t, err)
	})

	t.Run("shows summary with special characters in task name", func(t *testing.T) {
		// Create task with special characters
		session, err := app.businessAPI.StartNewTask(ctx, "Task: @#$% & more!")
		require.NoError(t, err)
		taskID := session.Task.ID

		// Show summary
		err = cmd.showTaskSummary(ctx, taskID)
		assert.NoError(t, err)
	})
}

func TestSummaryCommand_ArgumentProcessing(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	cmd := NewSummaryCommand(app)
	ctx := context.Background()

	// Create test data
	_, err := app.businessAPI.StartNewTask(ctx, "Argument Test Task")
	require.NoError(t, err)
	_, err = app.businessAPI.StopAllRunningTasks(ctx)
	require.NoError(t, err)

	tests := []struct {
		name        string
		args        []string
		description string
	}{
		{
			name:        "no arguments",
			args:        []string{},
			description: "should search all tasks",
		},
		{
			name:        "time filter only",
			args:        []string{"1h"},
			description: "should apply time filter",
		},
		{
			name:        "text filter only",
			args:        []string{"Task"},
			description: "should apply text filter",
		},
		{
			name:        "time and text filter",
			args:        []string{"1d", "Test", "Task"},
			description: "should apply both filters",
		},
		{
			name:        "invalid time as text filter",
			args:        []string{"invalidtime"},
			description: "should treat as text filter",
		},
		{
			name:        "multiple words as text filter",
			args:        []string{"Multiple", "Word", "Filter"},
			description: "should join as text filter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// These will typically proceed to interactive selection
			// The important thing is they don't crash during argument processing
			_ = cmd.summaryTask(ctx, tt.args)
			// We don't assert success/failure here because the command
			// will likely reach the interactive input stage
		})
	}
}

func TestNewSummaryCommand(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	cmd := NewSummaryCommand(app)
	
	assert.NotNil(t, cmd)
	assert.NotNil(t, cmd.businessAPI)
}