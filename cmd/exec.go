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
	"strconv"
	"strings"

	// external
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	// internal
	"github.com/louiss0/javascript-package-delegator/detect"
)

// ParseYarnMajor extracts the major version number from a yarn version string
func ParseYarnMajor(version string) int {
	major, _, ok := parseYarnMajorMinor(version)
	if !ok {
		return 0
	}
	return major
}

func parseYarnMajorMinor(version string) (int, int, bool) {
	version = strings.TrimSpace(version)
	if version == "" {
		return 0, 0, false
	}

	if after, ok := strings.CutPrefix(version, "berry-"); ok {
		version = after
	}

	// Skip any leading non-digit characters
	start := -1
	for i := 0; i < len(version); i++ {
		if version[i] >= '0' && version[i] <= '9' {
			start = i
			break
		}
	}
	if start == -1 {
		return 0, 0, false
	}

	version = version[start:]
	parts := strings.Split(version, ".")
	if len(parts) == 0 {
		return 0, 0, false
	}

	majorStr := leadingDigits(parts[0])
	if majorStr == "" {
		return 0, 0, false
	}

	major, err := strconv.Atoi(majorStr)
	if err != nil {
		return 0, 0, false
	}

	minor := 0
	if len(parts) > 1 {
		minorStr := leadingDigits(parts[1])
		if minorStr != "" {
			minor, err = strconv.Atoi(minorStr)
			if err != nil {
				return 0, 0, false
			}
		}
	}

	return major, minor, true
}

func leadingDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r < '0' || r > '9' {
			break
		}
		b.WriteRune(r)
	}
	return b.String()
}

func yarnSupportsDLX(version string) bool {
	major, minor, ok := parseYarnMajorMinor(version)
	if !ok {
		return false
	}
	if major >= 2 {
		return true
	}
	return major == 1 && minor >= 22
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
		Short: "Execute packages from the local project or a remote URL",
		Long: `Execute packages from the local project or a remote URL.
Execute local dependencies using the appropriate package manager's exec command.
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
