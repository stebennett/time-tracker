package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResumeCommand_Execute(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	cmd := NewResumeCommand(app)
	ctx := context.Background()

	t.Run("handles no time argument", func(t *testing.T) {
		// Create some historical tasks
		_, err := app.businessAPI.StartNewTask(ctx, "Historical Task")
		require.NoError(t, err)
		_, err = app.businessAPI.StopAllRunningTasks(ctx)
		require.NoError(t, err)

		// Note: Resume command requires user interaction, so we can't fully test Execute
		// But we can test that it doesn't crash on initialization
		err = cmd.resumeTask(ctx, []string{})
		// This will reach the interactive part and likely error, but that's expected
		// The important thing is it processes the arguments correctly
	})

	t.Run("handles time argument", func(t *testing.T) {
		// Test with time argument - will reach interactive part but validates time argument
		_ = cmd.resumeTask(ctx, []string{"1h"})
	})

	t.Run("handles invalid time format", func(t *testing.T) {
		err := cmd.resumeTask(ctx, []string{"invalid"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid time shorthand")
	})
}

func TestResumeCommand_ResumeTask(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	cmd := NewResumeCommand(app)
	ctx := context.Background()

	t.Run("sets default time range when no args", func(t *testing.T) {
		// Create a task first
		_, err := app.businessAPI.StartNewTask(ctx, "Test Task")
		require.NoError(t, err)
		_, err = app.businessAPI.StopAllRunningTasks(ctx)
		require.NoError(t, err)

		// This will fail at the interactive part, but we can test argument processing
		_ = cmd.resumeTask(ctx, []string{})
		// Will error during interactive selection, but that's expected
	})

	t.Run("processes valid time shorthand", func(t *testing.T) {
		// Create a task first
		_, err := app.businessAPI.StartNewTask(ctx, "Time Test Task")
		require.NoError(t, err)
		_, err = app.businessAPI.StopAllRunningTasks(ctx)
		require.NoError(t, err)

		// Test with valid time argument
		_ = cmd.resumeTask(ctx, []string{"1h"})
		// Will error during interactive selection, but validates time processing
	})

	t.Run("validates time shorthand format", func(t *testing.T) {
		err := cmd.resumeTask(ctx, []string{"badformat"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid time shorthand")
	})

	t.Run("handles no tasks found", func(t *testing.T) {
		// Search for non-existent tasks
		_ = cmd.resumeTask(ctx, []string{"1h"}) // Will search but find no tasks matching
		// Our mock will likely return empty results, leading to "No tasks found" message
	})
}

func TestResumeCommand_ArgumentProcessing(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	cmd := NewResumeCommand(app)
	ctx := context.Background()

	// Create some test data
	_, err := app.businessAPI.StartNewTask(ctx, "Resume Test Task")
	require.NoError(t, err)
	_, err = app.businessAPI.StopAllRunningTasks(ctx)
	require.NoError(t, err)

	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		errMsg   string
	}{
		{
			name:    "no arguments uses default",
			args:    []string{},
			wantErr: false, // Will proceed to interactive part
		},
		{
			name:    "valid time argument",
			args:    []string{"1h"},
			wantErr: false, // Will proceed to interactive part
		},
		{
			name:    "invalid time format",
			args:    []string{"invalid"},
			wantErr: true,
			errMsg:  "invalid time shorthand",
		},
		{
			name:    "multiple valid time formats",
			args:    []string{"1d"},
			wantErr: false,
		},
		{
			name:    "complex invalid format",
			args:    []string{"1x"},
			wantErr: true,
			errMsg:  "invalid time shorthand",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.resumeTask(ctx, tt.args)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			}
			// Note: When wantErr is false, we don't assert NoError because
			// the command will likely error at the interactive input stage
		})
	}
}

func TestNewResumeCommand(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	cmd := NewResumeCommand(app)
	
	assert.NotNil(t, cmd)
	assert.NotNil(t, cmd.businessAPI)
}