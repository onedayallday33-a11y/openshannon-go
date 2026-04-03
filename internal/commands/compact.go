package commands

import (
	"context"
	"fmt"

	"github.com/onedayallday33-a11y/openshannon-go/internal/agent"
)

type CompactCommand struct{}

func (c *CompactCommand) Name() string { return "compact" }
func (c *CompactCommand) Description() string { return "Truncate conversation history" }

func (c *CompactCommand) Execute(ctx context.Context, a *agent.Agent, args []string) (agent.DirectCommandResult, error) {
	oldLen := len(a.History)
	if oldLen <= 2 {
		return agent.DirectCommandResult{
			DirectOutput: "History is already compact.",
			IsHandled:    true,
		}, nil
	}

	// Keep only last 4 messages (2 turns)
	keep := 4
	if oldLen < keep {
		keep = oldLen
	}
	a.History = a.History[oldLen-keep:]

	return agent.DirectCommandResult{
		DirectOutput: fmt.Sprintf("History compacted from %d to %d messages.", oldLen, len(a.History)),
		IsHandled:    true,
	}, nil
}
