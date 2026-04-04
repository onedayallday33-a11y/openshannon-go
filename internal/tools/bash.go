package tools

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/onedayallday33-a11y/openshannon-go/internal/utils/permissions"
)

// BashTool implements the Tool interface for executing shell commands
type BashTool struct{}

// Name of the tool
func (t *BashTool) Name() string {
	return "Bash"
}

// Description of the tool
func (t *BashTool) Description() string {
	return "Run shell command"
}

// InputSchema for the tool
func (t *BashTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "The command to execute",
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "Optional timeout in milliseconds",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Clear, concise description of what this command does",
			},
		},
		"required": []string{"command"},
	}
}

// Execute the bash logic
func (t *BashTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	command, ok := args["command"].(string)
	if !ok {
		return nil, fmt.Errorf("command is required and must be a string")
	}

	timeoutMs := 30000 // Default 30s
	if v, exists := args["timeout"]; exists {
		if val, itOk := v.(float64); itOk {
			timeoutMs = int(val)
		} else if val, itOk := v.(int); itOk {
			timeoutMs = val
		}
	}

	cwd, _ := os.Getwd()

	// 1. Security Check
	allowed, msg := permissions.IsCommandSafe(command, cwd)
	if !allowed {
		return nil, fmt.Errorf("security violation: %s", msg)
	}

	// 2. Setup Context with Timeout
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	// 3. Prepare Command
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(execCtx, "cmd", "/c", command)
	} else {
		cmd = exec.CommandContext(execCtx, "bash", "-c", command)
	}

	// 4. Capture Output (Real-time and Combined)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = cmd.Stdout // Merge stderr into stdout like a typical PTY

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	var output strings.Builder
	reader := bufio.NewReader(stdout)

	// Stream reading
	done := make(chan bool)
	go func() {
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			output.WriteString(line)
		}
		done <- true
	}()

	// Wait for reader completion before Wait()
	// This ensures we get all output and avoid race conditions or closing pipes early
	select {
	case <-done:
	case <-execCtx.Done():
		// On timeout, the reader might still be blocked. 
		// Closing stdout will force the reader to exit.
		stdout.Close()
		<-done // Wait for reader to acknowledge exit
	}

	// Wait for command completion
	err = cmd.Wait()
	
	interrupted := false
	if execCtx.Err() == context.DeadlineExceeded {
		interrupted = true
		output.WriteString("\n<error>Command timed out</error>")
	}

	result := map[string]interface{}{
		"stdout":      output.String(),
		"stderr":      "", // Combined in stdout
		"interrupted": interrupted,
		"exit_code":   0,
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result["exit_code"] = exitError.ExitCode()
		} else if !interrupted {
			return nil, err
		}
	}

	return result, nil
}
