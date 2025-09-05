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

// BuildCreateCommand builds command line for running package create commands
func BuildCreateCommand(pm, yarnVersion, name string, args []string) (program string, argv []string, err error) {
	// Special handling for deno - requires URL as name
	if pm == "deno" {
		if name == "" {
			return "", nil, fmt.Errorf("deno create requires a URL as the first argument")
		}
		if !isURL(name) {
			return "", nil, fmt.Errorf("deno create requires a valid URL, got: %s", name)
		}
		return "deno", append([]string{"run", name}, args...), nil
	}

	// For all other package managers, require name and reject URLs
	if name == "" {
		return "", nil, fmt.Errorf("package name is required for create command")
	}
	if isURL(name) {
		return "", nil, fmt.Errorf("URLs are not supported for %s, use deno instead", pm)
	}

	// Normalize create bin - add "create-" prefix if not already present
	bin := name
	if !strings.HasPrefix(name, "create-") {
		bin = "create-" + name
	}

	// Build command based on package manager
	switch pm {
	case "npm":
		// npm exec create-<name> -- <args>
		argv = append([]string{"exec", bin, "--"}, args...)
		return "npm", argv, nil
	case "pnpm":
		// pnpm exec create-<name> <args>
		argv = append([]string{"exec", bin}, args...)
		return "pnpm", argv, nil
	case "yarn":
		yarnMajor := ParseYarnMajor(yarnVersion)
		if yarnMajor <= 1 {
			// Yarn v1 -> use npx: npx create-<name> <args>
			argv = append([]string{bin}, args...)
			return "npx", argv, nil
		} else {
			// Yarn v2+ -> use dlx: yarn dlx create-<name> <args>
			argv = append([]string{"dlx", bin}, args...)
			return "yarn", argv, nil
		}
	case "bun":
		// bunx create-<name> <args>
		argv = append([]string{bin}, args...)
		return "bunx", argv, nil
	default:
		return "", nil, fmt.Errorf("unsupported package manager: %s", pm)
	}
}

// NewCreateCmd creates a new Cobra command for the "create" functionality
func NewCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [name|url] [args...]",
		Short: "Scaffold a new project using create runners",
		Long: `Scaffold a new project using the appropriate package manager's create command.
This command delegates to the package manager's create functionality to bootstrap new projects.

Package Manager Behavior:
- npm: Runs 'npm exec create-<name> -- <args>'
- pnpm: Runs 'pnpm exec create-<name> <args>'
- yarn v1: Runs 'npx create-<name> <args>'
- yarn v2+: Runs 'yarn dlx create-<name> <args>'
- bun: Runs 'bunx create-<name> <args>'
- deno: Runs 'deno run <url> <args>' (expects URL as first argument)

Examples:
  jpd create react-app my-app
  jpd create vite@latest my-app -- --template react-swc
  jpd create next-app myapp --typescript --tailwind
  jpd -a deno create https://deno.land/x/fresh/init.ts my-fresh-app`,
		Aliases: []string{"c"},
		Args: func(cmd *cobra.Command, args []string) error {
			// All package managers (including deno) require at least one argument
			if len(args) < 1 {
				return fmt.Errorf("requires at least one argument (package name or URL for deno)")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			pm, _ := cmd.Flags().GetString(AGENT_FLAG)
			goEnv := getGoEnvFromCommandContext(cmd)
			cmdRunner := getCommandRunnerFromCommandContext(cmd)
			de := getDebugExecutorFromCommandContext(cmd)

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

			// Extract name and args
			name := ""
			packageArgs := []string{}
			if len(args) > 0 {
				name = args[0]
				packageArgs = args[1:]
			}

			// Build command for creating projects
			execCommand, cmdArgs, err := BuildCreateCommand(pm, yarnVersion, name, packageArgs)
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
