// Package cmd provides command-line interface implementations for the JavaScript package delegator.
package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

func NewUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [packages...]",
		Short: "Update packages using the detected package manager",
		Long: `Update packages to their latest versions using the appropriate package manager.
Equivalent to 'nup' command - detects npm, yarn, pnpm, or bun and runs the update command.

Examples:
  javascript-package-delegator update           # Update all packages
  javascript-package-delegator update lodash    # Update specific package
  javascript-package-delegator update -i        # Interactive update (where supported)
  javascript-package-delegator update -g typescript # Update global package`,
		Aliases: []string{"u", "up", "upgrade"},
		RunE: func(cmd *cobra.Command, args []string) error {
			pm, _ := cmd.Flags().GetString(AGENT_FLAG)
			cmdRunner := getCommandRunnerFromCommandContext(cmd)

			goEnv := getGoEnvFromCommandContext(cmd)
			de := getDebugExecutorFromCommandContext(cmd)

			goEnv.ExecuteIfModeIsProduction(func() {
				log.Info("Using package manager", "pm", pm)
			})

			// Get flags
			interactive, _ := cmd.Flags().GetBool("interactive")
			global, _ := cmd.Flags().GetBool("global")
			latest, _ := cmd.Flags().GetBool("latest")

			// Build command based on package manager and flags
			var cmdArgs []string
			switch pm {
			case "npm":
				if interactive {
					return fmt.Errorf("npm does not support interactive updates")
				}
				if len(args) == 0 {
					cmdArgs = []string{"update"}
				} else {
					cmdArgs = []string{"update"}
					cmdArgs = append(cmdArgs, args...)
				}
				if global {
					cmdArgs = append(cmdArgs, "--global")
				}
				if latest {
					cmdArgs = append(cmdArgs, "--latest")
				}

			case "yarn":
				if interactive {
					cmdArgs = []string{"upgrade-interactive"}
					if len(args) > 0 {
						cmdArgs = append(cmdArgs, args...)
					}
				} else {
					if len(args) == 0 {
						cmdArgs = []string{"upgrade"}
					} else {
						cmdArgs = []string{"upgrade"}
						cmdArgs = append(cmdArgs, args...)
					}
				}
				if global {
					cmdArgs = append(cmdArgs, "--global")
				}
				if latest {
					cmdArgs = append(cmdArgs, "--latest")
				}

			case "pnpm":
				if interactive {
					cmdArgs = []string{"update", "--interactive"}
					if len(args) > 0 {
						cmdArgs = append(cmdArgs, args...)
					}

				} else {
					if len(args) == 0 {
						cmdArgs = []string{"update"}
					} else {
						cmdArgs = []string{"update"}
						cmdArgs = append(cmdArgs, args...)
					}
				}

				if global {
					cmdArgs = append(cmdArgs, "--global")
				}
				if latest {
					cmdArgs = append(cmdArgs, "--latest")
				}

			case "bun":
				if interactive {
					return fmt.Errorf("bun does not support interactive updates")
				}

				if len(args) == 0 {
					cmdArgs = []string{"update"}
				} else {
					cmdArgs = []string{"update"}
					cmdArgs = append(cmdArgs, args...)
				}

				if global {
					cmdArgs = append(cmdArgs, "--global")
				}
				if latest {
					cmdArgs = append(cmdArgs, "--latest")
				}

			case "deno":
				return fmt.Errorf("deno does not support the update command")

			default:
				return fmt.Errorf("unsupported package manager: %s", pm)
			}

			// Execute the command
			de.LogJSCommandIfDebugIsTrue(pm, cmdArgs...)
			cmdRunner.Command(pm, cmdArgs...)

			goEnv.ExecuteIfModeIsProduction(func() {
				log.Info("Running command", "pm", pm, "args", strings.Join(cmdArgs, " "))
			})
			return cmdRunner.Run()
		},
	}

	// Add flags
	cmd.Flags().BoolP("interactive", "i", false, "Interactive update (where supported)")
	cmd.Flags().BoolP("global", "g", false, "Update global packages")
	cmd.Flags().BoolP("latest", "L", false, "Update to latest version (ignoring version ranges)")

	return cmd
}
