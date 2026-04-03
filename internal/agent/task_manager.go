package agent

import (
	"fmt"
	"sync"

	"github.com/onedayallday33-a11y/openshannon-go/internal/types"
)

// TaskManager handles the task list in memory
type TaskManager struct {
	mu     sync.RWMutex
	tasks  map[int]*types.Task
	nextID int
}

var (
	DefaultTaskManager *TaskManager
	once               sync.Once
)

// GetTaskManager returns the singleton instance
func GetTaskManager() *TaskManager {
	once.Do(func() {
		DefaultTaskManager = &TaskManager{
			tasks:  make(map[int]*types.Task),
			nextID: 1,
		}
	})
	return DefaultTaskManager
}

// CreateTask adds a new task to the list
func (tm *TaskManager) CreateTask(subject, description string) *types.Task {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	task := &types.Task{
		ID:          tm.nextID,
		Subject:     subject,
		Description: description,
		Status:      types.StatusPending,
	}

	tm.tasks[task.ID] = task
	tm.nextID++
	return task
}

// GetTask returns a task by ID
func (tm *TaskManager) GetTask(id int) (*types.Task, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	task, ok := tm.tasks[id]
	return task, ok
}

// ListTasks returns all tasks
func (tm *TaskManager) ListTasks() []*types.Task {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	list := make([]*types.Task, 0, len(tm.tasks))
	for _, t := range tm.tasks {
		list = append(list, t)
	}
	return list
}

// UpdateTaskStatus changes the status of a task
func (tm *TaskManager) UpdateTaskStatus(id int, status types.TaskStatus) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	task, ok := tm.tasks[id]
	if !ok {
		return fmt.Errorf("task %d not found", id)
	}

	task.Status = status
	return nil
}

// DeleteTask removes a task from the list
func (tm *TaskManager) DeleteTask(id int) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	delete(tm.tasks, id)
}
