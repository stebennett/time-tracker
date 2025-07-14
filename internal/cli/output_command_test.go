package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutputCommand_Execute(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	cmd := NewOutputCommand(app)
	ctx := context.Background()

	t.Run("outputs CSV format successfully", func(t *testing.T) {
		// Create some test data
		_, err := app.businessAPI.StartNewTask(ctx, "CSV Test Task")
		require.NoError(t, err)
		_, err = app.businessAPI.StopAllRunningTasks(ctx)
		require.NoError(t, err)

		// Output as CSV
		err = cmd.Execute(ctx, []string{"format=csv"})
		assert.NoError(t, err)
	})

	t.Run("handles empty data", func(t *testing.T) {
		// Output when no tasks exist
		err := cmd.Execute(ctx, []string{"format=csv"})
		assert.NoError(t, err)
	})

	t.Run("rejects invalid format", func(t *testing.T) {
		err := cmd.Execute(ctx, []string{"format=invalid"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported format")
	})

	t.Run("rejects malformed format argument", func(t *testing.T) {
		err := cmd.Execute(ctx, []string{"invalidformat"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid format option")
	})

	t.Run("requires format argument", func(t *testing.T) {
		err := cmd.Execute(ctx, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "usage: tt output format=csv")
	})

	t.Run("handles multiple arguments", func(t *testing.T) {
		// Should only use first argument
		err := cmd.Execute(ctx, []string{"format=csv", "extra", "args"})
		assert.NoError(t, err)
	})
}

func TestOutputCommand_OutputTasks(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	cmd := NewOutputCommand(app)
	ctx := context.Background()

	t.Run("processes valid format", func(t *testing.T) {
		err := cmd.outputTasks(ctx, []string{"format=csv"})
		assert.NoError(t, err)
	})

	t.Run("rejects no arguments", func(t *testing.T) {
		err := cmd.outputTasks(ctx, []string{})
		assert.Error(t, err)
	})

	t.Run("rejects invalid format prefix", func(t *testing.T) {
		err := cmd.outputTasks(ctx, []string{"type=csv"})
		assert.Error(t, err)
	})

	t.Run("rejects unsupported format", func(t *testing.T) {
		err := cmd.outputTasks(ctx, []string{"format=json"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported format")
	})
}

func TestOutputCommand_OutputCSV(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	cmd := NewOutputCommand(app)
	ctx := context.Background()

	t.Run("outputs CSV with running task", func(t *testing.T) {
		// Create a running task
		_, err := app.businessAPI.StartNewTask(ctx, "Running CSV Task")
		require.NoError(t, err)

		err = cmd.outputCSV(ctx)
		assert.NoError(t, err)
	})

	t.Run("outputs CSV with stopped task", func(t *testing.T) {
		// Create and stop a task
		_, err := app.businessAPI.StartNewTask(ctx, "Stopped CSV Task")
		require.NoError(t, err)
		_, err = app.businessAPI.StopAllRunningTasks(ctx)
		require.NoError(t, err)

		err = cmd.outputCSV(ctx)
		assert.NoError(t, err)
	})

	t.Run("outputs CSV with mixed tasks", func(t *testing.T) {
		// Create multiple tasks with different states
		_, err := app.businessAPI.StartNewTask(ctx, "Task 1")
		require.NoError(t, err)
		_, err = app.businessAPI.StopAllRunningTasks(ctx)
		require.NoError(t, err)

		_, err = app.businessAPI.StartNewTask(ctx, "Task 2")
		require.NoError(t, err)

		err = cmd.outputCSV(ctx)
		assert.NoError(t, err)
	})

	t.Run("outputs CSV with special characters in task names", func(t *testing.T) {
		// Create task with special characters
		_, err := app.businessAPI.StartNewTask(ctx, "Task with, comma and \"quotes\"")
		require.NoError(t, err)

		err = cmd.outputCSV(ctx)
		assert.NoError(t, err)
	})

	t.Run("handles empty data", func(t *testing.T) {
		// Stop all tasks and clear data would be ideal, but our mock doesn't support clearing
		// Test with no additional tasks
		err := cmd.outputCSV(ctx)
		assert.NoError(t, err)
	})
}

func TestNewOutputCommand(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	cmd := NewOutputCommand(app)
	
	assert.NotNil(t, cmd)
	assert.NotNil(t, cmd.businessAPI)
}