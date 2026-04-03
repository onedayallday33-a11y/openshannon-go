package tools

import (
	"context"
	"fmt"
	"os"

	"github.com/onedayallday33-a11y/openshannon-go/internal/lsp"
)

// LSPTool implements Tool for code intelligence
type LSPTool struct{}

func (t *LSPTool) Name() string { return "LSP" }
func (t *LSPTool) Description() string {
	return "Code intelligence operations: goToDefinition, findReferences, hover, documentSymbol"
}

func (t *LSPTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"operation": map[string]interface{}{
				"type": "string",
				"enum": []string{"goToDefinition", "findReferences", "hover", "documentSymbol"},
			},
			"file_path": map[string]interface{}{"type": "string"},
			"line":      map[string]interface{}{"type": "integer", "description": "1-based line"},
			"character": map[string]interface{}{"type": "integer", "description": "1-based char"},
		},
		"required": []string{"operation", "file_path"},
	}
}

func (t *LSPTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	op, _ := args["operation"].(string)
	path, _ := args["file_path"].(string)
	
	line := 1
	if v, ok := args["line"].(float64); ok {
		line = int(v)
	}
	char := 1
	if v, ok := args["character"].(float64); ok {
		char = int(v)
	}

	cwd, _ := os.Getwd()
	manager := lsp.GetLspManager(cwd)

	client, err := manager.GetClientForFile(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get LSP client: %v", err)
	}
	if client == nil {
		return nil, fmt.Errorf("no language server found for file: %s", path)
	}

	// 0-based position for protocol
	pos := lsp.Position{Line: line - 1, Character: char - 1}

	var method string
	var params interface{}

	switch op {
	case "goToDefinition":
		method = "textDocument/definition"
		params = map[string]interface{}{
			"textDocument": map[string]string{"uri": "file://" + path},
			"position":     pos,
		}
	case "findReferences":
		method = "textDocument/references"
		params = map[string]interface{}{
			"textDocument": map[string]string{"uri": "file://" + path},
			"position":     pos,
			"context":      map[string]interface{}{"includeDeclaration": true},
		}
	case "hover":
		method = "textDocument/hover"
		params = map[string]interface{}{
			"textDocument": map[string]string{"uri": "file://" + path},
			"position":     pos,
		}
	case "documentSymbol":
		method = "textDocument/documentSymbol"
		params = map[string]interface{}{
			"textDocument": map[string]string{"uri": "file://" + path},
		}
	}

	resp, err := client.Call(ctx, method, params)
	if err != nil {
		return nil, err
	}

	return resp.Result, nil
}
