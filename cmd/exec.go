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

// ParseYarnMajor extracts the major version number from a yarn version string
func ParseYarnMajor(version string) int {
	if version == "" {
		return 0
	}

	// Handle simple cases like "3" or "berry-3.1.0"
	if strings.HasPrefix(version, "berry-") {
		version = strings.TrimPrefix(version, "berry-")
	}

	// Extract first character and convert to int
	if len(version) > 0 && version[0] >= '1' && version[0] <= '9' {
		return int(version[0] - '0')
	}

	return 0 // unknown
}

// isURL checks if a string is a valid HTTP or HTTPS URL
func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// BuildExecCommand builds command line for running local dependencies
func BuildExecCommand(pm, yarnVersion, bin string, args []string) (program string, argv []string, err error) {
	if bin == "" {
		return "", nil, fmt.Errorf("binary name is required for exec command")
	}

	switch pm {
	case "npm":
		argv = append([]string{"exec", bin, "--"}, args...)
		return "npm", argv, nil
	case "pnpm":
		argv = append([]string{"exec", bin}, args...)
		return "pnpm", argv, nil
	case "yarn":
		argv = append([]string{bin}, args...)
		return "yarn", argv, nil
	case "bun":
		argv = append([]string{"x", bin}, args...)
		return "bun", argv, nil
	case "deno":
		argv = append([]string{"run", bin}, args...)
		return "deno", argv, nil
	default:
		return "", nil, fmt.Errorf("unsupported package manager: %s", pm)
	}
}

func NewExecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec <bin> [args...]",
		Short: "Execute local dependencies using package manager exec",
		Long: `Execute local dependencies using the appropriate package manager's exec command.
This runs binaries from locally installed packages in your project.

Examples:
  javascript-package-delegator exec eslint --version
  javascript-package-delegator exec ts-node src/index.ts
  javascript-package-delegator exec vite build
  javascript-package-delegator exec prettier --check .`,
		Aliases: []string{"e"},
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pm, _ := cmd.Flags().GetString(AGENT_FLAG)
			goEnv := getGoEnvFromCommandContext(cmd)
			cmdRunner := getCommandRunnerFromCommandContext(cmd)
			de := getDebugExecutorFromCommandContext(cmd)

			binaryName := args[0]
			binaryArgs := args[1:]

			goEnv.ExecuteIfModeIsProduction(func() {
				log.Info("Using package manager", "pm", pm)
			})

			// Get yarn version if needed
			yarnVersion := ""
			if pm == "yarn" {
				if version, err := detect.DetectYarnVersion(
					getYarnVersionRunnerCommandContext(cmd),
				); err == nil {
					yarnVersion = version
				}
			}

			// Build command for executing local dependencies
			execCommand, cmdArgs, err := BuildExecCommand(pm, yarnVersion, binaryName, binaryArgs)
			if err != nil {
				return err
			}

			// Execute the command
			cmdRunner.Command(execCommand, cmdArgs...)
			de.LogJSCommandIfDebugIsTrue(execCommand, cmdArgs...)

			goEnv.ExecuteIfModeIsProduction(func() {
				log.Info("Running command", "cmd", execCommand, "args", strings.Join(cmdArgs, " "))
			})

			return cmdRunner.Run()
		},
	}

	return cmd
}
