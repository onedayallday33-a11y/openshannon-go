package tools

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/onedayallday33-a11y/openshannon-go/internal/utils/permissions"
)

// FileEditTool implements the Tool interface for editing files
type FileEditTool struct{}

// Name of the tool
func (t *FileEditTool) Name() string {
	return "FileEdit"
}

// Description of the tool
func (t *FileEditTool) Description() string {
	return "Edit a file by replacing a string with another"
}

// InputSchema for the tool
func (t *FileEditTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file to edit",
			},
			"edits": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"old_string": map[string]interface{}{
							"type":        "string",
							"description": "The exact string to find in the file",
						},
						"new_string": map[string]interface{}{
							"type":        "string",
							"description": "The string to replace it with",
						},
						"replace_all": map[string]interface{}{
							"type":        "boolean",
							"description": "Whether to replace all occurrences",
						},
					},
					"required": []string{"old_string", "new_string"},
				},
			},
		},
		"required": []string{"file_path", "edits"},
	}
}

// Execute the file edit logic
func (t *FileEditTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	filePath, ok := args["file_path"].(string)
	if !ok {
		return nil, fmt.Errorf("file_path is required")
	}

	editsRaw, ok := args["edits"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("edits must be an array")
	}

	cwd, _ := os.Getwd()

	// 1. Security Check (CWD Restricted)
	allowed, msg := permissions.IsWriteAllowed(filePath, cwd)
	if !allowed {
		return nil, fmt.Errorf("security violation: %s", msg)
	}

	// 2. Read File
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}
	content := string(contentBytes)

	// 3. Apply Edits
	updatedContent := content
	numEditsApplied := 0

	for i, eRaw := range editsRaw {
		e, ok := eRaw.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("edit at index %d is invalid", i)
		}

		oldStr, _ := e["old_string"].(string)
		newStr, _ := e["new_string"].(string)
		replaceAll, _ := e["replace_all"].(bool)

		// Normalization: Handle line ending mismatches (\r\n vs \n)
		// We normalize both strings to \n for comparison, but try to keep file's original style for the replacement.
		
		normalizedOld := normalizeLineEndings(oldStr)
		normalizedContent := normalizeLineEndings(updatedContent)

		if !strings.Contains(normalizedContent, normalizedOld) {
			return nil, fmt.Errorf("edit at index %d failed: 'old_string' not found in file (hint: ensure exact match including whitespace)", i)
		}

		// Perform the replacement
		// If the file uses \r\n, we should probably stick to it, but for simplicity in V1, 
		// we'll replace the first/all occurrences found.
		// Note: Using the non-normalized content for final replacement if possible to preserve style.
		
		if replaceAll {
			updatedContent = strings.ReplaceAll(updatedContent, oldStr, newStr)
			numEditsApplied++
		} else {
			// Replace only once
			updatedContent = strings.Replace(updatedContent, oldStr, newStr, 1)
			numEditsApplied++
		}
	}

	// 4. Save File
	err = os.WriteFile(filePath, []byte(updatedContent), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %v", err)
	}

	return map[string]interface{}{
		"summary": fmt.Sprintf("Successfully applied %d edits to %s", numEditsApplied, filePath),
	}, nil
}

// normalizeLineEndings converts all \r\n to \n
func normalizeLineEndings(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}
