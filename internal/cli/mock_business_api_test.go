package cli

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"time-tracker/internal/api"
	"time-tracker/internal/domain"
	"time-tracker/internal/errors"
)

// mockBusinessAPI implements the BusinessAPI interface for testing
type mockBusinessAPI struct {
	tasks         map[int64]*domain.Task
	timeEntries   map[int64]*domain.TimeEntry
	nextTaskID    int64
	nextEntryID   int64
	currentTaskID *int64 // Track currently running task
}

// newMockBusinessAPI creates a new mock BusinessAPI instance
func newMockBusinessAPI() api.BusinessAPI {
	return &mockBusinessAPI{
		tasks:       make(map[int64]*domain.Task),
		timeEntries: make(map[int64]*domain.TimeEntry),
		nextTaskID:  1,
		nextEntryID: 1,
	}
}

func (m *mockBusinessAPI) StartNewTask(ctx context.Context, taskName string) (*api.TaskSession, error) {
	// Stop any running tasks first
	_, _ = m.StopAllRunningTasks(ctx)

	// Create new task
	task := &domain.Task{
		ID:       m.nextTaskID,
		TaskName: taskName,
	}
	m.tasks[task.ID] = task
	m.nextTaskID++

	// Create time entry
	now := time.Now()
	entry := &domain.TimeEntry{
		ID:        m.nextEntryID,
		TaskID:    task.ID,
		StartTime: now,
		EndTime:   nil, // Running
	}
	m.timeEntries[entry.ID] = entry
	m.nextEntryID++
	m.currentTaskID = &task.ID

	return &api.TaskSession{
		Task:     task,
		Duration: "running for 0m",
	}, nil
}

func (m *mockBusinessAPI) ResumeTask(ctx context.Context, taskID int64) (*api.TaskSession, error) {
	// Stop any running tasks first
	_, _ = m.StopAllRunningTasks(ctx)

	task, exists := m.tasks[taskID]
	if !exists {
		return nil, errors.NewNotFoundError("task", fmt.Sprintf("%d", taskID))
	}

	// Create new time entry
	now := time.Now()
	entry := &domain.TimeEntry{
		ID:        m.nextEntryID,
		TaskID:    taskID,
		StartTime: now,
		EndTime:   nil, // Running
	}
	m.timeEntries[entry.ID] = entry
	m.nextEntryID++
	m.currentTaskID = &taskID

	return &api.TaskSession{
		Task:     task,
		Duration: "running for 0m",
	}, nil
}

func (m *mockBusinessAPI) StopAllRunningTasks(ctx context.Context) ([]*domain.TimeEntry, error) {
	var stopped []*domain.TimeEntry
	now := time.Now()

	for _, entry := range m.timeEntries {
		if entry.EndTime == nil {
			entry.EndTime = &now
			stopped = append(stopped, entry)
		}
	}

	m.currentTaskID = nil
	return stopped, nil
}

func (m *mockBusinessAPI) DeleteTaskWithEntries(ctx context.Context, taskID int64) error {
	// Delete all time entries for this task
	for id, entry := range m.timeEntries {
		if entry.TaskID == taskID {
			delete(m.timeEntries, id)
		}
	}

	// Delete the task
	delete(m.tasks, taskID)

	if m.currentTaskID != nil && *m.currentTaskID == taskID {
		m.currentTaskID = nil
	}

	return nil
}

func (m *mockBusinessAPI) UpdateTaskName(ctx context.Context, taskID int64, newName string) (*domain.Task, error) {
	task, exists := m.tasks[taskID]
	if !exists {
		return nil, errors.NewNotFoundError("task", fmt.Sprintf("%d", taskID))
	}

	task.TaskName = newName
	return task, nil
}

func (m *mockBusinessAPI) GetCurrentSession(ctx context.Context) (*api.TaskSession, error) {
	if m.currentTaskID == nil {
		return nil, errors.NewNotFoundError("running task", "")
	}

	task, exists := m.tasks[*m.currentTaskID]
	if !exists {
		return nil, errors.NewNotFoundError("task", fmt.Sprintf("%d", *m.currentTaskID))
	}

	// Find running entry
	var runningEntry *domain.TimeEntry
	for _, entry := range m.timeEntries {
		if entry.TaskID == *m.currentTaskID && entry.EndTime == nil {
			runningEntry = entry
			break
		}
	}

	if runningEntry == nil {
		m.currentTaskID = nil
		return nil, errors.NewNotFoundError("running task", "")
	}

	duration := time.Since(runningEntry.StartTime)
	minutes := int(duration.Minutes())
	
	return &api.TaskSession{
		Task:     task,
		Duration: fmt.Sprintf("running for %dm", minutes),
	}, nil
}

func (m *mockBusinessAPI) GetTask(ctx context.Context, id int64) (*domain.Task, error) {
	task, exists := m.tasks[id]
	if !exists {
		return nil, errors.NewNotFoundError("task", fmt.Sprintf("%d", id))
	}
	return task, nil
}

