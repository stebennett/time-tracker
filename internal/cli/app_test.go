package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"time-tracker/internal/repository/sqlite"
)

func setupTestApp(t *testing.T) (*App, func()) {
	// Create data directory if it doesn't exist
	dataDir := "./data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Set up test database path
	dbPath := filepath.Join(dataDir, "tt.db")

	// Create repository instance
	repo, err := sqlite.New(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	app := &App{repo: repo}

	// Return cleanup function
	cleanup := func() {
		repo.Close()
		os.Remove(dbPath)
	}

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, cleanup := setupTestApp(t)
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
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// Create a fixed time for testing
	fixedTime := time.Date(2025, 6, 16, 11, 22, 1, 0, time.UTC)
	timeNow = func() time.Time { return fixedTime }

	// Create some test entries
	entries := []*sqlite.TimeEntry{
		{
			StartTime:   fixedTime.Add(-2 * time.Hour),
			EndTime:     &fixedTime,
			Description: "First task",
		},
		{
			StartTime:   fixedTime.Add(-1 * time.Hour),
			Description: "Second task",
		},
		{
			StartTime:   fixedTime.Add(-30 * time.Minute),
			EndTime:     &fixedTime,
			Description: "Third task",
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
			want:    "2025-06-16 09:22:01 - 2025-06-16 11:22:01 (2h 0m): First task\n" +
				"2025-06-16 10:22:01 - running: Second task\n" +
				"2025-06-16 10:52:01 - 2025-06-16 11:22:01 (0h 30m): Third task\n",
			wantErr: false,
		},
		{
			name:    "list last hour",
			args:    []string{"1h"},
			want:    "2025-06-16 10:22:01 - running: Second task\n" +
				"2025-06-16 10:52:01 - 2025-06-16 11:22:01 (0h 30m): Third task\n",
			wantErr: false,
		},
		{
			name:    "list with text filter",
			args:    []string{"task"},
			want:    "2025-06-16 09:22:01 - 2025-06-16 11:22:01 (2h 0m): First task\n" +
				"2025-06-16 10:22:01 - running: Second task\n" +
				"2025-06-16 10:52:01 - 2025-06-16 11:22:01 (0h 30m): Third task\n",
			wantErr: false,
		},
		{
			name:    "list with time and text filter",
			args:    []string{"1h", "Second"},
			want:    "2025-06-16 10:22:01 - running: Second task\n",
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

func TestNewApp(t *testing.T) {
	app, err := NewApp()
	if err != nil {
		t.Errorf("NewApp() error = %v", err)
	}
	if app == nil {
		t.Error("NewApp() returned nil")
	}
	if app.repo == nil {
		t.Error("NewApp() repository is nil")
	}
}

func TestShowCurrentTask(t *testing.T) {
	app, cleanup := setupTestApp(t)
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
				entry := &sqlite.TimeEntry{
					StartTime:   time.Now(),
					Description: "Test task",
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
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// Create test entries
	now := time.Now()
	entries := []*sqlite.TimeEntry{
		{
			StartTime:   now.Add(-2 * time.Hour),
			EndTime:     &now,
			Description: "First task",
		},
		{
			StartTime:   now.Add(-1 * time.Hour),
			Description: "Second task",
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
	expectedHeader := []string{"ID", "Start Time", "End Time", "Duration (hours)", "Description"}
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