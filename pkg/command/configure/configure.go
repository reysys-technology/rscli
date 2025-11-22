package configure

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "configure",
	Short: "Configure rscli credentials and settings",
	Long: "Configure rscli credentials and settings.\n\n" +
		"rscli can be configured using environment variables or a config file.\n\n" +
		"## Environment Variables\n\n" +
		"| Variable | Description |\n" +
		"|----------|-------------|\n" +
		"| RS_SECRET_ID | Your Reysys API secret ID (required) |\n" +
		"| RS_SECRET | Your Reysys API secret (required) |\n" +
		"| RS_BASE_URL | API base URL (default: http://localhost:9670) |\n\n" +
		"## Config File\n\n" +
		"Create a config file at ~/.reysys/config.yaml or ./config.yaml:\n\n" +
		"```yaml\n" +
		"secret_id: your-secret-id\n" +
		"secret: your-secret\n" +
		"base_url: https://api.reysys.com\n" +
		"```\n\n" +
		"## Priority\n\n" +
		"Environment variables take precedence over config file values.\n\n" +
		"## Example\n\n" +
		"```bash\n" +
		"# Using environment variables\n" +
		"export RS_SECRET_ID=my-id\n" +
		"export RS_SECRET=my-secret\n" +
		"rscli account get-account-information\n\n" +
		"# Using config file\n" +
		"mkdir -p ~/.reysys\n" +
		"cat > ~/.reysys/config.yaml << EOF\n" +
		"secret_id: my-id\n" +
		"secret: my-secret\n" +
		"EOF\n" +
		"rscli account get-account-information\n" +
		"```",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Long)
	},
}
