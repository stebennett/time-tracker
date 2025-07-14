package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListCommand_Execute(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	cmd := NewListCommand(app)
	ctx := context.Background()

	t.Run("lists all tasks when no arguments", func(t *testing.T) {
		// Create some test tasks
		_, err := app.businessAPI.StartNewTask(ctx, "Task 1")
		require.NoError(t, err)
		_, err = app.businessAPI.StopAllRunningTasks(ctx)
		require.NoError(t, err)

		_, err = app.businessAPI.StartNewTask(ctx, "Task 2")
		require.NoError(t, err)
		_, err = app.businessAPI.StopAllRunningTasks(ctx)
		require.NoError(t, err)

		// List all tasks
		err = cmd.Execute(ctx, []string{})
		assert.NoError(t, err)
	})

	t.Run("lists tasks with time filter", func(t *testing.T) {
		// Create a task
		_, err := app.businessAPI.StartNewTask(ctx, "Recent Task")
		require.NoError(t, err)

		// List tasks from last hour
		err = cmd.Execute(ctx, []string{"1h"})
		assert.NoError(t, err)
	})

	t.Run("lists tasks with text filter", func(t *testing.T) {
		// Create tasks with different names
		_, err := app.businessAPI.StartNewTask(ctx, "Meeting Task")
		require.NoError(t, err)
		_, err = app.businessAPI.StopAllRunningTasks(ctx)
		require.NoError(t, err)

		_, err = app.businessAPI.StartNewTask(ctx, "Coding Task")
		require.NoError(t, err)

		// List tasks containing "Meeting"
		err = cmd.Execute(ctx, []string{"Meeting"})
		assert.NoError(t, err)
	})

	t.Run("lists tasks with time and text filter", func(t *testing.T) {
		// Create a task
		_, err := app.businessAPI.StartNewTask(ctx, "Filtered Task")
		require.NoError(t, err)

		// List with both time and text filter
		err = cmd.Execute(ctx, []string{"1h", "Filtered"})
		assert.NoError(t, err)
	})

	t.Run("handles invalid time format", func(t *testing.T) {
		err := cmd.Execute(ctx, []string{"invalid_time"})
		// Should treat as text filter, not error
		assert.NoError(t, err)
	})

	t.Run("handles empty results", func(t *testing.T) {
		// Search for non-existent task
		err := cmd.Execute(ctx, []string{"NonExistentTask"})
		assert.NoError(t, err) // Should not error, just show "No tasks found"
	})
}

func TestListCommand_ListTasks(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	cmd := NewListCommand(app)
	ctx := context.Background()

	t.Run("processes no arguments correctly", func(t *testing.T) {
		err := cmd.listTasks(ctx, []string{})
		assert.NoError(t, err)
	})

	t.Run("processes time shorthand correctly", func(t *testing.T) {
		err := cmd.listTasks(ctx, []string{"1h"})
		assert.NoError(t, err)
	})

	t.Run("processes text filter correctly", func(t *testing.T) {
		err := cmd.listTasks(ctx, []string{"SomeTask"})
		assert.NoError(t, err)
	})

	t.Run("processes combined filters correctly", func(t *testing.T) {
		err := cmd.listTasks(ctx, []string{"1h", "Task", "Name"})
		assert.NoError(t, err)
	})
}

func TestListCommand_PrintTimeEntries(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	cmd := NewListCommand(app)
	ctx := context.Background()

	t.Run("prints empty results", func(t *testing.T) {
		entries, err := app.businessAPI.SearchTimeEntries(ctx, "", "NonExistent")
		require.NoError(t, err)
		
		err = cmd.printTimeEntries(ctx, entries)
		assert.NoError(t, err)
	})

	t.Run("prints entries with various states", func(t *testing.T) {
		// Create running and stopped tasks
		_, err := app.businessAPI.StartNewTask(ctx, "Running Task")
		require.NoError(t, err)
		
		// Get entries
		entries, err := app.businessAPI.SearchTimeEntries(ctx, "", "")
		require.NoError(t, err)
		
		err = cmd.printTimeEntries(ctx, entries)
		assert.NoError(t, err)
	})
}

func TestListCommand_ConfigMethods(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	cmd := NewListCommand(app)

	t.Run("getTimeFormat returns default when no config", func(t *testing.T) {
		format := cmd.getTimeFormat()
		assert.Equal(t, "2006-01-02 15:04:05", format)
	})

	t.Run("getRunningStatus returns default when no config", func(t *testing.T) {
		status := cmd.getRunningStatus()
		assert.Equal(t, "running", status)
	})

	t.Run("truncateTaskName handles long names", func(t *testing.T) {
		longName := "This is a very long task name that should be truncated"
		truncated := cmd.truncateTaskName(longName)
		// Without config, should return original name
		assert.Equal(t, longName, truncated)
	})
}

func TestNewListCommand(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	cmd := NewListCommand(app)
	
	assert.NotNil(t, cmd)
	assert.NotNil(t, cmd.businessAPI)
}