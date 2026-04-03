package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/onedayallday33-a11y/openshannon-go/internal/utils/permissions"
)

// FileWriteTool implements the Tool interface for writing whole files
type FileWriteTool struct{}

// Name of the tool
func (t *FileWriteTool) Name() string {
	return "FileWrite"
}

// Description of the tool
func (t *FileWriteTool) Description() string {
	return "Create or overwrite a file with new content"
}

// InputSchema for the tool
func (t *FileWriteTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "Absolute path to the file to write",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "The exact content to write to the file",
			},
		},
		"required": []string{"file_path", "content"},
	}
}

// Execute the file write logic
func (t *FileWriteTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	filePath, ok := args["file_path"].(string)
	if !ok {
		return nil, fmt.Errorf("file_path is required")
	}

	content, ok := args["content"].(string)
	if !ok {
		return nil, fmt.Errorf("content is required")
	}

	cwd, _ := os.Getwd()

	// 1. Security Check
	allowed, msg := permissions.IsWriteAllowed(filePath, cwd)
	if !allowed {
		return nil, fmt.Errorf("security violation: %s", msg)
	}

	// 2. Determine if it's create or update
	exists := true
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		exists = false
	}

	// 3. Ensure parent directories exist (MkdirAll)
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directories: %v", err)
	}

	// 4. Write Content
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %v", err)
	}

	status := "update"
	if !exists {
		status = "create"
	}

	return map[string]interface{}{
		"type":     status,
		"filePath": filePath,
		"message":  fmt.Sprintf("File %s successfully at %s", status, filePath),
	}, nil
}
