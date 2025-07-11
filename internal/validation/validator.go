package validation

import (
	"regexp"
	"strconv"
	"strings"
	"time"
	"time-tracker/internal/config"
)

// Validator provides common validation utilities
type Validator struct {
	timeShorthandRegex *regexp.Regexp
	config             *config.Config
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{
		timeShorthandRegex: regexp.MustCompile(`^(\d+)(m|h|d|w|mo|y)$`),
		config:             nil, // Use defaults
	}
}

// NewValidatorWithConfig creates a new validator instance with configuration
func NewValidatorWithConfig(cfg *config.Config) *Validator {
	return &Validator{
		timeShorthandRegex: regexp.MustCompile(`^(\d+)(m|h|d|w|mo|y)$`),
		config:             cfg,
	}
}

// IsNonEmptyString checks if a string is not empty after trimming whitespace
func (v *Validator) IsNonEmptyString(s string) bool {
	return strings.TrimSpace(s) != ""
}

// IsValidStringLength checks if a string length is within the specified range
func (v *Validator) IsValidStringLength(s string, min, max int) bool {
	length := len(strings.TrimSpace(s))
	return length >= min && length <= max
}

// IsValidTaskNameLength checks if a task name length is within configured limits
func (v *Validator) IsValidTaskNameLength(name string) bool {
	length := len(strings.TrimSpace(name))
	minLen := v.getTaskNameMinLength()
	maxLen := v.getTaskNameMaxLength()
	return length >= minLen && length <= maxLen
}

// IsValidTaskName checks if a task name contains only allowed characters
func (v *Validator) IsValidTaskName(name string) bool {
	// Allow alphanumeric characters, spaces, hyphens, underscores, and common punctuation
	// But explicitly reject newlines, tabs, and other control characters
	validChars := regexp.MustCompile(`^[a-zA-Z0-9 \-_.,!?()]+$`)
	return validChars.MatchString(name)
}

// IsValidTimeRange checks if start time is before end time
func (v *Validator) IsValidTimeRange(startTime time.Time, endTime *time.Time) bool {
	if endTime == nil {
		return true // Running task, no end time
	}
	return startTime.Before(*endTime)
}

// IsValidDuration checks if a duration is within reasonable bounds
func (v *Validator) IsValidDuration(duration time.Duration) bool {
	maxDuration := v.getMaxDuration()
	return duration > 0 && duration <= maxDuration
}

// IsValidTaskID checks if a task ID is valid (positive)
func (v *Validator) IsValidTaskID(id int64) bool {
	return id > 0
}

// IsValidTimeShorthand checks if a time shorthand format is valid
func (v *Validator) IsValidTimeShorthand(shorthand string) bool {
	matches := v.timeShorthandRegex.FindStringSubmatch(shorthand)
	if matches == nil {
		return false
	}
	
	// Check if the number is valid
	value, err := strconv.Atoi(matches[1])
	if err != nil || value <= 0 {
		return false
	}
	
	// Check if the unit is valid
	unit := matches[2]
	validUnits := []string{"m", "h", "d", "w", "mo", "y"}
	for _, validUnit := range validUnits {
		if unit == validUnit {
			return true
		}
	}
	
	return false
}

// IsReasonableDate checks if a date is within reasonable bounds
func (v *Validator) IsReasonableDate(t time.Time) bool {
	now := time.Now()
	// Allow dates from 10 years ago to 1 year in the future
	tenYearsAgo := now.AddDate(-10, 0, 0)
	oneYearFromNow := now.AddDate(1, 0, 0)
	
	return t.After(tenYearsAgo) && t.Before(oneYearFromNow)
}

// IsValidDateRange checks if a date range is logical
func (v *Validator) IsValidDateRange(startTime, endTime *time.Time) bool {
	if startTime == nil || endTime == nil {
		return true // One or both dates are nil, which is valid for open-ended ranges
	}
	return startTime.Before(*endTime) || startTime.Equal(*endTime)
}

// TrimAndValidateString trims whitespace and returns the cleaned string
func (v *Validator) TrimAndValidateString(s string) string {
	return strings.TrimSpace(s)
}

// getTaskNameMinLength returns configured minimum task name length or default
func (v *Validator) getTaskNameMinLength() int {
	if v.config != nil {
		return v.config.Validation.TaskNameMinLength
	}
	return 1 // Default minimum
}

// getTaskNameMaxLength returns configured maximum task name length or default
func (v *Validator) getTaskNameMaxLength() int {
	if v.config != nil {
		return v.config.Validation.TaskNameMaxLength
	}
	return 255 // Default maximum
}

// getMaxDuration returns configured maximum duration or default
func (v *Validator) getMaxDuration() time.Duration {
	if v.config != nil {
		return v.config.Validation.MaxDuration
	}
	return 24 * time.Hour // Default maximum
}