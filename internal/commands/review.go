package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/onedayallday33-a11y/openshannon-go/internal/agent"
)

type ReviewCommand struct{}

func (c *ReviewCommand) Name() string { return "review" }
func (c *ReviewCommand) Description() string { return "Review a file or diff" }

func (c *ReviewCommand) Execute(ctx context.Context, a *agent.Agent, args []string) (agent.DirectCommandResult, error) {
	if len(args) == 0 {
		return agent.DirectCommandResult{
			DirectOutput: "Usage: /review <file_path>",
			IsHandled:    true,
		}, nil
	}

	path := args[0]
	content, err := os.ReadFile(path)
	if err != nil {
		return agent.DirectCommandResult{}, fmt.Errorf("failed to read file for review: %v", err)
	}

	prompt := fmt.Sprintf("Please review the following file for potential bugs, security issues, and improvements. Provide constructive feedback:\n\nFile: %s\n```\n%s\n```", path, string(content))

	return agent.DirectCommandResult{
		PromptText: prompt,
		IsHandled:  true,
	}, nil
}
