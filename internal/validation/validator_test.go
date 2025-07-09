package validation

import (
	"testing"
	"time"
)

func TestValidator_IsNonEmptyString(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Empty string", "", false},
		{"Whitespace only", "   ", false},
		{"Tab and newline", "\t\n", false},
		{"Valid string", "hello", true},
		{"String with spaces", "hello world", true},
		{"String with leading/trailing spaces", "  hello  ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.IsNonEmptyString(tt.input)
			if result != tt.expected {
				t.Errorf("IsNonEmptyString(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidator_IsValidStringLength(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name     string
		input    string
		min      int
		max      int
		expected bool
	}{
		{"Empty string, min 1", "", 1, 10, false},
		{"Too short", "a", 2, 10, false},
		{"Too long", "very long string", 1, 5, false},
		{"Valid length", "hello", 1, 10, true},
		{"Exactly min", "ab", 2, 10, true},
		{"Exactly max", "hello", 1, 5, true},
		{"With leading/trailing spaces", "  hello  ", 1, 10, true}, // Should trim spaces
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.IsValidStringLength(tt.input, tt.min, tt.max)
			if result != tt.expected {
				t.Errorf("IsValidStringLength(%q, %d, %d) = %v, expected %v", tt.input, tt.min, tt.max, result, tt.expected)
			}
		})
	}
}

func TestValidator_IsValidTaskName(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid basic name", "Task 1", true},
		{"Name with hyphen", "Task-1", true},
		{"Name with underscore", "Task_1", true},
		{"Name with numbers", "Task123", true},
		{"Name with punctuation", "Task 1! (important)", true},
		{"Name with question mark", "Is this done?", true},
		{"Invalid characters", "Task@#$%", false},
		{"Name with newline", "Task\nname", false},
		{"Name with tab", "Task\tname", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.IsValidTaskName(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidTaskName(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidator_IsValidTimeRange(t *testing.T) {
	validator := NewValidator()
	
	startTime := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
	endTime := time.Date(2023, 1, 1, 11, 0, 0, 0, time.UTC)
	earlyTime := time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		start    time.Time
		end      *time.Time
		expected bool
	}{
		{"Valid range", startTime, &endTime, true},
		{"No end time (running)", startTime, nil, true},
		{"Invalid range (end before start)", startTime, &earlyTime, false},
		{"Same start and end", startTime, &startTime, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.IsValidTimeRange(tt.start, tt.end)
			if result != tt.expected {
				t.Errorf("IsValidTimeRange(%v, %v) = %v, expected %v", tt.start, tt.end, result, tt.expected)
			}
		})
	}
}

func TestValidator_IsValidDuration(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name     string
		duration time.Duration
		expected bool
	}{
		{"Valid duration", 1 * time.Hour, true},
		{"Zero duration", 0, false},
		{"Negative duration", -1 * time.Hour, false},
		{"Too long duration", 25 * time.Hour, false},
		{"Exactly 24 hours", 24 * time.Hour, true},
		{"Just under 24 hours", 23*time.Hour + 59*time.Minute, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.IsValidDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("IsValidDuration(%v) = %v, expected %v", tt.duration, result, tt.expected)
			}
		})
	}
}

func TestValidator_IsValidTaskID(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name     string
		id       int64
		expected bool
	}{
		{"Valid ID", 1, true},
		{"Zero ID", 0, false},
		{"Negative ID", -1, false},
		{"Large ID", 999999, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.IsValidTaskID(tt.id)
			if result != tt.expected {
				t.Errorf("IsValidTaskID(%d) = %v, expected %v", tt.id, result, tt.expected)
			}
		})
	}
}

func TestValidator_IsValidTimeShorthand(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid minutes", "30m", true},
		{"Valid hours", "2h", true},
		{"Valid days", "1d", true},
		{"Valid weeks", "2w", true},
		{"Valid months", "3mo", true},
		{"Valid years", "1y", true},
		{"Invalid format", "30", false},
		{"Invalid unit", "30x", false},
		{"Zero value", "0m", false},
		{"Negative value", "-1h", false},
		{"No number", "m", false},
		{"Multiple units", "1h30m", false},
		{"Decimal value", "1.5h", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.IsValidTimeShorthand(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidTimeShorthand(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidator_IsReasonableDate(t *testing.T) {
	validator := NewValidator()
	
	now := time.Now()
	tenYearsAgo := now.AddDate(-10, 0, 0)
	elevenYearsAgo := now.AddDate(-11, 0, 0)
	oneYearFromNow := now.AddDate(1, 0, 0)
	twoYearsFromNow := now.AddDate(2, 0, 0)

	tests := []struct {
		name     string
		date     time.Time
		expected bool
	}{
		{"Current time", now, true},
		{"One year ago", now.AddDate(-1, 0, 0), true},
		{"Ten years ago (boundary)", tenYearsAgo.Add(time.Hour), true},
		{"Eleven years ago", elevenYearsAgo, false},
		{"One year from now (boundary)", oneYearFromNow.Add(-time.Hour), true},
		{"Two years from now", twoYearsFromNow, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.IsReasonableDate(tt.date)
			if result != tt.expected {
				t.Errorf("IsReasonableDate(%v) = %v, expected %v", tt.date, result, tt.expected)
			}
		})
	}
}

func TestValidator_IsValidDateRange(t *testing.T) {
	validator := NewValidator()
	
	startTime := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
	endTime := time.Date(2023, 1, 1, 11, 0, 0, 0, time.UTC)
	earlyTime := time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		start    *time.Time
		end      *time.Time
		expected bool
	}{
		{"Valid range", &startTime, &endTime, true},
		{"Same start and end", &startTime, &startTime, true},
		{"Invalid range", &startTime, &earlyTime, false},
		{"Nil start", nil, &endTime, true},
		{"Nil end", &startTime, nil, true},
		{"Both nil", nil, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.IsValidDateRange(tt.start, tt.end)
			if result != tt.expected {
				t.Errorf("IsValidDateRange(%v, %v) = %v, expected %v", tt.start, tt.end, result, tt.expected)
			}
		})
	}
}

func TestValidator_TrimAndValidateString(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"No trimming needed", "hello", "hello"},
		{"Leading spaces", "  hello", "hello"},
		{"Trailing spaces", "hello  ", "hello"},
		{"Both sides", "  hello  ", "hello"},
		{"With tabs", "\thello\t", "hello"},
		{"With newlines", "\nhello\n", "hello"},
		{"Mixed whitespace", " \t\nhello\n\t ", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.TrimAndValidateString(tt.input)
			if result != tt.expected {
				t.Errorf("TrimAndValidateString(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}