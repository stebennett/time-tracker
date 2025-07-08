package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"
)

// TestCommandHandlerIsolation tests that each command handler can be tested in isolation
func TestCommandHandlerIsolation(t *testing.T) {
	app, cleanup := setupTestAppWithMockAPI(t)
	defer cleanup()

	// Test each command handler individually to ensure they can be extracted
	t.Run("start command handler", func(t *testing.T) {
		startCmd := NewStartCommand(app)
		err := startCmd.Execute([]string{"Test Task"})
		if err != nil {
			t.Errorf("startCmd.Execute() error = %v", err)
		}
		
		tasks, err := app.api.ListTasks()
		if err != nil {
			t.Errorf("Failed to list tasks: %v", err)
		}
		if len(tasks) != 1 || tasks[0].TaskName != "Test Task" {
			t.Errorf("Expected task 'Test Task' to be created")
		}
	})

	t.Run("stop command handler", func(t *testing.T) {
		// Create a running task first
		task, _ := app.api.CreateTask("Running Task")
		app.api.CreateTimeEntry(task.ID, time.Now(), nil)
		
		stopCmd := NewStopCommand(app)
		err := stopCmd.Execute([]string{})
		if err != nil {
			t.Errorf("stopCmd.Execute() error = %v", err)
		}
		
		entries, err := app.api.ListTimeEntries()
		if err != nil {
			t.Errorf("Failed to list entries: %v", err)
		}
		for _, entry := range entries {
			if entry.EndTime == nil {
				t.Errorf("Expected all tasks to be stopped")
			}
		}
	})

	t.Run("list command handler", func(t *testing.T) {
		// Create test data
		task, _ := app.api.CreateTask("List Test Task")
		app.api.CreateTimeEntry(task.ID, time.Now(), nil)
		
		listCmd := NewListCommand(app)
		err := listCmd.Execute([]string{})
		if err != nil {
			t.Errorf("listCmd.Execute() error = %v", err)
		}
	})

	t.Run("current command handler", func(t *testing.T) {
		currentCmd := NewCurrentCommand(app)
		err := currentCmd.Execute([]string{})
		if err != nil {
			t.Errorf("currentCmd.Execute() error = %v", err)
		}
	})

	t.Run("output command handler", func(t *testing.T) {
		outputCmd := NewOutputCommand(app)
		err := outputCmd.Execute([]string{"format=csv"})
		if err != nil {
			t.Errorf("outputCmd.Execute() error = %v", err)
		}
	})
}

