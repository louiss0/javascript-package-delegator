/*
Copyright © 2025 Shelton Louis

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
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
			pm := getPackageNameFromCommandContext(cmd)

			appEnv := getAppEnvFromCommandContext(cmd)

			if appEnv != _DEV {
				log.Infof("Using %s\n", pm)

			}

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
				cmdArgs = []string{"oudated"}

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
			execCmd := exec.Command(pm, cmdArgs...)
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr
			execCmd.Stdin = os.Stdin
			if appEnv != _DEV {

				log.Infof("Running: %s %s\n", pm, strings.Join(cmdArgs, " "))
			}
			return execCmd.Run()
		},
	}

	// Add flags
	cmd.Flags().BoolP("interactive", "i", false, "Interactive update (where supported)")
	cmd.Flags().BoolP("global", "g", false, "Update global packages")
	cmd.Flags().Bool("latest", false, "Update to latest version (ignoring version ranges)")

	return cmd
}
