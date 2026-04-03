package agent

import (
	"context"
	"os"
	"testing"

	"github.com/onedayallday33-a11y/openshannon-go/internal/toolapi"
	"github.com/onedayallday33-a11y/openshannon-go/internal/types"
	"github.com/stretchr/testify/assert"
)

// MockTool for testing
type MockTool struct{}

func (m *MockTool) Name() string        { return "MockTool" }
func (m *MockTool) Description() string { return "A mock tool" }
func (m *MockTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{"type": "object"}
}
func (m *MockTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	return "mock result", nil
}

func TestAgent_Loop(t *testing.T) {
	agent := NewAgent("test-session", types.AgentConfig{
		Model:    "test-model",
		System:   "System prompt",
		MaxTurns: 5,
		Tools:    []toolapi.Tool{&MockTool{}},
	})

	ctx := context.Background()

	t.Run("Initialize and Add Message", func(t *testing.T) {
		agent.AddMessage(types.Message{
			Role: types.RoleUser,
			Content: []types.ContentBlock{
				{Type: "text", Text: "Hello"},
			},
		})
		
		assert.Equal(t, 1, len(agent.History))
		assert.Equal(t, types.RoleUser, agent.History[0].Role)
	})

	t.Run("Handle Tool Use Logic", func(t *testing.T) {
		inputs := map[string]interface{}{"arg": "val"}
		
		err := os.WriteFile("test.txt", []byte("hello world"), 0644)
		assert.NoError(t, err)
		defer os.Remove("test.txt")

		result, err := agent.HandleToolUse(ctx, "call_1", "MockTool", inputs)
		assert.NoError(t, err)
		assert.Equal(t, "mock result", result)
	})
}
