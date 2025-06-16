package cli

import (
	"bytes"
	"os"
	"testing"
	"time"

	"time-tracker/internal/repository/sqlite"
)

func setupTestApp(t *testing.T) (*App, func()) {
	// Create a temporary file for the test database
	tmpfile, err := os.CreateTemp("", "testdb-*.db")
	if err != nil {
		t.Fatal(err)
	}

	// Create repository instance
	repo, err := sqlite.New(tmpfile.Name())
	if err != nil {
		os.Remove(tmpfile.Name())
		t.Fatal(err)
	}

	app := &App{repo: repo}

	// Return cleanup function
	cleanup := func() {
		repo.Close()
		os.Remove(tmpfile.Name())
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
			want:    "Started new task: stop now\n",
			wantErr: false,
		},
		{
			name:    "stop working is a new task",
			args:    []string{"stop", "working"},
			want:    "Started new task: stop working\n",
			wantErr: false,
		},
		{
			name:    "new task",
			args:    []string{"Working on feature X"},
			want:    "Started new task: Working on feature X\n",
			wantErr: false,
		},
		{
			name:    "multiple words task",
			args:    []string{"Working", "on", "feature", "X"},
			want:    "Started new task: Working on feature X\n",
			wantErr: false,
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