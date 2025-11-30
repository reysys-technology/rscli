package main

import (
	"fmt"
	"os"

	"github.com/reysys-technology/rscli/pkg/command"
)

func main() {
	root := command.Root()

	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
