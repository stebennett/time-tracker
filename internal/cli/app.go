package cli

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"time-tracker/internal/repository/sqlite"
)

// timeNow is a variable that can be replaced in tests
var timeNow = time.Now

// App represents the main CLI application
type App struct {
	repo sqlite.Repository
}

// GetDatabasePath returns the path to the SQLite database file
func GetDatabasePath() (string, error) {
	// Check for TT_DB environment variable
	if dbPath := os.Getenv("TT_DB"); dbPath != "" {
		return dbPath, nil
	}

	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Create .tt directory if it doesn't exist
	ttDir := filepath.Join(homeDir, ".tt")
	if err := os.MkdirAll(ttDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create .tt directory: %w", err)
	}

	// Return path to tt.db in .tt directory
	return filepath.Join(ttDir, "tt.db"), nil
}

// NewApp creates a new CLI application instance with dependency injection
func NewApp(repo sqlite.Repository) *App {
	return &App{
		repo: repo,
	}
}

// NewAppWithDefaultRepository creates a new CLI application instance with the default SQLite repository
// This maintains backward compatibility and is used for production
func NewAppWithDefaultRepository() (*App, error) {
	// Get database path
	dbPath, err := GetDatabasePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get database path: %w", err)
	}

	// Initialize SQLite repository
	repo, err := sqlite.New(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return &App{
		repo: repo,
	}, nil
}

