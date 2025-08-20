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

const WITH_SHORTHAND = "with-shorthand"

// NewCompletionCmd creates the parent 'completion' command and its subcommands
func NewCompletionCmd() *cobra.Command {
	// Use the custom_flags.FilePathFlag for the output file
	outputFileFlag := custom_flags.NewFilePathFlag("output")

	completionCmd := &cobra.Command{
		Use:   "completion <target>",
		Short: "Generate shell completions, Carapace specs, and Warp workflows",
		Long: `Generate shell completion scripts, Carapace specs, and Warp workflows for jpd.

Supported targets:
  bash, zsh, fish, powershell, nushell - Standard shell completions
  carapace                             - YAML spec for carapace-bin
  warp                                - Workflow YAML files for Warp terminal

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

Carapace:
		$ jpd completion carapace > jpd.yaml
		# Place in your carapace specs directory

Warp:
		$ jpd completion warp --output ./workflows/  # Generate individual workflow files
		$ jpd completion warp                        # Print multi-doc YAML to stdout

Examples:
		jpd completion bash                          # Print Bash completion to stdout
		jpd completion carapace                      # Print Carapace YAML spec
		jpd completion warp --output ./workflows/    # Generate Warp workflow files
		jpd completion warp                          # Print Warp workflows as multi-doc YAML

Note: --with-shorthand flag is ignored for carapace and warp targets.
`,
		DisableFlagsInUseLine: true, // Don't show global flags for completion command itself
		Args: func(cmd *cobra.Command, args []string) error {
			// Get supported shells from the generator for consistency
			generator := completion.NewGenerator()
			supportedShells := generator.SupportedShells()

			// Sort for consistent output and efficient lookup
			sort.Strings(supportedShells)

			// Generate a comma-separated list of supported shells using strings.Join
			supportedShellList := strings.Join(supportedShells, ", ")

			if len(args) != 1 {
				return custom_errors.CreateInvalidArgumentErrorWithMessage(
					fmt.Sprintf("requires exactly one argument representing the target. Supported targets are: %s", supportedShellList))
			}

			target := args[0]

			// Check if the target is supported using binary search (declarative and efficient)
			idx := sort.SearchStrings(supportedShells, target)
			if idx >= len(supportedShells) || supportedShells[idx] != target {
				return custom_errors.CreateInvalidArgumentErrorWithMessage(
					fmt.Sprintf("unsupported target: '%s'. Supported targets are: %s", target, supportedShellList))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			generator := completion.NewGenerator()

			withShorthand, err := cmd.Flags().GetBool(WITH_SHORTHAND)

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
	completionCmd.Flags().BoolP(WITH_SHORTHAND, "w", false, "Generate completion script with shorthand flags")

	return completionCmd
}
