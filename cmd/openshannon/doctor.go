package main

import (
	"fmt"
	"os"

	"github.com/onedayallday33-a11y/openshannon-go/internal/config"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system health and environment configurations",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running doctor checks...")
		
		allSpecsMet := true

		if config.UseOpenAI() {
			fmt.Println("✅ CLAUDE_CODE_USE_OPENAI is enabled")
			if key := config.OpenAIApiKey(); key == "" {
                 // OpenAIApiKey could be optional for Local Models, but we warn.
				fmt.Println("⚠️  OPENAI_API_KEY is not set (might be OK for local models)")
			} else {
				fmt.Println("✅ OPENAI_API_KEY is present")
			}
			fmt.Printf("ℹ️  Provider URL: %s\n", config.OpenAIBaseURL())
			fmt.Printf("ℹ️  Model: %s\n", config.OpenAIModel())

		} else {
			fmt.Println("ℹ️  CLAUDE_CODE_USE_OPENAI is not enabled")
		}

		if allSpecsMet {
			fmt.Println("\nConfiguration checks passed.")
		} else {
			fmt.Println("\nSome checks failed. Please review the missing variables.")
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
