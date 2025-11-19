package main

import (
	"fmt"
	"os"

	"github.com/reysys-technology/cli/pkg/account"
	"github.com/reysys-technology/cli/pkg/config"
	"github.com/reysys-technology/cli/pkg/trivy"

	"github.com/spf13/cobra"
)

var command = &cobra.Command{
	Use: "cli",
}

func init() {
	command.PersistentFlags().StringVar(&config.BaseURL, "url", "http://localhost:9670", "Base URL for API requests")
	command.AddCommand(account.Command)
	command.AddCommand(trivy.Command)
}

func main() {
	if err := config.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
