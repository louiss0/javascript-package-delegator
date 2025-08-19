package cmd

import (
	// standard library
	_ "embed" // Required for the embed directive
	"fmt"     // Import io package
	"sort"
	"strings"

	// external
	"github.com/spf13/cobra"

	// internal
	"github.com/louiss0/javascript-package-delegator/custom_errors"
	"github.com/louiss0/javascript-package-delegator/custom_flags" // Import the custom_flags package
	"github.com/louiss0/javascript-package-delegator/internal/completion"
)

//go:embed assets/jpd-extern.nu
var nushellCompletionScript string

const WITH_SHORTHANDS = "with-shorthands"

// NushellCompletionScript returns the embedded Nushell completion script content.
func NushellCompletionScript() string {
	return nushellCompletionScript
}

// NewCompletionCmd creates the parent 'completion' command and its subcommands
func NewCompletionCmd() *cobra.Command {
	// Use the custom_flags.FilePathFlag for the output file
	outputFileFlag := custom_flags.NewFilePathFlag("output")

	completionCmd := &cobra.Command{
		Use:   "completion <shell>",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for jpd.

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
		jpd completion bash                     # Print Bash completion to stdout
		jpd completion nushell                  # Print Nushell completion to stdout
		jpd completion nushell --output jpd_completions.nu # Save Nushell completion to a file
`,
		DisableFlagsInUseLine: true, // Don't show global flags for completion command itself
		Args: func(cmd *cobra.Command, args []string) error {
			// Define supported shells as a sorted slice for consistent output and efficient lookup.
			supportedShells := []string{
				"bash",
				"fish",
				"nushell",
				"powershell",
				"zsh",
			}

			// Generate a comma-separated list of supported shells using strings.Join
			supportedShellList := strings.Join(supportedShells, ", ")

			if len(args) != 1 {
				return custom_errors.CreateInvalidArgumentErrorWithMessage(
					fmt.Sprintf("requires exactly one argument representing the shell. Supported shells are: %s", supportedShellList))
			}

			shell := args[0]

			// Check if the shell is supported using binary search (declarative and efficient)
			// This requires importing the "sort" and "strings" packages.
			idx := sort.SearchStrings(supportedShells, shell)
			if idx >= len(supportedShells) || supportedShells[idx] != shell {
				return custom_errors.CreateInvalidArgumentErrorWithMessage(
					fmt.Sprintf("unsupported shell: '%s'. Supported shells are: %s", shell, supportedShellList))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			generator := completion.NewGenerator()

			withShorthand, err := cmd.Flags().GetBool(WITH_SHORTHANDS)

			if err != nil {
				return err
			}

			err = generator.GenerateCompletion(cmd, args[0], outputFileFlag.String(), withShorthand)

			if err != nil {
				return err
			}

			return nil
		},
	}

	// Bind the custom flag type
	completionCmd.Flags().VarP(&outputFileFlag, "output", "o", "Write completion script to a file instead of stdout")
	completionCmd.Flags().BoolP(WITH_SHORTHANDS, "w", false, "Generate completion script with shorthand flags")

	return completionCmd
}
