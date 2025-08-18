// Package completion provides shell completion generation logic for jpd commands.
// This package extracts completion generators from the cmd package to improve
// code organization and maintainability.
package completion

import (
	// standard library
	_ "embed" // Required for the embed directive
	"fmt"
	"io"

	// external
	"github.com/spf13/cobra"

	// internal
	"github.com/louiss0/javascript-package-delegator/shell_alias"
)

//go:embed assets/jpd-extern.nu
var nushellCompletionScript string

// Generator provides methods for generating shell completions with optional shorthand aliases.
type Generator interface {
	// GenerateCompletion generates completion script for the specified shell and writes it to the output writer.
	// If withShorthand is true, shorthand alias functions are appended to the completion output.
	GenerateCompletion(cmd *cobra.Command, shell string, outputWriter io.Writer, withShorthand bool) error

	// GetSupportedShells returns a list of supported shell names.
	GetSupportedShells() []string

	// GetDefaultAliasMapping returns the default alias mapping for shorthand generation.
	GetDefaultAliasMapping() map[string][]string
}

// generator is the concrete implementation of the Generator interface.
type generator struct {
	aliasGenerator shell_alias.Generator
}

// NewGenerator creates a new completion generator instance.
func NewGenerator() Generator {
	return &generator{
		aliasGenerator: shell_alias.NewGenerator(),
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
func (g *generator) GenerateCompletion(cmd *cobra.Command, shell string, outputWriter io.Writer, withShorthand bool) error {
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
		_, completionErr = fmt.Fprint(outputWriter, GetNushellCompletionScript())
		if completionErr != nil {
			completionErr = fmt.Errorf("failed to write Nushell completion script: %w", completionErr)
		}
	case "carapace":
		_, completionErr = fmt.Fprint(outputWriter, GenerateCarapaceBridge())
		if completionErr != nil {
			completionErr = fmt.Errorf("failed to write carapace completion bridge: %w", completionErr)
		}
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}

	if completionErr != nil {
		return completionErr
	}

	// If --with-shorthand flag is set, append alias functions
	if withShorthand {
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
		case "carapace":
			// Carapace can use bash-style aliases for now
			aliasBlock = g.aliasGenerator.GenerateBash(aliasMap)
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

// GetNushellCompletionScript returns the embedded Nushell completion script content.
func GetNushellCompletionScript() string {
	return nushellCompletionScript
}

// GenerateCarapaceBridge generates a carapace completion bridge script.
func GenerateCarapaceBridge() string {
	return `# carapace completion bridge for jpd
# 
# This script provides instructions for integrating jpd with carapace.
# Carapace is a multi-shell completion framework that can bridge completions
# across different shells.
#
# Requirements:
# - Install carapace-bin: https://github.com/rsteube/carapace-bin
#
# Setup Instructions:
#
# For Bash/Zsh:
# Add this to your shell rc file (~/.bashrc or ~/.zshrc):
#   source <(carapace _carapace)
#   eval "$(jpd completion carapace)"
#
# For Fish:
# Add this to your Fish config (~/.config/fish/config.fish):
#   carapace _carapace | source
#   jpd completion carapace | source
#
# For Nushell:
# Add this to your Nushell config:
#   $env.config.completions.external.enable = true
#   carapace _carapace nushell | save ~/.carapace/init.nu
#   source ~/.carapace/init.nu
#
# The carapace completion system will then provide intelligent completions
# for jpd commands across all supported shells.
#
# For more information about carapace, visit:
# https://rsteube.github.io/carapace/
`
}
