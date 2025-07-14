package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCommandRegistry(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	registry := NewCommandRegistry(app)
	
	assert.NotNil(t, registry)
	assert.NotNil(t, registry.commands)
}

func TestCommandRegistry_Execute(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	registry := NewCommandRegistry(app)
	ctx := context.Background()

	t.Run("executes start command", func(t *testing.T) {
		err := registry.Execute(ctx, "start", []string{"Test Task"})
		assert.NoError(t, err)

		// Verify task was created
		session, err := app.businessAPI.GetCurrentSession(ctx)
		require.NoError(t, err)
		assert.Equal(t, "Test Task", session.Task.TaskName)
	})

	t.Run("executes stop command", func(t *testing.T) {
		// Start a task first
		_, err := app.businessAPI.StartNewTask(ctx, "Task to Stop")
		require.NoError(t, err)

		// Stop it
		err = registry.Execute(ctx, "stop", []string{})
		assert.NoError(t, err)

		// Verify no task is running
		_, err = app.businessAPI.GetCurrentSession(ctx)
		assert.Error(t, err)
	})

	t.Run("executes current command", func(t *testing.T) {
		err := registry.Execute(ctx, "current", []string{})
		assert.NoError(t, err)
	})

	t.Run("executes list command", func(t *testing.T) {
		err := registry.Execute(ctx, "list", []string{})
		assert.NoError(t, err)
	})

	t.Run("executes output command", func(t *testing.T) {
		err := registry.Execute(ctx, "output", []string{"format=csv"})
		assert.NoError(t, err)
	})

	t.Run("handles unknown command", func(t *testing.T) {
		err := registry.Execute(ctx, "unknown", []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown command")
	})

	t.Run("handles empty command", func(t *testing.T) {
		err := registry.Execute(ctx, "", []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown command")
	})
}

func TestCommandRegistry_GetUsage(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	registry := NewCommandRegistry(app)
	
	usage := registry.GetUsage()
	assert.NotEmpty(t, usage)
	
	// Check that usage contains expected commands
	assert.Contains(t, usage, "start")
	assert.Contains(t, usage, "stop")
	assert.Contains(t, usage, "current")
	assert.Contains(t, usage, "list")
	assert.Contains(t, usage, "output")
	assert.Contains(t, usage, "resume")
	assert.Contains(t, usage, "summary")
	assert.Contains(t, usage, "delete")
}

func TestCommandRegistry_CommandsRegistered(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	registry := NewCommandRegistry(app)
	
	// Verify all expected commands are registered
	expectedCommands := []string{
		"start",
		"stop", 
		"current",
		"list",
		"output",
		"resume",
		"summary",
		"delete",
	}
	
	for _, cmd := range expectedCommands {
		t.Run("command "+cmd+" is registered", func(t *testing.T) {
			// Try to execute each command to verify it exists
			// We don't care about success/failure, just that it's recognized
			err := registry.Execute(context.Background(), cmd, []string{})
			// Should not get "unknown command" error
			if err != nil {
				assert.NotContains(t, err.Error(), "unknown command")
			}
		})
	}
}

func TestCommandRegistry_ErrorPropagation(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	registry := NewCommandRegistry(app)
	ctx := context.Background()

	t.Run("propagates command validation errors", func(t *testing.T) {
		// Start command without arguments should error
		err := registry.Execute(ctx, "start", []string{})
		assert.Error(t, err)
	})

	t.Run("propagates format errors", func(t *testing.T) {
		// Output command with invalid format should error
		err := registry.Execute(ctx, "output", []string{"format=invalid"})
		assert.Error(t, err)
	})
}