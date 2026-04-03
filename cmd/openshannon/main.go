package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/onedayallday33-a11y/openshannon-go/internal/agent"
	"github.com/onedayallday33-a11y/openshannon-go/internal/cli"
	"github.com/onedayallday33-a11y/openshannon-go/internal/commands"
	"github.com/onedayallday33-a11y/openshannon-go/internal/config"
	"github.com/onedayallday33-a11y/openshannon-go/internal/memory"
	"github.com/onedayallday33-a11y/openshannon-go/internal/profile"
	"github.com/onedayallday33-a11y/openshannon-go/internal/tools"
	"github.com/onedayallday33-a11y/openshannon-go/internal/types"
)

func main() {
	modelFlag := flag.String("model", "claude-3-5-sonnet-20241022", "LLM model to use")
	profileFlag := flag.String("profile", "", "Profile name to load")
	noTUIFlag := flag.Bool("no-tui", false, "Disable the premium TUI and use simple line-based REPL")
	flag.Parse()

	// 0. Initialize Config
	config.InitConfig()

	// 1. Initialize Registry & Commands
	disp := agent.GetDispatcher()
	disp.Register(&commands.HelpCommand{})
	disp.Register(&commands.DoctorCommand{})
	disp.Register(&commands.CompactCommand{})
	disp.Register(&commands.ReviewCommand{})
	disp.Register(&commands.CommitCommand{})
	disp.Register(&commands.ClearCommand{})
	disp.Register(&commands.ExitCommand{})
	disp.Register(&commands.ModelCommand{})

	// 2. Load Profile
	pm := profile.NewProfileManager()
	var activeProfile profile.Profile
	if *profileFlag != "" {
		p, err := pm.GetProfile(*profileFlag)
		if err != nil {
			fmt.Printf("Warning: Profile %s not found. Using defaults.\n", *profileFlag)
		} else {
			activeProfile = p
		}
	} else {
		p, _ := pm.GetDefaultProfile()
		activeProfile = p
	}

	// 3. Initialize Agent
	selectedModel := *modelFlag
	if config.UseOpenAI() {
		if envModel := config.OpenAIModel(); envModel != "" && envModel != "gpt-4o" {
			selectedModel = envModel
		}
	}

	agentCfg := types.AgentConfig{
		Model:    selectedModel,
		System:   "You are OpenShannon, an AI coding assistant. You have access to tools to help you build and debug software.",
		MaxTurns: 15,
	}
	
	if activeProfile.Model != "" {
		agentCfg.Model = activeProfile.Model
	}
	if activeProfile.System != "" {
		agentCfg.System = activeProfile.System
	}

	// Register Core Tools
	agentCfg.Tools = append(agentCfg.Tools, 
		&tools.FileReadTool{}, 
		&tools.FileWriteTool{}, 
		&tools.FileEditTool{},
		&tools.BashTool{},
		&tools.GrepTool{},
		&tools.GlobTool{},
		&tools.WebFetchTool{},
		&tools.WebSearchTool{},
	)
	
	sessionID := fmt.Sprintf("session_%d", time.Now().Unix())
	a := agent.NewAgent(sessionID, agentCfg)

	// Add Auto-save hook
	memManager := memory.NewPersistenceManager()
	a.OnTurnEnd = func(ag *agent.Agent) {
		_ = memManager.AutoSave(ag)
	}

	// 4. Start UI
	repl := cli.NewREPL(a)
	if *noTUIFlag {
		repl.Run()
	} else {
		if err := repl.LaunchTUI(); err != nil {
			fmt.Printf("Error launching TUI: %v\nFalling back to simple REPL...\n", err)
			repl.Run()
		}
	}
}