func (m *mockBusinessAPI) GetTaskSummary(ctx context.Context, taskID int64) (*api.TaskSummary, error) {
	task, exists := m.tasks[taskID]
	if !exists {
		return nil, errors.NewNotFoundError("task", fmt.Sprintf("%d", taskID))
	}

	var entries []*domain.TimeEntry
	var totalDuration time.Duration
	var runningCount int
	var firstEntry, lastEntry time.Time

	for _, entry := range m.timeEntries {
		if entry.TaskID == taskID {
			entries = append(entries, entry)
			if firstEntry.IsZero() || entry.StartTime.Before(firstEntry) {
				firstEntry = entry.StartTime
			}
			if entry.EndTime != nil {
				duration := entry.EndTime.Sub(entry.StartTime)
				totalDuration += duration
				if entry.EndTime.After(lastEntry) {
					lastEntry = *entry.EndTime
				}
			} else {
				runningCount++
				duration := time.Since(entry.StartTime)
				totalDuration += duration
				if time.Now().After(lastEntry) {
					lastEntry = time.Now()
				}
			}
		}
	}

	// Sort entries by start time
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].StartTime.Before(entries[j].StartTime)
	})

	hours := int(totalDuration.Hours())
	minutes := int(totalDuration.Minutes()) % 60

	return &api.TaskSummary{
		Task:         task,
		TimeEntries:  entries,
		TotalTime:    fmt.Sprintf("%dh %dm", hours, minutes),
		SessionCount: len(entries),
		RunningCount: runningCount,
		FirstEntry:   firstEntry,
		LastEntry:    lastEntry,
		IsRunning:    runningCount > 0,
	}, nil
}

func (m *mockBusinessAPI) ParseTimeRange(ctx context.Context, timeStr string) (*api.TimeRange, error) {
	// Simple implementation for testing
	if timeStr == "" {
		return nil, nil
	}
	
	now := time.Now()
	var start time.Time
	
	switch timeStr {
	case "1h":
		start = now.Add(-1 * time.Hour)
	case "1d":
		start = now.Add(-24 * time.Hour)
	case "1w":
		start = now.Add(-7 * 24 * time.Hour)
	default:
		return nil, fmt.Errorf("unsupported time range: %s", timeStr)
	}
	
	return &api.TimeRange{
		Start: start,
		End:   now,
	}, nil
}

func (m *mockBusinessAPI) SearchTasks(ctx context.Context, timeRange string, textFilter string, sortOrder api.SortOrder) ([]*api.TaskActivity, error) {
	var result []*api.TaskActivity
	
	// Get time range if specified
	var timeRangeObj *api.TimeRange
	if timeRange != "" {
		tr, err := m.ParseTimeRange(ctx, timeRange)
		if err != nil {
			return nil, err
		}
		timeRangeObj = tr
	}
	
	for _, task := range m.tasks {
		// Apply text filter
		if textFilter != "" && !strings.Contains(strings.ToLower(task.TaskName), strings.ToLower(textFilter)) {
			continue
		}
		
		// Find entries for this task
		var taskEntries []*domain.TimeEntry
		for _, entry := range m.timeEntries {
			if entry.TaskID == task.ID {
				// Apply time filter
				if timeRangeObj != nil {
					if entry.StartTime.Before(timeRangeObj.Start) || entry.StartTime.After(timeRangeObj.End) {
						continue
					}
				}
				taskEntries = append(taskEntries, entry)
			}
		}
		
		if len(taskEntries) == 0 {
			continue
		}
		
		// Find latest entry for LastWorked
		var latestEntry *domain.TimeEntry
		for _, entry := range taskEntries {
			if latestEntry == nil || entry.StartTime.After(latestEntry.StartTime) {
				latestEntry = entry
			}
		}
		
		result = append(result, &api.TaskActivity{
			Task:       task,
			LastWorked: latestEntry.StartTime,
		})
	}
	
	// Simple sorting
	if sortOrder == api.SortByName {
		sort.Slice(result, func(i, j int) bool {
			return result[i].Task.TaskName < result[j].Task.TaskName
		})
	}
	
	return result, nil
}

func (m *mockBusinessAPI) SearchTimeEntries(ctx context.Context, timeRange string, textFilter string) ([]*api.TimeEntryWithTask, error) {
	var result []*api.TimeEntryWithTask
	
	// Get time range if specified
	var timeRangeObj *api.TimeRange
	if timeRange != "" {
		tr, err := m.ParseTimeRange(ctx, timeRange)
		if err != nil {
			return nil, err
		}
		timeRangeObj = tr
	}
	
	for _, entry := range m.timeEntries {
		task := m.tasks[entry.TaskID]
		
		// Apply text filter
		if textFilter != "" && !strings.Contains(strings.ToLower(task.TaskName), strings.ToLower(textFilter)) {
			continue
		}
		
		// Apply time filter
		if timeRangeObj != nil {
			if entry.StartTime.Before(timeRangeObj.Start) || entry.StartTime.After(timeRangeObj.End) {
				continue
			}
		}
		
		// Calculate duration
		var duration string
		if entry.EndTime != nil {
			d := entry.EndTime.Sub(entry.StartTime)
			hours := int(d.Hours())
			minutes := int(d.Minutes()) % 60
			duration = fmt.Sprintf("%dh %dm", hours, minutes)
		} else {
			d := time.Since(entry.StartTime)
			hours := int(d.Hours())
			minutes := int(d.Minutes()) % 60
			duration = fmt.Sprintf("%dh %dm", hours, minutes)
		}
		
		result = append(result, &api.TimeEntryWithTask{
			TimeEntry: entry,
			Task:      task,
			Duration:  duration,
		})
	}
	
	return result, nil
}

func (m *mockBusinessAPI) GetDashboardData(ctx context.Context, timeRange string) (*api.DashboardData, error) {
	// Simple implementation for testing
	return &api.DashboardData{}, nil
}

func (m *mockBusinessAPI) GetTodayStatistics(ctx context.Context) (*api.DayStatistics, error) {
	// Simple implementation for testing
	return &api.DayStatistics{}, nil
}

// setupTestAppWithMockBusinessAPI creates a test app with mock BusinessAPI
func setupTestAppWithMockBusinessAPI(t *testing.T) (*App, func()) {
	mockAPI := newMockBusinessAPI()
	app := NewApp(mockAPI)
	
	cleanup := func() {
		// Nothing to clean up for mock
	}
	
	return app, cleanup
}