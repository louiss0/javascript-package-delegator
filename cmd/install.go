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
	"strings"

	"github.com/spf13/cobra"
)

func NewInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install [packages...]",
		Short: "Install packages using the detected package manager",
		Long: `Install packages using the appropriate package manager based on lock files.
Equivalent to 'ni' command - detects npm, yarn, pnpm, or bun and runs the appropriate install command.

Examples:
  node-package-delegator install           # Install all dependencies
  node-package-delegator install lodash    # Install lodash
  node-package-delegator install -D vitest # Install vitest as dev dependency
  node-package-delegator install -g typescript # Install globally`,
		Aliases: []string{"i", "add"},
		Run: func(cmd *cobra.Command, args []string) {
			if err := runInstall(args, cmd); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	// Add flags
	cmd.Flags().BoolP("dev", "D", false, "Install as dev dependency")
	cmd.Flags().BoolP("global", "g", false, "Install globally")
	cmd.Flags().BoolP("production", "P", false, "Install production dependencies only")
	cmd.Flags().Bool("frozen", false, "Install with frozen lockfile")
	cmd.Flags().BoolP("interactive", "i", false, "Interactive package selection")

	return cmd
}

func runInstall(packages []string, cmd *cobra.Command) error {
	pm, err := detectPackageManager()
	if err != nil {
		return fmt.Errorf("failed to detect package manager: %w", err)
	}

	fmt.Printf("Using %s\n", pm)

	// Build command based on package manager and flags
	var cmdArgs []string
	switch pm {
	case "npm":
		if len(packages) == 0 {
			cmdArgs = []string{"install"}
		} else {
			cmdArgs = []string{"install"}
			cmdArgs = append(cmdArgs, packages...)
		}
		if dev, _ := cmd.Flags().GetBool("dev"); dev {
			cmdArgs = append(cmdArgs, "--save-dev")
		}
		if global, _ := cmd.Flags().GetBool("global"); global {
			cmdArgs = append(cmdArgs, "--global")
		}
		if production, _ := cmd.Flags().GetBool("production"); production {
			cmdArgs = append(cmdArgs, "--omit=dev")
		}
		if frozen, _ := cmd.Flags().GetBool("frozen"); frozen {
			cmdArgs = append(cmdArgs, "--package-lock-only")
		}

	case "yarn":
		if len(packages) == 0 {
			cmdArgs = []string{"install"}
		} else {
			cmdArgs = []string{"add"}
			cmdArgs = append(cmdArgs, packages...)
		}
		if dev, _ := cmd.Flags().GetBool("dev"); dev {
			cmdArgs = append(cmdArgs, "--dev")
		}
		if global, _ := cmd.Flags().GetBool("global"); global {
			cmdArgs = append(cmdArgs, "--global")
		}
		if production, _ := cmd.Flags().GetBool("production"); production {
			cmdArgs = append(cmdArgs, "--production")
		}
		if frozen, _ := cmd.Flags().GetBool("frozen"); frozen {
			cmdArgs = append(cmdArgs, "--frozen-lockfile")
		}

	case "pnpm":
		if len(packages) == 0 {
			cmdArgs = []string{"install"}
		} else {
			cmdArgs = []string{"add"}
			cmdArgs = append(cmdArgs, packages...)
		}
		if dev, _ := cmd.Flags().GetBool("dev"); dev {
			cmdArgs = append(cmdArgs, "--save-dev")
		}
		if global, _ := cmd.Flags().GetBool("global"); global {
			cmdArgs = append(cmdArgs, "--global")
		}
		if production, _ := cmd.Flags().GetBool("production"); production {
			cmdArgs = append(cmdArgs, "--prod")
		}
		if frozen, _ := cmd.Flags().GetBool("frozen"); frozen {
			cmdArgs = append(cmdArgs, "--frozen-lockfile")
		}

	case "bun":
		if len(packages) == 0 {
			cmdArgs = []string{"install"}
		} else {
			cmdArgs = []string{"add"}
			cmdArgs = append(cmdArgs, packages...)
		}
		if dev, _ := cmd.Flags().GetBool("dev"); dev {
			cmdArgs = append(cmdArgs, "--development")
		}
		if global, _ := cmd.Flags().GetBool("global"); global {
			cmdArgs = append(cmdArgs, "--global")
		}
		if production, _ := cmd.Flags().GetBool("production"); production {
			cmdArgs = append(cmdArgs, "--production")
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

// detectPackageManager detects the package manager based on lock files
func detectPackageManager() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Check for lock files in order of preference
	lockFiles := map[string]string{
		"bun.lockb":      "bun",
		"pnpm-lock.yaml": "pnpm",
		"yarn.lock":      "yarn",
		"package-lock.json": "npm",
	}

	for lockFile, pm := range lockFiles {
		if _, err := os.Stat(filepath.Join(cwd, lockFile)); err == nil {
			return pm, nil
		}
	}

	// Default to npm if no lock file is found
	return "npm", nil
}

