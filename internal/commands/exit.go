package commands

import (
	"context"
	"os"

	"github.com/onedayallday33-a11y/openshannon-go/internal/agent"
)

type ExitCommand struct{}

func (e *ExitCommand) Name() string {
	return "exit"
}

func (e *ExitCommand) Description() string {
	return "Close the application"
}

func (e *ExitCommand) Execute(ctx context.Context, a *agent.Agent, args []string) (agent.DirectCommandResult, error) {
	os.Exit(0)
	return agent.DirectCommandResult{IsHandled: true}, nil
}
