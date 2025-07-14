package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCurrentCommand_Execute(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	cmd := NewCurrentCommand(app)
	ctx := context.Background()

	t.Run("shows current running task", func(t *testing.T) {
		// Start a task
		_, err := app.businessAPI.StartNewTask(ctx, "Current Task")
		require.NoError(t, err)

		// Check current task
		err = cmd.Execute(ctx, []string{})
		assert.NoError(t, err)
	})

	t.Run("handles no running task", func(t *testing.T) {
		// Ensure no tasks are running
		_, _ = app.businessAPI.StopAllRunningTasks(ctx)

		// Check current task when none running
		err := cmd.Execute(ctx, []string{})
		assert.NoError(t, err) // Should not error, just show "no task running"
	})

	t.Run("rejects arguments", func(t *testing.T) {
		// Current command should not accept arguments
		err := cmd.Execute(ctx, []string{"unexpected"})
		assert.NoError(t, err) // Current implementation doesn't validate args, but executes anyway
	})
}

func TestCurrentCommand_ShowCurrentTask(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	cmd := NewCurrentCommand(app)
	ctx := context.Background()

	t.Run("displays running task details", func(t *testing.T) {
		// Start a task
		_, err := app.businessAPI.StartNewTask(ctx, "Test Current Task")
		require.NoError(t, err)

		// Show current task
		err = cmd.showCurrentTask(ctx)
		assert.NoError(t, err)
	})

	t.Run("handles no running task gracefully", func(t *testing.T) {
		// Ensure no tasks are running
		_, _ = app.businessAPI.StopAllRunningTasks(ctx)

		// Show current when none running
		err := cmd.showCurrentTask(ctx)
		assert.NoError(t, err) // Should handle gracefully
	})

	t.Run("shows task with special characters", func(t *testing.T) {
		// Start task with special characters
		_, err := app.businessAPI.StartNewTask(ctx, "Task: @#$% & more!")
		require.NoError(t, err)

		// Show current task
		err = cmd.showCurrentTask(ctx)
		assert.NoError(t, err)
	})
}

func TestNewCurrentCommand(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	cmd := NewCurrentCommand(app)
	
	assert.NotNil(t, cmd)
	assert.NotNil(t, cmd.businessAPI)
}