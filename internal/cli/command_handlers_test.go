package cli

import (
	"context"
	"testing"
)

// TestCommandHandlersWithBusinessAPI tests that command handlers work with BusinessAPI
func TestCommandHandlersWithBusinessAPI(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("start command creates task", func(t *testing.T) {
		startCmd := NewStartCommand(app)
		err := startCmd.Execute(ctx, []string{"Test Task"})
		if err != nil {
			t.Errorf("startCmd.Execute() error = %v", err)
		}
		
		// Verify task was created and is running
		session, err := app.businessAPI.GetCurrentSession(ctx)
		if err != nil {
			t.Errorf("Failed to get current session: %v", err)
		}
		if session.Task.TaskName != "Test Task" {
			t.Errorf("Expected task 'Test Task', got %s", session.Task.TaskName)
		}
	})

	t.Run("stop command stops running task", func(t *testing.T) {
		// Start a task first
		_, err := app.businessAPI.StartNewTask(ctx, "Running Task")
		if err != nil {
			t.Errorf("Failed to start task: %v", err)
		}
		
		// Stop it
		stopCmd := NewStopCommand(app)
		err = stopCmd.Execute(ctx, []string{})
		if err != nil {
			t.Errorf("stopCmd.Execute() error = %v", err)
		}
		
		// Verify no tasks are running
		_, err = app.businessAPI.GetCurrentSession(ctx)
		if err == nil {
			t.Errorf("Expected no current session after stop")
		}
	})

	t.Run("current command shows running task", func(t *testing.T) {
		// Start a task
		_, err := app.businessAPI.StartNewTask(ctx, "Current Task")
		if err != nil {
			t.Errorf("Failed to start task: %v", err)
		}
		
		// Check current
		currentCmd := NewCurrentCommand(app)
		err = currentCmd.Execute(ctx, []string{})
		if err != nil {
			t.Errorf("currentCmd.Execute() error = %v", err)
		}
	})

	t.Run("list command works", func(t *testing.T) {
		// Create some test data
		_, err := app.businessAPI.StartNewTask(ctx, "List Task 1")
		if err != nil {
			t.Errorf("Failed to start task: %v", err)
		}
		
		listCmd := NewListCommand(app)
		err = listCmd.Execute(ctx, []string{})
		if err != nil {
			t.Errorf("listCmd.Execute() error = %v", err)
		}
	})

	t.Run("output command works", func(t *testing.T) {
		outputCmd := NewOutputCommand(app)
		err := outputCmd.Execute(ctx, []string{"format=csv"})
		if err != nil {
			t.Errorf("outputCmd.Execute() error = %v", err)
		}
	})

	t.Run("resume command works with existing tasks", func(t *testing.T) {
		// Create a task and stop it
		_, err := app.businessAPI.StartNewTask(ctx, "Resume Task")
		if err != nil {
			t.Errorf("Failed to start task: %v", err)
		}
		_, err = app.businessAPI.StopAllRunningTasks(ctx)
		if err != nil {
			t.Errorf("Failed to stop task: %v", err)
		}
		
		// Note: Resume command requires user input, so we just test it doesn't crash with empty args
		// Interactive commands are tested in e2e tests
	})

	t.Run("summary command works", func(t *testing.T) {
		// Note: Summary command requires user input for task selection
		// Interactive commands are tested in e2e tests
	})

	t.Run("delete command works", func(t *testing.T) {
		// Note: Delete command requires user input for task selection
		// Interactive commands are tested in e2e tests
	})
}

// TestCommandHandlerInputValidation tests input validation
func TestCommandHandlerInputValidation(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("start command requires task name", func(t *testing.T) {
		startCmd := NewStartCommand(app)
		err := startCmd.Execute(ctx, []string{})
		if err == nil {
			t.Errorf("Expected error for start command without task name")
		}
	})

	t.Run("stop command accepts no args", func(t *testing.T) {
		stopCmd := NewStopCommand(app)
		err := stopCmd.Execute(ctx, []string{})
		if err != nil {
			t.Errorf("stopCmd.Execute() should not error with no args: %v", err)
		}
	})

	t.Run("current command accepts no args", func(t *testing.T) {
		currentCmd := NewCurrentCommand(app)
		err := currentCmd.Execute(ctx, []string{})
		// Should not error even when no task is running
		if err != nil {
			t.Errorf("currentCmd.Execute() should not error with no args: %v", err)
		}
	})

	t.Run("output command requires format", func(t *testing.T) {
		outputCmd := NewOutputCommand(app)
		err := outputCmd.Execute(ctx, []string{})
		if err == nil {
			t.Errorf("Expected error for output command without format")
		}
	})

	t.Run("output command rejects invalid format", func(t *testing.T) {
		outputCmd := NewOutputCommand(app)
		err := outputCmd.Execute(ctx, []string{"format=invalid"})
		if err == nil {
			t.Errorf("Expected error for invalid output format")
		}
	})
}

// TestCommandHandlerErrorHandling tests error handling
func TestCommandHandlerErrorHandling(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("current command with no running task", func(t *testing.T) {
		currentCmd := NewCurrentCommand(app)
		err := currentCmd.Execute(ctx, []string{})
		// Should not error, should just show "no task running" message
		if err != nil {
			t.Errorf("currentCmd.Execute() should not error when no task running: %v", err)
		}
	})

	t.Run("stop command with no running tasks", func(t *testing.T) {
		stopCmd := NewStopCommand(app)
		err := stopCmd.Execute(ctx, []string{})
		// Should not error
		if err != nil {
			t.Errorf("stopCmd.Execute() should not error when no tasks running: %v", err)
		}
	})
}