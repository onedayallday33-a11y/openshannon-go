package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileReadTool_Execute(t *testing.T) {
	// Setup temporary test file
	tmpDir := t.TempDir()
	testFilePath := filepath.Join(tmpDir, "test.txt")
	content := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5"
	err := os.WriteFile(testFilePath, []byte(content), 0644)
	assert.NoError(t, err)

	tool := &FileReadTool{}

	t.Run("Read entire file", func(t *testing.T) {
		args := map[string]interface{}{"file_path": testFilePath}
		result, err := tool.Execute(context.Background(), args)
		require.NoError(t, err)
		assert.Contains(t, result.(string), "Line 1")
		assert.Contains(t, result.(string), "Line 5")
	})

	t.Run("Read with offset and limit", func(t *testing.T) {
		args := map[string]interface{}{
			"file_path": testFilePath,
			"offset":    2,
			"limit":     2,
		}
		result, err := tool.Execute(context.Background(), args)
		require.NoError(t, err)
		// Line 2 and Line 3 should be present
		assert.Contains(t, result.(string), "Line 2")
		assert.Contains(t, result.(string), "Line 3")
		// Line 1 and Line 4 should not be present
		assert.NotContains(t, result.(string), "Line 1")
		assert.NotContains(t, result.(string), "Line 4")
	})

	t.Run("Read non-existent file", func(t *testing.T) {
		args := map[string]interface{}{"file_path": "non_existent.txt"}
		_, err := tool.Execute(context.Background(), args)
		assert.Error(t, err)
	})

	t.Run("Permission denied reading .git/config", func(t *testing.T) {
		// Mock a .git directory structure
		gitDir := filepath.Join(tmpDir, ".git")
		os.Mkdir(gitDir, 0755)
		gitConfig := filepath.Join(gitDir, "config")
		os.WriteFile(gitConfig, []byte("config"), 0644)

		args := map[string]interface{}{"file_path": gitConfig}
		_, err := tool.Execute(context.Background(), args)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "permission denied")
	})
}
