package sqlite

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*SQLiteRepository, func()) {
	// Create a temporary file for the test database
	tmpfile, err := os.CreateTemp("", "testdb-*.db")
	require.NoError(t, err)
	
	// Create repository instance
	repo, err := New(tmpfile.Name())
	require.NoError(t, err)

	// Return cleanup function
	cleanup := func() {
		repo.Close()
		os.Remove(tmpfile.Name())
	}

	return repo, cleanup
}

func TestCreateTimeEntry(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	now := time.Now()
	entry := &TimeEntry{
		StartTime:   now,
		Description: "Test entry",
	}

	// Test creating entry
	err := repo.CreateTimeEntry(entry)
	require.NoError(t, err)
	assert.Greater(t, entry.ID, int64(0))

	// Verify entry was created
	retrieved, err := repo.GetTimeEntry(entry.ID)
	require.NoError(t, err)
	assert.Equal(t, entry.ID, retrieved.ID)
	assert.Equal(t, entry.StartTime.Unix(), retrieved.StartTime.Unix())
	assert.Equal(t, entry.Description, retrieved.Description)
	assert.Nil(t, retrieved.EndTime)
}

func TestGetTimeEntry(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	// Test getting non-existent entry
	_, err := repo.GetTimeEntry(999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Create and get entry
	now := time.Now()
	entry := &TimeEntry{
		StartTime:   now,
		Description: "Test entry",
	}
	err = repo.CreateTimeEntry(entry)
	require.NoError(t, err)

	retrieved, err := repo.GetTimeEntry(entry.ID)
	require.NoError(t, err)
	assert.Equal(t, entry.ID, retrieved.ID)
	assert.Equal(t, entry.StartTime.Unix(), retrieved.StartTime.Unix())
	assert.Equal(t, entry.Description, retrieved.Description)
}

func TestListTimeEntries(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	// Create multiple entries
	entries := []*TimeEntry{
		{StartTime: time.Now().Add(-2 * time.Hour), Description: "First entry"},
		{StartTime: time.Now().Add(-1 * time.Hour), Description: "Second entry"},
		{StartTime: time.Now(), Description: "Third entry"},
	}

	for _, entry := range entries {
		err := repo.CreateTimeEntry(entry)
		require.NoError(t, err)
	}

	// Test listing entries
	retrieved, err := repo.ListTimeEntries()
	require.NoError(t, err)
	assert.Len(t, retrieved, 3)

	// Verify order (ascending by start time)
	assert.True(t, retrieved[0].StartTime.Before(retrieved[1].StartTime))
	assert.True(t, retrieved[1].StartTime.Before(retrieved[2].StartTime))
}

func TestUpdateTimeEntry(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	// Create entry
	now := time.Now()
	entry := &TimeEntry{
		StartTime:   now,
		Description: "Original description",
	}
	err := repo.CreateTimeEntry(entry)
	require.NoError(t, err)

	// Update entry
	newTime := now.Add(time.Hour)
	endTime := now.Add(2 * time.Hour)
	entry.StartTime = newTime
	entry.EndTime = &endTime
	entry.Description = "Updated description"

	err = repo.UpdateTimeEntry(entry)
	require.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetTimeEntry(entry.ID)
	require.NoError(t, err)
	assert.Equal(t, newTime.Unix(), retrieved.StartTime.Unix())
	assert.Equal(t, endTime.Unix(), retrieved.EndTime.Unix())
	assert.Equal(t, "Updated description", retrieved.Description)

	// Test updating non-existent entry
	nonExistent := &TimeEntry{ID: 999}
	err = repo.UpdateTimeEntry(nonExistent)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeleteTimeEntry(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	// Create entry
	entry := &TimeEntry{
		StartTime:   time.Now(),
		Description: "Test entry",
	}
	err := repo.CreateTimeEntry(entry)
	require.NoError(t, err)

	// Delete entry
	err = repo.DeleteTimeEntry(entry.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = repo.GetTimeEntry(entry.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test deleting non-existent entry
	err = repo.DeleteTimeEntry(999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSearchTimeEntries(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	// Create test entries
	now := time.Now()
	startTime1 := now.Add(-2 * time.Hour)
	endTime1 := now
	startTime2 := now.Add(-1 * time.Hour)
	startTime3 := now
	endTime3 := now.Add(time.Hour)

	entries := []*TimeEntry{
		{
			StartTime:   startTime1,
			EndTime:     &endTime1,
			Description: "First meeting",
		},
		{
			StartTime:   startTime2,
			Description: "Second meeting",
		},
		{
			StartTime:   startTime3,
			EndTime:     &endTime3,
			Description: "Third meeting",
		},
	}

	for _, entry := range entries {
		err := repo.CreateTimeEntry(entry)
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
			name: "Search by description",
			opts: SearchOptions{
				Description: stringPtr("meeting"),
			},
			expected: 3,
		},
		{
			name: "Search by time range and description",
			opts: SearchOptions{
				StartTime:   &searchStart,
				EndTime:     &searchEnd,
				Description: stringPtr("First"),
			},
			expected: 1,
		},
		{
			name: "Search with no results",
			opts: SearchOptions{
				Description: stringPtr("nonexistent"),
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := repo.SearchTimeEntries(tt.opts)
			require.NoError(t, err)
			assert.Len(t, results, tt.expected)

			// Verify ascending order
			for i := 1; i < len(results); i++ {
				assert.True(t, results[i-1].StartTime.Before(results[i].StartTime))
			}
		})
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
} 