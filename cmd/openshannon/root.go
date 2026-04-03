package main

import (
	"context"

	"github.com/onedayallday33-a11y/openshannon-go/internal/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "openshannon",
	Short: "OpenShannon-Go is a Go-native 100% feature-complete implementation of OpenClaude",
	Long:  `OpenShannon-Go is a fast, standalone Go rewrite of the OpenClaude-TS tool with zero runtime dependencies.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Output default ketika dijalankan tanpa command
		cmd.Help()
	},
}

func Execute() error {
	return rootCmd.ExecuteContext(context.Background())
}

func init() {
	cobra.OnInitialize(config.InitConfig)
}
