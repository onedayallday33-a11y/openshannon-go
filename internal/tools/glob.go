package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/onedayallday33-a11y/openshannon-go/internal/utils/permissions"
)

const (
	// DefaultGlobLimit for the Glob tool
	DefaultGlobLimit = 100
)

// GlobTool implements the Tool interface for finding files by pattern
type GlobTool struct{}

// Name of the tool
func (t *GlobTool) Name() string {
	return "Glob"
}

// Description of the tool
func (t *GlobTool) Description() string {
	return "Find files by name pattern or wildcard"
}

// InputSchema for the tool
func (t *GlobTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "The glob pattern to match files against (e.g. **/*.go)",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "The directory to search in (defaults to CWD)",
			},
		},
		"required": []string{"pattern"},
	}
}

// Execute the glob logic
func (t *GlobTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	pattern, ok := args["pattern"].(string)
	if !ok {
		return nil, fmt.Errorf("pattern is required and must be a string")
	}

	searchPath := "."
	if v, exists := args["path"]; exists {
		if val, itOk := v.(string); itOk {
			searchPath = val
		}
	}

	absSearchPath, err := filepath.Abs(searchPath)
	if err != nil {
		return nil, err
	}

	cwd, _ := os.Getwd()

	// 1. Perform Globbing using doublestar (supports **)
	// We use OS-agnostic slash conversion for doublestar
	fsys := os.DirFS(absSearchPath)
	matches, err := doublestar.Glob(fsys, pattern)
	if err != nil {
		return nil, fmt.Errorf("glob error: %v", err)
	}

	// 2. Filter Results (Security & Limit)
	var finalFiles []string
	truncated := false
	
	// Sort to ensure deterministic results
	sort.Strings(matches)

	for _, relPath := range matches {
		if len(finalFiles) >= DefaultGlobLimit {
			truncated = true
			break
		}

		fullPath := filepath.Join(absSearchPath, relPath)
		
		// Check Permissions (Skip if denied)
		allowed, _ := permissions.IsReadAllowed(fullPath, cwd)
		if allowed {
			// Convert back to relative path from CWD for output (standard behavior)
			relToCwd, err := filepath.Rel(cwd, fullPath)
			if err == nil {
				finalFiles = append(finalFiles, relToCwd)
			} else {
				finalFiles = append(finalFiles, fullPath)
			}
		}
	}

	result := map[string]interface{}{
		"filenames": finalFiles,
		"numFiles":  len(finalFiles),
		"truncated": truncated,
	}

	return result, nil
}
