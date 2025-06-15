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
	"path/filepath"

	"github.com/spf13/cobra"
)

func NewAgentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Show the detected package manager agent",
		Long: `Show information about the detected package manager agent.
Equivalent to 'na' command - detects and displays the package manager being used.

This command shows which package manager would be used based on lock files in the current directory.

Examples:
  javascript-package-delegator agent    # Show detected package manager`,
		Aliases: []string{"a"},
		Run: func(cmd *cobra.Command, args []string) {
			if err := runAgent(args, cmd); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	return cmd
}

func runAgent(args []string, cmd *cobra.Command) error {
	pm, err := detectPackageManager()
	if err != nil {
		return fmt.Errorf("failed to detect package manager: %w", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Show detailed information
	fmt.Printf("Detected package manager: %s\n", pm)
	fmt.Printf("Working directory: %s\n", cwd)

	// Show which lock file was found
	lockFiles := map[string]string{
		"bun.lockb":         "bun",
		"pnpm-lock.yaml":    "pnpm",
		"yarn.lock":         "yarn",
		"package-lock.json": "npm",
	}

	foundLockFile := "none (defaulting to npm)"
	for lockFile, pmName := range lockFiles {
		if _, err := os.Stat(filepath.Join(cwd, lockFile)); err == nil && pmName == pm {
			foundLockFile = lockFile
			break
		}
	}

	fmt.Printf("Lock file: %s\n", foundLockFile)

	// Show version if available
	if version, err := getPackageManagerVersion(pm); err == nil {
		fmt.Printf("Version: %s\n", version)
	} else {
		fmt.Printf("Version: unable to detect (%v)\n", err)
	}

	return nil
}

func getPackageManagerVersion(pm string) (string, error) {
	var cmd *exec.Cmd
	switch pm {
	case "npm":
		cmd = exec.Command("npm", "--version")
	case "yarn":
		cmd = exec.Command("yarn", "--version")
	case "pnpm":
		cmd = exec.Command("pnpm", "--version")
	case "bun":
		cmd = exec.Command("bun", "--version")
	default:
		return "", fmt.Errorf("unsupported package manager: %s", pm)
	}

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}
