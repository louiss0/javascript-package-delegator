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
	"strings"

	"github.com/charmbracelet/log"
	"github.com/louiss0/javascript-package-delegator/detect"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func NewCleanInstallCmd(detectVolta func() bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clean-install",
		Short: "Clean install packages using the detected package manager",
		Long: `Clean install packages using the appropriate package manager with frozen lockfile.
Equivalent to 'nci' command - detects npm, yarn, pnpm, or bun and runs clean install.

This command is designed for CI environments and production builds where you want to install
exactly what's in the lockfile without updating it.

Examples:
  javascript-package-delegator clean-install     # Clean install all dependencies`,
		Aliases: []string{"ci"},
		RunE: func(cmd *cobra.Command, args []string) error {

			pm, _ := cmd.Flags().GetString(AGENT_FLAG)
			goEnv := getGoEnvFromCommandContext(cmd)
			cmdRunner := getCommandRunnerFromCommandContext(cmd)
			de := getDebugExecutorFromCommandContext(cmd)
			if dbg, _ := cmd.Flags().GetBool(_DEBUG_FLAG); dbg {
				de.LogDebugMessageIfDebugIsTrue("Command start", "name", "clean-install", "pm", pm)
			}

			// Build command based on package manager
			var cmdArgs []string
			switch pm {
			case "npm":
				cmdArgs = []string{"ci"}

			case "yarn":
				// Yarn v1 uses install --frozen-lockfile, v2+ uses install --immutable
				yarnVersion, err := detect.DetectYarnVersion(
					getYarnVersionRunnerCommandContext(cmd),
				)

				if err != nil || strings.HasPrefix(yarnVersion, "1.") {
					// Yarn v1 or unknown version
					cmdArgs = []string{"install", "--frozen-lockfile"}
				} else {
					// Yarn v2+
					cmdArgs = []string{"install", "--immutable"}
				}

			case "pnpm":
				cmdArgs = []string{"install", "--frozen-lockfile"}

			case "bun":
				cmdArgs = []string{"install", "--frozen-lockfile"}

			case "deno":
				return fmt.Errorf("%s doesn't support this command", "deno")

			default:
				return fmt.Errorf("unsupported package manager: %s", pm)
			}

			noVolta, error := cmd.Flags().GetBool(_NO_VOLTA_FLAG)

			if error != nil {
				return error
			}

			// shouldUseVoltaWithPackageManager is true if:
			// 1. Volta is detected on the system (detectVolta())
			// 2. The detected package manager (pm) is one of npm, pnpm, or yarn (lo.Contains checks this)
			// 3. The --no-volta flag was NOT provided (!noVolta)
			shouldUseVoltaWithPackageManager := detectVolta() &&
				lo.Contains([]string{detect.NPM, detect.PNPM, detect.YARN}, pm) &&
				!noVolta

			if shouldUseVoltaWithPackageManager {
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

			return cmdRunner.Run()
		},
	}

	cmd.Flags().Bool(_NO_VOLTA_FLAG, false, "Disable Volta integration for this command") // New flag for Volta opt-out

	return cmd
}
