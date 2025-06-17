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

	"github.com/louiss0/javascript-package-delegator/detect"
	"github.com/spf13/cobra"
)

func NewUninstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall <packages...>",
		Short: "Uninstall packages using the detected package manager",
		Long: `Uninstall packages using the appropriate package manager.
Equivalent to 'nun' command - detects npm, yarn, pnpm, or bun and runs the uninstall command.

Examples:
  javascript-package-delegator uninstall lodash       # Uninstall lodash
  javascript-package-delegator uninstall lodash react # Uninstall multiple packages
  javascript-package-delegator uninstall -g typescript # Uninstall global package`,
		Aliases: []string{"un", "remove", "rm"},
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := runUninstall(args, cmd); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	// Add flags
	cmd.Flags().BoolP("global", "g", false, "Uninstall global packages")

	return cmd
}

func runUninstall(packages []string, cmd *cobra.Command) error {
	pm, err := detect.JSPackageManager()
	if err != nil {
		return fmt.Errorf("failed to detect package manager: %w", err)
	}

	fmt.Printf("Using %s\n", pm)

	// Get flags
	global, _ := cmd.Flags().GetBool("global")

	// Build command based on package manager and flags
	var cmdArgs []string
	switch pm {
	case "npm":
		cmdArgs = []string{"uninstall"}
		cmdArgs = append(cmdArgs, packages...)
		if global {
			cmdArgs = append(cmdArgs, "--global")
		}

	case "yarn":
		cmdArgs = []string{"remove"}
		cmdArgs = append(cmdArgs, packages...)
		if global {
			cmdArgs = append(cmdArgs, "--global")
		}

	case "pnpm":
		cmdArgs = []string{"remove"}
		cmdArgs = append(cmdArgs, packages...)
		if global {
			cmdArgs = append(cmdArgs, "--global")
		}

	case "bun":
		cmdArgs = []string{"remove"}
		cmdArgs = append(cmdArgs, packages...)
		if global {
			cmdArgs = append(cmdArgs, "--global")
		}

	default:
		return fmt.Errorf("unsupported package manager: %s", pm)
	}

	// Execute the command
	execCmd := exec.Command(pm, cmdArgs...)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	execCmd.Stdin = os.Stdin

	fmt.Printf("Running: %s %s\n", pm, strings.Join(cmdArgs, " "))
	return execCmd.Run()
}
