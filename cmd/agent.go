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

	"github.com/louiss0/javascript-package-delegator/detect"
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
			if err := runAgent(args); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	return cmd
}

func runAgent(args []string) error {
	pm, err := detect.JSPackageManager()

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

	foundLockFile := "none (defaulting to npm)"
	for _, lockFileAndPackageName := range detect.LOCKFILES {

		lockFile := lockFileAndPackageName[0]
		packageName := lockFileAndPackageName[1]

		if _, err := os.Stat(filepath.Join(cwd, lockFile)); err == nil && packageName == pm {
			foundLockFile = lockFile
			break
		}
	}

	fmt.Printf("Lock file: %s\n", foundLockFile)

	commmand := exec.Command(pm, args...)

	commmand.Stderr = os.Stderr
	commmand.Stdin = os.Stdin
	commmand.Stdout = os.Stdout

	return commmand.Run()
}
