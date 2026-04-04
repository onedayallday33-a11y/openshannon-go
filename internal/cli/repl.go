package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/onedayallday33-a11y/openshannon-go/internal/agent"
	"github.com/onedayallday33-a11y/openshannon-go/internal/types"
	tea "github.com/charmbracelet/bubbletea"
)

// REPL structure for interactive loop
type REPL struct {
	Agent *agent.Agent
}

// NewREPL initializes a new REPL
func NewREPL(a *agent.Agent) *REPL {
	return &REPL{Agent: a}
}

// LaunchTUI starts the Bubble Tea based UI
func (r *REPL) LaunchTUI() error {
	p := tea.NewProgram(NewModel(r.Agent), tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}

// Run starts the interactive loop
func (r *REPL) Run() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("OpenShannon-Go REPL (type /help for commands)\n")
	fmt.Printf("Use '\\' at end of line for multi-line input.\n\n")

	// Setup signal handling
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigs)

	for {
		fmt.Printf("> ")
		var lines []string
		
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasSuffix(line, "\\") {
				lines = append(lines, strings.TrimSuffix(line, "\\"))
				fmt.Printf(">> ")
				continue
			}
			lines = append(lines, line)
			break
		}

		if err := scanner.Err(); err != nil {
			fmt.Printf("Scanner error: %v\n", err)
			break
		}

		// If scanner closed (EOF), exit REPL
		if len(lines) == 0 && scanner.Text() == "" {
			// Check if we just hit CTRL+C which interrupted Scan()
			select {
			case <-sigs:
				fmt.Println()
				continue // Just a new prompt
			default:
				fmt.Println("Goodbye!")
				return
			}
		}

		prompt := strings.Join(lines, "\n")
		if strings.TrimSpace(prompt) == "" {
			continue
		}

		if strings.ToLower(prompt) == "exit" || strings.ToLower(prompt) == "quit" {
			fmt.Println("Goodbye!")
			break
		}

		// Handle execution with context and cancellation support
		turnCtx, cancel := context.WithCancel(context.Background())
		
		// Wait for signal in background to cancel current turn
		go func() {
			select {
			case <-sigs:
				cancel()
			case <-turnCtx.Done():
			}
		}()

		output, err := r.Agent.Run(turnCtx, prompt, func(ev types.AgentEvent) {
			switch ev.Type {
			case types.EventTextDelta:
				fmt.Print(ev.Text)
			case types.EventToolStart:
				fmt.Printf("\n[Tool: %s] Calling...\n", ev.Tool.Name)
			case types.EventToolEnd:
				fmt.Printf("\n[Tool: %s] Completed.\n", ev.Tool.Name)
			case types.EventThinkingStart:
				fmt.Printf("\n[Thinking] ")
			}
		})

		if err != nil {
			if err == context.Canceled {
				fmt.Println("\n[Cancelled]")
			} else {
				fmt.Printf("Error: %v\n", err)
			}
		} else {
			if output == "" {
				// If no text was produced but agent finished
				fmt.Println()
			}
		}
		
		cancel() // cleanup context
		// Drain signal if any was sent during execution but not caught by goroutine
		select {
		case <-sigs:
		default:
		}
		fmt.Println()
	}
}
