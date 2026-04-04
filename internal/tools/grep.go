package tools

import (
	"bufio"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

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

type grepFileResult struct {
	path        string
	matchInFile bool
	matches     []grepMatchLine
	numMatches  int
}

// Execute the grep logic (Parallel implementation)
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

	// Parallel logic
	numWorkers := runtime.NumCPU()
	if numWorkers < 1 {
		numWorkers = 1
	}

	pathsChan := make(chan string, 128)
	resultsChan := make(chan grepFileResult, 128)
	var wg sync.WaitGroup
	var stopFlag int32

	// Workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case path, ok := <-pathsChan:
					if !ok {
						return
					}
					if atomic.LoadInt32(&stopFlag) == 1 {
						continue
					}

					res := t.searchInFile(path, cwd, re, outputMode, before, after)
					
					// Re-check stopFlag before sending to avoid blocking on collection
					if atomic.LoadInt32(&stopFlag) == 1 && outputMode != "count" {
						continue
					}

					select {
					case <-ctx.Done():
						return
					case resultsChan <- res:
					}
				}
			}
		}()
	}

	// Result Collector
	var matchFiles []string
	var results []string
	var totalMatches int64
	
	doneChan := make(chan bool)
	go func() {
		for res := range resultsChan {
			if res.matchInFile {
				matchFiles = append(matchFiles, res.path)
				atomic.AddInt64(&totalMatches, int64(res.numMatches))
				
				if outputMode == "content" {
					for _, m := range res.matches {
						marker := ":"
						if m.context {
							marker = "-"
						}
						results = append(results, fmt.Sprintf("%s%s%d%s%s", res.path, marker, m.lineNum, marker, m.content))
					}
				}

				// Check head limit
				if atomic.LoadInt64(&totalMatches) >= int64(DefaultGrepHeadLimit) && outputMode != "count" {
					atomic.StoreInt32(&stopFlag, 1)
				}
			}
		}
		doneChan <- true
	}()

	// Traversal
	err = filepath.WalkDir(absPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		
		// Periodic check for stopFlag to abort traversal early
		if atomic.LoadInt32(&stopFlag) == 1 && outputMode != "count" {
			return filepath.SkipDir // Or just return nil, but SkipDir might be faster if we were in a deep tree
		}

		if d.IsDir() {
			allowed, _ := permissions.IsReadAllowed(path, cwd)
			if !allowed {
				return fs.SkipDir
			}
			return nil
		}

		allowed, _ := permissions.IsReadAllowed(path, cwd)
		if !allowed {
			return nil
		}

		info, err := d.Info()
		if err == nil && info.Size() > DefaultMaxFileSize {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case pathsChan <- path:
		}
		return nil
	})

	close(pathsChan)
	wg.Wait()
	close(resultsChan)
	<-doneChan

	result := map[string]interface{}{
		"filenames": matchFiles,
		"numFiles":  len(matchFiles),
	}
	
	if outputMode == "content" {
		result["content"] = strings.Join(results, "\n")
	} else if outputMode == "count" {
		result["numMatches"] = int(totalMatches)
	}

	return result, nil
}

func (t *GrepTool) searchInFile(path, cwd string, re *regexp.Regexp, outputMode string, before, after int) grepFileResult {
	f, err := os.Open(path)
	if err != nil {
		return grepFileResult{}
	}
	defer f.Close()

	relPath, _ := filepath.Rel(cwd, path)
	scanner := bufio.NewScanner(f)
	
	var fileMatches []grepMatchLine
	var history []string
	matchInFile := false
	numMatches := 0
	lineNum := 0
	linesToCapture := 0

	for scanner.Scan() {
		lineNum++
		lineContent := scanner.Text()
		
		if re.MatchString(lineContent) {
			matchInFile = true
			numMatches++
			
			if outputMode == "content" && before > 0 {
				startLine := lineNum - len(history)
				for i, h := range history {
					fileMatches = append(fileMatches, grepMatchLine{startLine + i, h, true})
				}
			}
			history = nil
			
			if outputMode == "content" {
				fileMatches = append(fileMatches, grepMatchLine{lineNum, lineContent, false})
			}
			linesToCapture = after
		} else {
			if outputMode == "content" {
				if linesToCapture > 0 {
					fileMatches = append(fileMatches, grepMatchLine{lineNum, lineContent, true})
					linesToCapture--
				} else if before > 0 {
					history = append(history, lineContent)
					if len(history) > before {
						history = history[1:]
					}
				}
			}
		}

		// Note: We don't check global head limit here to avoid too much atomic activity,
		// but the worker loop checks stopFlag.
	}

	return grepFileResult{
		path:        relPath,
		matchInFile: matchInFile,
		matches:     fileMatches,
		numMatches:  numMatches,
	}
}
