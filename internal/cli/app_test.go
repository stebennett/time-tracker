package cli

import (
	"bytes"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"time-tracker/internal/repository/sqlite"
)

func setupTestAppWithInMemoryRepo(t *testing.T) (*App, func()) {
	repo, err := sqlite.New(":memory:")
	if err != nil {
		t.Fatalf("failed to create in-memory repo: %v", err)
	}
	app := NewApp(repo)
	cleanup := func() { repo.Close() }
	return app, cleanup
}

func TestApp_Run(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr bool
	}{
		{
			name:    "empty args",
			args:    []string{},
			want:    "",
			wantErr: true,
		},
		{
			name:    "stop command",
			args:    []string{"stop"},
			want:    "All running tasks have been stopped\n",
			wantErr: false,
		},
		{
			name:    "stop now is a new task",
			args:    []string{"stop", "now"},
			want:    "All running tasks have been stopped\nStarted new task: stop now\n",
			wantErr: false,
		},
		{
			name:    "stop working is a new task",
			args:    []string{"stop", "working"},
			want:    "All running tasks have been stopped\nStarted new task: stop working\n",
			wantErr: false,
		},
		{
			name:    "new task",
			args:    []string{"Working on feature X"},
			want:    "All running tasks have been stopped\nStarted new task: Working on feature X\n",
			wantErr: false,
		},
		{
			name:    "multiple words task",
			args:    []string{"Working", "on", "feature", "X"},
			want:    "All running tasks have been stopped\nStarted new task: Working on feature X\n",
			wantErr: false,
		},
		{
			name:    "list command",
			args:    []string{"list"},
			want:    "No tasks found\n",
			wantErr: false,
		},
		{
			name:    "list with invalid time format",
			args:    []string{"list", "invalid"},
			want:    "No tasks found\n",
			wantErr: false,
		},
		{
			name:    "current command with no running task",
			args:    []string{"current"},
			want:    "No task is currently running\n",
			wantErr: false,
		},
		{
			name:    "output command without format",
			args:    []string{"output"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "output command with invalid format",
			args:    []string{"output", "format=invalid"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "summary command with no tasks",
			args:    []string{"summary"},
			want:    "No tasks found matching the criteria.\n",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, cleanup := setupTestAppWithInMemoryRepo(t)
			defer cleanup()

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run app
			err := app.Run(tt.args)

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Read captured output
			var buf bytes.Buffer
			buf.ReadFrom(r)
			got := buf.String()

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("App.Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check output
			if !tt.wantErr && got != tt.want {
				t.Errorf("App.Run() output = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseTimeShorthand(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
	}{
		{
			name:    "valid minutes",
			input:   "30m",
			want:    30 * time.Minute,
			wantErr: false,
		},
		{
			name:    "valid hours",
			input:   "2h",
			want:    2 * time.Hour,
			wantErr: false,
		},
		{
			name:    "valid days",
			input:   "1d",
			want:    24 * time.Hour,
			wantErr: false,
		},
		{
			name:    "valid weeks",
			input:   "2w",
			want:    14 * 24 * time.Hour,
			wantErr: false,
		},
		{
			name:    "valid months",
			input:   "3mo",
			want:    90 * 24 * time.Hour,
			wantErr: false,
		},
		{
			name:    "valid years",
			input:   "1y",
			want:    365 * 24 * time.Hour,
			wantErr: false,
		},
		{
			name:    "invalid format",
			input:   "invalid",
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid number",
			input:   "abc",
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid unit",
			input:   "1x",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTimeShorthand(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTimeShorthand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseTimeShorthand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListTasks(t *testing.T) {
	app, cleanup := setupTestAppWithInMemoryRepo(t)
	defer cleanup()

	// Create a fixed time for testing
	fixedTime := time.Date(2025, 6, 16, 11, 22, 1, 0, time.UTC)
	timeNow = func() time.Time { return fixedTime }

	// Create some test tasks
	task1 := &sqlite.Task{TaskName: "First task"}
	task2 := &sqlite.Task{TaskName: "Second task"}
	task3 := &sqlite.Task{TaskName: "Third task"}
	app.repo.CreateTask(task1)
	app.repo.CreateTask(task2)
	app.repo.CreateTask(task3)

	// Create some test entries
	entries := []*sqlite.TimeEntry{
		{
			StartTime: fixedTime.Add(-2 * time.Hour),
			EndTime:   &fixedTime,
			TaskID:    task1.ID,
		},
		{
			StartTime: fixedTime.Add(-1 * time.Hour),
			TaskID:    task2.ID,
		},
		{
			StartTime: fixedTime.Add(-30 * time.Minute),
			EndTime:   &fixedTime,
			TaskID:    task3.ID,
		},
	}

	for _, entry := range entries {
		err := app.repo.CreateTimeEntry(entry)
		if err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr bool
	}{
		{
			name:    "list all",
			args:    []string{},
			want:    "2025-06-16 09:22:01 - 2025-06-16 11:22:01 (2h 0m): First task\n2025-06-16 10:52:01 - 2025-06-16 11:22:01 (0h 30m): Third task\n2025-06-16 10:22:01 - running (1h 0m): Second task\n",
			wantErr: false,
		},
		{
			name:    "list last hour",
			args:    []string{"1h"},
			want:    "2025-06-16 10:52:01 - 2025-06-16 11:22:01 (0h 30m): Third task\n2025-06-16 10:22:01 - running (1h 0m): Second task\n",
			wantErr: false,
		},
		{
			name:    "list with text filter",
			args:    []string{"task"},
			want:    "2025-06-16 09:22:01 - 2025-06-16 11:22:01 (2h 0m): First task\n2025-06-16 10:52:01 - 2025-06-16 11:22:01 (0h 30m): Third task\n2025-06-16 10:22:01 - running (1h 0m): Second task\n",
			wantErr: false,
		},
		{
			name:    "list with time and text filter",
			args:    []string{"1h", "Second"},
			want:    "2025-06-16 10:22:01 - running (1h 0m): Second task\n",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run list command
			err := app.listTasks(tt.args)

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Read captured output
			var buf bytes.Buffer
			buf.ReadFrom(r)
			got := buf.String()

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("listTasks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check output
			if !tt.wantErr && got != tt.want {
				t.Errorf("listTasks() output = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShowCurrentTask(t *testing.T) {
	app, cleanup := setupTestAppWithInMemoryRepo(t)
	defer cleanup()

	tests := []struct {
		name       string
		setupTask  bool
		wantPrefix string
		wantErr    bool
	}{
		{
			name:       "no running task",
			setupTask:  false,
			wantPrefix: "No task is currently running",
			wantErr:    false,
		},
		{
			name:       "has running task",
			setupTask:  true,
			wantPrefix: "Current task: Test task (running for",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupTask {
				// Create a running task
				task := &sqlite.Task{TaskName: "Test task"}
				app.repo.CreateTask(task)
				entry := &sqlite.TimeEntry{
					StartTime: time.Now(),
					TaskID:    task.ID,
				}
				err := app.repo.CreateTimeEntry(entry)
				if err != nil {
					t.Fatalf("Failed to create test entry: %v", err)
				}
			}

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run current command
			err := app.showCurrentTask()

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Read captured output
			var buf bytes.Buffer
			buf.ReadFrom(r)
			got := buf.String()

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("showCurrentTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check output prefix
			if !tt.wantErr && !strings.HasPrefix(got, tt.wantPrefix) {
				t.Errorf("showCurrentTask() output = %v, want prefix %v", got, tt.wantPrefix)
			}
		})
	}
}

func TestOutputCSV(t *testing.T) {
	app, cleanup := setupTestAppWithInMemoryRepo(t)
	defer cleanup()

	// Create test tasks
	task1 := &sqlite.Task{TaskName: "First task"}
	task2 := &sqlite.Task{TaskName: "Second task"}
	app.repo.CreateTask(task1)
	app.repo.CreateTask(task2)

	// Create test entries
	now := time.Now()
	entries := []*sqlite.TimeEntry{
		{
			StartTime: now.Add(-2 * time.Hour),
			EndTime:   &now,
			TaskID:    task1.ID,
		},
		{
			StartTime: now.Add(-1 * time.Hour),
			TaskID:    task2.ID,
		},
	}

	for _, entry := range entries {
		err := app.repo.CreateTimeEntry(entry)
		if err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run output command
	err := app.outputCSV()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	got := buf.String()

	// Check error
	if err != nil {
		t.Errorf("outputCSV() error = %v", err)
		return
	}

	// Check CSV format
	lines := strings.Split(strings.TrimSpace(got), "\n")
	if len(lines) != 3 { // Header + 2 entries
		t.Errorf("outputCSV() got %d lines, want 3", len(lines))
		return
	}

	// Check header
	header := strings.Split(lines[0], ",")
	expectedHeader := []string{"ID", "Start Time", "End Time", "Duration (hours)", "Task Name"}
	if len(header) != len(expectedHeader) {
		t.Errorf("outputCSV() header has %d columns, want %d", len(header), len(expectedHeader))
		return
	}

	// Check first entry
	firstEntry := strings.Split(lines[1], ",")
	if len(firstEntry) != len(expectedHeader) {
		t.Errorf("outputCSV() first entry has %d columns, want %d", len(firstEntry), len(expectedHeader))
		return
	}

	// Check that duration is approximately 2 hours
	duration, err := strconv.ParseFloat(firstEntry[3], 64)
	if err != nil {
		t.Errorf("outputCSV() failed to parse duration: %v", err)
		return
	}
	if duration < 1.9 || duration > 2.1 {
		t.Errorf("outputCSV() duration = %.2f, want approximately 2.0", duration)
	}
}

func TestDuplicateTaskNames(t *testing.T) {
	app, cleanup := setupTestAppWithInMemoryRepo(t)
	defer cleanup()

	// Create two tasks with the same name
	taskName := "Duplicate Task"
	err := app.createNewTask(taskName)
	if err != nil {
		t.Fatalf("Failed to create first task: %v", err)
	}
	time.Sleep(10 * time.Millisecond) // Ensure different start times
	err = app.createNewTask(taskName)
	if err != nil {
		t.Fatalf("Failed to create second task: %v", err)
	}

	entries, err := app.repo.ListTimeEntries()
	if err != nil {
		t.Fatalf("Failed to list time entries: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("Expected 2 time entries, got %d", len(entries))
	}

	task1, err := app.repo.GetTask(entries[0].TaskID)
	if err != nil {
		t.Fatalf("Failed to get task 1: %v", err)
	}
	task2, err := app.repo.GetTask(entries[1].TaskID)
	if err != nil {
		t.Fatalf("Failed to get task 2: %v", err)
	}
	if task1.TaskName != taskName || task2.TaskName != taskName {
		t.Fatalf("Expected both tasks to have name %q, got %q and %q", taskName, task1.TaskName, task2.TaskName)
	}
	if task1.ID == task2.ID {
		t.Fatalf("Expected different task IDs for duplicate names, got %d", task1.ID)
	}
}

func TestMultipleRunningTasksAndStop(t *testing.T) {
	app, cleanup := setupTestAppWithInMemoryRepo(t)
	defer cleanup()

	// Create two running tasks (simulate by direct repo usage)
	task1 := &sqlite.Task{TaskName: "Task 1"}
	task2 := &sqlite.Task{TaskName: "Task 2"}
	app.repo.CreateTask(task1)
	app.repo.CreateTask(task2)
	entry1 := &sqlite.TimeEntry{StartTime: time.Now().Add(-2 * time.Hour), TaskID: task1.ID}
	entry2 := &sqlite.TimeEntry{StartTime: time.Now().Add(-1 * time.Hour), TaskID: task2.ID}
	app.repo.CreateTimeEntry(entry1)
	app.repo.CreateTimeEntry(entry2)

	// Stop all running tasks
	err := app.stopRunningTasks()
	if err != nil {
		t.Fatalf("Failed to stop running tasks: %v", err)
	}

	entries, err := app.repo.ListTimeEntries()
	if err != nil {
		t.Fatalf("Failed to list time entries: %v", err)
	}
	for _, entry := range entries {
		if entry.EndTime == nil {
			t.Fatalf("Expected all tasks to be stopped, but found running task with ID %d", entry.ID)
		}
	}
}

func TestSearchByPartialTaskName(t *testing.T) {
	app, cleanup := setupTestAppWithInMemoryRepo(t)
	defer cleanup()

	task1 := &sqlite.Task{TaskName: "Alpha Project"}
	task2 := &sqlite.Task{TaskName: "Beta Project"}
	task3 := &sqlite.Task{TaskName: "Alpha Test"}
	app.repo.CreateTask(task1)
	app.repo.CreateTask(task2)
	app.repo.CreateTask(task3)
	app.repo.CreateTimeEntry(&sqlite.TimeEntry{StartTime: time.Now(), TaskID: task1.ID})
	app.repo.CreateTimeEntry(&sqlite.TimeEntry{StartTime: time.Now(), TaskID: task2.ID})
	app.repo.CreateTimeEntry(&sqlite.TimeEntry{StartTime: time.Now(), TaskID: task3.ID})

	// Search for "Alpha"
	alpha := "Alpha"
	opts := sqlite.SearchOptions{TaskName: &alpha}
	results, err := app.repo.SearchTimeEntries(opts)
	if err != nil {
		t.Fatalf("Failed to search time entries: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("Expected 2 results for partial search 'Alpha', got %d", len(results))
	}

	// Search for "Project"
	project := "Project"
	opts = sqlite.SearchOptions{TaskName: &project}
	results, err = app.repo.SearchTimeEntries(opts)
	if err != nil {
		t.Fatalf("Failed to search time entries: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("Expected 2 results for partial search 'Project', got %d", len(results))
	}
}

// Helper to run resume with injected stdin
func runResumeWithInput(app *App, args []string, input string) (output string, err error) {
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	inR, inW, _ := os.Pipe()
	os.Stdin = inR
	inW.Write([]byte(input + "\n"))
	inW.Close()

	err = app.resumeTask(args)

	w.Close()
	os.Stdout = oldStdout
	os.Stdin = oldStdin

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output = buf.String()
	return
}

func TestResumeFeature_Acceptance(t *testing.T) {
	app, cleanup := setupTestAppWithInMemoryRepo(t)
	defer cleanup()

	// Use a fixed time for determinism
	fixedTime := time.Date(2025, 6, 16, 12, 0, 0, 0, time.UTC)
	timeNow = func() time.Time { return fixedTime }

	// Create tasks and entries for today and previous days
	task1 := &sqlite.Task{TaskName: "Alpha"}
	task2 := &sqlite.Task{TaskName: "Beta"}
	task3 := &sqlite.Task{TaskName: "Gamma"}
	app.repo.CreateTask(task1)
	app.repo.CreateTask(task2)
	app.repo.CreateTask(task3)
	// Today
	app.repo.CreateTimeEntry(&sqlite.TimeEntry{StartTime: fixedTime.Add(-2 * time.Hour), EndTime: &fixedTime, TaskID: task1.ID})
	app.repo.CreateTimeEntry(&sqlite.TimeEntry{StartTime: fixedTime.Add(-1 * time.Hour), TaskID: task2.ID})
	// Previous day
	app.repo.CreateTimeEntry(&sqlite.TimeEntry{StartTime: fixedTime.Add(-26 * time.Hour), EndTime: &fixedTime, TaskID: task3.ID})

	// 1. Resume with default (today), select task 1 (Beta)
	output, err := runResumeWithInput(app, []string{}, "1")
	if err != nil {
		t.Fatalf("resumeTask failed: %v", err)
	}
	if !strings.Contains(output, "Select a task to resume:") || !strings.Contains(output, "Beta") || !strings.Contains(output, "Resumed task: Beta") {
		t.Errorf("unexpected output: %s", output)
	}
	// Check that a new time entry for Beta was created and any running task is stopped
	entries, _ := app.repo.ListTimeEntries()
	var found bool
	for _, e := range entries {
		if e.TaskID == task2.ID && e.StartTime.Equal(fixedTime) {
			found = true
		}
		if e.EndTime == nil && e.TaskID != task2.ID {
			t.Errorf("unexpected running task: %v", e)
		}
	}
	if !found {
		t.Errorf("expected new time entry for Beta at %v", fixedTime)
	}

	// 2. Resume with custom duration (3h), select task 2 (Alpha)
	output, err = runResumeWithInput(app, []string{"3h"}, "2")
	if err != nil {
		t.Fatalf("resumeTask failed: %v", err)
	}
	if !strings.Contains(output, "Alpha") || !strings.Contains(output, "Resumed task: Alpha") {
		t.Errorf("unexpected output: %s", output)
	}

	// 3. Resume and quit with 'q'
	output, err = runResumeWithInput(app, []string{}, "q")
	if err != nil {
		t.Fatalf("resumeTask failed: %v", err)
	}
	if !strings.Contains(output, "Resume cancelled.") {
		t.Errorf("expected cancel message, got: %s", output)
	}
}

// TestAppWithDependencyInjection demonstrates using dependency injection with a mock repository
func TestAppWithDependencyInjection(t *testing.T) {
	app, cleanup := setupTestAppWithInMemoryRepo(t)
	defer cleanup()

	// Verify the app was created
	if app == nil {
		t.Fatal("Expected app to be created, got nil")
	}

	// Verify the repository was injected
	if app.repo == nil {
		t.Fatal("Expected repository to be injected, got nil")
	}

	// Test that we can use the app with the in-memory repository
	taskName := "Test Task with DI"
	err := app.createNewTask(taskName)
	if err != nil {
		t.Fatalf("Expected no error creating task, got: %v", err)
	}

	// Verify the task was created in the repository
	tasks, err := app.repo.ListTasks()
	if err != nil {
		t.Fatalf("Expected no error listing tasks, got: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}

	if tasks[0].TaskName != taskName {
		t.Fatalf("Expected task name '%s', got '%s'", taskName, tasks[0].TaskName)
	}
}

// TestAppWithMockRepository tests that the app works with a mock repository
func TestAppWithMockRepository(t *testing.T) {
	app, cleanup := setupTestAppWithInMemoryRepo(t)
	defer cleanup()

	// Test creating a task
	taskName := "Test Task"
	err := app.createNewTask(taskName)
	if err != nil {
		t.Fatalf("Expected no error creating task, got: %v", err)
	}

	// Verify the task was created in the repository
	tasks, err := app.repo.ListTasks()
	if err != nil {
		t.Fatalf("Expected no error listing tasks, got: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}

	if tasks[0].TaskName != taskName {
		t.Fatalf("Expected task name '%s', got '%s'", taskName, tasks[0].TaskName)
	}
}

// TestAppWithMockRepositoryListTasks tests listing tasks with mock repository
func TestAppWithMockRepositoryListTasks(t *testing.T) {
	app, cleanup := setupTestAppWithInMemoryRepo(t)
	defer cleanup()

	// Pre-populate with some test data
	testTask := &sqlite.Task{
		TaskName: "Test Task",
	}
	app.repo.CreateTask(testTask)

	// Test listing tasks
	err := app.listTasks([]string{})
	if err != nil {
		t.Fatalf("Expected no error listing tasks, got: %v", err)
	}
}

// TestAppWithMockRepositoryHelper demonstrates using the setupTestAppWithMock helper
func TestAppWithMockRepositoryHelper(t *testing.T) {
	app, cleanup := setupTestAppWithInMemoryRepo(t)
	defer cleanup()

	// Test creating multiple tasks
	taskNames := []string{"Task 1", "Task 2", "Task 3"}
	for _, taskName := range taskNames {
		err := app.createNewTask(taskName)
		if err != nil {
			t.Fatalf("Expected no error creating task '%s', got: %v", taskName, err)
		}
	}

	// Verify all tasks were created in the repository
	tasks, err := app.repo.ListTasks()
	if err != nil {
		t.Fatalf("Expected no error listing tasks, got: %v", err)
	}

	if len(tasks) != len(taskNames) {
		t.Fatalf("Expected %d tasks, got %d", len(taskNames), len(tasks))
	}

	// Verify task names match
	for i, task := range tasks {
		if task.TaskName != taskNames[i] {
			t.Fatalf("Expected task name '%s', got '%s'", taskNames[i], task.TaskName)
		}
	}
}

func dynamicUnderline(taskName string) string {
	return strings.Repeat("=", len(taskName)+12)
}

func TestSummaryFeature(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		input   string
		setup   func(app *App)
		want    func() string
		wantErr bool
	}{
		{
			name:    "summary with no tasks",
			args:    []string{"summary"},
			setup:   func(app *App) {},
			want:    func() string { return "No tasks found matching the criteria.\n" },
			wantErr: false,
		},
		{
			name: "summary with single task - direct display",
			args: []string{"summary", "coding"},
			setup: func(app *App) {
				repo := app.repo
				task := &sqlite.Task{TaskName: "coding project"}
				repo.CreateTask(task)
				start1 := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
				end1 := time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC)
				entry1 := &sqlite.TimeEntry{TaskID: task.ID, StartTime: start1, EndTime: &end1}
				repo.CreateTimeEntry(entry1)
				start2 := time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC)
				end2 := time.Date(2024, 1, 1, 16, 0, 0, 0, time.UTC)
				entry2 := &sqlite.TimeEntry{TaskID: task.ID, StartTime: start2, EndTime: &end2}
				repo.CreateTimeEntry(entry2)
			},
			want: func() string {
				taskName := "coding project"
				return "\nSummary for: " + taskName + "\n" + dynamicUnderline(taskName) + "\n" +
					"Start Time           End Time             Duration        Status\n" +
					"---------------------------------------------------------------------------\n" +
					"2024-01-01 09:00:00  2024-01-01 11:00:00  2h 0m           Completed\n" +
					"2024-01-01 14:00:00  2024-01-01 16:00:00  2h 0m           Completed\n" +
					"---------------------------------------------------------------------------\n" +
					"Total Sessions: 2\n" +
					"Time Range: 2024-01-01 09:00:00 to 2024-01-01 16:00:00\n" +
					"Total Time: 4h 0m\n"
			},
			wantErr: false,
		},
		{
			name:  "summary with multiple tasks - user selection",
			args:  []string{"summary", "project"},
			input: "1\n",
			setup: func(app *App) {
				repo := app.repo
				task1 := &sqlite.Task{TaskName: "project A"}
				task2 := &sqlite.Task{TaskName: "project B"}
				repo.CreateTask(task1)
				repo.CreateTask(task2)
				start1 := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
				end1 := time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC)
				entry1 := &sqlite.TimeEntry{TaskID: task1.ID, StartTime: start1, EndTime: &end1}
				repo.CreateTimeEntry(entry1)
				start2 := time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC)
				end2 := time.Date(2024, 1, 1, 16, 0, 0, 0, time.UTC)
				entry2 := &sqlite.TimeEntry{TaskID: task2.ID, StartTime: start2, EndTime: &end2}
				repo.CreateTimeEntry(entry2)
			},
			want: func() string {
				prompt := "Select a task to summarize:\n1. project A\n2. project B\nEnter number to summarize, or 'q' to quit: "
				taskName := "project A"
				return prompt + "\nSummary for: " + taskName + "\n" + dynamicUnderline(taskName) + "\n" +
					"Start Time           End Time             Duration        Status\n" +
					"---------------------------------------------------------------------------\n" +
					"2024-01-01 09:00:00  2024-01-01 11:00:00  2h 0m           Completed\n" +
					"---------------------------------------------------------------------------\n" +
					"Total Sessions: 1\n" +
					"Time Range: 2024-01-01 09:00:00 to 2024-01-01 11:00:00\n" +
					"Total Time: 2h 0m\n"
			},
			wantErr: false,
		},
		{
			name: "summary with time filter",
			args: []string{"summary", "2h", "coding"},
			setup: func(app *App) {
				timeNow = func() time.Time { return time.Now() }
				repo := app.repo
				task := &sqlite.Task{TaskName: "coding project"}
				repo.CreateTask(task)
				start := timeNow().Add(-1 * time.Hour)
				end := timeNow().Add(-30 * time.Minute)
				entry := &sqlite.TimeEntry{TaskID: task.ID, StartTime: start, EndTime: &end}
				repo.CreateTimeEntry(entry)
			},
			want: func() string {
				return "Summary for: coding project\n======================\nStart Time           End Time             Duration        Status\n---------------------------------------------------------------------------\n"
			},
			wantErr: false,
		},
		{
			name: "summary with running session",
			args: []string{"summary", "running"},
			setup: func(app *App) {
				// Create a task
				repo := app.repo
				task := &sqlite.Task{TaskName: "running task"}
				repo.CreateTask(task)
				// Create a running session (no end time)
				start := time.Now().Add(-1 * time.Hour)
				entry := &sqlite.TimeEntry{TaskID: task.ID, StartTime: start, EndTime: nil}
				repo.CreateTimeEntry(entry)
			},
			want: func() string {
				return "Summary for: running task\n=====================\nStart Time           End Time             Duration        Status\n---------------------------------------------------------------------------\n"
			},
			wantErr: false,
		},
		{
			name:  "summary user cancels selection",
			args:  []string{"summary", "project"},
			input: "q\n",
			setup: func(app *App) {
				// Create multiple tasks
				repo := app.repo
				task1 := &sqlite.Task{TaskName: "project A"}
				task2 := &sqlite.Task{TaskName: "project B"}
				repo.CreateTask(task1)
				repo.CreateTask(task2)
				// Add time entries for both tasks so they appear in the selection
				now := time.Now()
				entry1 := &sqlite.TimeEntry{TaskID: task1.ID, StartTime: now.Add(-2 * time.Hour), EndTime: &now}
				entry2 := &sqlite.TimeEntry{TaskID: task2.ID, StartTime: now.Add(-1 * time.Hour), EndTime: &now}
				repo.CreateTimeEntry(entry1)
				repo.CreateTimeEntry(entry2)
			},
			want: func() string {
				return "Select a task to summarize:\n1. project A\n2. project B\nEnter number to summarize, or 'q' to quit: Summary cancelled.\n"
			},
			wantErr: false,
		},
		{
			name:  "summary with time filter shows all entries for matching task",
			args:  []string{"summary", "2h", "coding"},
			input: "1\n",
			setup: func(app *App) {
				// Create a task with entries both inside and outside the 2h window
				repo := app.repo
				task := &sqlite.Task{TaskName: "coding project"}
				repo.CreateTask(task)

				// Use fixed time for deterministic test
				baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
				timeNow = func() time.Time { return baseTime }

				// Entry within the 2h window (1 hour ago)
				start1 := baseTime.Add(-1 * time.Hour)
				end1 := baseTime.Add(-30 * time.Minute)
				entry1 := &sqlite.TimeEntry{TaskID: task.ID, StartTime: start1, EndTime: &end1}
				repo.CreateTimeEntry(entry1)

				// Entry outside the 2h window (3 hours ago)
				start2 := baseTime.Add(-3 * time.Hour)
				end2 := baseTime.Add(-2 * time.Hour)
				entry2 := &sqlite.TimeEntry{TaskID: task.ID, StartTime: start2, EndTime: &end2}
				repo.CreateTimeEntry(entry2)

				// Entry outside the 2h window (5 hours ago)
				start3 := baseTime.Add(-5 * time.Hour)
				end3 := baseTime.Add(-4 * time.Hour)
				entry3 := &sqlite.TimeEntry{TaskID: task.ID, StartTime: start3, EndTime: &end3}
				repo.CreateTimeEntry(entry3)
			},
			want: func() string {
				taskName := "coding project"
				prompt := "Select a task to summarize:\n1. coding project\nEnter number to summarize, or 'q' to quit: "
				return prompt + "\nSummary for: " + taskName + "\n" + dynamicUnderline(taskName) + "\n" +
					"Start Time           End Time             Duration        Status\n" +
					"---------------------------------------------------------------------------\n" +
					"2024-01-01 07:00:00  2024-01-01 08:00:00  1h 0m           Completed\n" +
					"2024-01-01 09:00:00  2024-01-01 10:00:00  1h 0m           Completed\n" +
					"2024-01-01 11:00:00  2024-01-01 11:30:00  0h 30m           Completed\n" +
					"---------------------------------------------------------------------------\n" +
					"Total Sessions: 3\n" +
					"Time Range: 2024-01-01 07:00:00 to 2024-01-01 11:30:00\n" +
					"Total Time: 2h 30m\n"
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, cleanup := setupTestAppWithInMemoryRepo(t)
			defer cleanup()
			if tt.setup != nil {
				tt.setup(app)
			}
			oldStdout := os.Stdout
			oldStdin := os.Stdin
			r, w, _ := os.Pipe()
			os.Stdout = w
			if tt.input != "" {
				stdinR, stdinW, _ := os.Pipe()
				os.Stdin = stdinR
				go func() {
					stdinW.WriteString(tt.input)
					stdinW.Close()
				}()
			}
			err := app.Run(tt.args)
			w.Close()
			os.Stdout = oldStdout
			if tt.input != "" {
				os.Stdin = oldStdin
			}
			var buf bytes.Buffer
			buf.ReadFrom(r)
			got := buf.String()
			if (err != nil) != tt.wantErr {
				t.Errorf("summaryTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if strings.Contains(tt.name, "running") || strings.Contains(tt.name, "time filter") {
				if !strings.Contains(got, "Summary for:") {
					t.Errorf("summaryTask() output doesn't contain expected header, got: %v", got)
				}
			} else {
				want := tt.want()
				if !tt.wantErr && got != want {
					t.Errorf("summaryTask() output = %v, want %v", got, want)
				}
			}
		})
	}
}

func TestShowTaskSummary(t *testing.T) {
	tests := []struct {
		name    string
		taskID  int64
		setup   func(app *App)
		want    func() string
		wantErr bool
	}{
		{
			name:    "task not found",
			taskID:  999,
			setup:   func(app *App) {},
			want:    func() string { return "" },
			wantErr: true,
		},
		{
			name:   "task with no time entries",
			taskID: 1,
			setup: func(app *App) {
				task := &sqlite.Task{ID: 1, TaskName: "test task"}
				app.repo.CreateTask(task)
			},
			want:    func() string { return "" },
			wantErr: true,
		},
		{
			name:   "task with completed sessions",
			taskID: 1,
			setup: func(app *App) {
				task := &sqlite.Task{ID: 1, TaskName: "test task"}
				app.repo.CreateTask(task)
				start1 := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
				end1 := time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC)
				entry1 := &sqlite.TimeEntry{
					TaskID:    task.ID,
					StartTime: start1,
					EndTime:   &end1,
				}
				app.repo.CreateTimeEntry(entry1)
			},
			want: func() string {
				taskName := "test task"
				return "\nSummary for: " + taskName + "\n" + dynamicUnderline(taskName) + "\n" +
					"Start Time           End Time             Duration        Status\n" +
					"---------------------------------------------------------------------------\n" +
					"2024-01-01 09:00:00  2024-01-01 11:00:00  2h 0m           Completed\n" +
					"---------------------------------------------------------------------------\n" +
					"Total Sessions: 1\n" +
					"Time Range: 2024-01-01 09:00:00 to 2024-01-01 11:00:00\n" +
					"Total Time: 2h 0m\n"
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, cleanup := setupTestAppWithInMemoryRepo(t)
			defer cleanup()
			if tt.setup != nil {
				tt.setup(app)
			}
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			err := app.showTaskSummary(tt.taskID)
			w.Close()
			os.Stdout = oldStdout
			var buf bytes.Buffer
			buf.ReadFrom(r)
			got := buf.String()
			want := tt.want()
			if (err != nil) != tt.wantErr {
				t.Errorf("showTaskSummary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != want {
				t.Errorf("showTaskSummary() output = %v, want %v", got, want)
			}
		})
	}
}
