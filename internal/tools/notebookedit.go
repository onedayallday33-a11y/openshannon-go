package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/onedayallday33-a11y/openshannon-go/internal/utils/permissions"
)

// JupyterNotebook represents the structure of an .ipynb file
type JupyterNotebook struct {
	Cells          []NotebookCell         `json:"cells"`
	Metadata       map[string]interface{} `json:"metadata"`
	Nbformat       int                    `json:"nbformat"`
	NbformatMinor  int                    `json:"nbformat_minor"`
}

// NotebookCell represents a single cell in a notebook
type NotebookCell struct {
	CellType       string                 `json:"cell_type"`
	Metadata       map[string]interface{} `json:"metadata"`
	Source         []string               `json:"source"`
	Outputs        []interface{}          `json:"outputs,omitempty"`
	ExecutionCount *int                   `json:"execution_count,omitempty"`
}

// NotebookEditTool implements the Tool interface for editing notebooks
type NotebookEditTool struct{}

// Name of the tool
func (t *NotebookEditTool) Name() string {
	return "NotebookEdit"
}

// Description of the tool
func (t *NotebookEditTool) Description() string {
	return "Edit a Jupyter Notebook (.ipynb) file by inserting, removing, or replacing cells"
}

// InputSchema for the tool
func (t *NotebookEditTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "The absolute path to the .ipynb file",
			},
			"action": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"insert", "remove", "update"},
				"description": "The action to perform on the notebook",
			},
			"index": map[string]interface{}{
				"type":        "integer",
				"description": "The 0-based index of the cell to act on",
			},
			"cell_type": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"code", "markdown"},
				"description": "Type of the cell for 'insert' or 'update' actions",
			},
			"source": map[string]interface{}{
				"type":        "string",
				"description": "The source content of the cell",
			},
		},
		"required": []string{"file_path", "action", "index"},
	}
}

// Execute the notebook edit logic
func (t *NotebookEditTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	filePath, _ := args["file_path"].(string)
	action, _ := args["action"].(string)
	
	var index int
	if v, ok := args["index"].(float64); ok {
		index = int(v)
	} else if v, ok := args["index"].(int); ok {
		index = v
	}

	cwd, _ := os.Getwd()
	// Security Check
	allowed, msg := permissions.IsWriteAllowed(filePath, cwd)
	if !allowed {
		return nil, fmt.Errorf("security violation: %s", msg)
	}

	// 1. Read Notebook
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read notebook: %v", err)
	}

	var nb JupyterNotebook
	if err := json.Unmarshal(data, &nb); err != nil {
		return nil, fmt.Errorf("invalid notebook JSON: %v", err)
	}

	// 2. Apply Action
	switch action {
	case "remove":
		if index < 0 || index >= len(nb.Cells) {
			return nil, fmt.Errorf("index %d out of bounds (length %d)", index, len(nb.Cells))
		}
		nb.Cells = append(nb.Cells[:index], nb.Cells[index+1:]...)

	case "insert", "update":
		cellType, _ := args["cell_type"].(string)
		source, _ := args["source"].(string)
		
		newCell := NotebookCell{
			CellType: cellType,
			Metadata: make(map[string]interface{}),
			Source:   []string{source},
		}
		if cellType == "code" {
			newCell.Outputs = []interface{}{}
			zero := 0
			newCell.ExecutionCount = &zero
		}

		if action == "update" {
			if index < 0 || index >= len(nb.Cells) {
				return nil, fmt.Errorf("index %d out of bounds", index)
			}
			nb.Cells[index] = newCell
		} else { // insert
			if index < 0 || index > len(nb.Cells) {
				index = len(nb.Cells)
			}
			nb.Cells = append(nb.Cells[:index], append([]NotebookCell{newCell}, nb.Cells[index:]...)...)
		}
	}

	// 3. Write back
	output, err := json.MarshalIndent(nb, "", " ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal notebook: %v", err)
	}

	if err := os.WriteFile(filePath, output, 0644); err != nil {
		return nil, fmt.Errorf("failed to write notebook: %v", err)
	}

	return map[string]interface{}{
		"status": "success",
		"action": action,
		"index":  index,
	}, nil
}
