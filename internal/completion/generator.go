// Package completion provides shell completion generation logic for jpd commands.
// This package extracts completion generators from the cmd package to improve
// code organization and maintainability.
package completion

import (
	// standard library
	"fmt"
	"io"
	"os"
	"strings"

	// external
	"github.com/spf13/cobra"

	// internal
	"github.com/louiss0/javascript-package-delegator/internal/integrations"
)

// Generator provides methods for generating shell completions with optional shorthand aliases.
type Generator interface {
	// GenerateCompletion generates completion script for the specified shell and writes it to the output writer.
	// If withShorthand is true, shorthand alias functions are appended to the completion output.
	GenerateCompletion(cmd *cobra.Command, shell string, filename string, withShorthand bool) error

	// GetSupportedShells returns a list of supported shell names.
	GetSupportedShells() []string

	// GetDefaultAliasMapping returns the default alias mapping for shorthand generation.
	GetDefaultAliasMapping() map[string][]string
}

// generator is the concrete implementation of the Generator interface.
type generator struct {
	aliasGenerator        integrations.AliasGenerator
	warpGenerator         integrations.WarpGenerator
	carapaceSpecGenerator integrations.CarapaceSpecGenerator
}

// NewGenerator creates a new completion generator instance.
func NewGenerator() Generator {
	return &generator{
		aliasGenerator:        integrations.NewAliasGenerator(),
		warpGenerator:         integrations.NewWarpGenerator(),
		carapaceSpecGenerator: integrations.NewCarapaceSpecGenerator(),
	}
}

// GetSupportedShells returns the list of supported shells.
func (g *generator) GetSupportedShells() []string {
	return []string{
		"bash",
		"carapace",
		"fish",
		"nushell",
		"powershell",
		"warp",
		"zsh",
	}
}

// GetDefaultAliasMapping returns the default alias mapping for jpd commands.
func (g *generator) GetDefaultAliasMapping() map[string][]string {
	return map[string][]string{
		"install":       {"jpi", "jpadd", "jpd-install"},
		"run":           {"jpr", "jpd-run"},
		"exec":          {"jpe", "jpd-exec"},
		"dlx":           {"jpx", "jpd-dlx"},
		"update":        {"jpu", "jpup", "jpupgrade", "jpd-update"},
		"uninstall":     {"jpun", "jprm", "jpremove", "jpd-uninstall"},
		"clean-install": {"jpci", "jpd-clean-install"},
		"agent":         {"jpa", "jpd-agent"},
	}
}

// GenerateCompletion generates completion script for the specified shell.
func (g *generator) GenerateCompletion(cmd *cobra.Command, shell string, filename string, withShorthand bool) error {
	var outputWriter io.Writer
	var file *os.File // To hold the *os.File if we create one

	// Determine output destination based on filename
	if filename == "" {
		outputWriter = cmd.OutOrStdout()
	} else {
		// Create or open the specified file for writing
		f, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("failed to create output file %s: %w", filename, err)
		}
		outputWriter = f
		file = f // Store the file handle to defer closing
	}

	// Ensure the file is closed if it was opened
	if file != nil {
		defer func() {
			if cerr := file.Close(); cerr != nil {
				// Log the close error to stderr, as we might already be returning another error.
				fmt.Fprintf(os.Stderr, "warning: failed to close completion file %s: %v\n", filename, cerr)
			}
		}()
	}

	// Generate base completion script for the shell
	var completionErr error
	switch shell {
	case "bash":
		completionErr = cmd.GenBashCompletionV2(outputWriter, false)
	case "zsh":
		completionErr = cmd.GenZshCompletion(outputWriter)
	case "fish":
		completionErr = cmd.GenFishCompletion(outputWriter, false)
	case "powershell":
		completionErr = cmd.GenPowerShellCompletionWithDesc(outputWriter)
	case "nushell":
		_, completionErr = fmt.Fprint(outputWriter, integrations.GetNushellCompletionScript())
		if completionErr != nil {
			completionErr = fmt.Errorf("failed to write Nushell completion script: %w", completionErr)
		}
	case "carapace":
		spec, err := g.carapaceSpecGenerator.GenerateYAMLSpec(cmd)
		if err != nil {
			completionErr = fmt.Errorf("failed to generate Carapace YAML spec: %w", err)
		} else {
			_, completionErr = fmt.Fprint(outputWriter, spec)
			if completionErr != nil {
				completionErr = fmt.Errorf("failed to write Carapace YAML spec: %w", completionErr)
			}
		}
	case "warp":
		// Handle warp workflow generation
		if filename == "" {
			// Output multi-doc YAML to stdout
			multiDoc, err := g.warpGenerator.RenderJPDWorkflowsMultiDoc()
			if err != nil {
				completionErr = fmt.Errorf("failed to generate Warp workflows multi-doc: %w", err)
			} else {
				_, completionErr = fmt.Fprint(outputWriter, multiDoc)
				if completionErr != nil {
					completionErr = fmt.Errorf("failed to write Warp workflows multi-doc: %w", completionErr)
				}
			}
		} else {
			// Check if filename is a directory or ends with /
			fileInfo, err := os.Stat(filename)
			isDir := err == nil && fileInfo.IsDir()
			endsWithSlash := strings.HasSuffix(filename, "/")

			if isDir || endsWithSlash {
				// Generate individual workflow files in directory
				if file != nil {
					file.Close() // Close the single file since we're writing multiple files
					file = nil
				}
				completionErr = g.warpGenerator.GenerateJPDWorkflows(filename)
				if completionErr != nil {
					completionErr = fmt.Errorf("failed to generate Warp workflow files: %w", completionErr)
				}
			} else {
				// Error: filename looks like a file, not a directory
				completionErr = fmt.Errorf("warp target requires a directory (not a file) when using --filename; use a directory path ending with '/' or omit --filename for stdout output")
			}
		}
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}

	if completionErr != nil {
		return completionErr
	}

	// If --with-shorthand flag is set, append alias functions
	// Note: --with-shorthand is ignored for carapace and warp targets
	if withShorthand && shell != "carapace" && shell != "warp" {
		aliasMap := g.GetDefaultAliasMapping()

		// Generate alias block based on shell type
		var aliasBlock string
		switch shell {
		case "bash":
			aliasBlock = g.aliasGenerator.GenerateBash(aliasMap)
		case "zsh":
			aliasBlock = g.aliasGenerator.GenerateZsh(aliasMap)
		case "fish":
			aliasBlock = g.aliasGenerator.GenerateFish(aliasMap)
		case "nushell":
			aliasBlock = g.aliasGenerator.GenerateNushell(aliasMap)
		case "powershell":
			aliasBlock = g.aliasGenerator.GeneratePowerShell(aliasMap)
		}

		// Append alias block to the output
		if aliasBlock != "" {
			_, err := fmt.Fprint(outputWriter, "\n", aliasBlock)
			if err != nil {
				return fmt.Errorf("failed to write alias block: %w", err)
			}
		}
	}

	return nil
}
