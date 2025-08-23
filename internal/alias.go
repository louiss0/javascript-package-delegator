// Package integrations provides functionality for generating shell aliases, Warp workflows,
// and Carapace specs for jpd subcommands across different shells and integrations.
package integrations

import (
	"fmt"
	"strings"
)

// AliasGenerator provides methods for generating shell alias functions with completion wiring.
type AliasGenerator interface {
	// GenerateBash generates bash alias functions and completion wiring.
	// Input: map[string][]string where keys are canonical subcommand names
	// and values are lists of shorthand names to generate functions for.
	GenerateBash(aliases map[string][]string) string

	// GenerateZsh generates zsh alias functions and completion wiring.
	GenerateZsh(aliases map[string][]string) string

	// GenerateFish generates fish alias functions and completion wiring.
	GenerateFish(aliases map[string][]string) string

	// GenerateNushell generates nushell alias functions and completion wiring.
	GenerateNushell(aliases map[string][]string) string

	// GeneratePowerShell generates PowerShell alias functions and completion wiring.
	GeneratePowerShell(aliases map[string][]string) string
}

// aliasGenerator is the concrete implementation of the AliasGenerator interface.
type aliasGenerator struct{}

// NewAliasGenerator creates a new instance of the shell alias generator.
func NewAliasGenerator() AliasGenerator {
	return &aliasGenerator{}
}

// GenerateBash generates bash alias functions and completion wiring.
func (g *aliasGenerator) GenerateBash(aliases map[string][]string) string {
	var result strings.Builder

	// Add header comment
	result.WriteString("# jpd shorthand aliases\n")

	// Add guard clause
	result.WriteString("command -v jpd > /dev/null || return 0\n\n")

	// Generate functions for each alias
	for subcommand, aliasNames := range aliases {
		for _, aliasName := range aliasNames {
			// Generate function
			result.WriteString(fmt.Sprintf("function %s() { command jpd %s \"$@\"; }\n", aliasName, subcommand))
			// Generate completion wiring
			result.WriteString(fmt.Sprintf("complete -F __start_jpd %s\n", aliasName))
			result.WriteString("\n")
		}
	}

	return result.String()
}

// GenerateZsh generates zsh alias functions and completion wiring.
func (g *aliasGenerator) GenerateZsh(aliases map[string][]string) string {
	var result strings.Builder

	// Add header comment
	result.WriteString("# jpd shorthand aliases\n")

	// Add guard clause
	result.WriteString("(( $+commands[jpd] )) || return\n\n")

	// Generate functions for each alias
	for subcommand, aliasNames := range aliases {
		for _, aliasName := range aliasNames {
			// Generate function
			result.WriteString(fmt.Sprintf("%s() { jpd %s \"$@\"; }\n", aliasName, subcommand))
			// Generate completion wiring
			result.WriteString(fmt.Sprintf("compdef _jpd %s\n", aliasName))
			result.WriteString("\n")
		}
	}

	return result.String()
}

// GenerateFish generates fish alias functions and completion wiring.
func (g *aliasGenerator) GenerateFish(aliases map[string][]string) string {
	var result strings.Builder

	// Add header comment
	result.WriteString("# jpd shorthand aliases\n\n")

	// Generate functions for each alias
	for subcommand, aliasNames := range aliases {
		for _, aliasName := range aliasNames {
			// Generate function
			result.WriteString(fmt.Sprintf("function %s\n", aliasName))
			result.WriteString(fmt.Sprintf("    jpd %s $argv\n", subcommand))
			result.WriteString("end\n")
			// Generate completion wiring
			result.WriteString(fmt.Sprintf("complete -c %s -w jpd\n", aliasName))
			result.WriteString("\n")
		}
	}

	return result.String()
}

// GenerateNushell generates nushell alias functions and completion wiring.
func (g *aliasGenerator) GenerateNushell(aliases map[string][]string) string {
	var result strings.Builder

	// Add header comment
	result.WriteString("# jpd shorthand aliases\n\n")

	// Generate functions for each alias
	for subcommand, aliasNames := range aliases {
		for _, aliasName := range aliasNames {
			// Generate extern declaration
			result.WriteString(fmt.Sprintf("export extern \"%s\" [\n", aliasName))
			result.WriteString("    ...args: string\n")
			result.WriteString("]\n")
			// Generate function definition
			result.WriteString(fmt.Sprintf("export def %s [...args] {\n", aliasName))
			result.WriteString(fmt.Sprintf("    jpd %s $args\n", subcommand))
			result.WriteString("}\n\n")
		}
	}

	return result.String()
}

// GeneratePowerShell generates PowerShell alias functions and completion wiring.
func (g *aliasGenerator) GeneratePowerShell(aliases map[string][]string) string {
	var result strings.Builder

	// Add header comment
	result.WriteString("# jpd shorthand aliases\n\n")

	// Add guard clause - check if jpd command exists
	result.WriteString("if (-not (Get-Command jpd -ErrorAction SilentlyContinue)) {\n")
	result.WriteString("    return\n")
	result.WriteString("}\n\n")

	// Generate functions for each alias
	for subcommand, aliasNames := range aliases {
		for _, aliasName := range aliasNames {
			// Generate function
			result.WriteString(fmt.Sprintf("function %s {\n", aliasName))
			result.WriteString(fmt.Sprintf("    jpd %s @args\n", subcommand))
			result.WriteString("}\n")
			// Generate completion registration
			result.WriteString(fmt.Sprintf("Register-ArgumentCompleter -CommandName '%s' -ScriptBlock {\n", aliasName))
			result.WriteString("    param($commandName, $parameterName, $wordToComplete, $commandAst, $fakeBoundParameters)\n")
			result.WriteString("    $completions = @()\n")
			result.WriteString("    # Add basic jpd argument completion here\n")
			result.WriteString("    return $completions\n")
			result.WriteString("}\n\n")
		}
	}

	return result.String()
}
