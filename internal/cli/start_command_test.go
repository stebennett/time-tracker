package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartCommand_Execute(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	cmd := NewStartCommand(app)
	ctx := context.Background()

	t.Run("successful task creation", func(t *testing.T) {
		err := cmd.Execute(ctx, []string{"Test Task"})
		assert.NoError(t, err)

		// Verify task was created and is running
		session, err := app.businessAPI.GetCurrentSession(ctx)
		require.NoError(t, err)
		assert.Equal(t, "Test Task", session.Task.TaskName)
	})

	t.Run("task creation with multiple words", func(t *testing.T) {
		err := cmd.Execute(ctx, []string{"Multi", "Word", "Task", "Name"})
		assert.NoError(t, err)

		session, err := app.businessAPI.GetCurrentSession(ctx)
		require.NoError(t, err)
		assert.Equal(t, "Multi Word Task Name", session.Task.TaskName)
	})

	t.Run("task creation with special characters", func(t *testing.T) {
		err := cmd.Execute(ctx, []string{"Task with special chars: @#$%"})
		assert.NoError(t, err)

		session, err := app.businessAPI.GetCurrentSession(ctx)
		require.NoError(t, err)
		assert.Equal(t, "Task with special chars: @#$%", session.Task.TaskName)
	})

	t.Run("stopping previous task when starting new one", func(t *testing.T) {
		// Start first task
		err := cmd.Execute(ctx, []string{"First Task"})
		assert.NoError(t, err)

		// Start second task - should stop first
		err = cmd.Execute(ctx, []string{"Second Task"})
		assert.NoError(t, err)

		// Verify only second task is running
		session, err := app.businessAPI.GetCurrentSession(ctx)
		require.NoError(t, err)
		assert.Equal(t, "Second Task", session.Task.TaskName)
	})

	t.Run("error cases", func(t *testing.T) {
		// No arguments
		err := cmd.Execute(ctx, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "usage: tt start")

		// Empty task name
		err = cmd.Execute(ctx, []string{""})
		// Should not error - empty string is still a valid task name
		assert.NoError(t, err)
	})
}

func TestStartCommand_CreateNewTask(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	cmd := NewStartCommand(app)
	ctx := context.Background()

	t.Run("creates task when none running", func(t *testing.T) {
		err := cmd.createNewTask(ctx, "New Task")
		assert.NoError(t, err)

		session, err := app.businessAPI.GetCurrentSession(ctx)
		require.NoError(t, err)
		assert.Equal(t, "New Task", session.Task.TaskName)
	})

	t.Run("stops running task and creates new one", func(t *testing.T) {
		// Start a task first
		_, err := app.businessAPI.StartNewTask(ctx, "Running Task")
		require.NoError(t, err)

		// Create new task - should stop previous
		err = cmd.createNewTask(ctx, "New Task")
		assert.NoError(t, err)

		session, err := app.businessAPI.GetCurrentSession(ctx)
		require.NoError(t, err)
		assert.Equal(t, "New Task", session.Task.TaskName)
	})
}

func TestNewStartCommand(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	cmd := NewStartCommand(app)
	
	assert.NotNil(t, cmd)
	assert.NotNil(t, cmd.businessAPI)
	assert.NotNil(t, cmd.errorHandler)
}