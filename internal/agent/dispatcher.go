package agent

import (
	"context"
	"strings"
	"sync"
)

// CommandResult represents the outcome of a slash command
// (Redefining here briefly to avoid circular import if needed, 
// or importing from a shared types package. Let's use internal/commands/interface)
// Actually, to avoid circular import (Agent <-> Commands), 
// we will define the Dispatcher to return a simple structure.

type DirectCommandResult struct {
	DirectOutput string
	PromptText   string
	IsHandled    bool
}

type SlashCommand interface {
	Name() string
	Description() string
	Execute(ctx context.Context, a *Agent, args []string) (DirectCommandResult, error)
}

type CommandDispatcher struct {
	mu       sync.Mutex
	registry map[string]SlashCommand
}

var (
	defaultDispatcher *CommandDispatcher
	dispOnce          sync.Once
)

func GetDispatcher() *CommandDispatcher {
	dispOnce.Do(func() {
		defaultDispatcher = &CommandDispatcher{
			registry: make(map[string]SlashCommand),
		}
	})
	return defaultDispatcher
}

func (d *CommandDispatcher) Register(cmd SlashCommand) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.registry[cmd.Name()] = cmd
}

// GetRegisteredCommands returns all commands for UI suggestions
func (d *CommandDispatcher) GetRegisteredCommands() []SlashCommand {
	d.mu.Lock()
	defer d.mu.Unlock()
	cmds := make([]SlashCommand, 0, len(d.registry))
	for _, cmd := range d.registry {
		cmds = append(cmds, cmd)
	}
	return cmds
}

// Dispatch checks if input is a slash command and executes it
func (d *CommandDispatcher) Dispatch(ctx context.Context, a *Agent, input string) (DirectCommandResult, error) {
	if !strings.HasPrefix(input, "/") {
		return DirectCommandResult{IsHandled: false}, nil
	}

	parts := strings.Fields(input[1:])
	if len(parts) == 0 {
		return DirectCommandResult{IsHandled: false}, nil
	}

	cmdName := parts[0]
	args := parts[1:]

	d.mu.Lock()
	cmd, ok := d.registry[cmdName]
	d.mu.Unlock()

	if !ok {
		return DirectCommandResult{
			DirectOutput: "Unknown command: /" + cmdName + ". Type /help for list.",
			IsHandled:    true,
		}, nil
	}

	return cmd.Execute(ctx, a, args)
}
