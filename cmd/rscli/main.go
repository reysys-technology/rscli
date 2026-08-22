package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/reysys-technology/rscli/pkg/command"
)

// exitCoder lets a command choose its own exit code. A pipeline has to be able
// to tell "your artifact did not meet the policy" from "the tool could not do
// its job" — if both are exit 1, the only way to keep a pipeline green during
// an outage is to stop gating, and that is how a security control disappears.
type exitCoder interface {
	ExitCode() int
}

func main() {
	root := command.Root()

	if err := root.Execute(); err != nil {
		var coder exitCoder
		if errors.As(err, &coder) {
			// The command already printed the detail. Keep this to one line so
			// it does not bury it in the CI log.
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(coder.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
