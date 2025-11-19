package trivy

import (
	uploadtrivycontainerimagescan "github.com/reysys-technology/cli/pkg/trivy/upload-trivy-container-image-scan"

	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use: "trivy",
}

func init() {
	Command.AddCommand(uploadtrivycontainerimagescan.Command)
}
