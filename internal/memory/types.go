package memory

import (
	"time"

	"github.com/onedayallday33-a11y/openshannon-go/internal/agent"
	"github.com/onedayallday33-a11y/openshannon-go/internal/types"
)

// SessionState represents the full state of an agent session at a point in time
type SessionState struct {
	ID        string          `json:"id"`
	Messages  []types.Message `json:"messages"`
	Tasks     []*types.Task   `json:"tasks"`
	CWD       string          `json:"cwd"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// Snapshot creates a session state from a running agent and task manager
func Snapshot(id string, agentMessages []types.Message, taskManager *agent.TaskManager) SessionState {
	tasks := taskManager.ListTasks()
	
	return SessionState{
		ID:        id,
		Messages:  agentMessages,
		Tasks:     tasks,
		UpdatedAt: time.Now(),
	}
}
