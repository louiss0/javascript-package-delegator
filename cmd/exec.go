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

// Package cmd provides command-line interface implementations for the JavaScript package delegator.
package cmd

import (
	// standard library
	"fmt"
	"strings"

	// external
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	// internal
	"github.com/louiss0/javascript-package-delegator/detect"
)

func NewExecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec <package> [args...]",
		Short: "Execute packages using the detected package manager",
		Long: `Execute packages using the appropriate package manager's execute command.
Equivalent to 'nlx' command - detects npm, yarn, pnpm, or bun and runs npx/yarn dlx/pnpm dlx/bunx.

Examples:
  javascript-package-delegator exec create-react-app my-app
  javascript-package-delegator exec @angular/cli new my-project
  javascript-package-delegator exec typescript --version
  javascript-package-delegator exec prettier --check .`,
		Aliases: []string{"e"},
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pm, _ := cmd.Flags().GetString(AGENT_FLAG)
			goEnv := getGoEnvFromCommandContext(cmd)
			cmdRunner := getCommandRunnerFromCommandContext(cmd)
			de := getDebugExecutorFromCommandContext(cmd)

			packageName := args[0]
			packageArgs := args[1:]

			goEnv.ExecuteIfModeIsProduction(func() {
				log.Infof("Using %s\n", pm)
			})

			// Build command based on package manager
			var execCommand string
			var cmdArgs []string

			switch pm {
			case "npm":
				execCommand = "npx"
				cmdArgs = []string{packageName}
				cmdArgs = append(cmdArgs, packageArgs...)

			case "yarn":
				// Check if it's Yarn v1 or v2+
				yarnVersion, err := detect.DetectYarnVersion(
					getYarnVersionRunnerCommandContext(cmd),
				)

				if err != nil {
					// Fallback to yarn v1 style
					execCommand = "yarn"
					cmdArgs = []string{packageName}
					cmdArgs = append(cmdArgs, packageArgs...)
				} else if strings.HasPrefix(yarnVersion, "1.") {
					// Yarn v1
					execCommand = "yarn"
					cmdArgs = []string{packageName}
					cmdArgs = append(cmdArgs, packageArgs...)
				} else {
					// Yarn v2+
					execCommand = "yarn"
					cmdArgs = []string{"dlx", packageName}
					cmdArgs = append(cmdArgs, packageArgs...)
				}

			case "pnpm":
				execCommand = "pnpm"
				cmdArgs = []string{"dlx", packageName}
				cmdArgs = append(cmdArgs, packageArgs...)

			case "bun":
				execCommand = "bunx"
				cmdArgs = []string{packageName}
				cmdArgs = append(cmdArgs, packageArgs...)

			case "deno":
				return fmt.Errorf("deno does not have a dlx/x equivalent")

			default:
				return fmt.Errorf("unsupported package manager: %s", pm)
			}

			// Execute the command
			cmdRunner.Command(execCommand, cmdArgs...)
			de.LogJSCommandIfDebugIsTrue(execCommand, cmdArgs...)

			goEnv.ExecuteIfModeIsProduction(func() {
				log.Infof("Running: %s %s\n", execCommand, strings.Join(cmdArgs, " "))
			})

			return cmdRunner.Run()
		},
	}

	return cmd
}
