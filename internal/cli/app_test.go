package cli

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"time-tracker/internal/repository/sqlite"
)

// mockAPI implements the api.API interface for testing
type mockAPI struct {
	tasks       map[int64]*sqlite.Task
	timeEntries map[int64]*sqlite.TimeEntry
	nextTaskID  int64
	nextEntryID int64
}

func newMockAPI() *mockAPI {
	return &mockAPI{
		tasks:       make(map[int64]*sqlite.Task),
		timeEntries: make(map[int64]*sqlite.TimeEntry),
		nextTaskID:  1,
		nextEntryID: 1,
	}
}

func (m *mockAPI) CreateTask(name string) (*sqlite.Task, error) {
	task := &sqlite.Task{
		ID:       m.nextTaskID,
		TaskName: name,
	}
	m.tasks[task.ID] = task
	m.nextTaskID++
	return task, nil
}

func (m *mockAPI) GetTask(id int64) (*sqlite.Task, error) {
	task, exists := m.tasks[id]
	if !exists {
		return nil, fmt.Errorf("task not found: %d", id)
	}
	return task, nil
}

func (m *mockAPI) ListTasks() ([]*sqlite.Task, error) {
	tasks := make([]*sqlite.Task, 0, len(m.tasks))
	for _, task := range m.tasks {
		tasks = append(tasks, task)
	}
	// Sort by ID for consistent ordering
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ID < tasks[j].ID
	})
	return tasks, nil
}

func (m *mockAPI) UpdateTask(id int64, name string) error {
	task, exists := m.tasks[id]
	if !exists {
		return fmt.Errorf("task not found: %d", id)
	}
	task.TaskName = name
	return nil
}

func (m *mockAPI) DeleteTask(id int64) error {
	if _, exists := m.tasks[id]; !exists {
		return fmt.Errorf("task not found: %d", id)
	}
	delete(m.tasks, id)
	return nil
}

func (m *mockAPI) CreateTimeEntry(taskID int64, startTime time.Time, endTime *time.Time) (*sqlite.TimeEntry, error) {
	entry := &sqlite.TimeEntry{
		ID:        m.nextEntryID,
		TaskID:    taskID,
		StartTime: startTime,
		EndTime:   endTime,
	}
	m.timeEntries[entry.ID] = entry
	m.nextEntryID++
	return entry, nil
}

func (m *mockAPI) GetTimeEntry(id int64) (*sqlite.TimeEntry, error) {
	entry, exists := m.timeEntries[id]
	if !exists {
		return nil, fmt.Errorf("time entry not found: %d", id)
	}
	return entry, nil
}

func (m *mockAPI) ListTimeEntries() ([]*sqlite.TimeEntry, error) {
	entries := make([]*sqlite.TimeEntry, 0, len(m.timeEntries))
	for _, entry := range m.timeEntries {
		entries = append(entries, entry)
	}
	return entries, nil
}

