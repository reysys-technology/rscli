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
		// main prints the error and picks the exit code, so cobra must not also
		// print it. Usage is for misuse of flags, not for a failed upload — in a
		// CI log the usage block buries the one line that matters.
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	command.AddCommand(account.Command)
	command.AddCommand(configure.Command)
	command.AddCommand(trivy.Command)

	return command
}
