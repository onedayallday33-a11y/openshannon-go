package commands

import (
	"context"

	"github.com/onedayallday33-a11y/openshannon-go/internal/agent"
)

// CommandResult represents the outcome of a slash command
type CommandResult struct {
	DirectOutput string // Output to print directly to user
	PromptText   string // If non-empty, this text is sent to the LLM agent
	IsHandled    bool   // True if command was found and executed
}

// SlashCommand defines the interface for all /commands
type SlashCommand interface {
	Name() string
	Description() string
	Execute(ctx context.Context, a *agent.Agent, args []string) (CommandResult, error)
}
