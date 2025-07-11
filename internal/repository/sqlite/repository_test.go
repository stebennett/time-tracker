package sqlite

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*SQLiteRepository, func()) {
	// Create data directory if it doesn't exist
	dataDir := "./data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Set up test database path
	dbPath := filepath.Join(dataDir, "tt.db")

	// Create repository instance
	repo, err := New(dbPath)
	require.NoError(t, err)

	// Return cleanup function
	cleanup := func() {
		repo.Close()
		os.Remove(dbPath)
	}

	return repo, cleanup
}

func TestCreateTimeEntry(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	task := &Task{TaskName: "Test entry"}
	err := repo.CreateTask(context.Background(), task)
	require.NoError(t, err)
	assert.Greater(t, task.ID, int64(0))

	now := time.Now()
	entry := &TimeEntry{
		StartTime: now,
		TaskID:    task.ID,
	}

	// Test creating entry
	err = repo.CreateTimeEntry(context.Background(), entry)
	require.NoError(t, err)
	assert.Greater(t, entry.ID, int64(0))

	// Verify entry was created
	retrieved, err := repo.GetTimeEntry(context.Background(), entry.ID)
	require.NoError(t, err)
	assert.Equal(t, entry.ID, retrieved.ID)
	assert.Equal(t, entry.StartTime.Unix(), retrieved.StartTime.Unix())
	assert.Equal(t, entry.TaskID, retrieved.TaskID)
	assert.Nil(t, retrieved.EndTime)
}

