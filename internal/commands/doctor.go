package commands

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/onedayallday33-a11y/openshannon-go/internal/agent"
)

type DoctorCommand struct{}

func (c *DoctorCommand) Name() string { return "doctor" }
func (c *DoctorCommand) Description() string { return "Check environment health" }

func (c *DoctorCommand) Execute(ctx context.Context, a *agent.Agent, args []string) (agent.DirectCommandResult, error) {
	var sb strings.Builder
	sb.WriteString("OpenShannon Doctor - Health Check\n")

	tools := []string{"go", "git", "gopls", "python", "pyright-langserver"}
	for _, tool := range tools {
		_, err := exec.LookPath(tool)
		status := "✓ Found"
		if err != nil {
			status = "✗ Not Found"
		}
		sb.WriteString(fmt.Sprintf("  - %-18s: %s\n", tool, status))
	}

	return agent.DirectCommandResult{
		DirectOutput: sb.String(),
		IsHandled:    true,
	}, nil
}
