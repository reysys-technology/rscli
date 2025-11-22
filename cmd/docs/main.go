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

	if err := genTree(root, root, outputDir, true); err != nil {
		log.Fatal(err)
	}
}

func genTree(cmd, root *cobra.Command, dir string, isRoot bool) error {
	if cmd.HasSubCommands() {
		var subDir string
		if isRoot {
			subDir = dir
		} else {
			subDir = filepath.Join(dir, cmd.Name())
			if err := os.MkdirAll(subDir, 0755); err != nil {
				return err
			}
		}

		for _, c := range cmd.Commands() {
			if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
				continue
			}
			if err := genTree(c, root, subDir, false); err != nil {
				return err
			}
		}
		return nil
	}

	filename := filepath.Join(dir, cmd.Name()+".md")

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// Front matter (YAML for Docusaurus)
	frontMatter := fmt.Sprintf("---\ntitle: %q\ndescription: %q\n---\n\n", cmd.Name(), cmd.Short)
	if _, err := f.WriteString(frontMatter); err != nil {
		return err
	}

	// Get current command's path parts (excluding root)
	var currentParts []string
	for p := cmd; p.Parent() != nil; p = p.Parent() {
		currentParts = append([]string{p.Name()}, currentParts...)
	}

	// Link handler that generates correct relative paths
	linkHandler := func(name string) string {
		base := strings.TrimSuffix(name, filepath.Ext(name))
		parts := strings.Split(base, "_")
		targetParts := parts[1:] // Strip root

		// Find target command to check if it has subcommands
		target := root
		for _, part := range targetParts {
			for _, c := range target.Commands() {
				if c.Name() == part {
					target = c
					break
				}
			}
		}

		// Find common prefix length
		commonLen := 0
		for i := 0; i < len(currentParts) && i < len(targetParts); i++ {
			if currentParts[i] == targetParts[i] {
				commonLen++
			} else {
				break
			}
		}

		// Calculate relative path
		ups := len(currentParts) - commonLen
		downs := targetParts[commonLen:]

		var result string
		if ups > 0 {
			result = strings.Repeat("../", ups)
		}
		if len(downs) > 0 {
			result += strings.Join(downs, "/")
		}

		// Docusaurus uses .md extensions
		if result == "" {
			return target.Name() + ".md"
		}
		return result + ".md"
	}

	// Generate content
	buf := new(bytes.Buffer)
	if err := doc.GenMarkdownCustom(cmd, buf, linkHandler); err != nil {
		return err
	}

	// Strip first header line
	content := buf.Bytes()
	if idx := bytes.Index(content, []byte("\n")); idx != -1 {
		content = content[idx+1:]
	}

	// Remove SEE ALSO section (Docusaurus handles navigation)
	if idx := bytes.Index(content, []byte("### SEE ALSO")); idx != -1 {
		content = content[:idx]
	}

	_, err = f.Write(content)
	return err
}
