package main

import (
	"fmt"
	"os"

	"github.com/reysys-technology/rscli/pkg/command"
	"github.com/reysys-technology/rscli/pkg/config"
)

var version = "dev"

func main() {
	if err := config.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	root := command.Root(version)

	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
