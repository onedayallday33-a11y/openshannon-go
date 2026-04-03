package tools

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/onedayallday33-a11y/openshannon-go/internal/utils/permissions"
)

const (
	// DefaultMaxLinesToRead for the Read tool
	DefaultMaxLinesToRead = 2000
)

// FileReadTool implements the Tool interface for reading files
type FileReadTool struct{}

// Name of the tool
func (t *FileReadTool) Name() string {
	return "Read"
}

// Description of the tool
func (t *FileReadTool) Description() string {
	return "Read a file from the local filesystem."
}

// InputSchema for the tool
func (t *FileReadTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "The absolute path to the file to read",
			},
			"offset": map[string]interface{}{
				"type":        "integer",
				"description": "The line number to start reading from. (1-indexed)",
				"default":     1,
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "The number of lines to read.",
				"default":     DefaultMaxLinesToRead,
			},
		},
		"required": []string{"file_path"},
	}
}

// Execute the reading logic
func (t *FileReadTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	filePath, ok := args["file_path"].(string)
	if !ok {
		return nil, fmt.Errorf("file_path is required and must be a string")
	}

	offset := 1
	if v, exists := args["offset"]; exists {
		if val, itOk := v.(float64); itOk {
			offset = int(val)
		} else if val, itOk := v.(int); itOk {
			offset = val
		}
	}

	limit := DefaultMaxLinesToRead
	if v, exists := args["limit"]; exists {
		if val, itOk := v.(float64); itOk {
			limit = int(val)
		} else if val, itOk := v.(int); itOk {
			limit = val
		}
	}

	cwd, _ := os.Getwd()
	// Security Check
	allowed, msg := permissions.IsReadAllowed(filePath, cwd)
	if !allowed {
		return nil, fmt.Errorf("permission denied: %s", msg)
	}

	// Basic Binary File Extension Check (from openclaude-ts)
	if isBinaryExtension(filePath) {
		return nil, fmt.Errorf("this tool cannot read binary files")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var result strings.Builder
	scanner := bufio.NewScanner(file)
	currentLine := 0
	linesRead := 0

	for scanner.Scan() {
		currentLine++
		if currentLine < offset {
			continue
		}

		if linesRead >= limit {
			break
		}

		// cat -n format: "     1  line content" 
		line := scanner.Text()
		result.WriteString(fmt.Sprintf("%6d  %s\n", currentLine, line))
		linesRead++
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if result.Len() == 0 && currentLine > 0 && currentLine < offset {
		return fmt.Sprintf("Warning: file has %d lines, but provided offset is %d.", currentLine, offset), nil
	}

	return result.String(), nil
}

// isBinaryExtension checks if file extension is likely binary
func isBinaryExtension(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".exe", ".bin", ".pyc", ".pyd", ".dll", ".so", ".dylib", ".wasm", ".o", ".a", ".out":
		return true
	case ".zip", ".tar", ".gz", ".7z", ".rar", ".bz2", ".xz":
		return true
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".pdf":
		// These are binary but OpenClaude-TS handles them; however, 
		// for our first text-only Go implementation we block it unless 
		// logic for them is implemented.
		return true
	}
	return false
}
