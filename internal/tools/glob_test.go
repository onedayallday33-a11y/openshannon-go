package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGlobTool_Execute(t *testing.T) {
	// Setup temporary test directory structure
	tmpDir := t.TempDir()
	
	// Save current CWD and restore it later
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	
	// Change CWD to tmpDir for the duration of the test
	err := os.Chdir(tmpDir)
	assert.NoError(t, err)
	
	// Create some files
	os.MkdirAll(filepath.Join("src", "api"), 0755)
	os.WriteFile(filepath.Join("src", "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join("src", "api", "v1.go"), []byte("package api"), 0644)
	os.WriteFile(filepath.Join("README.md"), []byte("# Title"), 0644)
	
	// Create a forbidden file
	gitDir := ".git"
	os.Mkdir(gitDir, 0755)
	os.WriteFile(filepath.Join(gitDir, "config"), []byte("git config"), 0644)

	tool := &GlobTool{}
	ctx := context.Background()

	t.Run("Find all .go files recursively", func(t *testing.T) {
		args := map[string]interface{}{
			"pattern": "**/*.go",
			"path":    ".",
		}
		result, err := tool.Execute(ctx, args)
		require.NoError(t, err)
		
		resMap := result.(map[string]interface{})
		filenames := resMap["filenames"].([]string)
		
		assert.Equal(t, 2, len(filenames))
		
		// Map back to slash for consistent testing cross-platform
		var slashFilenames []string
		for _, f := range filenames {
			slashFilenames = append(slashFilenames, filepath.ToSlash(f))
		}
		
		assert.Contains(t, slashFilenames, "src/main.go")
		assert.Contains(t, slashFilenames, "src/api/v1.go")
	})

	t.Run("Security filter: should not find files in .git", func(t *testing.T) {
		args := map[string]interface{}{
			"pattern": "**/*",
			"path":    ".",
		}
		result, err := tool.Execute(ctx, args)
		require.NoError(t, err)
		
		resMap := result.(map[string]interface{})
		filenames := resMap["filenames"].([]string)
		
		for _, f := range filenames {
			assert.NotContains(t, f, ".git", "Result should not contain hidden .git folder")
		}
	})

	t.Run("No matches found", func(t *testing.T) {
		args := map[string]interface{}{
			"pattern": "*.txt",
			"path":    tmpDir,
		}
		result, err := tool.Execute(ctx, args)
		require.NoError(t, err)
		
		resMap := result.(map[string]interface{})
		assert.Equal(t, 0, resMap["numFiles"])
	})
}