// Run executes the CLI application with the given arguments
func (a *App) Run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: tt start \"your text here\" or tt stop or tt list [time] [text] or tt current or tt output format=csv or tt summary [time] [text] or tt resume or tt delete")
	}

	// Handle different commands
	switch args[0] {
	case "start":
		if len(args) < 2 {
			return fmt.Errorf("usage: tt start \"your text here\"")
		}
		text := strings.Join(args[1:], " ")
		return a.createNewTask(text)
	case "stop":
		if len(args) == 1 {
			return a.stopRunningTasks()
		}
		// If there are additional arguments, treat as an error
		return fmt.Errorf("usage: tt stop")
	case "list":
		return a.listTasks(args[1:])
	case "current":
		return a.showCurrentTask()
	case "output":
		return a.outputTasks(args[1:])
	case "resume":
		return a.resumeTask(args[1:])
	case "summary":
		return a.summaryTask(args[1:])
	case "delete":
		return a.deleteTask(args[1:])
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

// parseTimeShorthand parses time shorthand like "30m", "2h", "1d", etc.
func parseTimeShorthand(shorthand string) (time.Duration, error) {
	re := regexp.MustCompile(`^(\d+)(m|h|d|w|mo|y)$`)
	matches := re.FindStringSubmatch(shorthand)
	if matches == nil {
		return 0, fmt.Errorf("invalid time format: %s", shorthand)
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("invalid number in time format: %s", shorthand)
	}

	unit := matches[2]
	var duration time.Duration

	switch unit {
	case "m":
		duration = time.Duration(value) * time.Minute
	case "h":
		duration = time.Duration(value) * time.Hour
	case "d":
		duration = time.Duration(value) * 24 * time.Hour
	case "w":
		duration = time.Duration(value) * 7 * 24 * time.Hour
	case "mo":
		duration = time.Duration(value) * 30 * 24 * time.Hour
	case "y":
		duration = time.Duration(value) * 365 * 24 * time.Hour
	default:
		return 0, fmt.Errorf("invalid time unit: %s", unit)
	}

	return duration, nil
}

// listTasks handles the list command with various options
func (a *App) listTasks(args []string) error {
	opts := sqlite.SearchOptions{}

	// If no arguments, list all tasks
	if len(args) == 0 {
		entries, err := a.repo.ListTimeEntries()
		if err != nil {
			return fmt.Errorf("failed to list tasks: %w", err)
		}
		return a.printEntries(entries)
	}

	// Check if first argument is a time shorthand
	if duration, err := parseTimeShorthand(args[0]); err == nil {
		// Time shorthand found, set time range
		now := timeNow()
		startTime := now.Add(-duration)
		opts.StartTime = &startTime
		opts.EndTime = &now

		// If there are more arguments, use them as search text
		if len(args) > 1 {
			text := strings.Join(args[1:], " ")
			opts.TaskName = &text
		}
	} else {
		// No time shorthand, treat all arguments as search text
		text := strings.Join(args, " ")
		opts.TaskName = &text
	}

	// Search for entries
	entries, err := a.repo.SearchTimeEntries(opts)
	if err != nil {
		return fmt.Errorf("failed to search tasks: %w", err)
	}

	return a.printEntries(entries)
}

// printEntries prints one line per task in the format:
// startTime - endTime (duration): taskName
// Where startTime and endTime are from the last time entry for the task, and endTime is 'running' if the entry is running.
func (a *App) printEntries(entries []*sqlite.TimeEntry) error {
	if len(entries) == 0 {
		fmt.Println("No tasks found")
		return nil
	}

	// Group entries by TaskID and find the last entry for each task
	type lastEntryInfo struct {
		TaskName  string
		StartTime time.Time
		EndTime   *time.Time
		IsRunning bool
	}
	lastEntryMap := make(map[int64]*lastEntryInfo)
	for _, entry := range entries {
		task, err := a.repo.GetTask(entry.TaskID)
		if err != nil {
			return fmt.Errorf("failed to get task for entry %d: %w", entry.ID, err)
		}
		info, ok := lastEntryMap[entry.TaskID]
		if !ok || entry.StartTime.After(info.StartTime) {
			lastEntryMap[entry.TaskID] = &lastEntryInfo{
				TaskName:  task.TaskName,
				StartTime: entry.StartTime,
				EndTime:   entry.EndTime,
				IsRunning: entry.EndTime == nil,
			}
		}
	}

	// Collect and sort: non-running tasks by StartTime ascending, running tasks last
	var runningInfos, finishedInfos []*lastEntryInfo
	for _, info := range lastEntryMap {
		if info.IsRunning {
			runningInfos = append(runningInfos, info)
		} else {
			finishedInfos = append(finishedInfos, info)
		}
	}
	sort.Slice(finishedInfos, func(i, j int) bool {
		return finishedInfos[i].StartTime.Before(finishedInfos[j].StartTime)
	})
	sort.Slice(runningInfos, func(i, j int) bool {
		return runningInfos[i].StartTime.Before(runningInfos[j].StartTime)
	})
	infos := append(finishedInfos, runningInfos...)
	for _, info := range infos {
		startStr := info.StartTime.Format("2006-01-02 15:04:05")
		var endStr string
		var duration time.Duration
		if info.IsRunning {
			endStr = "running"
			duration = timeNow().Sub(info.StartTime)
		} else if info.EndTime != nil {
			endStr = info.EndTime.Format("2006-01-02 15:04:05")
			duration = info.EndTime.Sub(info.StartTime)
		}
		hours := int(duration.Hours())
		minutes := int(duration.Minutes()) % 60
		fmt.Printf("%s - %s (%dh %dm): %s\n", startStr, endStr, hours, minutes, info.TaskName)
	}

	return nil
}

// stopRunningTasks marks all running tasks as complete
func (a *App) stopRunningTasks() error {
	// Search for tasks with no end time
	opts := sqlite.SearchOptions{}
	entries, err := a.repo.SearchTimeEntries(opts)
	if err != nil {
		return fmt.Errorf("failed to search for running tasks: %w", err)
	}

	now := timeNow()
	for _, entry := range entries {
		if entry.EndTime == nil {
			entry.EndTime = &now
			if err := a.repo.UpdateTimeEntry(entry); err != nil {
				return fmt.Errorf("failed to update task %d: %w", entry.ID, err)
			}
		}
	}

	fmt.Println("All running tasks have been stopped")
	return nil
}

// createNewTask creates a new task
func (a *App) createNewTask(taskName string) error {
	// First, stop any running tasks
	if err := a.stopRunningTasks(); err != nil {
		return err
	}

	// Always create a new task
	task := &sqlite.Task{TaskName: taskName}
	if err := a.repo.CreateTask(task); err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	// Create new time entry
	now := timeNow()
	entry := &sqlite.TimeEntry{
		StartTime: now,
		TaskID:    task.ID,
	}

	if err := a.repo.CreateTimeEntry(entry); err != nil {
		return fmt.Errorf("failed to create new task: %w", err)
	}

	fmt.Printf("Started new task: %s\n", taskName)
	return nil
}

// showCurrentTask displays the currently running task
func (a *App) showCurrentTask() error {
	// Search for tasks with no end time
	opts := sqlite.SearchOptions{}
	entries, err := a.repo.SearchTimeEntries(opts)
	if err != nil {
		return fmt.Errorf("failed to search for running tasks: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("No task is currently running")
		return nil
	}

	// Get the most recent running task
	entry := entries[0]
	duration := timeNow().Sub(entry.StartTime)
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60

	task, err := a.repo.GetTask(entry.TaskID)
	if err != nil {
		return fmt.Errorf("failed to get task for entry %d: %w", entry.ID, err)
	}

	fmt.Printf("Current task: %s (running for %dh %dm)\n",
		task.TaskName, hours, minutes)
	return nil
}

// outputTasks outputs tasks in the specified format
func (a *App) outputTasks(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: tt output format=csv")
	}

	// Parse format option
	format := args[0]
	if !strings.HasPrefix(format, "format=") {
		return fmt.Errorf("invalid format option: %s", format)
	}

	format = strings.TrimPrefix(format, "format=")
	switch format {
	case "csv":
		return a.outputCSV()
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// outputCSV outputs all tasks in CSV format
func (a *App) outputCSV() error {
	// Get all tasks
	entries, err := a.repo.ListTimeEntries()
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	// Create CSV writer
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write header
	header := []string{"ID", "Start Time", "End Time", "Duration (hours)", "Task Name"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write entries
	for _, entry := range entries {
		// Format start time
		startTime := entry.StartTime.Format(time.RFC3339)

		// Format end time
		var endTime string
		var duration float64
		if entry.EndTime != nil {
			endTime = entry.EndTime.Format(time.RFC3339)
			duration = entry.EndTime.Sub(entry.StartTime).Hours()
		}

		task, err := a.repo.GetTask(entry.TaskID)
		if err != nil {
			return fmt.Errorf("failed to get task for entry %d: %w", entry.ID, err)
		}

		// Write row
		row := []string{
			strconv.FormatInt(entry.ID, 10),
			startTime,
			endTime,
			fmt.Sprintf("%.2f", duration),
			task.TaskName,
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// resumeTask implements the resume scenario
func (a *App) resumeTask(args []string) error {
	// Determine time range (default: today)
	var startTime time.Time
	now := timeNow()
	if len(args) > 0 {
		dur, err := parseTimeShorthand(args[0])
		if err != nil {
			return fmt.Errorf("invalid time shorthand: %v", err)
		}
		startTime = now.Add(-dur)
	} else {
		y, m, d := now.Date()
		startTime = time.Date(y, m, d, 0, 0, 0, 0, now.Location())
	}

	// Find all time entries in the period, most recent first
	opts := sqlite.SearchOptions{StartTime: &startTime, EndTime: &now}
	entries, err := a.repo.SearchTimeEntries(opts)
	if err != nil {
		return fmt.Errorf("failed to search time entries: %w", err)
	}
	if len(entries) == 0 {
		fmt.Println("No tasks found in the selected period.")
		return nil
	}

	// Group by task, show most recent entry for each task
	taskMap := make(map[int64]*sqlite.TimeEntry)
	for i := len(entries) - 1; i >= 0; i-- { // reverse for most recent
		entry := entries[i]
		if _, ok := taskMap[entry.TaskID]; !ok {
			taskMap[entry.TaskID] = entry
		}
	}

	// Build a slice for display
	var taskIDs []int64
	for id := range taskMap {
		taskIDs = append(taskIDs, id)
	}
	// Sort by most recent start time
	sort.Slice(taskIDs, func(i, j int) bool {
		return taskMap[taskIDs[i]].StartTime.After(taskMap[taskIDs[j]].StartTime)
	})

	fmt.Println("Select a task to resume:")
	for i, id := range taskIDs {
		task, _ := a.repo.GetTask(id)
		last := taskMap[id].StartTime.Format("2006-01-02 15:04:05")
		fmt.Printf("%d. %s (last worked: %s)\n", i+1, task.TaskName, last)
	}
	fmt.Print("Enter number to resume, or 'q' to quit: ")

	// Read user input
	var input string
	fmt.Fscanln(os.Stdin, &input)
	if input == "q" || input == "Q" {
		fmt.Println("Resume cancelled.")
		return nil
	}
	idx, err := strconv.Atoi(input)
	if err != nil || idx < 1 || idx > len(taskIDs) {
		return fmt.Errorf("invalid selection")
	}
	selectedTaskID := taskIDs[idx-1]

	// Stop any running tasks
	if err := a.stopRunningTasks(); err != nil {
		return err
	}

	// Create a new time entry for the selected task
	entry := &sqlite.TimeEntry{
		StartTime: timeNow(),
		TaskID:    selectedTaskID,
	}
	if err := a.repo.CreateTimeEntry(entry); err != nil {
		return fmt.Errorf("failed to resume task: %w", err)
	}
	task, _ := a.repo.GetTask(selectedTaskID)
	fmt.Printf("Resumed task: %s\n", task.TaskName)
	return nil
}

// summaryTask implements the summary command
func (a *App) summaryTask(args []string) error {
	// Determine time range and search text
	var startTime *time.Time
	var searchText string
	now := timeNow()

	if len(args) > 0 {
		// Check if first argument is a time shorthand
		if duration, err := parseTimeShorthand(args[0]); err == nil {
			// Time shorthand found, set time range
			start := now.Add(-duration)
			startTime = &start

			// If there are more arguments, use them as search text
			if len(args) > 1 {
				searchText = strings.Join(args[1:], " ")
			}
		} else {
			// No time shorthand, treat all arguments as search text
			searchText = strings.Join(args, " ")
		}
	}

	// Find tasks that match the criteria
	var matchingTaskIDs []int64

	if startTime != nil {
		// If time filter is specified, find tasks that have ANY entries in the time window
		timeFilterOpts := sqlite.SearchOptions{
			StartTime: startTime,
			EndTime:   &now,
		}
		if searchText != "" {
			timeFilterOpts.TaskName = &searchText
		}

		timeFilterEntries, err := a.repo.SearchTimeEntries(timeFilterOpts)
		if err != nil {
			return fmt.Errorf("failed to search time entries: %w", err)
		}

		// Get unique task IDs from entries in the time window
		taskIDSet := make(map[int64]bool)
		for _, entry := range timeFilterEntries {
			taskIDSet[entry.TaskID] = true
		}

		// Convert to slice
		for taskID := range taskIDSet {
			matchingTaskIDs = append(matchingTaskIDs, taskID)
		}
	} else if searchText != "" {
		// Only text filter, find tasks by name
		textFilterOpts := sqlite.SearchOptions{
			TaskName: &searchText,
		}

		textFilterEntries, err := a.repo.SearchTimeEntries(textFilterOpts)
		if err != nil {
			return fmt.Errorf("failed to search time entries: %w", err)
		}

		// Get unique task IDs
		taskIDSet := make(map[int64]bool)
		for _, entry := range textFilterEntries {
			taskIDSet[entry.TaskID] = true
		}

		// Convert to slice
		for taskID := range taskIDSet {
			matchingTaskIDs = append(matchingTaskIDs, taskID)
		}
	} else {
		// No filters, get all tasks
		allEntries, err := a.repo.ListTimeEntries()
		if err != nil {
			return fmt.Errorf("failed to list time entries: %w", err)
		}

		// Get unique task IDs
		taskIDSet := make(map[int64]bool)
		for _, entry := range allEntries {
			taskIDSet[entry.TaskID] = true
		}

		// Convert to slice
		for taskID := range taskIDSet {
			matchingTaskIDs = append(matchingTaskIDs, taskID)
		}
	}

	if len(matchingTaskIDs) == 0 {
		fmt.Println("No tasks found matching the criteria.")
		return nil
	}

	// Sort task IDs by task name for consistent ordering
	sort.Slice(matchingTaskIDs, func(i, j int) bool {
		taskI, _ := a.repo.GetTask(matchingTaskIDs[i])
		taskJ, _ := a.repo.GetTask(matchingTaskIDs[j])
		return taskI.TaskName < taskJ.TaskName
	})

	// If only one task, show its summary directly
	if len(matchingTaskIDs) == 1 {
		return a.showTaskSummary(matchingTaskIDs[0])
	}

	// Multiple tasks found, let user choose
	fmt.Println("Select a task to summarize:")
	for i, taskID := range matchingTaskIDs {
		task, _ := a.repo.GetTask(taskID)
		fmt.Printf("%d. %s\n", i+1, task.TaskName)
	}
	fmt.Print("Enter number to summarize, or 'q' to quit: ")

	// Read user input
	var input string
	fmt.Fscanln(os.Stdin, &input)
	if input == "q" || input == "Q" {
		fmt.Println("Summary cancelled.")
		return nil
	}
	idx, err := strconv.Atoi(input)
	if err != nil || idx < 1 || idx > len(matchingTaskIDs) {
		return fmt.Errorf("invalid selection")
	}
	selectedTaskID := matchingTaskIDs[idx-1]

	return a.showTaskSummary(selectedTaskID)
}

// showTaskSummary displays a detailed summary for a specific task
func (a *App) showTaskSummary(taskID int64) error {
	// Get task details
	task, err := a.repo.GetTask(taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Get all time entries for this task
	opts := sqlite.SearchOptions{TaskID: &taskID}
	entries, err := a.repo.SearchTimeEntries(opts)
	if err != nil {
		return fmt.Errorf("failed to get time entries: %w", err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("no time entries found for task")
	}

	// Sort entries by start time
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].StartTime.Before(entries[j].StartTime)
	})

	// Calculate summary statistics
	var earliestStart time.Time
	var latestEnd time.Time
	var totalDuration time.Duration
	var runningSessions int

	earliestStart = entries[0].StartTime
	latestEnd = entries[0].StartTime // Initialize with first start time

	for _, entry := range entries {
		// Update earliest start
		if entry.StartTime.Before(earliestStart) {
			earliestStart = entry.StartTime
		}

		// Calculate duration and update latest end
		if entry.EndTime != nil {
			duration := entry.EndTime.Sub(entry.StartTime)
			totalDuration += duration

			if entry.EndTime.After(latestEnd) {
				latestEnd = *entry.EndTime
			}
		} else {
			// Running session
			runningSessions++
			currentDuration := timeNow().Sub(entry.StartTime)
			totalDuration += currentDuration

			// For running sessions, use current time as latest
			if timeNow().After(latestEnd) {
				latestEnd = timeNow()
			}
		}
	}

	// Print summary header
	fmt.Printf("\nSummary for: %s\n", task.TaskName)
	fmt.Println(strings.Repeat("=", len(task.TaskName)+12))

	// Print table header
	fmt.Printf("%-20s %-20s %-15s %s\n", "Start Time", "End Time", "Duration", "Status")
	fmt.Println(strings.Repeat("-", 75))

	// Print each session
	for _, entry := range entries {
		startStr := entry.StartTime.Format("2006-01-02 15:04:05")
		var endStr, durationStr, status string

		if entry.EndTime != nil {
			endStr = entry.EndTime.Format("2006-01-02 15:04:05")
			duration := entry.EndTime.Sub(entry.StartTime)
			hours := int(duration.Hours())
			minutes := int(duration.Minutes()) % 60
			durationStr = fmt.Sprintf("%dh %dm", hours, minutes)
			status = "Completed"
		} else {
			endStr = "running"
			duration := timeNow().Sub(entry.StartTime)
			hours := int(duration.Hours())
			minutes := int(duration.Minutes()) % 60
			durationStr = fmt.Sprintf("%dh %dm", hours, minutes)
			status = "Running"
		}

		fmt.Printf("%-20s %-20s %-15s %s\n", startStr, endStr, durationStr, status)
	}

	// Print summary footer
	fmt.Println(strings.Repeat("-", 75))

	// Format total duration
	totalHours := int(totalDuration.Hours())
	totalMinutes := int(totalDuration.Minutes()) % 60

	// Format time range
	earliestStr := earliestStart.Format("2006-01-02 15:04:05")
	latestStr := latestEnd.Format("2006-01-02 15:04:05")

	fmt.Printf("Total Sessions: %d", len(entries))
	if runningSessions > 0 {
		fmt.Printf(" (%d running)", runningSessions)
	}
	fmt.Printf("\n")
	fmt.Printf("Time Range: %s to %s\n", earliestStr, latestStr)
	fmt.Printf("Total Time: %dh %dm\n", totalHours, totalMinutes)

	return nil
}

// deleteTask implements the delete command
func (a *App) deleteTask(args []string) error {
	// Determine time range (default: last 24h or user-supplied duration)
	now := timeNow()
	var startTime time.Time
	var filterText string

	if len(args) > 0 {
		if dur, err := parseTimeShorthand(args[0]); err == nil {
			startTime = now.Add(-dur)
			if len(args) > 1 {
				filterText = strings.Join(args[1:], " ")
			}
		} else {
			// Not a valid duration, treat as text filter
			filterText = strings.Join(args, " ")
			startTime = now.Add(-24 * time.Hour)
		}
	} else {
		startTime = now.Add(-24 * time.Hour)
	}

	// Build search options like listTasks
	opts := sqlite.SearchOptions{StartTime: &startTime, EndTime: &now}
	if filterText != "" {
		opts.TaskName = &filterText
	}

	entries, err := a.repo.SearchTimeEntries(opts)
	if err != nil {
		return fmt.Errorf("failed to search time entries: %w", err)
	}
	if len(entries) == 0 {
		fmt.Println("No tasks found to delete.")
		return nil
	}

	// Group by task, get the last entry for each task
	taskMap := make(map[int64]*sqlite.TimeEntry)
	for _, entry := range entries {
		if existing, ok := taskMap[entry.TaskID]; !ok || entry.StartTime.After(existing.StartTime) {
			taskMap[entry.TaskID] = entry
		}
	}

	// Build a slice for display and sort by last worked time ascending
	taskIDs := make([]int64, 0, len(taskMap))
	for id := range taskMap {
		taskIDs = append(taskIDs, id)
	}
	sort.Slice(taskIDs, func(i, j int) bool {
		return taskMap[taskIDs[i]].StartTime.Before(taskMap[taskIDs[j]].StartTime)
	})

	if len(taskIDs) == 0 {
		fmt.Println("No tasks found to delete.")
		return nil
	}

	fmt.Println("Select a task to delete:")
	for i, id := range taskIDs {
		task, _ := a.repo.GetTask(id)
		last := taskMap[id].StartTime.Format("2006-01-02 15:04:05")
		fmt.Printf("%d. %s (last worked: %s)\n", i+1, task.TaskName, last)
	}
	fmt.Print("Enter number to delete, or 'q' to quit: ")

	// Read user input
	var input string
	fmt.Fscanln(os.Stdin, &input)
	if input == "q" || input == "Q" {
		fmt.Println("Delete cancelled.")
		return nil
	}
	idx, err := strconv.Atoi(input)
	if err != nil || idx < 1 || idx > len(taskIDs) {
		return fmt.Errorf("invalid selection")
	}
	selectedTaskID := taskIDs[idx-1]
	task, _ := a.repo.GetTask(selectedTaskID)

	// Delete all time entries for the task
	entryOpts := sqlite.SearchOptions{TaskID: &selectedTaskID}
	allEntries, err := a.repo.SearchTimeEntries(entryOpts)
	if err != nil {
		return fmt.Errorf("failed to get time entries for task: %w", err)
	}
	for _, entry := range allEntries {
		err := a.repo.DeleteTimeEntry(entry.ID)
		if err != nil {
			return fmt.Errorf("failed to delete time entry %d: %w", entry.ID, err)
		}
	}

	// Delete the task itself
	if err := a.repo.DeleteTask(selectedTaskID); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	fmt.Printf("Deleted task: %s\n", task.TaskName)
	return nil
}
