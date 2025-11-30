package command

import (
	"runtime/debug"

	"github.com/reysys-technology/rscli/pkg/command/account"
	"github.com/reysys-technology/rscli/pkg/command/configure"
	"github.com/reysys-technology/rscli/pkg/command/trivy"

	"github.com/spf13/cobra"
)

func Root() *cobra.Command {
	buildInfo, _ := debug.ReadBuildInfo()
	version := "unknown"
	if buildInfo != nil && buildInfo.Main.Version != "" {
		version = buildInfo.Main.Version
	}
	command := &cobra.Command{
		Use:     "rscli",
		Version: version,
	}
	command.AddCommand(account.Command)
	command.AddCommand(configure.Command)
	command.AddCommand(trivy.Command)

	return command
}