func (m *mockAPI) SearchTimeEntries(opts sqlite.SearchOptions) ([]*sqlite.TimeEntry, error) {
	var entries []*sqlite.TimeEntry

	for _, entry := range m.timeEntries {
		// Filter by time range
		if opts.StartTime != nil && entry.StartTime.Before(*opts.StartTime) {
			continue
		}
		if opts.EndTime != nil && entry.StartTime.After(*opts.EndTime) {
			continue
		}

		// Filter by task ID
		if opts.TaskID != nil && entry.TaskID != *opts.TaskID {
			continue
		}

		// Filter by task name
		if opts.TaskName != nil {
			task, exists := m.tasks[entry.TaskID]
			if !exists || !strings.Contains(strings.ToLower(task.TaskName), strings.ToLower(*opts.TaskName)) {
				continue
			}
		}

		// If no filters specified, only return running tasks
		if opts.StartTime == nil && opts.EndTime == nil && opts.TaskID == nil && opts.TaskName == nil {
			if entry.EndTime != nil {
				continue
			}
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func (m *mockAPI) UpdateTimeEntry(id int64, startTime time.Time, endTime *time.Time, taskID int64) error {
	entry, exists := m.timeEntries[id]
	if !exists {
		return fmt.Errorf("time entry not found: %d", id)
	}
	entry.StartTime = startTime
	entry.EndTime = endTime
	entry.TaskID = taskID
	return nil
}

func (m *mockAPI) DeleteTimeEntry(id int64) error {
	if _, exists := m.timeEntries[id]; !exists {
		return fmt.Errorf("time entry not found: %d", id)
	}
	delete(m.timeEntries, id)
	return nil
}

func (m *mockAPI) StartTask(taskID int64) (*sqlite.TimeEntry, error) {
	// Stop any running tasks
	now := time.Now()
	for _, entry := range m.timeEntries {
		if entry.EndTime == nil {
			entry.EndTime = &now
		}
	}

	// Create new entry
	return m.CreateTimeEntry(taskID, time.Now(), nil)
}

func (m *mockAPI) StopTask(entryID int64) error {
	entry, exists := m.timeEntries[entryID]
	if !exists {
		return fmt.Errorf("time entry not found: %d", entryID)
	}
	if entry.EndTime != nil {
		return fmt.Errorf("task already stopped")
	}
	now := time.Now()
	entry.EndTime = &now
	return nil
}

func (m *mockAPI) ResumeTask(taskID int64) (*sqlite.TimeEntry, error) {
	// Stop any running tasks
	now := time.Now()
	for _, entry := range m.timeEntries {
		if entry.EndTime == nil {
			entry.EndTime = &now
		}
	}

	// Create new entry
	return m.CreateTimeEntry(taskID, time.Now(), nil)
}

func (m *mockAPI) GetCurrentlyRunningTask() (*sqlite.TimeEntry, error) {
	for _, entry := range m.timeEntries {
		if entry.EndTime == nil {
			return entry, nil
		}
	}
	return nil, fmt.Errorf("no running task")
}

func (m *mockAPI) ListTodayTasks() ([]*sqlite.Task, error) {
	// For simplicity, return all tasks
	return m.ListTasks()
}

func setupTestAppWithMockAPI(t *testing.T) (*App, func()) {
	mockAPI := newMockAPI()
	app := NewApp(mockAPI)
	cleanup := func() {}
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
			name:    "stop with extra args (error)",
			args:    []string{"stop", "now"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "start new task",
			args:    []string{"start", "Working on feature X"},
			want:    "All running tasks have been stopped\nStarted new task: Working on feature X\n",
			wantErr: false,
		},
		{
			name:    "start multiple words task",
			args:    []string{"start", "Working", "on", "feature", "X"},
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
		{
			name:    "unknown command",
			args:    []string{"foobar"},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, cleanup := setupTestAppWithMockAPI(t)
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
	app, cleanup := setupTestAppWithMockAPI(t)
	defer cleanup()

	// Create a fixed time for testing
	fixedTime := time.Date(2025, 6, 16, 11, 22, 1, 0, time.UTC)
	timeNow = func() time.Time { return fixedTime }

	// Create some test tasks using the API
	task1, _ := app.api.CreateTask("First task")
	task2, _ := app.api.CreateTask("Second task")
	task3, _ := app.api.CreateTask("Third task")

	// Create some test entries using the API
	app.api.CreateTimeEntry(task1.ID, fixedTime.Add(-2*time.Hour), &fixedTime)
	app.api.CreateTimeEntry(task2.ID, fixedTime.Add(-1*time.Hour), nil) // Running task
	app.api.CreateTimeEntry(task3.ID, fixedTime.Add(-30*time.Minute), &fixedTime)

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
			listCmd := NewListCommand(app)
			err := listCmd.Execute(tt.args)

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
	app, cleanup := setupTestAppWithMockAPI(t)
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
				// Create a running task using the API
				task, _ := app.api.CreateTask("Test task")
				app.api.CreateTimeEntry(task.ID, time.Now(), nil)
			}

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run current command
			currentCmd := NewCurrentCommand(app)
			err := currentCmd.Execute([]string{})

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
	app, cleanup := setupTestAppWithMockAPI(t)
	defer cleanup()

	// Create test tasks using the API
	task1, _ := app.api.CreateTask("First task")
	task2, _ := app.api.CreateTask("Second task")

	// Create test entries using the API
	now := time.Now()
	app.api.CreateTimeEntry(task1.ID, now.Add(-2*time.Hour), &now)
	app.api.CreateTimeEntry(task2.ID, now.Add(-1*time.Hour), nil) // Running task

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run output command
	outputCmd := NewOutputCommand(app)
	err := outputCmd.outputCSV()

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

	// Check that duration is approximately 2 hours (only if end time is set)
	if firstEntry[2] != "" { // Only check duration if end time is set
		duration, err := strconv.ParseFloat(firstEntry[3], 64)
		if err != nil {
			t.Errorf("outputCSV() failed to parse duration: %v", err)
			return
		}
		if duration < 1.9 || duration > 2.1 {
			t.Errorf("outputCSV() duration = %.2f, want approximately 2.0", duration)
		}
	}
}

func TestDuplicateTaskNames(t *testing.T) {
	app, cleanup := setupTestAppWithMockAPI(t)
	defer cleanup()

	// Create two tasks with the same name
	taskName := "Duplicate Task"
	startCmd := NewStartCommand(app)
	err := startCmd.Execute([]string{taskName})
	if err != nil {
		t.Fatalf("Failed to create first task: %v", err)
	}
	time.Sleep(10 * time.Millisecond) // Ensure different start times
	startCmd = NewStartCommand(app)
	err = startCmd.Execute([]string{taskName})
	if err != nil {
		t.Fatalf("Failed to create second task: %v", err)
	}

	entries, err := app.api.ListTimeEntries()
	if err != nil {
		t.Fatalf("Failed to list time entries: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("Expected 2 time entries, got %d", len(entries))
	}

	task1, err := app.api.GetTask(entries[0].TaskID)
	if err != nil {
		t.Fatalf("Failed to get task 1: %v", err)
	}
	task2, err := app.api.GetTask(entries[1].TaskID)
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
	app, cleanup := setupTestAppWithMockAPI(t)
	defer cleanup()

	// Create two running tasks using the API
	task1, _ := app.api.CreateTask("Task 1")
	task2, _ := app.api.CreateTask("Task 2")
	app.api.CreateTimeEntry(task1.ID, time.Now().Add(-2*time.Hour), nil)
	app.api.CreateTimeEntry(task2.ID, time.Now().Add(-1*time.Hour), nil)

	// Stop all running tasks
	stopCmd := NewStopCommand(app)
	err := stopCmd.Execute([]string{})
	if err != nil {
		t.Fatalf("Failed to stop running tasks: %v", err)
	}

	entries, err := app.api.ListTimeEntries()
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
	app, cleanup := setupTestAppWithMockAPI(t)
	defer cleanup()

	task1, _ := app.api.CreateTask("Alpha Project")
	task2, _ := app.api.CreateTask("Beta Project")
	task3, _ := app.api.CreateTask("Alpha Test")
	app.api.CreateTimeEntry(task1.ID, time.Now(), nil)
	app.api.CreateTimeEntry(task2.ID, time.Now(), nil)
	app.api.CreateTimeEntry(task3.ID, time.Now(), nil)

	// Search for "Alpha"
	alpha := "Alpha"
	opts := sqlite.SearchOptions{TaskName: &alpha}
	results, err := app.api.SearchTimeEntries(opts)
	if err != nil {
		t.Fatalf("Failed to search time entries: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("Expected 2 results for partial search 'Alpha', got %d", len(results))
	}

	// Search for "Project"
	project := "Project"
	opts = sqlite.SearchOptions{TaskName: &project}
	results, err = app.api.SearchTimeEntries(opts)
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

	resumeCmd := NewResumeCommand(app)
	err = resumeCmd.Execute(args)

	w.Close()
	os.Stdout = oldStdout
	os.Stdin = oldStdin

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output = buf.String()
	return
}

func TestResumeFeature_Acceptance(t *testing.T) {
	app, cleanup := setupTestAppWithMockAPI(t)
	defer cleanup()

	// Use a fixed time for determinism
	fixedTime := time.Date(2025, 6, 16, 12, 0, 0, 0, time.UTC)
	timeNow = func() time.Time { return fixedTime }

	// Create tasks and entries for today and previous days using the API
	task1, _ := app.api.CreateTask("Alpha")
	task2, _ := app.api.CreateTask("Beta")
	task3, _ := app.api.CreateTask("Gamma")
	// Today
	app.api.CreateTimeEntry(task1.ID, fixedTime.Add(-2*time.Hour), &fixedTime)
	app.api.CreateTimeEntry(task2.ID, fixedTime.Add(-1*time.Hour), nil)
	// Previous day
	app.api.CreateTimeEntry(task3.ID, fixedTime.Add(-26*time.Hour), &fixedTime)

	// 1. Resume with default (today), select task 1 (Beta)
	output, err := runResumeWithInput(app, []string{}, "1")
	if err != nil {
		t.Fatalf("resumeCmd.Execute failed: %v", err)
	}
	if !strings.Contains(output, "Select a task to resume:") || !strings.Contains(output, "Beta") || !strings.Contains(output, "Resumed task: Beta") {
		t.Errorf("unexpected output: %s", output)
	}
	// Check that a new time entry for Beta was created and any running task is stopped
	entries, _ := app.api.ListTimeEntries()
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
		t.Fatalf("resumeCmd.Execute failed: %v", err)
	}
	if !strings.Contains(output, "Alpha") || !strings.Contains(output, "Resumed task: Alpha") {
		t.Errorf("unexpected output: %s", output)
	}

	// 3. Resume and quit with 'q'
	output, err = runResumeWithInput(app, []string{}, "q")
	if err != nil {
		t.Fatalf("resumeCmd.Execute failed: %v", err)
	}
	if !strings.Contains(output, "Resume cancelled.") {
		t.Errorf("expected cancel message, got: %s", output)
	}
}

// TestAppWithDependencyInjection demonstrates using dependency injection with a mock repository
func TestAppWithDependencyInjection(t *testing.T) {
	app, cleanup := setupTestAppWithMockAPI(t)
	defer cleanup()

	// Verify the app was created
	if app == nil {
		t.Fatal("Expected app to be created, got nil")
	}

	// Verify the API was injected
	if app.api == nil {
		t.Fatal("Expected API to be injected, got nil")
	}

	// Test that we can use the app with the mock API
	taskName := "Test Task with DI"
	startCmd := NewStartCommand(app)
	err := startCmd.Execute([]string{taskName})
	if err != nil {
		t.Fatalf("Expected no error creating task, got: %v", err)
	}

	// Verify the task was created in the API
	tasks, err := app.api.ListTasks()
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
	app, cleanup := setupTestAppWithMockAPI(t)
	defer cleanup()

	// Test creating a task
	taskName := "Test Task"
	startCmd := NewStartCommand(app)
	err := startCmd.Execute([]string{taskName})
	if err != nil {
		t.Fatalf("Expected no error creating task, got: %v", err)
	}

	// Verify the task was created in the API
	tasks, err := app.api.ListTasks()
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
	app, cleanup := setupTestAppWithMockAPI(t)
	defer cleanup()

	// Pre-populate with some test data
	_, _ = app.api.CreateTask("Test Task")

	// Test listing tasks
	listCmd := NewListCommand(app)
	err := listCmd.Execute([]string{})
	if err != nil {
		t.Fatalf("Expected no error listing tasks, got: %v", err)
	}
}

// TestAppWithMockRepositoryHelper demonstrates using the setupTestAppWithMock helper
func TestAppWithMockRepositoryHelper(t *testing.T) {
	app, cleanup := setupTestAppWithMockAPI(t)
	defer cleanup()

	// Test creating multiple tasks
	taskNames := []string{"Task 1", "Task 2", "Task 3"}
	for _, taskName := range taskNames {
		startCmd := NewStartCommand(app)
	err := startCmd.Execute([]string{taskName})
		if err != nil {
			t.Fatalf("Expected no error creating task '%s', got: %v", taskName, err)
		}
	}

	// Verify all tasks were created in the API
	tasks, err := app.api.ListTasks()
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
