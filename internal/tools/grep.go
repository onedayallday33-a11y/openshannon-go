package tools

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/onedayallday33-a11y/openshannon-go/internal/utils/permissions"
)

const (
	// DefaultMaxFileSize for Grep (1MB)
	DefaultMaxFileSize = 1024 * 1024
	// DefaultGrepHeadLimit
	DefaultGrepHeadLimit = 250
)

// GrepTool implements the Tool interface for searching file contents
type GrepTool struct{}

// Name of the tool
func (t *GrepTool) Name() string {
	return "Search"
}

// Description of the tool
func (t *GrepTool) Description() string {
	return "Search for file contents with regex (recursive)"
}

// InputSchema for the tool
func (t *GrepTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "The regular expression pattern to search for",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "File or directory to search in (defaults to CWD)",
			},
			"output_mode": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"content", "files_with_matches", "count"},
				"description": "Output mode: 'content' (lines), 'files_with_matches' (paths), 'count' (counts)",
			},
			"-i": map[string]interface{}{
				"type":        "boolean",
				"description": "Case insensitive search",
			},
			"-B": map[string]interface{}{
				"type":        "integer",
				"description": "Lines to show before (context before)",
			},
			"-A": map[string]interface{}{
				"type":        "integer",
				"description": "Lines to show after (context after)",
			},
			"-C": map[string]interface{}{
				"type":        "integer",
				"description": "Context lines (both before and after)",
			},
		},
		"required": []string{"pattern"},
	}
}

type grepMatchLine struct {
	lineNum int
	content string
	context bool // true if it's a context line, false if it's a match
}

// Execute the grep logic
func (t *GrepTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	patternStr, _ := args["pattern"].(string)
	searchPath := "."
	if p, ok := args["path"].(string); ok {
		searchPath = p
	}

	outputMode := "files_with_matches"
	if m, ok := args["output_mode"].(string); ok {
		outputMode = m
	}

	caseInsensitive := false
	if i, ok := args["-i"].(bool); ok {
		caseInsensitive = i
	}

	// Context handling
	var before, after int
	if c, ok := args["-C"].(float64); ok {
		before, after = int(c), int(c)
	} else if c, ok := args["-C"].(int); ok {
		before, after = c, c
	} else {
		if b, ok := args["-B"].(float64); ok {
			before = int(b)
		} else if b, ok := args["-B"].(int); ok {
			before = b
		}
		if a, ok := args["-A"].(float64); ok {
			after = int(a)
		} else if a, ok := args["-A"].(int); ok {
			after = a
		}
	}

	// Prepare Regex
	finalPattern := patternStr
	if caseInsensitive {
		finalPattern = "(?i)" + finalPattern
	}
	re, err := regexp.Compile(finalPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %v", err)
	}

	absPath, err := filepath.Abs(searchPath)
	if err != nil {
		return nil, err
	}

	cwd, _ := os.Getwd()
	
	results := []string{}
	matchFiles := []string{}
	totalMatches := 0
	
	// Traversal
	err = filepath.WalkDir(absPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}
		if d.IsDir() {
			// Skip dangerous directories (logic handled in IsReadAllowed, but good to prune early)
			allowed, _ := permissions.IsReadAllowed(path, cwd)
			if !allowed {
				return fs.SkipDir
			}
			return nil
		}

		// Security Check
		allowed, _ := permissions.IsReadAllowed(path, cwd)
		if !allowed {
			return nil
		}

		// File Size Check (1MB Skip)
		info, err := d.Info()
		if err == nil && info.Size() > DefaultMaxFileSize {
			return nil
		}

		// Open and scan
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()

		relPath, _ := filepath.Rel(cwd, path)
		
		scanner := bufio.NewScanner(f)
		lineNum := 0
		
		var fileMatches []grepMatchLine
		var history []string // Circular buffer for context before
		
		matchInFile := false
		linesToCapture := 0 // Tracks how many lines after a match to capture

		for scanner.Scan() {
			lineNum++
			lineContent := scanner.Text()
			
			if re.MatchString(lineContent) {
				matchInFile = true
				totalMatches++
				
				// Capture context before if needed
				if outputMode == "content" && before > 0 {
					startLine := lineNum - len(history)
					for i, h := range history {
						fileMatches = append(fileMatches, grepMatchLine{startLine + i, h, true})
					}
				}
				history = nil // Clear history after match
				
				// Capture current match
				if outputMode == "content" {
					fileMatches = append(fileMatches, grepMatchLine{lineNum, lineContent, false})
				}
				
				// Reset after counter
				linesToCapture = after
			} else {
				if outputMode == "content" {
					if linesToCapture > 0 {
						fileMatches = append(fileMatches, grepMatchLine{lineNum, lineContent, true})
						linesToCapture--
					} else if before > 0 {
						// Keep history for next potential match
						history = append(history, lineContent)
						if len(history) > before {
							history = history[1:] // Shift
						}
					}
				}
			}
			
			// Stop if reached excessive matches (head_limit)
			if totalMatches >= DefaultGrepHeadLimit && outputMode != "count" {
				break
			}
		}

		if matchInFile {
			matchFiles = append(matchFiles, relPath)
			if outputMode == "content" {
				for _, m := range fileMatches {
					marker := ":"
					if m.context {
						marker = "-"
					}
					results = append(results, fmt.Sprintf("%s%s%d%s%s", relPath, marker, m.lineNum, marker, m.content))
				}
			} else if outputMode == "count" {
				// Count is handled via totalMatches and matchFiles.length
			}
		}

		if totalMatches >= DefaultGrepHeadLimit && outputMode != "count" {
			return io.EOF // Special return to stop recursion
		}
		return nil
	})

	if err != nil && err.Error() != "EOF" {
		// Note: WalkDir returns EOF if I returned it
	}

	result := map[string]interface{}{
		"filenames": matchFiles,
		"numFiles":  len(matchFiles),
	}
	
	if outputMode == "content" {
		result["content"] = strings.Join(results, "\n")
	} else if outputMode == "count" {
		result["numMatches"] = totalMatches
	}

	return result, nil
}
