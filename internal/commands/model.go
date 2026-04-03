package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/onedayallday33-a11y/openshannon-go/internal/agent"
)

type ModelCommand struct{}

func (m *ModelCommand) Name() string {
	return "model"
}

func (m *ModelCommand) Description() string {
	return "Show or change the current LLM model"
}

func (m *ModelCommand) Execute(ctx context.Context, a *agent.Agent, args []string) (agent.DirectCommandResult, error) {
	if len(args) == 0 {
		return agent.DirectCommandResult{
			DirectOutput: fmt.Sprintf("Current model: %s", a.Config.Model),
			IsHandled:    true,
		}, nil
	}

	newModel := strings.TrimSpace(args[0])
	oldModel := a.Config.Model
	a.Config.Model = newModel

	return agent.DirectCommandResult{
		DirectOutput: fmt.Sprintf("Model changed from '%s' to '%s'.", oldModel, newModel),
		IsHandled:    true,
	}, nil
}
