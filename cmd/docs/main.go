package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/reysys-technology/rscli/pkg/command"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var version = "dev"

func main() {
	root := command.Root(version)
	root.DisableAutoGenTag = true

	outputDir := "dist/docs"
	if len(os.Args) > 1 {
		outputDir = os.Args[1]
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatal(err)
	}

	if err := genMarkdownTreeStructure(root, outputDir); err != nil {
		log.Fatal(err)
	}
}

func genMarkdownTreeStructure(cmd *cobra.Command, dir string) error {
	var filename string

	if cmd.HasSubCommands() {
		// Commands with subcommands become sections with _index.md
		subDir := filepath.Join(dir, cmd.Name())
		if err := os.MkdirAll(subDir, 0755); err != nil {
			return err
		}
		filename = filepath.Join(subDir, "_index.md")

		for _, c := range cmd.Commands() {
			if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
				continue
			}
			if err := genMarkdownTreeStructure(c, subDir); err != nil {
				return err
			}
		}
	} else {
		// Leaf commands are regular .md files
		filename = filepath.Join(dir, cmd.Name()+".md")
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// Convert rscli_account_list.md to rscli/account/list.md
	linkHandler := func(name string) string {
		base := strings.TrimSuffix(name, filepath.Ext(name))
		return strings.ReplaceAll(base, "_", "/") + ".md"
	}

	buf := new(bytes.Buffer)
	if err := doc.GenMarkdownCustom(cmd, buf, linkHandler); err != nil {
		return err
	}

	// Strip the first header line (## command path)
	content := buf.Bytes()
	if idx := bytes.Index(content, []byte("\n")); idx != -1 {
		content = content[idx+1:]
	}

	// Add Hugo front matter
	frontMatter := fmt.Sprintf(`---
title: "%s"
description: "%s"
---

`, cmd.Name(), cmd.Short)

	if _, err := f.WriteString(frontMatter); err != nil {
		return err
	}

	_, err = f.Write(content)
	return err
}
