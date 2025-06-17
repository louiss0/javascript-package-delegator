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
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/louiss0/javascript-package-delegator/detect"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

type PackageJSON struct {
	Scripts map[string]string `json:"scripts"`
}

func NewRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [script] [args...]",
		Short: "Run scripts using the detected package manager",
		Long: `Run package.json scripts using the appropriate package manager.
Equivalent to 'nr' command - detects npm, yarn, pnpm, or bun and runs the script.

Examples:
  javascript-package-delegator run             # List available scripts
  javascript-package-delegator run dev         # Run dev script
  javascript-package-delegator run build --prod # Run build script with args
  javascript-package-delegator run test -- --watch # Run test with npm-style args`,
		Aliases: []string{"r"},
		Run: func(cmd *cobra.Command, args []string) {
			if err := runScript(args, cmd); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	// Add flags
	cmd.Flags().Bool("if-present", false, "Run script only if it exists")

	return cmd
}

func runScript(args []string, cmd *cobra.Command) error {
	pm, err := detect.JSPackageManager()
	if err != nil {
		return fmt.Errorf("failed to detect package manager: %w", err)
	}

	// If no script name provided, list available scripts
	if len(args) == 0 {
		return listScripts()
	}

	scriptName := args[0]
	scriptArgs := args[1:]

	// Check if script exists when --if-present flag is used
	ifPresent, _ := cmd.Flags().GetBool("if-present")
	if ifPresent {
		if exists, err := scriptExists(scriptName); err != nil {
			return err
		} else if !exists {
			fmt.Printf("Script '%s' not found, skipping\n", scriptName)
			return nil
		}
	}

	fmt.Printf("Using %s\n", pm)

	// Build command based on package manager
	var cmdArgs []string
	switch pm {
	case "npm":
		cmdArgs = []string{"run", scriptName}
		if len(scriptArgs) > 0 {
			cmdArgs = append(cmdArgs, "--")
			cmdArgs = append(cmdArgs, scriptArgs...)
		}
		if ifPresent {
			cmdArgs = append([]string{"run", "--if-present", scriptName}, scriptArgs...)
		}

	case "yarn":
		cmdArgs = []string{"run", scriptName}
		cmdArgs = append(cmdArgs, scriptArgs...)

	case "pnpm":
		cmdArgs = []string{"run", scriptName}
		if len(scriptArgs) > 0 {
			cmdArgs = append(cmdArgs, "--")
			cmdArgs = append(cmdArgs, scriptArgs...)
		}
		if ifPresent {
			cmdArgs = append([]string{"run", "--if-present", scriptName}, scriptArgs...)
		}

	case "bun":
		cmdArgs = []string{"run", scriptName}
		cmdArgs = append(cmdArgs, scriptArgs...)

	case "deno":
		cmdArgs = []string{"task", scriptName}

		if lo.Contains(scriptArgs, "--eval") {

			return fmt.Errorf("Don't pass %s  here use the exce command instead", "--eval")
		}

		cmdArgs = append(cmdArgs, scriptArgs...)

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

func listScripts() error {

	pkg, err := readPackageJSON()
	if err != nil {
		return err
	}

	if len(pkg.Scripts) == 0 {
		fmt.Println("No scripts found in package.json")
		return nil
	}

	fmt.Println("Available scripts:")
	for name, command := range pkg.Scripts {
		fmt.Printf("  %s: %s\n", name, command)
	}

	return nil
}

func scriptExists(scriptName string) (bool, error) {
	pkg, err := readPackageJSON()
	if err != nil {
		return false, err
	}

	_, exists := pkg.Scripts[scriptName]
	return exists, nil
}

func readPackageJSON() (*PackageJSON, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	packageJSONPath := filepath.Join(cwd, "package.json")
	data, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read package.json: %w", err)
	}

	var pkg PackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("failed to parse package.json: %w", err)
	}

	return &pkg, nil
}
