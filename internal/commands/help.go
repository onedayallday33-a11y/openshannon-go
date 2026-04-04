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
	registeredCmds := agent.GetDispatcher().GetRegisteredCommands()
	
	// Convert to sortable slice
	type cmdInfo struct{ name, desc string }
	cmds := make([]cmdInfo, 0, len(registeredCmds))
	for _, rc := range registeredCmds {
		cmds = append(cmds, cmdInfo{rc.Name(), rc.Description()})
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
