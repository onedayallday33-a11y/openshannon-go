package tools

import (
	"context"
	"strconv"

	"github.com/onedayallday33-a11y/openshannon-go/internal/agent"
	"github.com/onedayallday33-a11y/openshannon-go/internal/types"
)

// TaskCreateTool implements Tool
type TaskCreateTool struct{}

func (t *TaskCreateTool) Name() string { return "TaskCreate" }
func (t *TaskCreateTool) Description() string { return "Create a new item in the task list" }
func (t *TaskCreateTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"subject":     map[string]interface{}{"type": "string", "description": "Short title"},
			"description": map[string]interface{}{"type": "string", "description": "What to do"},
		},
		"required": []string{"subject", "description"},
	}
}
func (t *TaskCreateTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	sub, _ := args["subject"].(string)
	desc, _ := args["description"].(string)
	task := agent.GetTaskManager().CreateTask(sub, desc)
	return map[string]interface{}{"id": task.ID, "status": "created"}, nil
}

// TaskListTool implements Tool
type TaskListTool struct{}

func (t *TaskListTool) Name() string { return "TaskList" }
func (t *TaskListTool) Description() string { return "List all tasks" }
func (t *TaskListTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{"type": "object"}
}
func (t *TaskListTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	tasks := agent.GetTaskManager().ListTasks()
	return tasks, nil
}

// TaskUpdateTool implements Tool
type TaskUpdateTool struct{}

func (t *TaskUpdateTool) Name() string { return "TaskUpdate" }
func (t *TaskUpdateTool) Description() string { return "Update task status (pending, in_progress, completed, blocked)" }
func (t *TaskUpdateTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id":     map[string]interface{}{"type": "integer"},
			"status": map[string]interface{}{"type": "string"},
		},
		"required": []string{"id", "status"},
	}
}
func (t *TaskUpdateTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	var id int
	switch v := args["id"].(type) {
	case float64:
		id = int(v)
	case int:
		id = v
	case string:
		id, _ = strconv.Atoi(v)
	}

	status, _ := args["status"].(string)
	err := agent.GetTaskManager().UpdateTaskStatus(id, types.TaskStatus(status))
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"id": id, "status": "updated"}, nil
}

// TaskDeleteTool implements Tool
type TaskDeleteTool struct{}

func (t *TaskDeleteTool) Name() string { return "TaskDelete" }
func (t *TaskDeleteTool) Description() string { return "Remove a task" }
func (t *TaskDeleteTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{"id": map[string]interface{}{"type": "integer"}},
		"required": []string{"id"},
	}
}
func (t *TaskDeleteTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	var id int
	if v, ok := args["id"].(float64); ok {
		id = int(v)
	} else if v, ok := args["id"].(int); ok {
		id = v
	}
	agent.GetTaskManager().DeleteTask(id)
	return map[string]interface{}{"id": id, "status": "deleted"}, nil
}
