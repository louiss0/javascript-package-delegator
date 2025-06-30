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

	"github.com/charmbracelet/log"
	"github.com/louiss0/javascript-package-delegator/custom_errors"
	"github.com/louiss0/javascript-package-delegator/detect"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

// NewInstallCmd creates a new Cobra command for the "install" functionality.
// This command delegates to the appropriate JavaScript package manager (npm, Yarn, pnpm, Bun, or Deno)
// to install project dependencies or specific packages.
// It also includes optional Volta integration to ensure consistent toolchain usage.
func NewInstallCmd(detectVolta func() bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install [packages...]",
		Short: "Install packages using the detected package manager",
		Long: `Install packages using the appropriate package manager based on lock files.
Equivalent to 'ni' command - detects npm, yarn, pnpm, bun, or deno and runs the appropriate install command.

Examples:
  jpd install           # Install all dependencies
  jpd install lodash    # Install lodash
  jpd install -D vitest # Install vitest as dev dependency
  jpd install -g typescript # Install globally
  jpd install --no-volta # Install packages bypassing Volta, even if installed
`,
		Aliases: []string{"i", "add"},
		RunE: func(cmd *cobra.Command, args []string) error {
			pm, _ := cmd.Flags().GetString(AGENT_FLAG)
			goEnv := getGoEnvFromCommandContext(cmd)
			cmdRunner := getCommandRunnerFromCommandContext(cmd)

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

				if len(args) == 0 {

					return fmt.Errorf("For deno one or more packages is required")
				}

				if production, _ := cmd.Flags().GetBool("production"); production {
					return custom_errors.CreateInvalidFlagErrorWithMessage(
						custom_errors.FlagName("production"),
						"Deno doesn't support prod!",
					)
				}

				if global, _ := cmd.Flags().GetBool("global"); global {

					cmdArgs = append([]string{"install"}, args...)
					break
				}

				cmdArgs = append([]string{"add"}, args...)

				if dev, _ := cmd.Flags().GetBool("dev"); dev {
					cmdArgs = append(cmdArgs, "--dev")
				}

			default:
				return fmt.Errorf("unsupported package manager: %s", pm)
			}

			if detectVolta() && lo.Contains([]string{detect.NPM, detect.PNPM, detect.YARN}, pm) {

				completeVoltaCommand := lo.Flatten([][]string{
					detect.VOLTA_RUN_COMMNAD,
					{pm},
					cmdArgs,
				})
				cmdRunner.Command(completeVoltaCommand[0], completeVoltaCommand[1:]...)

				goEnv.ExecuteIfModeIsProduction(func() {

					log.Info("Executing this ", "command", completeVoltaCommand)

				})
			} else {

				cmdRunner.Command(pm, cmdArgs...)

				goEnv.ExecuteIfModeIsProduction(func() {

					log.Info("Executing this ", "command", append([]string{pm}, cmdArgs...))

				})
			}

			// Execute the command
			return cmdRunner.Run()
		},
	}

	// Add flags
	cmd.Flags().BoolP("dev", "D", false, "Install as dev dependency")
	cmd.Flags().BoolP("global", "g", false, "Install globally")
	cmd.Flags().BoolP("production", "P", false, "Install production dependencies only")
	cmd.Flags().Bool("frozen", false, "Install with frozen lockfile")
	cmd.Flags().BoolP("interactive", "i", false, "Interactive package selection")
	cmd.Flags().Bool("no-volta", false, "Disable Volta integration for this command") // New flag for Volta opt-out

	return cmd
}
