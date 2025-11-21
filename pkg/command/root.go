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

	command.AddCommand(account.Command)
	command.AddCommand(trivy.Command)

	return command
}
