package account

import (
	getaccountinformation "github.com/reysys-technology/rscli/pkg/command/account/get-account-information"

	"github.com/spf13/cobra"
)

var Command = func() *cobra.Command {
	cmd := &cobra.Command{
		Use: "account",
	}
	cmd.AddCommand(getaccountinformation.Command)
	return cmd
}()
