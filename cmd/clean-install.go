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
	"github.com/spf13/cobra"
)

func NewCleanInstallCmd() *cobra.Command {
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

			pm := getPackageNameFromCommandContext(cmd)
			goMode := getGoModeFromCommandContext(cmd)
			// Build command based on package manager
			var cmdArgs []string
			switch pm {
			case "npm":
				cmdArgs = []string{"ci"}

			case "yarn":
				// Yarn v1 uses install --frozen-lockfile, v2+ uses install --immutable
				yarnVersion, err := getYarnVersion()
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

			// Execute the command
			cmdRunner := getCommandRunnerFromCommandContext(cmd)
			cmdRunner.Command(pm, cmdArgs...)

			if goMode != _DEV {
				log.Infof("Running: %s %s\n", pm, strings.Join(cmdArgs, " "))
			}

			return cmdRunner.Run()
		},
	}

	return cmd
}

func getYarnVersion() (string, error) {

	cmd := exec.Command("yarn", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", nil
	}
	return strings.TrimSpace(string(output)), nil
}
