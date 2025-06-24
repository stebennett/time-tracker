package config

import (
	"os"
	"testing"

	"time-tracker/internal/repository/sqlite"
)

func TestGetEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     Environment
	}{
		{
			name:     "development environment",
			envValue: "development",
			want:     Development,
		},
		{
			name:     "testing environment",
			envValue: "testing",
			want:     Testing,
		},
		{
			name:     "production environment",
			envValue: "production",
			want:     Production,
		},
		{
			name:     "empty environment defaults to production",
			envValue: "",
			want:     Production,
		},
		{
			name:     "invalid environment defaults to production",
			envValue: "invalid",
			want:     Production,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envValue != "" {
				os.Setenv("TT_ENV", tt.envValue)
				defer os.Unsetenv("TT_ENV")
			} else {
				os.Unsetenv("TT_ENV")
			}

			got := GetEnvironment()
			if got != tt.want {
				t.Errorf("GetEnvironment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewRepositoryFactory(t *testing.T) {
	factory := NewRepositoryFactory(Development)
	if factory == nil {
		t.Fatal("NewRepositoryFactory() returned nil")
	}
	if factory.env != Development {
		t.Errorf("NewRepositoryFactory() env = %v, want %v", factory.env, Development)
	}
}

func TestRepositoryFactory_CreateRepository(t *testing.T) {
	tests := []struct {
		name    string
		env     Environment
		wantErr bool
	}{
		{
			name:    "development environment",
			env:     Development,
			wantErr: false,
		},
		{
			name:    "testing environment",
			env:     Testing,
			wantErr: false,
		},
		{
			name:    "production environment",
			env:     Production,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := NewRepositoryFactory(tt.env)
			repo, err := factory.CreateRepository()

			if (err != nil) != tt.wantErr {
				t.Errorf("RepositoryFactory.CreateRepository() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && repo == nil {
				t.Error("RepositoryFactory.CreateRepository() returned nil repository")
				return
			}

			if repo != nil {
				defer repo.Close()
			}
		})
	}
}

func TestRepositoryFactory_CreateRepository_Development(t *testing.T) {
	// Clean up any existing tt.db file before the test
	dbPath := "tt.db"
	if _, err := os.Stat(dbPath); err == nil {
		os.Remove(dbPath)
	}

	factory := NewRepositoryFactory(Development)
	repo, err := factory.CreateRepository()
	if err != nil {
		t.Fatalf("CreateRepository() error = %v", err)
	}
	defer repo.Close()

	// Test that we can use the repository
	repo.CreateTask(&sqlite.Task{TaskName: "Test Task"})

	tasks, err := repo.ListTasks()
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}
	if tasks == nil {
		t.Error("ListTasks() returned nil")
	}

	// Clean up the tt.db file after the test
	defer func() {
		if _, err := os.Stat(dbPath); err == nil {
			os.Remove(dbPath)
		}
	}()
}

func TestRepositoryFactory_CreateRepository_Testing(t *testing.T) {
	factory := NewRepositoryFactory(Testing)
	repo, err := factory.CreateRepository()
	if err != nil {
		t.Fatalf("CreateRepository() error = %v", err)
	}
	defer repo.Close()

	// Test that we can use the repository
	repo.CreateTask(&sqlite.Task{TaskName: "Test Task"})

	tasks, err := repo.ListTasks()
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}
	if tasks == nil {
		t.Error("ListTasks() returned nil")
	}
}
