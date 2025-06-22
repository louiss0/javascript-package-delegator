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
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

func NewInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install [packages...]",
		Short: "Install packages using the detected package manager",
		Long: `Install packages using the appropriate package manager based on lock files.
Equivalent to 'ni' command - detects npm, yarn, pnpm, bun, or deno and runs the appropriate install command.

Examples:
  jpd install           # Install all dependencies
  jpd install lodash    # Install lodash
  jpd install -D vitest # Install vitest as dev dependency
  jpd install -g typescript # Install globally`,
		Aliases: []string{"i", "add"},
		RunE: func(cmd *cobra.Command, args []string) error {
			pm := getPackageNameFromCommandContext(cmd)

			goEnv := getGoEnvFromCommandContext(cmd)

			goEnv.ExecuteIfModeIsProduction(func() {

				log.Infof("Using %s\n", pm)
			})

			// Build command based on package manager and flags
			var cmdArgs []string
			switch pm {
			case "npm":
				if len(args) == 0 {
					cmdArgs = []string{"install"}
				} else {
					cmdArgs = []string{"install"}
					cmdArgs = append(cmdArgs, args...)
				}
				if dev, _ := cmd.Flags().GetBool("dev"); dev {
					cmdArgs = append(cmdArgs, "--save-dev")
				}
				if global, _ := cmd.Flags().GetBool("global"); global {
					cmdArgs = append(cmdArgs, "--global")
				}
				if production, _ := cmd.Flags().GetBool("production"); production {
					cmdArgs = append(cmdArgs, "--omit=dev")
				}
				if frozen, _ := cmd.Flags().GetBool("frozen"); frozen {
					cmdArgs = append(cmdArgs, "--package-lock-only")
				}

			case "yarn":
				if len(args) == 0 {
					cmdArgs = []string{"install"}
				} else {
					cmdArgs = []string{"add"}
					cmdArgs = append(cmdArgs, args...)
				}
				if dev, _ := cmd.Flags().GetBool("dev"); dev {
					cmdArgs = append(cmdArgs, "--dev")
				}
				if global, _ := cmd.Flags().GetBool("global"); global {
					cmdArgs = append(cmdArgs, "--global")
				}
				if production, _ := cmd.Flags().GetBool("production"); production {
					cmdArgs = append(cmdArgs, "--production")
				}
				if frozen, _ := cmd.Flags().GetBool("frozen"); frozen {
					cmdArgs = append(cmdArgs, "--frozen-lockfile")
				}

			case "pnpm":
				if len(args) == 0 {
					cmdArgs = []string{"install"}
				} else {
					cmdArgs = []string{"add"}
					cmdArgs = append(cmdArgs, args...)
				}
				if dev, _ := cmd.Flags().GetBool("dev"); dev {
					cmdArgs = append(cmdArgs, "--save-dev")
				}
				if global, _ := cmd.Flags().GetBool("global"); global {
					cmdArgs = append(cmdArgs, "--global")
				}
				if production, _ := cmd.Flags().GetBool("production"); production {
					cmdArgs = append(cmdArgs, "--prod")
				}
				if frozen, _ := cmd.Flags().GetBool("frozen"); frozen {
					cmdArgs = append(cmdArgs, "--frozen-lockfile")
				}

			case "bun":
				if len(args) == 0 {
					cmdArgs = []string{"install"}
				} else {
					cmdArgs = []string{"add"}
					cmdArgs = append(cmdArgs, args...)
				}
				if dev, _ := cmd.Flags().GetBool("dev"); dev {
					cmdArgs = append(cmdArgs, "--development")
				}
				if global, _ := cmd.Flags().GetBool("global"); global {
					cmdArgs = append(cmdArgs, "--global")
				}
				if production, _ := cmd.Flags().GetBool("production"); production {
					cmdArgs = append(cmdArgs, "--production")
				}

			case "deno":
				// Deno doesn't have traditional "install" - it downloads deps on run
				// But we can cache dependencies
				if len(args) == 0 {
					// Check for deps.ts or mod.ts file to cache
					if _, err := os.Stat("deps.ts"); err == nil {
						cmdArgs = []string{"cache", "deps.ts"}
					} else if _, err := os.Stat("mod.ts"); err == nil {
						cmdArgs = []string{"cache", "mod.ts"}
					} else {
						return fmt.Errorf("deno: no deps.ts or mod.ts file found to cache")
					}
				} else {
					// For specific packages, advise user to add them to imports
					return fmt.Errorf("deno doesn't support installing packages directly. Add imports to your code or deps.ts file")
				}
				// Note: Deno ignores most flags as it doesn't have traditional package management

			default:
				return fmt.Errorf("unsupported package manager: %s", pm)
			}

			// Execute the command
			cmdRunner := getCommandRunnerFromCommandContext(cmd)
			cmdRunner.Command(pm, cmdArgs...)

			goEnv.ExecuteIfModeIsProduction(func() {
				log.Infof("Running: %s %s\n", pm, strings.Join(cmdArgs, " "))

			})
			return cmdRunner.Run()
		},
	}

	// Add flags
	cmd.Flags().BoolP("dev", "D", false, "Install as dev dependency")
	cmd.Flags().BoolP("global", "g", false, "Install globally")
	cmd.Flags().BoolP("production", "P", false, "Install production dependencies only")
	cmd.Flags().Bool("frozen", false, "Install with frozen lockfile")
	cmd.Flags().BoolP("interactive", "i", false, "Interactive package selection")

	return cmd
}
