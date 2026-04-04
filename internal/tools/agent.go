package tools

import (
	"context"
	"fmt"

	"github.com/onedayallday33-a11y/openshannon-go/internal/agent"
	"github.com/onedayallday33-a11y/openshannon-go/internal/toolapi"
	"github.com/onedayallday33-a11y/openshannon-go/internal/types"
)

// AgentTool implements the Tool interface to spawn sub-agents
type AgentTool struct {
	// Model to use for the sub-agent
	Model string
	// BaseTools to provide to the sub-agent
	BaseTools []toolapi.Tool
}

// Name of the tool
func (t *AgentTool) Name() string {
	return "Agent"
}

// Description of the tool
func (t *AgentTool) Description() string {
	return "Start a sub-agent for a specific task"
}

// InputSchema for the tool
func (t *AgentTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Task description for the sub-agent",
			},
		},
		"required": []string{"message"},
	}
}

// Execute the sub-agent logic
func (t *AgentTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	message, ok := args["message"].(string)
	if !ok {
		return nil, fmt.Errorf("message is required")
	}

	// 1. Create Sub-Agent with a sub-session ID
	subAgent := agent.NewAgent("sub-agent", types.AgentConfig{
		Model:    t.Model,
		MaxTurns: 10,
		System:   "You are a helpful sub-agent assistant. Complete the objective and return the result as plain text.",
		Tools:    t.BaseTools,
	})

	// 2. Run Sub-Agent loop
	result, err := subAgent.Run(ctx, message, nil)
	if err != nil {
		return nil, fmt.Errorf("sub-agent failed: %v", err)
	}

	return result, nil
}
