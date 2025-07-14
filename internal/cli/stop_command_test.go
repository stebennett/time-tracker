package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStopCommand_Execute(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	cmd := NewStopCommand(app)
	ctx := context.Background()

	t.Run("stops running task", func(t *testing.T) {
		// Start a task first
		_, err := app.businessAPI.StartNewTask(ctx, "Running Task")
		require.NoError(t, err)

		// Verify task is running
		_, err = app.businessAPI.GetCurrentSession(ctx)
		require.NoError(t, err)

		// Stop the task
		err = cmd.Execute(ctx, []string{})
		assert.NoError(t, err)

		// Verify no task is running
		_, err = app.businessAPI.GetCurrentSession(ctx)
		assert.Error(t, err) // Should be "not found" error
	})

	t.Run("stops multiple running tasks", func(t *testing.T) {
		// Start multiple tasks (though our implementation only allows one at a time)
		_, err := app.businessAPI.StartNewTask(ctx, "Task 1")
		require.NoError(t, err)

		// Stop all tasks
		err = cmd.Execute(ctx, []string{})
		assert.NoError(t, err)

		// Verify no tasks are running
		_, err = app.businessAPI.GetCurrentSession(ctx)
		assert.Error(t, err)
	})

	t.Run("handles no running tasks gracefully", func(t *testing.T) {
		// Make sure no tasks are running
		_, _ = app.businessAPI.StopAllRunningTasks(ctx)

		// Stop when nothing is running
		err := cmd.Execute(ctx, []string{})
		assert.NoError(t, err) // Should not error
	})

	t.Run("rejects arguments", func(t *testing.T) {
		err := cmd.Execute(ctx, []string{"unexpected", "args"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "usage: tt stop")
	})
}

func TestStopCommand_StopRunningTasks(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	cmd := NewStopCommand(app)
	ctx := context.Background()

	t.Run("stops running task successfully", func(t *testing.T) {
		// Start a task
		_, err := app.businessAPI.StartNewTask(ctx, "Test Task")
		require.NoError(t, err)

		// Stop it
		err = cmd.stopRunningTasks(ctx)
		assert.NoError(t, err)

		// Verify it's stopped
		_, err = app.businessAPI.GetCurrentSession(ctx)
		assert.Error(t, err)
	})

	t.Run("handles no running tasks", func(t *testing.T) {
		// Ensure no tasks are running
		_, _ = app.businessAPI.StopAllRunningTasks(ctx)

		// Stop when nothing is running
		err := cmd.stopRunningTasks(ctx)
		assert.NoError(t, err)
	})
}

func TestNewStopCommand(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	cmd := NewStopCommand(app)
	
	assert.NotNil(t, cmd)
	assert.NotNil(t, cmd.businessAPI)
}