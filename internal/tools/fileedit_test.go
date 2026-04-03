package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileEditTool_Execute(t *testing.T) {
	// Setup temporary test directory structure
	tmpDir := t.TempDir()
	
	// Save current CWD and restore it later
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	
	// Change CWD to tmpDir for the duration of the test
	err := os.Chdir(tmpDir)
	assert.NoError(t, err)

	filePath := "code.go"
	initialContent := "package main\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n"
	os.WriteFile(filePath, []byte(initialContent), 0644)

	tool := &FileEditTool{}
	ctx := context.Background()

	t.Run("Successful single edit", func(t *testing.T) {
		args := map[string]interface{}{
			"file_path": filePath,
			"edits": []interface{}{
				map[string]interface{}{
					"old_string": "fmt.Println(\"hello\")",
					"new_string": "fmt.Println(\"world\")",
				},
			},
		}
		result, err := tool.Execute(ctx, args)
		require.NoError(t, err)
		
		resMap := result.(map[string]interface{})
		assert.Contains(t, resMap["summary"].(string), "Successfully applied 1 edits")
		
		// Verify content
		newContent, _ := os.ReadFile(filePath)
		assert.Contains(t, string(newContent), "world")
		assert.NotContains(t, string(newContent), "\"hello\"")
	})

	t.Run("Multiple edits in one call", func(t *testing.T) {
		os.WriteFile(filePath, []byte(initialContent), 0644)
		args := map[string]interface{}{
			"file_path": filePath,
			"edits": []interface{}{
				map[string]interface{}{
					"old_string": "package main",
					"new_string": "package main_updated",
				},
				map[string]interface{}{
					"old_string": "func main()",
					"new_string": "func myMain()",
				},
			},
		}
		_, err := tool.Execute(ctx, args)
		require.NoError(t, err)
		
		newContent, _ := os.ReadFile(filePath)
		assert.Contains(t, string(newContent), "package main_updated")
		assert.Contains(t, string(newContent), "func myMain()")
	})

	t.Run("Fail: string not found", func(t *testing.T) {
		args := map[string]interface{}{
			"file_path": filePath,
			"edits": []interface{}{
				map[string]interface{}{
					"old_string": "non-existent-string",
					"new_string": "something",
				},
			},
		}
		_, err := tool.Execute(ctx, args)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "old_string' not found")
	})

	t.Run("Security violation: writing outside CWD", func(t *testing.T) {
		// Try to write to a parent directory (tmpDir's parent)
		parentFile := filepath.Join("..", "hack.txt")
		args := map[string]interface{}{
			"file_path": parentFile,
			"edits": []interface{}{
				map[string]interface{}{
					"old_string": "anything",
					"new_string": "anything",
				},
			},
		}
		_, err := tool.Execute(ctx, args)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "security violation")
	})
}
