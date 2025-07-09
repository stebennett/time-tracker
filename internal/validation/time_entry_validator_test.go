package validation

import (
	"testing"
	"time"
	"time-tracker/internal/domain"
)

func TestTimeEntryValidator_ValidateTimeEntryForCreation(t *testing.T) {
	validator := NewTimeEntryValidator()
	
	now := time.Now()
	future := now.Add(1 * time.Hour)
	past := now.Add(-1 * time.Hour)
	tooFuture := now.Add(25 * time.Hour)

	tests := []struct {
		name        string
		taskID      int64
		startTime   time.Time
		endTime     *time.Time
		expectError bool
	}{
		{"Valid running entry", 1, now, nil, false},
		{"Valid completed entry", 1, past, &now, false},
		{"Invalid task ID", 0, now, nil, true},
		{"Zero start time", 1, time.Time{}, nil, true},
		{"End before start", 1, now, &past, true},
		{"Too long duration", 1, now, &tooFuture, true},
		{"Valid short duration", 1, now, &future, false},
		{"Same start and end", 1, now, &now, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateTimeEntryForCreation(tt.taskID, tt.startTime, tt.endTime)
			
			if tt.expectError && err == nil {
				t.Errorf("ValidateTimeEntryForCreation(%d, %v, %v) expected error but got nil", tt.taskID, tt.startTime, tt.endTime)
			} else if !tt.expectError && err != nil {
				t.Errorf("ValidateTimeEntryForCreation(%d, %v, %v) expected no error but got %v", tt.taskID, tt.startTime, tt.endTime, err)
			}
		})
	}
}

func TestTimeEntryValidator_ValidateTimeEntryForUpdate(t *testing.T) {
	validator := NewTimeEntryValidator()
	
	now := time.Now()
	past := now.Add(-1 * time.Hour)

	tests := []struct {
		name        string
		id          int64
		taskID      int64
		startTime   time.Time
		endTime     *time.Time
		expectError bool
	}{
		{"Valid update", 1, 1, past, &now, false},
		{"Invalid entry ID", 0, 1, past, &now, true},
		{"Invalid task ID", 1, 0, past, &now, true},
		{"Valid running entry", 1, 1, now, nil, false},
		{"End before start", 1, 1, now, &past, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateTimeEntryForUpdate(tt.id, tt.taskID, tt.startTime, tt.endTime)
			
			if tt.expectError && err == nil {
				t.Errorf("ValidateTimeEntryForUpdate(%d, %d, %v, %v) expected error but got nil", tt.id, tt.taskID, tt.startTime, tt.endTime)
			} else if !tt.expectError && err != nil {
				t.Errorf("ValidateTimeEntryForUpdate(%d, %d, %v, %v) expected no error but got %v", tt.id, tt.taskID, tt.startTime, tt.endTime, err)
			}
		})
	}
}

func TestTimeEntryValidator_ValidateTimeEntry(t *testing.T) {
	validator := NewTimeEntryValidator()
	
	now := time.Now()
	past := now.Add(-1 * time.Hour)

	tests := []struct {
		name        string
		timeEntry   domain.TimeEntry
		expectError bool
	}{
		{"Valid running entry", domain.TimeEntry{ID: 1, TaskID: 1, StartTime: now, EndTime: nil}, false},
		{"Valid completed entry", domain.TimeEntry{ID: 1, TaskID: 1, StartTime: past, EndTime: &now}, false},
		{"Invalid task ID", domain.TimeEntry{ID: 1, TaskID: 0, StartTime: now, EndTime: nil}, true},
		{"Invalid start time", domain.TimeEntry{ID: 1, TaskID: 1, StartTime: time.Time{}, EndTime: nil}, true},
		{"End before start", domain.TimeEntry{ID: 1, TaskID: 1, StartTime: now, EndTime: &past}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateTimeEntry(tt.timeEntry)
			
			if tt.expectError && err == nil {
				t.Errorf("ValidateTimeEntry(%+v) expected error but got nil", tt.timeEntry)
			} else if !tt.expectError && err != nil {
				t.Errorf("ValidateTimeEntry(%+v) expected no error but got %v", tt.timeEntry, err)
			}
		})
	}
}

func TestTimeEntryValidator_ValidateSearchOptions(t *testing.T) {
	validator := NewTimeEntryValidator()
	
	now := time.Now()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)
	veryFuture := now.AddDate(2, 0, 0)
	taskName := "Test Task"
	emptyTaskName := ""

	tests := []struct {
		name        string
		opts        domain.SearchOptions
		expectError bool
	}{
		{"Empty options", domain.SearchOptions{}, false},
		{"Valid date range", domain.SearchOptions{StartTime: &past, EndTime: &future}, false},
		{"Invalid date range", domain.SearchOptions{StartTime: &future, EndTime: &past}, true},
		{"Valid task ID", domain.SearchOptions{TaskID: intPtr(1)}, false},
		{"Invalid task ID", domain.SearchOptions{TaskID: intPtr(0)}, true},
		{"Valid task name", domain.SearchOptions{TaskName: &taskName}, false},
		{"Empty task name", domain.SearchOptions{TaskName: &emptyTaskName}, true},
		{"Future date too far", domain.SearchOptions{StartTime: &veryFuture}, true},
		{"Same start and end", domain.SearchOptions{StartTime: &now, EndTime: &now}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateSearchOptions(tt.opts)
			
			if tt.expectError && err == nil {
				t.Errorf("ValidateSearchOptions(%+v) expected error but got nil", tt.opts)
			} else if !tt.expectError && err != nil {
				t.Errorf("ValidateSearchOptions(%+v) expected no error but got %v", tt.opts, err)
			}
		})
	}
}

func TestTimeEntryValidator_ValidateTimeShorthand(t *testing.T) {
	validator := NewTimeEntryValidator()

	tests := []struct {
		name        string
		shorthand   string
		expectError bool
	}{
		{"Valid minutes", "30m", false},
		{"Valid hours", "2h", false},
		{"Valid days", "1d", false},
		{"Valid weeks", "2w", false},
		{"Valid months", "3mo", false},
		{"Valid years", "1y", false},
		{"Invalid format", "30", true},
		{"Invalid unit", "30x", true},
		{"Zero value", "0m", true},
		{"Negative value", "-1h", true},
		{"No number", "m", true},
		{"Multiple units", "1h30m", true},
		{"Decimal value", "1.5h", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateTimeShorthand(tt.shorthand)
			
			if tt.expectError && err == nil {
				t.Errorf("ValidateTimeShorthand(%q) expected error but got nil", tt.shorthand)
			} else if !tt.expectError && err != nil {
				t.Errorf("ValidateTimeShorthand(%q) expected no error but got %v", tt.shorthand, err)
			}
		})
	}
}

func TestTimeEntryValidator_ValidateTimeEntryID(t *testing.T) {
	validator := NewTimeEntryValidator()

	tests := []struct {
		name        string
		id          int64
		expectError bool
	}{
		{"Valid ID", 1, false},
		{"Zero ID", 0, true},
		{"Negative ID", -1, true},
		{"Large ID", 999999, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateTimeEntryID(tt.id)
			
			if tt.expectError && err == nil {
				t.Errorf("ValidateTimeEntryID(%d) expected error but got nil", tt.id)
			} else if !tt.expectError && err != nil {
				t.Errorf("ValidateTimeEntryID(%d) expected no error but got %v", tt.id, err)
			}
		})
	}
}

// Helper function to create int64 pointer
func intPtr(i int64) *int64 {
	return &i
}