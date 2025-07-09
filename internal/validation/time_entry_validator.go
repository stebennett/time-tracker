package validation

import (
	"time"
	"time-tracker/internal/domain"
)

// TimeEntryValidator provides validation for TimeEntry-related operations
type TimeEntryValidator struct {
	validator     *Validator
	taskValidator *TaskValidator
}

// NewTimeEntryValidator creates a new time entry validator
func NewTimeEntryValidator() *TimeEntryValidator {
	return &TimeEntryValidator{
		validator:     NewValidator(),
		taskValidator: NewTaskValidator(),
	}
}

// ValidateTimeEntryForCreation validates a time entry for creation
func (tev *TimeEntryValidator) ValidateTimeEntryForCreation(taskID int64, startTime time.Time, endTime *time.Time) error {
	validationError := NewValidationError()
	
	// Validate task ID
	if !tev.validator.IsValidTaskID(taskID) {
		validationError.AddInvalidValueError("task_id", taskID, "must be a positive integer")
	}
	
	// Validate start time
	if startTime.IsZero() {
		validationError.AddRequiredError("start_time")
	} else if !tev.validator.IsReasonableDate(startTime) {
		validationError.AddInvalidValueError("start_time", startTime, "must be within reasonable date range")
	}
	
	// Validate end time if provided
	if endTime != nil {
		if !tev.validator.IsReasonableDate(*endTime) {
			validationError.AddInvalidValueError("end_time", *endTime, "must be within reasonable date range")
		}
		
		// Validate time range
		if !tev.validator.IsValidTimeRange(startTime, endTime) {
			validationError.AddInvalidRangeError("time_range", map[string]time.Time{
				"start": startTime,
				"end":   *endTime,
			}, "end time must be after start time")
		}
		
		// Validate duration
		duration := endTime.Sub(startTime)
		if !tev.validator.IsValidDuration(duration) {
			validationError.AddInvalidValueError("duration", duration, "must be positive and less than 24 hours")
		}
	}
	
	if validationError.HasErrors() {
		return validationError
	}
	
	return nil
}

// ValidateTimeEntryForUpdate validates a time entry for update
func (tev *TimeEntryValidator) ValidateTimeEntryForUpdate(id int64, taskID int64, startTime time.Time, endTime *time.Time) error {
	validationError := NewValidationError()
	
	// Validate time entry ID
	if !tev.validator.IsValidTaskID(id) { // Using same validation as task ID
		validationError.AddInvalidValueError("time_entry_id", id, "must be a positive integer")
	}
	
	// Validate the time entry data
	if timeEntryErr := tev.ValidateTimeEntryForCreation(taskID, startTime, endTime); timeEntryErr != nil {
		if timeEntryValidationErr, ok := timeEntryErr.(*ValidationError); ok {
			validationError.Errors = append(validationError.Errors, timeEntryValidationErr.Errors...)
		}
	}
	
	if validationError.HasErrors() {
		return validationError
	}
	
	return nil
}

// ValidateTimeEntry validates a domain.TimeEntry object
func (tev *TimeEntryValidator) ValidateTimeEntry(timeEntry domain.TimeEntry) error {
	validationError := NewValidationError()
	
	// Validate using the domain model's IsValid method first
	if !timeEntry.IsValid() {
		validationError.AddInvalidValueError("time_entry", timeEntry, "fails basic validation")
	}
	
	// Perform additional validation
	if timeEntryErr := tev.ValidateTimeEntryForCreation(timeEntry.TaskID, timeEntry.StartTime, timeEntry.EndTime); timeEntryErr != nil {
		if timeEntryValidationErr, ok := timeEntryErr.(*ValidationError); ok {
			validationError.Errors = append(validationError.Errors, timeEntryValidationErr.Errors...)
		}
	}
	
	if validationError.HasErrors() {
		return validationError
	}
	
	return nil
}

// ValidateSearchOptions validates search options for time entries
func (tev *TimeEntryValidator) ValidateSearchOptions(opts domain.SearchOptions) error {
	validationError := NewValidationError()
	
	// Validate start time if provided
	if opts.StartTime != nil {
		if !tev.validator.IsReasonableDate(*opts.StartTime) {
			validationError.AddInvalidValueError("start_time", *opts.StartTime, "must be within reasonable date range")
		}
	}
	
	// Validate end time if provided
	if opts.EndTime != nil {
		if !tev.validator.IsReasonableDate(*opts.EndTime) {
			validationError.AddInvalidValueError("end_time", *opts.EndTime, "must be within reasonable date range")
		}
	}
	
	// Validate date range if both provided
	if !tev.validator.IsValidDateRange(opts.StartTime, opts.EndTime) {
		validationError.AddInvalidRangeError("date_range", map[string]interface{}{
			"start": opts.StartTime,
			"end":   opts.EndTime,
		}, "end time must be after or equal to start time")
	}
	
	// Validate task ID if provided
	if opts.TaskID != nil && !tev.validator.IsValidTaskID(*opts.TaskID) {
		validationError.AddInvalidValueError("task_id", *opts.TaskID, "must be a positive integer")
	}
	
	// Validate task name if provided
	if opts.TaskName != nil {
		trimmedName := tev.validator.TrimAndValidateString(*opts.TaskName)
		if !tev.validator.IsNonEmptyString(trimmedName) {
			validationError.AddInvalidValueError("task_name", *opts.TaskName, "must not be empty")
		} else if !tev.validator.IsValidStringLength(trimmedName, 1, 255) {
			validationError.AddInvalidLengthError("task_name", *opts.TaskName, 1, 255)
		}
	}
	
	if validationError.HasErrors() {
		return validationError
	}
	
	return nil
}

// ValidateTimeShorthand validates time shorthand format (e.g., "30m", "2h", "1d")
func (tev *TimeEntryValidator) ValidateTimeShorthand(shorthand string) error {
	if !tev.validator.IsValidTimeShorthand(shorthand) {
		validationError := NewValidationError()
		validationError.AddInvalidFormatError("time_shorthand", shorthand, "30m, 2h, 1d, 2w, 3mo, 1y")
		return validationError
	}
	return nil
}

// ValidateTimeEntryID validates a time entry ID
func (tev *TimeEntryValidator) ValidateTimeEntryID(id int64) error {
	if !tev.validator.IsValidTaskID(id) { // Using same validation as task ID
		validationError := NewValidationError()
		validationError.AddInvalidValueError("time_entry_id", id, "must be a positive integer")
		return validationError
	}
	return nil
}