func TestGetTimeEntry(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	// Test getting non-existent entry
	_, err := repo.GetTimeEntry(context.Background(), 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Create and get entry
	task := &Task{TaskName: "Test entry"}
	err = repo.CreateTask(context.Background(), task)
	require.NoError(t, err)

	now := time.Now()
	entry := &TimeEntry{
		StartTime: now,
		TaskID:    task.ID,
	}
	err = repo.CreateTimeEntry(context.Background(), entry)
	require.NoError(t, err)

	retrieved, err := repo.GetTimeEntry(context.Background(), entry.ID)
	require.NoError(t, err)
	assert.Equal(t, entry.ID, retrieved.ID)
	assert.Equal(t, entry.StartTime.Unix(), retrieved.StartTime.Unix())
	assert.Equal(t, entry.TaskID, retrieved.TaskID)
}

func TestListTimeEntries(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	task := &Task{TaskName: "Test task"}
	err := repo.CreateTask(context.Background(), task)
	require.NoError(t, err)

	// Create multiple entries
	entries := []*TimeEntry{
		{StartTime: time.Now().Add(-2 * time.Hour), TaskID: task.ID},
		{StartTime: time.Now().Add(-1 * time.Hour), TaskID: task.ID},
		{StartTime: time.Now(), TaskID: task.ID},
	}

	for _, entry := range entries {
		err := repo.CreateTimeEntry(context.Background(), entry)
		require.NoError(t, err)
	}

	// Test listing entries
	retrieved, err := repo.ListTimeEntries(context.Background())
	require.NoError(t, err)
	assert.Len(t, retrieved, 3)

	// Verify order (ascending by start time)
	assert.True(t, retrieved[0].StartTime.Before(retrieved[1].StartTime))
	assert.True(t, retrieved[1].StartTime.Before(retrieved[2].StartTime))
}

func TestUpdateTimeEntry(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	task := &Task{TaskName: "Original task"}
	err := repo.CreateTask(context.Background(), task)
	require.NoError(t, err)

	// Create entry
	now := time.Now()
	entry := &TimeEntry{
		StartTime: now,
		TaskID:    task.ID,
	}
	err = repo.CreateTimeEntry(context.Background(), entry)
	require.NoError(t, err)

	// Update entry
	newTime := now.Add(time.Hour)
	endTime := now.Add(2 * time.Hour)
	entry.StartTime = newTime
	entry.EndTime = &endTime
	entry.TaskID = task.ID

	err = repo.UpdateTimeEntry(context.Background(), entry)
	require.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetTimeEntry(context.Background(), entry.ID)
	require.NoError(t, err)
	assert.Equal(t, newTime.Unix(), retrieved.StartTime.Unix())
	assert.Equal(t, endTime.Unix(), retrieved.EndTime.Unix())
	assert.Equal(t, task.ID, retrieved.TaskID)

	// Test updating non-existent entry
	nonExistent := &TimeEntry{ID: 999, TaskID: task.ID}
	err = repo.UpdateTimeEntry(context.Background(), nonExistent)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeleteTimeEntry(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	task := &Task{TaskName: "Test task"}
	err := repo.CreateTask(context.Background(), task)
	require.NoError(t, err)

	// Create entry
	entry := &TimeEntry{
		StartTime: time.Now(),
		TaskID:    task.ID,
	}
	err = repo.CreateTimeEntry(context.Background(), entry)
	require.NoError(t, err)

	// Delete entry
	err = repo.DeleteTimeEntry(context.Background(), entry.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = repo.GetTimeEntry(context.Background(), entry.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test deleting non-existent entry
	err = repo.DeleteTimeEntry(context.Background(), 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSearchTimeEntries(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	// Create test tasks
	task1 := &Task{TaskName: "First meeting"}
	task2 := &Task{TaskName: "Second meeting"}
	task3 := &Task{TaskName: "Third meeting"}
	require.NoError(t, repo.CreateTask(context.Background(), task1))
	require.NoError(t, repo.CreateTask(context.Background(), task2))
	require.NoError(t, repo.CreateTask(context.Background(), task3))

	now := time.Now()
	startTime1 := now.Add(-2 * time.Hour)
	endTime1 := now
	startTime2 := now.Add(-1 * time.Hour)
	startTime3 := now
	endTime3 := now.Add(time.Hour)

	entries := []*TimeEntry{
		{
			StartTime: startTime1,
			EndTime:   &endTime1,
			TaskID:    task1.ID,
		},
		{
			StartTime: startTime2,
			TaskID:    task2.ID,
		},
		{
			StartTime: startTime3,
			EndTime:   &endTime3,
			TaskID:    task3.ID,
		},
	}

	for _, entry := range entries {
		err := repo.CreateTimeEntry(context.Background(), entry)
		require.NoError(t, err)
	}

	searchStart := now.Add(-3 * time.Hour)
	searchEnd := now.Add(time.Hour)

	tests := []struct {
		name     string
		opts     SearchOptions
		expected int
	}{
		{
			name: "Search by time range",
			opts: SearchOptions{
				StartTime: &searchStart,
				EndTime:   &searchEnd,
			},
			expected: 3,
		},
		{
			name: "Search by task name",
			opts: SearchOptions{
				TaskName: stringPtr("meeting"),
			},
			expected: 3,
		},
		{
			name: "Search by time range and task name",
			opts: SearchOptions{
				StartTime: &searchStart,
				EndTime:   &searchEnd,
				TaskName:  stringPtr("First"),
			},
			expected: 1,
		},
		{
			name: "Search with no results",
			opts: SearchOptions{
				TaskName: stringPtr("nonexistent"),
			},
			expected: 0,
		},
		{
			name:     "Search for running tasks",
			opts:     SearchOptions{},
			expected: 1, // Only the second entry has no end time
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := repo.SearchTimeEntries(context.Background(), tt.opts)
			require.NoError(t, err)
			assert.Len(t, results, tt.expected)

			// Verify ascending order
			for i := 1; i < len(results); i++ {
				assert.True(t, results[i-1].StartTime.Before(results[i].StartTime))
			}

			// For running tasks test, verify the result has no end time and correct task
			if tt.name == "Search for running tasks" {
				assert.Len(t, results, 1)
				assert.Nil(t, results[0].EndTime)
				assert.Equal(t, task2.ID, results[0].TaskID)
			}
		})
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

func TestTimeFormatting(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a task
	task := &Task{TaskName: "Test task"}
	err := repo.CreateTask(context.Background(), task)
	require.NoError(t, err)

	// Create a time entry with a specific time
	testTime := time.Date(2025, 6, 23, 11, 47, 24, 890799237, time.UTC)
	entry := &TimeEntry{
		StartTime: testTime,
		TaskID:    task.ID,
	}

	err = repo.CreateTimeEntry(context.Background(), entry)
	require.NoError(t, err)

	// Retrieve the entry
	retrieved, err := repo.GetTimeEntry(context.Background(), entry.ID)
	require.NoError(t, err)

	// Verify the time is stored correctly
	expectedRFC3339 := "2025-06-23T11:47:24Z"
	assert.Equal(t, expectedRFC3339, retrieved.StartTime.Format(time.RFC3339))

	// Verify the time values are equal (ignoring monotonic clock)
	assert.Equal(t, testTime.Unix(), retrieved.StartTime.Unix())
}