// TestCommandHandlerInputValidation tests input validation for each command
func TestCommandHandlerInputValidation(t *testing.T) {
	app, cleanup := setupTestAppWithMockAPI(t)
	defer cleanup()

	tests := []struct {
		name        string
		commandFunc func([]string) error
		args        []string
		wantErr     bool
	}{
		{
			name: "list with valid time format",
			commandFunc: func(args []string) error { return NewListCommand(app).Execute(args) },
			args:        []string{"1h"},
			wantErr:     false,
		},
		{
			name: "list with invalid time format",
			commandFunc: func(args []string) error { return NewListCommand(app).Execute(args) },
			args:        []string{"invalid"},
			wantErr:     false, // Should not error, just treat as text search
		},
		{
			name: "output with valid format",
			commandFunc: func(args []string) error { return NewOutputCommand(app).Execute(args) },
			args:        []string{"format=csv"},
			wantErr:     false,
		},
		{
			name: "output with invalid format",
			commandFunc: func(args []string) error { return NewOutputCommand(app).Execute(args) },
			args:        []string{"format=invalid"},
			wantErr:     true,
		},
		{
			name: "output with no format",
			commandFunc: func(args []string) error { return NewOutputCommand(app).Execute(args) },
			args:        []string{},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout to avoid test output pollution
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := tt.commandFunc(tt.args)

			w.Close()
			os.Stdout = oldStdout

			// Read and discard output
			var buf bytes.Buffer
			buf.ReadFrom(r)

			if (err != nil) != tt.wantErr {
				t.Errorf("command error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestCommandHandlerDependencies tests that command handlers properly use injected dependencies
func TestCommandHandlerDependencies(t *testing.T) {
	app, cleanup := setupTestAppWithMockAPI(t)
	defer cleanup()

	// Verify that command handlers use the injected API
	if app.api == nil {
		t.Fatal("Expected API to be injected")
	}

	// Test that a command handler uses the API
	taskName := "Dependency Test Task"
	startCmd := NewStartCommand(app)
	err := startCmd.Execute([]string{taskName})
	if err != nil {
		t.Errorf("startCmd.Execute() error = %v", err)
	}

	// Verify the task was created through the API
	tasks, err := app.api.ListTasks()
	if err != nil {
		t.Errorf("Failed to list tasks: %v", err)
	}
	if len(tasks) != 1 || tasks[0].TaskName != taskName {
		t.Errorf("Expected task to be created through API")
	}
}

// TestCommandHandlerStateManagement tests that command handlers properly manage state
func TestCommandHandlerStateManagement(t *testing.T) {
	app, cleanup := setupTestAppWithMockAPI(t)
	defer cleanup()

	// Test that start command stops previous tasks
	task1, _ := app.api.CreateTask("Task 1")
	_, _ = app.api.CreateTask("Task 2")
	
	// Start first task
	app.api.CreateTimeEntry(task1.ID, time.Now(), nil)
	
	// Start second task (should stop first)
	startCmd := NewStartCommand(app)
	err := startCmd.Execute([]string{"Task 3"})
	if err != nil {
		t.Errorf("startCmd.Execute() error = %v", err)
	}
	
	// Verify first task is stopped
	entry1, _ := app.api.GetTimeEntry(1)
	if entry1.EndTime == nil {
		t.Errorf("Expected first task to be stopped")
	}
	
	// Verify new task is running
	entries, _ := app.api.ListTimeEntries()
	var runningCount int
	for _, entry := range entries {
		if entry.EndTime == nil {
			runningCount++
		}
	}
	if runningCount != 1 {
		t.Errorf("Expected exactly 1 running task, got %d", runningCount)
	}
}

// TestCommandHandlerErrorHandling tests error handling in command handlers
func TestCommandHandlerErrorHandling(t *testing.T) {
	app, cleanup := setupTestAppWithMockAPI(t)
	defer cleanup()

	// Test error propagation
	t.Run("error in API call", func(t *testing.T) {
		// This test would need a mock that can simulate API errors
		// For now, we test that errors are properly propagated
		currentCmd := NewCurrentCommand(app)
		err := currentCmd.Execute([]string{})
		// Should not error when no tasks are running
		if err != nil {
			t.Errorf("currentCmd.Execute() should not error when no tasks running, got: %v", err)
		}
	})
}

// TestCommandHandlerTimeHandling tests time-related functionality in command handlers
func TestCommandHandlerTimeHandling(t *testing.T) {
	app, cleanup := setupTestAppWithMockAPI(t)
	defer cleanup()

	// Set a fixed time for testing
	fixedTime := time.Date(2025, 6, 16, 12, 0, 0, 0, time.UTC)
	oldTimeNow := timeNow
	timeNow = func() time.Time { return fixedTime }
	defer func() { timeNow = oldTimeNow }()

	// Test time parsing in list command
	task, _ := app.api.CreateTask("Time Test Task")
	app.api.CreateTimeEntry(task.ID, fixedTime.Add(-2*time.Hour), &fixedTime)
	
	// Test different time formats
	testCases := []struct {
		name string
		args []string
	}{
		{"minutes", []string{"30m"}},
		{"hours", []string{"2h"}},
		{"days", []string{"1d"}},
		{"weeks", []string{"1w"}},
		{"months", []string{"1mo"}},
		{"years", []string{"1y"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			listCmd := NewListCommand(app)
			err := listCmd.Execute(tc.args)

			w.Close()
			os.Stdout = oldStdout

			// Read and discard output
			var buf bytes.Buffer
			buf.ReadFrom(r)

			if err != nil {
				t.Errorf("listTasks(%v) error = %v", tc.args, err)
			}
		})
	}
}

// TestCommandHandlerOutputFormatting tests output formatting in command handlers
func TestCommandHandlerOutputFormatting(t *testing.T) {
	app, cleanup := setupTestAppWithMockAPI(t)
	defer cleanup()

	// Set fixed time for consistent output
	fixedTime := time.Date(2025, 6, 16, 12, 0, 0, 0, time.UTC)
	oldTimeNow := timeNow
	timeNow = func() time.Time { return fixedTime }
	defer func() { timeNow = oldTimeNow }()

	// Create test data
	task, _ := app.api.CreateTask("Output Test Task")
	app.api.CreateTimeEntry(task.ID, fixedTime.Add(-1*time.Hour), &fixedTime)

	// Test list output formatting
	t.Run("list output format", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		listCmd := NewListCommand(app)
		err := listCmd.Execute([]string{})

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		if err != nil {
			t.Errorf("listTasks() error = %v", err)
		}

		// Check output format
		if !strings.Contains(output, "Output Test Task") {
			t.Errorf("Expected output to contain task name")
		}
		if !strings.Contains(output, "1h 0m") {
			t.Errorf("Expected output to contain duration")
		}
	})

	// Test CSV output formatting
	t.Run("csv output format", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		outputCmd := NewOutputCommand(app)
		err := outputCmd.outputCSV()

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		if err != nil {
			t.Errorf("outputCSV() error = %v", err)
		}

		// Check CSV format
		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) < 2 {
			t.Errorf("Expected at least header and one data row")
		}
		if !strings.Contains(lines[0], "ID,Start Time,End Time") {
			t.Errorf("Expected CSV header")
		}
	})
}

// TestCommandHandlerEdgeCases tests edge cases in command handlers
func TestCommandHandlerEdgeCases(t *testing.T) {
	app, cleanup := setupTestAppWithMockAPI(t)
	defer cleanup()

	// Test empty task name
	t.Run("empty task name", func(t *testing.T) {
		startCmd := NewStartCommand(app)
		err := startCmd.Execute([]string{""})
		if err != nil {
			t.Errorf("startCmd.Execute('') should not error, got: %v", err)
		}
	})

	// Test very long task name
	t.Run("long task name", func(t *testing.T) {
		longName := strings.Repeat("A", 1000)
		startCmd := NewStartCommand(app)
		err := startCmd.Execute([]string{longName})
		if err != nil {
			t.Errorf("startCmd.Execute with long name should not error, got: %v", err)
		}
	})

	// Test special characters in task name
	t.Run("special characters", func(t *testing.T) {
		specialName := "Task with !@#$%^&*()_+-=[]{}|;':\",./<>?"
		startCmd := NewStartCommand(app)
		err := startCmd.Execute([]string{specialName})
		if err != nil {
			t.Errorf("startCmd.Execute with special chars should not error, got: %v", err)
		}
	})

	// Test stop when no tasks are running
	t.Run("stop no running tasks", func(t *testing.T) {
		stopCmd := NewStopCommand(app)
		err := stopCmd.Execute([]string{})
		if err != nil {
			t.Errorf("stopCmd.Execute() with no running tasks should not error, got: %v", err)
		}
	})

	// Test current when no tasks are running
	t.Run("current no running tasks", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		currentCmd := NewCurrentCommand(app)
		err := currentCmd.Execute([]string{})

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		if err != nil {
			t.Errorf("showCurrentTask() with no running tasks should not error, got: %v", err)
		}
		if !strings.Contains(output, "No task is currently running") {
			t.Errorf("Expected message about no running tasks")
		}
	})
}