package main

import (
	"fmt"
	"os"

	"cli/pkg/account"
	"cli/pkg/config"

	"github.com/spf13/cobra"
)

var command = &cobra.Command{
	Use: "cli",
}

func init() {
	command.PersistentFlags().StringVar(&config.BaseURL, "url", "http://localhost:9670", "Base URL for API requests")
	command.AddCommand(account.Command)
}

func main() {
	// Initialize configuration using Viper
	if err := config.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
