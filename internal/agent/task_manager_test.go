package agent

import (
	"testing"

	"github.com/onedayallday33-a11y/openshannon-go/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestTaskManager(t *testing.T) {
	tm := GetTaskManager()

	t.Run("Create Task", func(t *testing.T) {
		task := tm.CreateTask("Test Task", "Do something")
		assert.Equal(t, "Test Task", task.Subject)
		assert.Equal(t, types.StatusPending, task.Status)
	})

	t.Run("Get and List Tasks", func(t *testing.T) {
		tasks := tm.ListTasks()
		assert.GreaterOrEqual(t, len(tasks), 1)
		
		task, ok := tm.GetTask(tasks[0].ID)
		assert.True(t, ok)
		assert.NotNil(t, task)
	})

	t.Run("Update Status", func(t *testing.T) {
		tasks := tm.ListTasks()
		id := tasks[0].ID
		err := tm.UpdateTaskStatus(id, types.StatusCompleted)
		assert.NoError(t, err)

		task, _ := tm.GetTask(id)
		assert.Equal(t, types.StatusCompleted, task.Status)
	})

	t.Run("Delete Task", func(t *testing.T) {
		task := tm.CreateTask("Delete Me", "Bye")
		id := task.ID
		tm.DeleteTask(id)

		_, ok := tm.GetTask(id)
		assert.False(t, ok)
	})
}
