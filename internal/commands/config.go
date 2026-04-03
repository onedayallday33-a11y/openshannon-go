package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/onedayallday33-a11y/openshannon-go/internal/agent"
	"github.com/onedayallday33-a11y/openshannon-go/internal/config"
	"github.com/spf13/viper"
)

type ConfigCommand struct{}

func (c *ConfigCommand) Name() string        { return "config" }
func (c *ConfigCommand) Description() string { return "View or edit configuration" }

func (c *ConfigCommand) Execute(ctx context.Context, a *agent.Agent, args []string) (agent.DirectCommandResult, error) {
	if len(args) == 0 {
		// Show all config
		keys := viper.AllKeys()
		var sb strings.Builder
		sb.WriteString("Current Configuration:\n\n")
		for _, k := range keys {
			sb.WriteString(fmt.Sprintf("%-20s: %v\n", k, viper.Get(k)))
		}
		return agent.DirectCommandResult{
			DirectOutput: sb.String(),
			IsHandled:    true,
		}, nil
	}

	if len(args) == 1 {
		// Show specific key
		key := args[0]
		val := viper.Get(key)
		if val == nil {
			return agent.DirectCommandResult{
				DirectOutput: fmt.Sprintf("Config key '%s' not found.", key),
				IsHandled:    true,
			}, nil
		}
		return agent.DirectCommandResult{
			DirectOutput: fmt.Sprintf("%s: %v", key, val),
			IsHandled:    true,
		}, nil
	}

	if len(args) >= 2 {
		// Set key
		key := args[0]
		val := strings.Join(args[1:], " ")
		viper.Set(key, val)
		err := config.SaveConfig() // Assuming this exists or works with Viper
		if err != nil {
			return agent.DirectCommandResult{
				DirectOutput: fmt.Sprintf("Updated %s to %s (temporarily, failed to save: %v)", key, val, err),
				IsHandled:    true,
			}, nil
		}
		return agent.DirectCommandResult{
			DirectOutput: fmt.Sprintf("Updated %s to %s and saved.", key, val),
			IsHandled:    true,
		}, nil
	}

	return agent.DirectCommandResult{IsHandled: false}, nil
}
