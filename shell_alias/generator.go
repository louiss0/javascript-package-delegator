// Package shell_alias provides functionality for generating shell alias functions
// and completion wiring for jpd subcommands across different shell types.
package shell_alias

import (
	"fmt"
	"strings"
)

// Shell represents the supported shell types for alias generation.
type Shell string

// Supported shell types
const (
	Bash       Shell = "bash"
	Zsh        Shell = "zsh"
	Fish       Shell = "fish"
	Nushell    Shell = "nushell"
	PowerShell Shell = "powershell"
	// Carapace is handled via separate bridging function
)

// AliasSpec represents a specification for alias generation.
type AliasSpec struct {
	Name    string   // Canonical subcommand name (e.g., "install")
	Aliases []string // List of shorthand names (e.g., ["jpi", "jpadd", "jpd-install"])
}

// Generator provides methods for generating shell alias functions with completion wiring.
type Generator interface {
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

// generator is the concrete implementation of the Generator interface.
type generator struct{}

// NewGenerator creates a new instance of the shell alias generator.
func NewGenerator() Generator {
	return &generator{}
}

// GenerateBash generates bash alias functions and completion wiring.
func (g *generator) GenerateBash(aliases map[string][]string) string {
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
func (g *generator) GenerateZsh(aliases map[string][]string) string {
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
func (g *generator) GenerateFish(aliases map[string][]string) string {
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
func (g *generator) GenerateNushell(aliases map[string][]string) string {
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
func (g *generator) GeneratePowerShell(aliases map[string][]string) string {
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
