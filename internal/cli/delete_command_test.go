package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteCommand_BusinessAPIIntegration(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()
	
	ctx := context.Background()
	
	// Create multiple tasks using BusinessAPI
	_, err := app.businessAPI.StartNewTask(ctx, "Task 1")
	assert.NoError(t, err)
	task1Session, err := app.businessAPI.GetCurrentSession(ctx)
	assert.NoError(t, err)
	_, err = app.businessAPI.StopAllRunningTasks(ctx)
	assert.NoError(t, err)
	
	_, err = app.businessAPI.StartNewTask(ctx, "Task 2")
	assert.NoError(t, err)
	task2Session, err := app.businessAPI.GetCurrentSession(ctx)
	assert.NoError(t, err)
	_, err = app.businessAPI.StopAllRunningTasks(ctx)
	assert.NoError(t, err)
	
	_, err = app.businessAPI.StartNewTask(ctx, "Task 3")
	assert.NoError(t, err)
	_, err = app.businessAPI.StopAllRunningTasks(ctx)
	assert.NoError(t, err)
	
	// Verify initial state - should have 3 tasks with time entries
	entries, err := app.businessAPI.SearchTimeEntries(ctx, "", "")
	assert.NoError(t, err)
	assert.Len(t, entries, 3, "Should have 3 time entries initially")
	
	tasks, err := app.businessAPI.SearchTasks(ctx, "", "", "name")
	assert.NoError(t, err)
	assert.Len(t, tasks, 3, "Should have 3 tasks initially")
	
	// Test that delete removes a specific task and its entries
	err = app.businessAPI.DeleteTaskWithEntries(ctx, task2Session.Task.ID)
	assert.NoError(t, err)
	
	// Verify task2 is gone
	_, err = app.businessAPI.GetTask(ctx, task2Session.Task.ID)
	assert.Error(t, err, "Task 2 should be deleted")
	
	// Verify only task2's entries are gone
	remainingEntries, err := app.businessAPI.SearchTimeEntries(ctx, "", "")
	assert.NoError(t, err)
	assert.Len(t, remainingEntries, 2, "Should have 2 time entries after deletion")
	
	// Verify other tasks still exist
	_, err = app.businessAPI.GetTask(ctx, task1Session.Task.ID)
	assert.NoError(t, err, "Task 1 should still exist")
	
	remainingTasks, err := app.businessAPI.SearchTasks(ctx, "", "", "name")
	assert.NoError(t, err)
	assert.Len(t, remainingTasks, 2, "Should have 2 tasks after deletion")
}

func TestDeleteCommand_ErrorHandling(t *testing.T) {
	app, cleanup := setupTestAppWithMockBusinessAPI(t)
	defer cleanup()
	
	ctx := context.Background()
	
	// Test deleting non-existent task
	err := app.businessAPI.DeleteTaskWithEntries(ctx, 999)
	// This should not error in our mock implementation
	// In a real implementation, it might error or be a no-op
	assert.NoError(t, err)
}