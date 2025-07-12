package services

import (
	"context"
	"time"
	"time-tracker/internal/domain"
)

// TimeRange represents a time period with start and end times
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// TaskSession represents a task with its current time entry and duration info
type TaskSession struct {
	Task      *domain.Task      `json:"task"`
	TimeEntry *domain.TimeEntry `json:"time_entry"`
	Duration  string            `json:"duration"` // Human-readable duration
}

// TaskActivity represents task metadata with activity information
type TaskActivity struct {
	Task         *domain.Task `json:"task"`
	LastWorked   time.Time    `json:"last_worked"`
	TotalTime    string       `json:"total_time"`    // Human-readable total duration
	SessionCount int          `json:"session_count"`
	IsRunning    bool         `json:"is_running"`
}

// TaskSummary represents comprehensive analysis of a specific task
type TaskSummary struct {
	Task         *domain.Task        `json:"task"`
	TimeEntries  []*domain.TimeEntry `json:"time_entries"`
	TotalTime    string              `json:"total_time"`
	SessionCount int                 `json:"session_count"`
	RunningCount int                 `json:"running_count"`
	FirstEntry   time.Time           `json:"first_entry"`
	LastEntry    time.Time           `json:"last_entry"`
	IsRunning    bool                `json:"is_running"`
}

// DashboardData represents all data needed for a dashboard view
type DashboardData struct {
	RunningTask *TaskSession    `json:"running_task"`
	RecentTasks []*TaskActivity `json:"recent_tasks"`
	TodayStats  *DayStatistics  `json:"today_stats"`
}

// DayStatistics represents summary statistics for a specific day
type DayStatistics struct {
	TotalTime      string `json:"total_time"`
	TaskCount      int    `json:"task_count"`
	SessionCount   int    `json:"session_count"`
	CompletedCount int    `json:"completed_count"`
}

// TimeEntryWithTask represents a time entry with its associated task
type TimeEntryWithTask struct {
	TimeEntry *domain.TimeEntry `json:"time_entry"`
	Task      *domain.Task      `json:"task"`
	Duration  string            `json:"duration"`
}

// SearchCriteria represents criteria for searching tasks and time entries
type SearchCriteria struct {
	TimeRange   *TimeRange `json:"time_range,omitempty"`
	TextFilter  string     `json:"text_filter,omitempty"`
	TaskID      *int64     `json:"task_id,omitempty"`
	RunningOnly bool       `json:"running_only,omitempty"`
}

// SortOrder defines how task results should be sorted
type SortOrder string

const (
	SortByRecentFirst SortOrder = "recent_first" // Most recently worked (default)
	SortByOldestFirst SortOrder = "oldest_first" // Least recently worked (good for cleanup)
	SortByName        SortOrder = "name"         // Alphabetical by task name
	SortByDuration    SortOrder = "duration"     // By total time spent
)

// ActivityAnalysis represents detailed analysis of task activity patterns
type ActivityAnalysis struct {
	TotalDuration    time.Duration `json:"total_duration"`
	AverageDuration  time.Duration `json:"average_duration"`
	LongestSession   time.Duration `json:"longest_session"`
	ShortestSession  time.Duration `json:"shortest_session"`
	SessionCount     int           `json:"session_count"`
	ProductiveHours  []int         `json:"productive_hours"` // Hours of day (0-23) when most active
}

// TimeService handles time-related operations and calculations
type TimeService interface {
	// Time parsing and validation
	ParseTimeRange(timeStr string) (*TimeRange, error)
	ValidateTimeEntry(taskID int64, start time.Time, end *time.Time) error
	
	// Duration operations
	CalculateDuration(start time.Time, end *time.Time) string
	FormatDuration(duration time.Duration) string
	CalculateRunningDuration(startTime time.Time) string
	
	// Running task management
	GetRunningEntries(ctx context.Context) ([]*domain.TimeEntry, error)
	StopRunningEntries(ctx context.Context) ([]*domain.TimeEntry, error)
	CreateTimeEntry(ctx context.Context, taskID int64) (*domain.TimeEntry, error)
	
	// Time range operations
	IsToday(t time.Time) bool
	GetTodayRange() *TimeRange
	GetDateRange(date time.Time) *TimeRange
}

// TaskService handles task lifecycle and workflow operations
type TaskService interface {
	// Task CRUD operations
	CreateTask(ctx context.Context, name string) (*domain.Task, error)
	GetTask(ctx context.Context, id int64) (*domain.Task, error)
	UpdateTask(ctx context.Context, id int64, name string) (*domain.Task, error)
	DeleteTaskWithEntries(ctx context.Context, id int64) error
	
	// Task workflow operations
	StartNewTask(ctx context.Context, name string) (*TaskSession, error)
	ResumeTask(ctx context.Context, id int64) (*TaskSession, error)
	GetCurrentSession(ctx context.Context) (*TaskSession, error)
	
	// Task session management
	CreateTaskSession(task *domain.Task, entry *domain.TimeEntry) *TaskSession
	StopAllRunningTasks(ctx context.Context) ([]*domain.TimeEntry, error)
}

// SearchService handles search and discovery operations
type SearchService interface {
	// Task search operations
	SearchTasks(ctx context.Context, criteria SearchCriteria) ([]*TaskActivity, error)
	SearchTimeEntries(ctx context.Context, criteria SearchCriteria) ([]*TimeEntryWithTask, error)
	
	// Filter and sort operations
	FilterTasksByTime(tasks []*TaskActivity, timeRange *TimeRange) []*TaskActivity
	SortTasks(tasks []*TaskActivity, order SortOrder) []*TaskActivity
	SortTimeEntries(entries []*TimeEntryWithTask, order SortOrder) []*TimeEntryWithTask
	
	// Discovery operations
	GetRecentTasks(ctx context.Context, timeRange *TimeRange, limit int) ([]*TaskActivity, error)
	FindTasksWithActivity(ctx context.Context, timeRange *TimeRange) ([]*TaskActivity, error)
}

// ReportingService handles analytics and reporting operations
type ReportingService interface {
	// Task analysis
	GetTaskSummary(ctx context.Context, id int64) (*TaskSummary, error)
	AnalyzeTaskActivity(entries []*domain.TimeEntry) *ActivityAnalysis
	CalculateTaskStatistics(ctx context.Context, id int64) (*TaskSummary, error)
	
	// Dashboard and analytics
	GetDashboardData(ctx context.Context, timeRange string) (*DashboardData, error)
	GetDayStatistics(ctx context.Context, date time.Time) (*DayStatistics, error)
	GetTodayStatistics(ctx context.Context) (*DayStatistics, error)
	
	// Aggregation operations
	AggregateTaskData(entries []*domain.TimeEntry) map[int64]*TaskActivity
	CalculateTotalDuration(entries []*domain.TimeEntry) time.Duration
	FormatStatistics(stats *ActivityAnalysis) *DayStatistics
}

// ServiceContainer manages all services and their dependencies
type ServiceContainer struct {
	TimeService      TimeService
	TaskService      TaskService
	SearchService    SearchService
	ReportingService ReportingService
}