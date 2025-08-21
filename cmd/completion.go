package cmd

import (
	// standard library
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	// external
	"github.com/spf13/cobra"

	// internal
	"github.com/louiss0/javascript-package-delegator/custom_errors"
	"github.com/louiss0/javascript-package-delegator/custom_flags"
	"github.com/louiss0/javascript-package-delegator/internal/integrations"
)

const WITH_SHORTHAND = "with-shorthand"

// getSupportedShells returns the list of supported shell completions
func getSupportedShells() []string {
	return []string{
		"bash",
		"fish",
		"nushell",
		"powershell",
		"zsh",
	}
}

// getDefaultAliasMapping returns the default alias mapping for jpd commands
func getDefaultAliasMapping() map[string][]string {
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

// NewCompletionCmd creates the parent 'completion' command and its subcommands
func NewCompletionCmd() *cobra.Command {
	// Use the custom_flags.FilePathFlag for the output file
	outputFileFlag := custom_flags.NewFilePathFlag("output")

	completionCmd := &cobra.Command{
		Use:   "completion <shell>",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for jpd.

Supported shells:
  bash, zsh, fish, powershell, nushell

To install completion for your shell, run:

Bash:
		$ jpd completion bash > /etc/bash_completion.d/jpd

Zsh:
		# To load completions for each session, run:
		$ jpd completion zsh > "${fpath[1]}/_jpd"
		# You will need to start a new shell for this setup to take effect.

Fish:
		$ jpd completion fish > ~/.config/fish/completions/jpd.fish

PowerShell:
		PS> jpd completion powershell | Out-String | Invoke-Expression

Nushell:
		$ jpd completion nushell > ~/.config/nushell/completions/jpd_completions.nu
		# Then add 'source ~/.config/nushell/completions/jpd_completions.nu' to your env.nu or config.nu

Examples:
		jpd completion bash                          # Print Bash completion to stdout
		jpd completion zsh --output completions.zsh  # Save Zsh completion to file
		jpd completion fish --with-shorthand         # Include alias functions
`,
		DisableFlagsInUseLine: true, // Don't show global flags for completion command itself
		Args: func(cmd *cobra.Command, args []string) error {
			supportedShells := getSupportedShells()

			// Sort for consistent output and efficient lookup
			sort.Strings(supportedShells)

			// Generate a comma-separated list of supported shells using strings.Join
			supportedShellList := strings.Join(supportedShells, ", ")

			if len(args) != 1 {
				return custom_errors.CreateInvalidArgumentErrorWithMessage(
					fmt.Sprintf("requires exactly one argument representing the shell. Supported shells are: %s", supportedShellList))
			}

			shell := args[0]

			// Check if the shell is supported using binary search (declarative and efficient)
			idx := sort.SearchStrings(supportedShells, shell)
			if idx >= len(supportedShells) || supportedShells[idx] != shell {
				return custom_errors.CreateInvalidArgumentErrorWithMessage(
					fmt.Sprintf("unsupported shell: '%s'. Supported shells are: %s", shell, supportedShellList))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			shell := args[0]
			filename := outputFileFlag.String()

			withShorthand, err := cmd.Flags().GetBool(WITH_SHORTHAND)
			if err != nil {
				return err
			}

			return generateCompletion(cmd, shell, filename, withShorthand)
		},
	}

	// Bind the custom flag type
	completionCmd.Flags().VarP(&outputFileFlag, "output", "o", "Write completion script to a file instead of stdout")
	completionCmd.Flags().BoolP(WITH_SHORTHAND, "w", false, "Generate completion script with shorthand flags")

	return completionCmd
}

// generateCompletion generates completion script for the specified shell
func generateCompletion(cmd *cobra.Command, shell string, filename string, withShorthand bool) error {
	var outputWriter io.Writer
	var file *os.File

	// Determine output destination based on filename
	if filename == "" {
		outputWriter = cmd.OutOrStdout()
	} else {
		// Ensure the directory exists before creating the file
		dir := filepath.Dir(filename)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		// Create or open the specified file for writing
		f, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("failed to create output file %s: %w", filename, err)
		}
		outputWriter = f
		file = f
	}

	// Ensure the file is closed if it was opened
	if file != nil {
		defer func() {
			if cerr := file.Close(); cerr != nil {
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
		_, completionErr = fmt.Fprint(outputWriter, integrations.NushellCompletionScript())
		if completionErr != nil {
			completionErr = fmt.Errorf("failed to write Nushell completion script: %w", completionErr)
		}
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}

	if completionErr != nil {
		return completionErr
	}

	// If --with-shorthand flag is set, append alias functions
	if withShorthand {
		aliasGenerator := integrations.NewAliasGenerator()
		aliasMap := getDefaultAliasMapping()

		// Generate alias block based on shell type
		var aliasBlock string
		switch shell {
		case "bash":
			aliasBlock = aliasGenerator.GenerateBash(aliasMap)
		case "zsh":
			aliasBlock = aliasGenerator.GenerateZsh(aliasMap)
		case "fish":
			aliasBlock = aliasGenerator.GenerateFish(aliasMap)
		case "nushell":
			aliasBlock = aliasGenerator.GenerateNushell(aliasMap)
		case "powershell":
			aliasBlock = aliasGenerator.GeneratePowerShell(aliasMap)
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
