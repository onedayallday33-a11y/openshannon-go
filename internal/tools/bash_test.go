package tools

import (
	"context"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBashTool_Execute(t *testing.T) {
	tool := &BashTool{}
	ctx := context.Background()

	t.Run("Execute simple echo", func(t *testing.T) {
		args := map[string]interface{}{"command": "echo hello_world"}
		result, err := tool.Execute(ctx, args)
		require.NoError(t, err)
		
		resMap := result.(map[string]interface{})
		assert.Contains(t, resMap["stdout"].(string), "hello_world")
		assert.False(t, resMap["interrupted"].(bool))
	})

	t.Run("Execute with timeout", func(t *testing.T) {
		// Use a long-running command and a short timeout
		var cmd string
		if runtime.GOOS == "windows" {
			cmd = "ping -n 6 127.0.0.1" // Runs for approx 5 seconds
		} else {
			cmd = "sleep 10"
		}

		args := map[string]interface{}{
			"command": cmd,
			"timeout": 500, // 500ms
		}
		result, err := tool.Execute(ctx, args)
		assert.NoError(t, err) // We handle timeout internally, it doesn't return err.

		resMap := result.(map[string]interface{})
		assert.True(t, resMap["interrupted"].(bool))
		assert.Contains(t, resMap["stdout"].(string), "<error>Command timed out</error>")
	})

	t.Run("Security violation: rm -rf /", func(t *testing.T) {
		args := map[string]interface{}{"command": "rm -rf /"}
		_, err := tool.Execute(ctx, args)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "security violation")
	})

	t.Run("Security violation: redirection to .bashrc", func(t *testing.T) {
		// Mock home dir or just use a known blocked file
		args := map[string]interface{}{"command": "echo hacked > .bashrc"}
		_, err := tool.Execute(ctx, args)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "security violation")
	})

	t.Run("Security violation: NTFS Stream", func(t *testing.T) {
		args := map[string]interface{}{"command": "echo hello > file.txt:stream"}
		_, err := tool.Execute(ctx, args)
		assert.Error(t, err)
		assert.Contains(t, strings.ToLower(err.Error()), "suspicious windows path pattern")
	})
}
