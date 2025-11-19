package main

import (
	"fmt"
	"os"

	"github.com/reysys-technology/rscli/pkg/account"
	"github.com/reysys-technology/rscli/pkg/config"
	"github.com/reysys-technology/rscli/pkg/trivy"

	"github.com/spf13/cobra"
)

var version = "dev"

var command = &cobra.Command{
	Use:     "rscli",
	Version: version,
}

func init() {
	command.PersistentFlags().StringVar(&config.BaseURL, "url", "http://localhost:9670", "Base URL for API requests")
	command.AddCommand(account.Command)
	command.AddCommand(trivy.Command)
}

func main() {
	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
