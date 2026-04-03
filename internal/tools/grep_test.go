package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGrepTool_Execute(t *testing.T) {
	// Setup temporary test directory structure
	tmpDir := t.TempDir()
	
	// Create some files
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("hello world\nthis is a test\nline 3\nline 4"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("another hello\nsecond line"), 0644)
	
	// Large file (2MB) should be skipped
	largeContent := strings.Repeat("A", 1024*1024+10) // > 1MB
	os.WriteFile(filepath.Join(tmpDir, "large.log"), []byte(largeContent), 0644)

	tool := &GrepTool{}
	ctx := context.Background()

	t.Run("Basic search - count mode", func(t *testing.T) {
		args := map[string]interface{}{
			"pattern":     "hello",
			"path":        tmpDir,
			"output_mode": "count",
		}
		result, err := tool.Execute(ctx, args)
		require.NoError(t, err)
		
		resMap := result.(map[string]interface{})
		assert.Equal(t, 2, resMap["numMatches"])
		assert.Equal(t, 2, resMap["numFiles"])
	})

	t.Run("Case insensitive search", func(t *testing.T) {
		args := map[string]interface{}{
			"pattern":     "HELLO",
			"path":        tmpDir,
			"output_mode": "count",
			"-i":          true,
		}
		result, err := tool.Execute(ctx, args)
		require.NoError(t, err)
		
		resMap := result.(map[string]interface{})
		assert.Equal(t, 2, resMap["numMatches"])
	})

	t.Run("Content mode with context (-A 1)", func(t *testing.T) {
		args := map[string]interface{}{
			"pattern":     "world",
			"path":        tmpDir,
			"output_mode": "content",
			"-A":          1,
		}
		result, err := tool.Execute(ctx, args)
		require.NoError(t, err)
		
		resMap := result.(map[string]interface{})
		content := resMap["content"].(string)
		
		// Expected hello world (match) + this is a test (context after)
		assert.Contains(t, content, "hello world")
		assert.Contains(t, content, "this is a test")
		assert.NotContains(t, content, "line 3")
	})

	t.Run("File size limit check", func(t *testing.T) {
		// Searching for 'A' which is in the large file, but it should be skipped
		args := map[string]interface{}{
			"pattern":     "A",
			"path":        tmpDir,
			"output_mode": "count",
		}
		result, err := tool.Execute(ctx, args)
		require.NoError(t, err)
		
		resMap := result.(map[string]interface{})
		assert.Equal(t, 0, resMap["numMatches"], "Large file should have been skipped")
	})
}
