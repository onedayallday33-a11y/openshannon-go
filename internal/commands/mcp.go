package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/onedayallday33-a11y/openshannon-go/internal/agent"
)

type MCPCommand struct{}

func (c *MCPCommand) Name() string        { return "mcp" }
func (c *MCPCommand) Description() string { return "Show MCP server and tool status" }

func (c *MCPCommand) Execute(ctx context.Context, a *agent.Agent, args []string) (agent.DirectCommandResult, error) {
	var mcpTools []string
	servers := make(map[string]int)

	for name := range a.Tools {
		if strings.HasPrefix(name, "mcp__") {
			parts := strings.Split(name, "__")
			if len(parts) >= 2 {
				serverName := parts[1]
				servers[serverName]++
			}
			mcpTools = append(mcpTools, name)
		}
	}

	if len(mcpTools) == 0 {
		return agent.DirectCommandResult{
			DirectOutput: "No MCP servers or tools currently registered.",
			IsHandled:    true,
		}, nil
	}

	var sb strings.Builder
	sb.WriteString("🔌 Registered MCP Servers:\n\n")
	for server, count := range servers {
		sb.WriteString(fmt.Sprintf("- %s (%d tools)\n", server, count))
	}

	return agent.DirectCommandResult{
		DirectOutput: sb.String(),
		IsHandled:    true,
	}, nil
}
