package account

import (
	getaccountinformation "github.com/reysys-technology/cli/pkg/account/get-account-information"

	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use: "account",
}

func init() {
	Command.AddCommand(getaccountinformation.Command)
}
