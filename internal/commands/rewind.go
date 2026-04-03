package commands

import (
	"context"
	"fmt"
	"strconv"

	"github.com/onedayallday33-a11y/openshannon-go/internal/agent"
	"github.com/onedayallday33-a11y/openshannon-go/internal/types"
)

type RewindCommand struct{}

func (c *RewindCommand) Name() string        { return "rewind" }
func (c *RewindCommand) Description() string { return "Rewind conversation and/or code" }

func (c *RewindCommand) Execute(ctx context.Context, a *agent.Agent, args []string) (agent.DirectCommandResult, error) {
	if len(a.History) == 0 {
		return agent.DirectCommandResult{
			DirectOutput: "Nothing to rewind.",
			IsHandled:    true,
		}, nil
	}

	numTurns := 1
	if len(args) > 0 {
		if n, err := strconv.Atoi(args[0]); err == nil && n > 0 {
			numTurns = n
		}
	}

	// Basic implementation: truncate last N assistant/user turns
	// A turn usually consists of User prompt + Assistant response + potentially tool results.
	// For simplicity, let's remove messages until we've removed N user prompts.
	
	removedUserPrompts := 0
	lastUserIndex := -1
	for i := len(a.History) - 1; i >= 0; i-- {
		if a.History[i].Role == types.RoleUser {
			removedUserPrompts++
			if removedUserPrompts == numTurns {
				lastUserIndex = i
				break
			}
		}
	}

	if lastUserIndex == -1 {
		a.ResetHistory()
		return agent.DirectCommandResult{
			DirectOutput: "Conversation cleared (rewound past the beginning).",
			IsHandled:    true,
		}, nil
	}

	a.History = a.History[:lastUserIndex]

	return agent.DirectCommandResult{
		DirectOutput: fmt.Sprintf("Rewound %d turn(s).", numTurns),
		IsHandled:    true,
	}, nil
}
