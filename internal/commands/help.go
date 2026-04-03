package commands

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/onedayallday33-a11y/openshannon-go/internal/agent"
)

type HelpCommand struct {
	// We'll need a way for Help to see all commands.
	// We can pass them in or use a registry.
}

func (c *HelpCommand) Name() string { return "help" }
func (c *HelpCommand) Description() string { return "Show available commands" }

func (c *HelpCommand) Execute(ctx context.Context, a *agent.Agent, args []string) (agent.DirectCommandResult, error) {
	// Note: In a real app, we'd fetch this from the current dispatcher's registry.
	// For now, let's list the ones we're implementing.
	
	cmds := []struct{ name, desc string }{
		{"help", "Show available commands"},
		{"doctor", "Check environment health"},
		{"compact", "Truncate conversation history"},
		{"review", "Review a file or diff"},
		{"commit", "Generate a git commit message"},
	}

	sort.Slice(cmds, func(i, j int) bool { return cmds[i].name < cmds[j].name })

	var sb strings.Builder
	sb.WriteString("Available commands:\n")
	for _, cmd := range cmds {
		sb.WriteString(fmt.Sprintf("  /%s - %s\n", cmd.name, cmd.desc))
	}

	return agent.DirectCommandResult{
		DirectOutput: sb.String(),
		IsHandled:    true,
	}, nil
}
