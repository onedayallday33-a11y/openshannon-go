package commands

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/onedayallday33-a11y/openshannon-go/internal/agent"
)

type DiffCommand struct{}

func (c *DiffCommand) Name() string        { return "diff" }
func (c *DiffCommand) Description() string { return "Show git diff or changes" }

func (c *DiffCommand) Execute(ctx context.Context, a *agent.Agent, args []string) (agent.DirectCommandResult, error) {
	gitArgs := []string{"diff"}
	if len(args) > 0 {
		gitArgs = append(gitArgs, args...)
	}

	cmd := exec.Command("git", gitArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(out), "Not a git repository") {
			return agent.DirectCommandResult{
				DirectOutput: "Not a git repository (or any of the parent directories): .git",
				IsHandled:    true,
			}, nil
		}
		return agent.DirectCommandResult{
			DirectOutput: fmt.Sprintf("Error running git diff: %v\n%s", err, string(out)),
			IsHandled:    true,
		}, nil
	}

	diffText := string(out)
	if diffText == "" {
		return agent.DirectCommandResult{
			DirectOutput: "No changes found.",
			IsHandled:    true,
		}, nil
	}

	// Limit output size for TUI
	lines := strings.Split(diffText, "\n")
	if len(lines) > 500 {
		diffText = strings.Join(lines[:500], "\n") + "\n\n... (output truncated)"
	}

	return agent.DirectCommandResult{
		DirectOutput: diffText,
		IsHandled:    true,
	}, nil
}
