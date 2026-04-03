package commands

import (
	"context"
	"fmt"

	"github.com/onedayallday33-a11y/openshannon-go/internal/agent"
)

type StatsCommand struct{}

func (c *StatsCommand) Name() string        { return "stats" }
func (c *StatsCommand) Description() string { return "Show session token usage" }

func (c *StatsCommand) Execute(ctx context.Context, a *agent.Agent, args []string) (agent.DirectCommandResult, error) {
	var totalInput, totalOutput int
	var turns int

	for _, m := range a.History {
		if m.Usage != nil {
			totalInput += m.Usage.InputTokens
			totalOutput += m.Usage.OutputTokens
			turns++
		}
	}

	output := "📊 Session Statistics:\n\n"
	output += fmt.Sprintf("Total Turns: %d\n", turns)
	output += fmt.Sprintf("Input Tokens:  %d\n", totalInput)
	output += fmt.Sprintf("Output Tokens: %d\n", totalOutput)
	output += fmt.Sprintf("Total Tokens:  %d\n", totalInput+totalOutput)

	if totalInput+totalOutput == 0 {
		output = "No token usage recorded for this session yet."
	}

	return agent.DirectCommandResult{
		DirectOutput: output,
		IsHandled:    true,
	}, nil
}
