package config

import (
	"context"
	"os"
	"testing"

	"time-tracker/internal/repository/sqlite"
)

func TestCreateRepository(t *testing.T) {
	// Create a temporary directory for testing to avoid home directory issues
	tmpDir := t.TempDir()
	
	// Set up environment variables for testing
	originalDbDir := os.Getenv("TT_DB_DIR")
	os.Setenv("TT_DB_DIR", tmpDir)
	defer func() {
		if originalDbDir != "" {
			os.Setenv("TT_DB_DIR", originalDbDir)
		} else {
			os.Unsetenv("TT_DB_DIR")
		}
	}()

	// Load configuration
	loader := NewLoader()
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Test repository creation
	repo, err := CreateRepository(cfg)
	if err != nil {
		t.Errorf("CreateRepository() error = %v", err)
		return
	}

	if repo == nil {
		t.Error("CreateRepository() returned nil repository")
		return
	}

	defer repo.Close()

	// Test that we can use the repository
	err = repo.CreateTask(context.Background(), &sqlite.Task{TaskName: "Test Task"})
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	tasks, err := repo.ListTasks(context.Background())
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}
	if tasks == nil {
		t.Error("ListTasks() returned nil")
	}
}

func TestCreateTestRepository(t *testing.T) {
	repo, err := CreateTestRepository()
	if err != nil {
		t.Fatalf("CreateTestRepository() error = %v", err)
	}
	defer repo.Close()

	// Test that we can use the repository
	err = repo.CreateTask(context.Background(), &sqlite.Task{TaskName: "Test Task"})
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	tasks, err := repo.ListTasks(context.Background())
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}
	if tasks == nil {
		t.Error("ListTasks() returned nil")
	}
}
