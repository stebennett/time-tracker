package domain

// Task represents a task in the domain model.
// This is a pure domain model without database-specific concerns.
type Task struct {
	ID       int64
	TaskName string
}

// NewTask creates a new Task with the given name.
func NewTask(name string) Task {
	return Task{
		TaskName: name,
	}
}

// IsValid checks if the task has valid data.
func (t Task) IsValid() bool {
	return t.TaskName != ""
}

// String returns the task name for display purposes.
func (t Task) String() string {
	return t.TaskName
}