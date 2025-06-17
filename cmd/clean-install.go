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
		Run: func(cmd *cobra.Command, args []string) {
			if err := runCleanInstall(args, cmd); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	return cmd
}

func runCleanInstall(args []string, cmd *cobra.Command) error {
	pm, err := detect.JSPackageManager()
	if err != nil {
		return fmt.Errorf("failed to detect package manager: %w", err)
	}

	fmt.Printf("Using %s\n", pm)

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
