package commands

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/onedayallday33-a11y/openshannon-go/internal/agent"
)

type CommitCommand struct{}

func (c *CommitCommand) Name() string { return "commit" }
func (c *CommitCommand) Description() string { return "Generate a git commit message" }

func (c *CommitCommand) Execute(ctx context.Context, a *agent.Agent, args []string) (agent.DirectCommandResult, error) {
	// Execute git diff --cached
	out, err := exec.Command("git", "diff", "--cached").Output()
	if err != nil {
		return agent.DirectCommandResult{
			DirectOutput: "Failed to run git diff. Ensure you are in a git repo and have staged changes.",
			IsHandled:    true,
		}, nil
	}

	diffText := string(out)
	if diffText == "" {
		return agent.DirectCommandResult{
			DirectOutput: "No staged changes found. Use 'git add' first.",
			IsHandled:    true,
		}, nil
	}

	prompt := fmt.Sprintf("Analyze these staged changes and suggest a concise, professional Git commit message following conventional commits format:\n\n```patch\n%s\n```", diffText)

	return agent.DirectCommandResult{
		PromptText: prompt,
		IsHandled:  true,
	}, nil
}
