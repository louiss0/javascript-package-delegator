package cmd

import (
	_ "embed" // Required for the embed directive
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/louiss0/javascript-package-delegator/custom_errors"
	"github.com/louiss0/javascript-package-delegator/custom_flags" // Import the custom_flags package
	"github.com/spf13/cobra"
)

//go:embed assets/jpd-extern.nu
var nushellCompletionScript string

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
			shell := args[0]

			// Map shell types to file extensions
			extensionMap := map[string]string{
				"bash":       "bash",
				"zsh":        "zsh",
				"fish":       "fish",
				"powershell": "ps1",
				"nushell":    "nu",
			}

			// Determine the final output filename
			// Access the string value from the FilePathFlag
			finalOutputFile := outputFileFlag.String()
			if finalOutputFile == "" {
				// Get the file extension for the current shell
				ext, ok := extensionMap[shell]
				if !ok {
					// Fallback if shell not in map (should be caught by switch statement later)
					ext = ""
				}
				// Default filename: jpd_<shell>_completion.<extension>
				// Example: jpd_bash_completion.bash, jpd_nushell_completion.nu
				if ext != "" {
					finalOutputFile = fmt.Sprintf("jpd_%s_completion.%s", shell, ext)
				} else {
					// If extension is unknown, still create a file, without an extension
					finalOutputFile = fmt.Sprintf("jpd_%s_completion", shell)
				}
			}

			// Always create the output file
			file, err := os.Create(finalOutputFile)
			if err != nil {
				return fmt.Errorf("failed to create output file '%s': %w", finalOutputFile, err)
			}
			defer file.Close()
			outputWriter := file // All output will now go to this file

			// Get the absolute path for the success message
			absPath, err := filepath.Abs(finalOutputFile)
			if err != nil {
				// If we can't get absolute path, use the original filename
				absPath = finalOutputFile
			}

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
				_, completionErr = fmt.Fprint(outputWriter, NushellCompletionScript())
				if completionErr != nil {
					completionErr = fmt.Errorf("failed to write Nushell completion script: %w", completionErr)
				}
			default:
				completionErr = fmt.Errorf("unsupported shell: %s. Supported shells are: bash, zsh, fish, powershell, nushell", shell)
			}

			if completionErr != nil {
				return completionErr
			}

			// Print success message with full path
			fmt.Fprintf(cmd.OutOrStdout(), "Successfully generated %s completion script at %s\n", shell, absPath)
			return nil
		},
	}

	// Bind the custom flag type
	completionCmd.Flags().VarP(&outputFileFlag, "output", "o", "Write completion script to a file instead of stdout")

	return completionCmd
}
