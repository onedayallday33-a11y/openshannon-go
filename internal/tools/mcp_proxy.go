package tools

import (
	"context"
	"fmt"

	"github.com/onedayallday33-a11y/openshannon-go/internal/mcp"
)

// MCPProxyTool wraps an MCP-provided tool to satisfy our Tool interface
type MCPProxyTool struct {
	Client      *mcp.McpClient
	McpName     string
	DescriptionStr string
	Schema      map[string]interface{}
}

// Name returns the prefixed tool name (e.g., mcp__server__tool)
func (t *MCPProxyTool) Name() string {
	return fmt.Sprintf("mcp__%s__%s", t.Client.Name, t.McpName)
}

// Description of the tool
func (t *MCPProxyTool) Description() string {
	return t.DescriptionStr
}

// InputSchema returns the tool's JSON schema
func (t *MCPProxyTool) InputSchema() map[string]interface{} {
	return t.Schema
}

// Execute forwards the call to the MCP client
func (t *MCPProxyTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	result, err := t.Client.CallTool(ctx, t.McpName, args)
	if err != nil {
		return nil, fmt.Errorf("MCP call failed: %v", err)
	}

	// MCP results are often { "content": [...] }
	return result, nil
}
