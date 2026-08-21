package getaccountinformation

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/reysys-technology/rscli/pkg"
	"github.com/reysys-technology/rscli/pkg/api"

	"github.com/spf13/cobra"
)

const describePath = "/account.describe"

var Command = &cobra.Command{
	Use:   "get-account-information",
	Short: "Show the account the current credentials belong to",
	Long: "Show the account the current credentials belong to.\n\n" +
		"Useful as a connectivity check: it proves the credentials are valid and\n" +
		"that the configured base URL is reachable, without uploading anything.",
	RunE: run,
}

func run(cmd *cobra.Command, args []string) error {
	cfg, err := pkg.Load()
	if err != nil {
		return err
	}
	client, err := api.New(cmd.Context(), cfg)
	if err != nil {
		return err
	}

	body, err := client.PostJSON(cmd.Context(), describePath, []byte("{}"))
	if err != nil {
		return err
	}

	// Re-indent so the output is readable in a terminal and in a CI log.
	var pretty bytes.Buffer
	if json.Indent(&pretty, body, "", "  ") == nil {
		fmt.Println(pretty.String())
		return nil
	}
	fmt.Println(string(body))
	return nil
}
