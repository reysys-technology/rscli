package uploadtrivycontainerimagescan

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/reysys-technology/rscli/pkg"
	"github.com/reysys-technology/rscli/pkg/api"

	"github.com/spf13/cobra"
)

// ingestPath is the backend's Trivy ingest endpoint. It accepts a raw Trivy
// JSON report for either artifact type — a container image or a source
// repository — and decides which from the report's own ArtifactType field.
const ingestPath = "/trivy.json.ingest"

var scanFilePath string

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
			"Credentials come from RS_CLIENT_ID and RS_CLIENT_SECRET; see `rscli configure`.",
		RunE: run,
	}
	cmd.Flags().StringVarP(&scanFilePath, "file", "f", "", "Path to the Trivy JSON scan result file (required)")
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

	if _, err := client.PostJSON(cmd.Context(), ingestPath, scanData); err != nil {
		return err
	}

	fmt.Printf("Uploaded %s (%s) to %s\n", report.ArtifactName, report.ArtifactType, cfg.BaseURL)
	return nil
}
