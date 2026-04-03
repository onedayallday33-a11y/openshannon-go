package commands

import (
	"context"

	"github.com/onedayallday33-a11y/openshannon-go/internal/agent"
)

type ClearCommand struct{}

func (c *ClearCommand) Name() string {
	return "clear"
}

func (c *ClearCommand) Description() string {
	return "Clear the current conversation history"
}

func (c *ClearCommand) Execute(ctx context.Context, a *agent.Agent, args []string) (agent.DirectCommandResult, error) {
	a.ResetHistory()
	return agent.DirectCommandResult{
		DirectOutput: "Conversation history cleared.",
		IsHandled:    true,
	}, nil
}
