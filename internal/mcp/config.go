package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// LoadConfigs merges MCP configs from user global and project local scopes
func LoadConfigs(cwd string) (map[string]McpServerConfig, error) {
	allServers := make(map[string]McpServerConfig)

	// 1. User Global Config (e.g. ~/.openshannon/mcp_config.json)
	home, _ := os.UserHomeDir()
	globalPath := filepath.Join(home, ".openshannon", "mcp_config.json")
	if servers, err := readMcpFile(globalPath); err == nil {
		for name, config := range servers {
			allServers[name] = config
		}
	}

	// 2. Project Local Config (.mcp.json in CWD)
	projectPath := filepath.Join(cwd, ".mcp.json")
	if servers, err := readMcpFile(projectPath); err == nil {
		for name, config := range servers {
			// Project level overrides global if same name
			allServers[name] = config
		}
	}

	return allServers, nil
}

func readMcpFile(path string) (map[string]McpServerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var mcpJson McpJsonConfig
	if err := json.Unmarshal(data, &mcpJson); err != nil {
		// Fallback for flat structure if needed, but per spec it's { "mcpServers": { ... } }
		return nil, fmt.Errorf("invalid MCP config at %s: %v", path, err)
	}

	return mcpJson.McpServers, nil
}
