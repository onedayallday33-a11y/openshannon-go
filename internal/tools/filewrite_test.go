package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileWriteTool_Execute(t *testing.T) {
	// Setup temporary test directory structure
	tmpDir := t.TempDir()
	
	// Save current CWD and restore it later
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	
	// Change CWD to tmpDir for the duration of the test
	err := os.Chdir(tmpDir)
	assert.NoError(t, err)

	tool := &FileWriteTool{}
	ctx := context.Background()

	t.Run("Create new file in new subdirectory", func(t *testing.T) {
		newPath := filepath.Join("new_dir", "sub_dir", "hello.txt")
		content := "hello world"
		
		args := map[string]interface{}{
			"file_path": newPath,
			"content":   content,
		}
		
		result, err := tool.Execute(ctx, args)
		require.NoError(t, err)
		
		resMap := result.(map[string]interface{})
		assert.Equal(t, "create", resMap["type"])
		
		// Verify file
		data, _ := os.ReadFile(newPath)
		assert.Equal(t, content, string(data))
	})

	t.Run("Overwrite existing file", func(t *testing.T) {
		path := "existing.txt"
		os.WriteFile(path, []byte("old content"), 0644)
		
		newContent := "new content"
		args := map[string]interface{}{
			"file_path": path,
			"content":   newContent,
		}
		
		result, err := tool.Execute(ctx, args)
		require.NoError(t, err)
		
		resMap := result.(map[string]interface{})
		assert.Equal(t, "update", resMap["type"])
		
		data, _ := os.ReadFile(path)
		assert.Equal(t, newContent, string(data))
	})

	t.Run("Security violation: write outside CWD", func(t *testing.T) {
		// Mock a path outside the temp dir (CWD)
		badPath := filepath.Join("..", "evil.txt")
		args := map[string]interface{}{
			"file_path": badPath,
			"content":   "hack",
		}
		
		_, err := tool.Execute(ctx, args)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "security violation")
	})
}
