package main

import (
	"fmt"
	"os"

	"github.com/reysys-technology/rscli/pkg/command"
)

var version = "dev"

func main() {
	root := command.Root(version)

	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
