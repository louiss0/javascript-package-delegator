// Package cmd contains the Cobra commands for the javascript-package-delegator CLI.
// It defines the various subcommands and their logic for delegating to JavaScript package managers.
package cmd

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// NewAgentCmd creates a new Cobra command for the "agent" functionality.
// This command detects and displays the JavaScript package manager being used
// in the current project, equivalent to the `na` command in `@antfu/ni`.
// It is useful for quickly identifying which package manager `jpd` would delegate to.
//
// The command's `RunE` function retrieves the detected package manager name from the
// command's persistent flags (which is populated by the root command's PersistentPreRunE
// logic) and then executes a command to show information about that package manager.
func NewAgentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Show the detected package manager agent",
		Long: `Show information about the detected package manager agent.
Equivalent to 'na' command - detects and displays the package manager being used.

This command shows which package manager would be used based on lock files in the current directory.

Examples:
  jpd agent    # Show detected package manager
  jpd agent -a yarn # Explicitly show yarn's agent info (e.g., its version or help)
`,
		Aliases: []string{"a"},
		// Allow passing through unknown flags (e.g., flags intended for the underlying package manager)
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Retrieve the detected package manager name from the command's flags.
			// This flag is populated by the root command's PersistentPreRunE logic.
			pm, err := cmd.Flags().GetString(AGENT_FLAG)
			if err != nil {
				return fmt.Errorf("failed to get agent flag: %w", err)
			}

			// Validate that the package manager was actually detected
			if pm == "" {
				return fmt.Errorf("no package manager detected; please ensure you have a lock file (package-lock.json, yarn.lock, pnpm-lock.yaml, bun.lockb, deno.json, etc.) in your project directory")
			}

			// Get the environment configuration to determine if logging should be verbose.
			goEnv := getGoEnvFromCommandContext(cmd)
			de := getDebugExecutorFromCommandContext(cmd)

			// Log the detected package manager in production mode.
			goEnv.ExecuteIfModeIsProduction(func() {
				log.Infof("Detected package manager, now executing command: %s\n", pm)
			})

			// Obtain the command runner from the context, which handles external process execution.
			cmdRunner := getCommandRunnerFromCommandContext(cmd)

			// If the user passed --version to this subcommand, Cobra will consume it as a local flag
			// and it won't appear in args. Re-add it so we pass it through to the underlying tool.
			if v, _ := cmd.Flags().GetBool("version"); v {
				args = append(args, "--version")
			}

			// Prepare the command to be executed. For the 'agent' command, it typically
			// runs the package manager itself. Any additional arguments provided to 'jpd agent'
			// are passed directly to the detected package manager.
			de.LogJSCommandIfDebugIsTrue(pm, args...)
			cmdRunner.Command(pm, args...)

			// Execute the command and return any error.
			return cmdRunner.Run()
		},
	}

	// Define a local --version flag so "jpd agent --version" is accepted
	cmd.Flags().Bool("version", false, "Show underlying package manager version")

	return cmd
}
