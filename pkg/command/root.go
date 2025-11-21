package command

import (
	"github.com/reysys-technology/rscli/pkg/command/account"
	"github.com/reysys-technology/rscli/pkg/command/trivy"

	"github.com/spf13/cobra"
)

func Root(version string) *cobra.Command {
	command := &cobra.Command{
		Use:     "rscli",
		Version: version,
	}

	// URL flag removed - use RS_BASE_URL env var or config file
	command.AddCommand(account.Command)
	command.AddCommand(trivy.Command)

	return command
}
