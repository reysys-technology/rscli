package trivy

import (
	uploadtrivycontainerimagescan "github.com/reysys-technology/rscli/pkg/command/trivy/upload-trivy-container-image-scan"

	"github.com/spf13/cobra"
)

var Command = func() *cobra.Command {
	cmd := &cobra.Command{
		Use: "trivy",
	}
	cmd.AddCommand(uploadtrivycontainerimagescan.Command)
	return cmd
}()
