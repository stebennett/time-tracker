package api

import (
	"context"
	"testing"
	"time"
	"time-tracker/internal/repository/sqlite"
)

func setupTestAPI(t *testing.T) (API, func()) {
	repo, err := sqlite.New(":memory:")
	if err != nil {
		t.Fatalf("failed to create in-memory repo: %v", err)
	}
	api := New(repo)
	cleanup := func() { repo.Close() }
	return api, cleanup
}

func TestAPI_CRUD_Task(t *testing.T) {
	api, cleanup := setupTestAPI(t)
	defer cleanup()

	// Create
	task, err := api.CreateTask(context.Background(), "Test Task")
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}
	if task.ID == 0 || task.TaskName != "Test Task" {
		t.Errorf("unexpected task: %+v", task)
	}

	// Get
	got, err := api.GetTask(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if got.ID != task.ID || got.TaskName != task.TaskName {
		t.Errorf("GetTask returned wrong task: %+v", got)
	}

	// List
	tasks, err := api.ListTasks(context.Background())
	if err != nil {
		t.Fatalf("ListTasks failed: %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(tasks))
	}

	// Update
	err = api.UpdateTask(context.Background(), task.ID, "Updated Task")
	if err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}
	updated, _ := api.GetTask(context.Background(), task.ID)
	if updated.TaskName != "Updated Task" {
		t.Errorf("UpdateTask did not update name: %+v", updated)
	}

	// Delete
	err = api.DeleteTask(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("DeleteTask failed: %v", err)
	}
	_, err = api.GetTask(context.Background(), task.ID)
	if err == nil {
		t.Errorf("expected error after DeleteTask, got nil")
	}
}

func TestAPI_CRUD_TimeEntry(t *testing.T) {
	api, cleanup := setupTestAPI(t)
	defer cleanup()

	task, _ := api.CreateTask(context.Background(), "Entry Task")
	start := time.Now().Add(-1 * time.Hour)
	end := time.Now()

	// Create
	entry, err := api.CreateTimeEntry(context.Background(), task.ID, start, &end)
	if err != nil {
		t.Fatalf("CreateTimeEntry failed: %v", err)
	}
	if entry.ID == 0 || entry.TaskID != task.ID {
		t.Errorf("unexpected entry: %+v", entry)
	}

	// Get
	got, err := api.GetTimeEntry(context.Background(), entry.ID)
	if err != nil {
		t.Fatalf("GetTimeEntry failed: %v", err)
	}
	if got.ID != entry.ID || got.TaskID != entry.TaskID {
		t.Errorf("GetTimeEntry returned wrong entry: %+v", got)
	}

	// List
	entries, err := api.ListTimeEntries(context.Background())
	if err != nil {
		t.Fatalf("ListTimeEntries failed: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}

	// Update
	newStart := start.Add(-30 * time.Minute)
	newEnd := end.Add(30 * time.Minute)
	err = api.UpdateTimeEntry(context.Background(), entry.ID, newStart, &newEnd, task.ID)
	if err != nil {
		t.Fatalf("UpdateTimeEntry failed: %v", err)
	}
	updated, _ := api.GetTimeEntry(context.Background(), entry.ID)
	if updated.StartTime.Unix() != newStart.Unix() || updated.EndTime.Unix() != newEnd.Unix() {
		t.Errorf("UpdateTimeEntry did not update times: %+v", updated)
	}

	// Delete
	err = api.DeleteTimeEntry(context.Background(), entry.ID)
	if err != nil {
		t.Fatalf("DeleteTimeEntry failed: %v", err)
	}
	_, err = api.GetTimeEntry(context.Background(), entry.ID)
	if err == nil {
		t.Errorf("expected error after DeleteTimeEntry, got nil")
	}
}

// TODO: Add tests for business logic: StartTask, StopTask, ResumeTask, GetCurrentlyRunningTask, ListTodayTasks, etc.

func TestAPI_BusinessLogic_StartStopCurrent(t *testing.T) {
	api, cleanup := setupTestAPI(t)
	defer cleanup()

	// Start a new task
	task, _ := api.CreateTask(context.Background(), "Work on feature")
	entry, err := api.StartTask(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("StartTask failed: %v", err)
	}
	if entry.TaskID != task.ID || entry.EndTime != nil {
		t.Errorf("unexpected started entry: %+v", entry)
	}

	// Get currently running task
	running, err := api.GetCurrentlyRunningTask(context.Background())
	if err != nil {
		t.Fatalf("GetCurrentlyRunningTask failed: %v", err)
	}
	if running.TaskID != task.ID || running.EndTime != nil {
		t.Errorf("unexpected running entry: %+v", running)
	}

	// Stop the running task
	err = api.StopTask(context.Background(), running.ID)
	if err != nil {
		t.Fatalf("StopTask failed: %v", err)
	}
	stopped, _ := api.GetTimeEntry(context.Background(), running.ID)
	if stopped.EndTime == nil {
		t.Errorf("expected EndTime to be set after StopTask")
	}

	// No running task now
	_, err = api.GetCurrentlyRunningTask(context.Background())
	if err == nil {
		t.Errorf("expected error when no running task")
	}
}

func TestAPI_BusinessLogic_ResumeTask(t *testing.T) {
	api, cleanup := setupTestAPI(t)
	defer cleanup()

	task, _ := api.CreateTask(context.Background(), "Resume Me")
	// Simulate a previous completed entry
	start := time.Now().Add(-2 * time.Hour)
	end := time.Now().Add(-1 * time.Hour)
	api.CreateTimeEntry(context.Background(), task.ID, start, &end)

	// Resume the task
	resumed, err := api.ResumeTask(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("ResumeTask failed: %v", err)
	}
	if resumed.TaskID != task.ID || resumed.EndTime != nil {
		t.Errorf("unexpected resumed entry: %+v", resumed)
	}
}

func TestAPI_BusinessLogic_ListTodayTasks(t *testing.T) {
	api, cleanup := setupTestAPI(t)
	defer cleanup()

	today := time.Now()
	task1, _ := api.CreateTask(context.Background(), "Today Task 1")
	task2, _ := api.CreateTask(context.Background(), "Today Task 2")
	api.CreateTimeEntry(context.Background(), task1.ID, today.Add(-2*time.Hour), nil)
	api.CreateTimeEntry(context.Background(), task2.ID, today.Add(-1*time.Hour), nil)

	tasks, err := api.ListTodayTasks(context.Background())
	if err != nil {
		t.Fatalf("ListTodayTasks failed: %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks for today, got %d", len(tasks))
	}
}
