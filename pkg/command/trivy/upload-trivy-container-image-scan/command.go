package uploadtrivycontainerimagescan

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/reysys-technology/rscli/pkg"
	"github.com/reysys-technology/rscli/pkg/api"
	"github.com/reysys-technology/rscli/pkg/gate"

	"github.com/spf13/cobra"
)

// ingestPath is the backend's Trivy ingest endpoint. It accepts a raw Trivy
// JSON report for either artifact type — a container image or a source
// repository — and decides which from the report's own ArtifactType field.
const ingestPath = "/trivy.json.ingest"

var (
	scanFilePath string
	enforceGate  bool
)

var Command = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload-trivy-container-image-scan",
		Short: "Upload a Trivy JSON report to Reysys",
		Long: "Upload a Trivy JSON report to Reysys.\n\n" +
			"Produce the report with Trivy's JSON format, then upload it:\n\n" +
			"  trivy image --format json -o scan.json <image>\n" +
			"  rscli trivy upload-trivy-container-image-scan -f scan.json\n\n" +
			"Repository reports work the same way:\n\n" +
			"  trivy repo --format json -o scan.json .\n\n" +
			"Credentials come from RS_CLIENT_ID and RS_CLIENT_SECRET; see `rscli configure`.\n\n" +
			"With --gate, the command exits non-zero when the scan does not meet the policy\n" +
			"configured in the console, so the pipeline step fails:\n\n" +
			"  0  passed, or no policy is enforced for this target\n" +
			"  1  the tool could not do its job (bad config, upload failed, or Reysys\n" +
			"     could not evaluate the scan) — never a statement about your code\n" +
			"  2  the artifact did not meet the policy\n" +
			"  3  the scan could not be judged, and the policy says not to allow that\n\n" +
			"Without --gate the verdict is printed and the command still succeeds, so a\n" +
			"team can watch what the gate would do before letting it block anything.",
		RunE:          run,
		SilenceUsage:  true,
		SilenceErrors: false,
	}
	cmd.Flags().StringVarP(&scanFilePath, "file", "f", "", "Path to the Trivy JSON scan result file (required)")
	cmd.Flags().BoolVar(&enforceGate, "gate", false, "Fail this command when the scan does not meet the configured policy")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}()

func run(cmd *cobra.Command, args []string) error {
	scanData, err := os.ReadFile(scanFilePath)
	if err != nil {
		return fmt.Errorf("reading scan file: %w", err)
	}

	// Fail on a malformed report here rather than sending it: the ingest
	// endpoint answers 200 for several unusable-input cases, so a bad file
	// would otherwise look like a successful upload.
	var report struct {
		ArtifactName string `json:"ArtifactName"`
		ArtifactType string `json:"ArtifactType"`
	}
	if err := json.Unmarshal(scanData, &report); err != nil {
		return fmt.Errorf("%s is not valid Trivy JSON: %w", scanFilePath, err)
	}
	if report.ArtifactType != "container_image" && report.ArtifactType != "repository" {
		return fmt.Errorf(
			"%s has ArtifactType %q; the backend accepts only \"container_image\" or \"repository\". "+
				"Produce it with `trivy image --format json` or `trivy repo --format json`",
			scanFilePath, report.ArtifactType)
	}

	cfg, err := pkg.Load()
	if err != nil {
		return err
	}
	client, err := api.New(cmd.Context(), cfg)
	if err != nil {
		return err
	}

	responseBody, err := client.PostJSON(cmd.Context(), ingestPath, scanData)
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "Uploaded %s (%s) to %s\n", report.ArtifactName, report.ArtifactType, cfg.BaseURL)

	// A backend that predates the gate returns an empty body, and an older one
	// returns something we do not recognise. Neither is an error: no verdict
	// means no gate, and the upload still succeeded.
	var response gate.Response
	if len(responseBody) > 0 {
		_ = json.Unmarshal(responseBody, &response)
	}
	response.Gate.Render(out)

	if code := response.Gate.ExitCode(enforceGate); code != gate.ExitPassed {
		return &exitCodeError{code: code, verdict: response.Gate}
	}
	return nil
}

// exitCodeError carries a specific exit code out of RunE. The message is
// deliberately terse: the verdict itself was already printed in full above, and
// repeating it under an "Error:" prefix would bury the part that matters.
type exitCodeError struct {
	code    int
	verdict *gate.Verdict
}

func (e *exitCodeError) Error() string {
	if e.verdict != nil && e.verdict.Action == gate.ActionUnevaluated {
		return "the scan could not be judged against the policy"
	}
	return "the artifact did not meet the policy"
}

func (e *exitCodeError) ExitCode() int { return e.code }
