package configure

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "configure",
	Short: "How to give rscli credentials and an endpoint",
	Long: "Configure rscli.\n\n" +
		"rscli authenticates with an OAuth2 client-credentials pair. Provision one in\n" +
		"the console under Account Info; the secret is shown once, so copy it then.\n\n" +
		"## Environment variables\n\n" +
		"| Variable | Description |\n" +
		"|----------|-------------|\n" +
		"| RS_CLIENT_ID | API client id (required) |\n" +
		"| RS_CLIENT_SECRET | API client secret (required) |\n" +
		"| RS_BASE_URL | API base URL (default: https://api.reysys.com). Must be https |\n" +
		"| RS_TOKEN_URL | Token endpoint (default: https://accounts.reysys.com/realms/accounts/protocol/openid-connect/token). Must be https |\n" +
		"| RS_INSECURE_SKIP_VERIFY | Skip TLS verification. Local development only — never in CI |\n" +
		"| RS_HTTP_TIMEOUT | How long to wait for one API request (default: 10m) |\n\n" +
		"The ingest is synchronous, so an upload waits for the server to store the\n" +
		"whole report — not for the bytes to travel. A large image report can take\n" +
		"minutes. If an upload times out, the server may still have finished: check\n" +
		"the console before assuming it did not, and raise RS_HTTP_TIMEOUT (for\n" +
		"example 20m) if it recurs.\n\n" +
		"RS_SECRET_ID and RS_SECRET are accepted as aliases for RS_CLIENT_ID and\n" +
		"RS_CLIENT_SECRET, so existing pipelines keep working.\n\n" +
		"## Config file\n\n" +
		"~/.reysys/config.yaml or ./config.yaml:\n\n" +
		"```yaml\n" +
		"client_id: your-client-id\n" +
		"client_secret: your-client-secret\n" +
		"base_url: https://api.reysys.com\n" +
		"```\n\n" +
		"Environment variables win over the file.\n\n" +
		"Both URLs must be https (http is allowed only on loopback, for local\n" +
		"development). Your credentials go to the token URL and the resulting\n" +
		"access token goes to the base URL, so whoever controls either one\n" +
		"receives something worth having.\n\n" +
		"## In a pipeline\n\n" +
		"Keep the credentials in the CI provider's secret store, never in the\n" +
		"repository. GitHub Actions:\n\n" +
		"```yaml\n" +
		"- run: trivy image --format json -o scan.json \"$IMAGE\"\n" +
		"- run: rscli trivy upload-trivy-container-image-scan -f scan.json\n" +
		"  env:\n" +
		"    RS_CLIENT_ID: ${{ secrets.RS_CLIENT_ID }}\n" +
		"    RS_CLIENT_SECRET: ${{ secrets.RS_CLIENT_SECRET }}\n" +
		"```\n\n" +
		"GitLab CI:\n\n" +
		"```yaml\n" +
		"scan:\n" +
		"  script:\n" +
		"    - trivy image --format json -o scan.json \"$IMAGE\"\n" +
		"    - rscli trivy upload-trivy-container-image-scan -f scan.json\n" +
		"  variables:\n" +
		"    RS_CLIENT_ID: $RS_CLIENT_ID\n" +
		"    RS_CLIENT_SECRET: $RS_CLIENT_SECRET\n" +
		"```\n\n" +
		"## Check it works\n\n" +
		"```bash\n" +
		"rscli account get-account-information\n" +
		"```",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Long)
	},
}
