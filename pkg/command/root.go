package command

import (
	"github.com/reysys-technology/rscli/pkg/command/account"
	"github.com/reysys-technology/rscli/pkg/command/trivy"
	"github.com/reysys-technology/rscli/pkg/config"

	"github.com/spf13/cobra"
)

func Root(version string) *cobra.Command {
	command := &cobra.Command{
		Use:     "rscli",
		Version: version,
	}

	command.PersistentFlags().StringVar(&config.BaseURL, "url", config.BaseURL, "Base URL for API requests (default from RS_BASE_URL or http://localhost:9670)")
	command.AddCommand(account.Command)
	command.AddCommand(trivy.Command)

	return command
}
