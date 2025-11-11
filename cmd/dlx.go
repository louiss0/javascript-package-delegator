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

// BuildDLXCommand builds command line for running temporary packages
func BuildDLXCommand(pm, yarnVersion, pkgOrURL string, args []string) (program string, argv []string, err error) {
	if pkgOrURL == "" {
		return "", nil, fmt.Errorf("package name or URL is required for dlx command")
	}

	switch pm {
	case "npm":
		argv = append([]string{pkgOrURL}, args...)
		return "npx", argv, nil
	case "pnpm":
		argv = append([]string{"dlx", pkgOrURL}, args...)
		return "pnpm", argv, nil
	case "yarn":
		if yarnSupportsDLX(yarnVersion) || yarnVersion == "" {
			argv = append([]string{"dlx", pkgOrURL}, args...)
			return "yarn", argv, nil
		}
		return "", nil, fmt.Errorf("yarn version %s does not support dlx", yarnVersion)
	case "bun":
		argv = append([]string{pkgOrURL}, args...)
		return "bunx", argv, nil
	case "deno":
		if !isURL(pkgOrURL) {
			return "", nil, fmt.Errorf("deno dlx requires a URL")
		}
		argv = append([]string{"run", pkgOrURL}, args...)
		return "deno", argv, nil
	default:
		return "", nil, fmt.Errorf("unsupported package manager: %s", pm)
	}
}

func NewDlxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dlx <package> [args...]",
		Short: "Execute temporary packages without installing",
		Long: `Execute packages temporarily without installing them first.
Downloads and runs packages on demand without permanent installation.

Examples:
  javascript-package-delegator dlx create-react-app my-app
  javascript-package-delegator dlx @angular/cli new my-project
  javascript-package-delegator dlx typescript --version
  javascript-package-delegator dlx prettier --check .`,
		Aliases: []string{"x"},
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pm, _ := cmd.Flags().GetString(AGENT_FLAG)
			goEnv := getGoEnvFromCommandContext(cmd)
			de := getDebugExecutorFromCommandContext(cmd)
			cmdRunner := getCommandRunnerFromCommandContext(cmd)

			packageName := args[0]
			packageArgs := args[1:]

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

			// Build command for running temporary packages
			execCommand, cmdArgs, err := BuildDLXCommand(pm, yarnVersion, packageName, packageArgs)
			if err != nil {
				return err
			}

			// Execute the command
			de.LogJSCommandIfDebugIsTrue(execCommand, cmdArgs...)
			cmdRunner.Command(execCommand, cmdArgs...)

			goEnv.ExecuteIfModeIsProduction(func() {
				log.Info("Running command", "cmd", execCommand, "args", strings.Join(cmdArgs, " "))
			})

			return cmdRunner.Run()
		},
	}

	return cmd
}
