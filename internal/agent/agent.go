package agent

import (
	"github.com/onedayallday33-a11y/openshannon-go/internal/history"
	"github.com/onedayallday33-a11y/openshannon-go/internal/toolapi"
	"github.com/onedayallday33-a11y/openshannon-go/internal/types"
)

// Agent represents the autonomous assistant
type Agent struct {
	ID        string
	Config    types.AgentConfig
	History     []types.Message
	Tools       map[string]toolapi.Tool
	FileHistory *history.FileHistory
	OnTurnEnd   func(*Agent) // Hook for persistence or UI updates
}

// NewAgent creates a new agent instance
func NewAgent(id string, config types.AgentConfig) *Agent {
	toolMap := make(map[string]toolapi.Tool)
	for _, t := range config.Tools {
		toolMap[t.Name()] = t
	}

	return &Agent{
		ID:          id,
		Config:      config,
		History:     []types.Message{},
		Tools:       toolMap,
		FileHistory: history.NewFileHistory(id),
	}
}

// NewAgentWithHistory restores an agent from a previous session
func NewAgentWithHistory(id string, config types.AgentConfig, history []types.Message) *Agent {
	agent := NewAgent(id, config)
	agent.History = history
	return agent
}

// AddMessage adds a message to history
func (a *Agent) AddMessage(m types.Message) {
	a.History = append(a.History, m)
}

// SetSystemPrompt updates the system instructions
func (a *Agent) SetSystemPrompt(s string) {
	a.Config.System = s
}

// ResetHistory clears all conversation history
func (a *Agent) ResetHistory() {
	a.History = []types.Message{}
}
