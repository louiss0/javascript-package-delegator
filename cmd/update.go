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
		Long: `Update packages using the appropriate package manager.
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
				log.Infof("Using %s\n", pm)
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
					// For npm, we need to use install with @latest
					if len(args) > 0 {
						cmdArgs = []string{"install"}
						for _, pkg := range args {
							cmdArgs = append(cmdArgs, pkg+"@latest")
						}
						if global {
							cmdArgs = append(cmdArgs, "--global")
						}
					}
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
				cmdArgs = []string{"outdated"}

				if interactive {
					cmdArgs = append(cmdArgs, "-i")
				}

				if global {
					cmdArgs = append(cmdArgs, "--global")
				}

				if latest {
					if len(args) > 0 {

						cmdArgs = append(cmdArgs, "--latest")

						cmdArgs = append(cmdArgs, args...)

					} else {
						cmdArgs = append(cmdArgs, "--latest")
					}
				}

			default:
				return fmt.Errorf("unsupported package manager: %s", pm)
			}

			// Execute the command
			de.LogJSCommandIfDebugIsTrue(pm, cmdArgs...)
			cmdRunner.Command(pm, cmdArgs...)

			goEnv.ExecuteIfModeIsProduction(func() {
				log.Infof("Running: %s %s\n", pm, strings.Join(cmdArgs, " "))
			})
			return cmdRunner.Run()
		},
	}

	// Add flags
	cmd.Flags().BoolP("interactive", "i", false, "Interactive update (where supported)")
	cmd.Flags().BoolP("global", "g", false, "Update global packages")
	cmd.Flags().Bool("latest", false, "Update to latest version (ignoring version ranges)")

	return cmd
}
