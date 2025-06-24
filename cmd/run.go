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
	"path/filepath"
	"strings"

	// "github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/log"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

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
		RunE: func(cmd *cobra.Command, args []string) error {
			pm := getJS_PackageManagerNameFromCommandContext(cmd)

			goEnv := getGoEnvFromCommandContext(cmd)

			// If no script name provided, list available scripts

			if pm == "deno" {

				if len(args) == 0 {
					pkg, err := readDenoJSON()

					if err != nil {
						return err
					}

					if len(pkg.Tasks) == 0 {
						fmt.Fprint(cmd.OutOrStdout(), "No tasks found in deno.json")
						return nil
					}

					if goEnv.IsDevelopmentMode() {

						fmt.Fprintf(
							cmd.OutOrStdout(),
							"Here are the tasks %s",
							strings.Join(lo.Keys(pkg.Tasks), ","),
						)
						return nil

					}

					log.Info("Available tasks:")

					scriptEntries := lo.Entries(pkg.Tasks)

					scriptTableScaffold := table.New().
						Headers("name", "task")

					lo.ForEach(scriptEntries, func(item lo.Entry[string, string], index int) {

						scriptTableScaffold.Rows([]string{item.Key, item.Value})

					})

					scriptTable := lipgloss.NewStyle().Render(
						scriptTableScaffold.Render(),
					)
					fmt.Println(scriptTable)

					return nil
				}

			} else {

				if len(args) == 0 {
					pkg, err := readPackageJSON()

					if err != nil {
						return err
					}

					if len(pkg.Scripts) == 0 {
						fmt.Fprint(cmd.OutOrStdout(), "No scripts found in package.json")
						return nil
					}

					if goEnv.IsDevelopmentMode() {

						fmt.Fprintf(
							cmd.OutOrStdout(),
							"Here are the scripts %s",
							strings.Join(lo.Keys(pkg.Scripts), ","),
						)
						return nil

					}

					log.Info("Available scripts:")

					scriptEntries := lo.Entries(pkg.Scripts)

					scriptTableScaffold := table.New().
						Headers("name", "script")

					lo.ForEach(scriptEntries, func(item lo.Entry[string, string], index int) {

						scriptTableScaffold.Rows([]string{item.Key, item.Value})

					})

					scriptTable := lipgloss.NewStyle().Render(
						scriptTableScaffold.Render(),
					)
					fmt.Println(scriptTable)

					return nil
				}

			}

			scriptName := args[0]
			scriptArgs := args[1:]

			// Check if script exists when --if-present flag is used
			ifPresent, _ := cmd.Flags().GetBool("if-present")
			if ifPresent {
				pkg, err := readPackageJSON()
				if err != nil {
					return err
				}
				if _, exists := pkg.Scripts[scriptName]; !exists {
					goEnv.ExecuteIfModeIsProduction(func() {
						fmt.Printf("Script '%s' not found, skipping\n", scriptName)
					})
					return nil
				}
			}

			goEnv.ExecuteIfModeIsProduction(func() {
				log.Infof("Using %s\n", pm)

			})
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

					return fmt.Errorf("Don't pass %s here use the exec command instead", "--eval")

				}

				cmdArgs = append(cmdArgs, scriptArgs...)

			default:
				return fmt.Errorf("unsupported package manager: %s", pm)
			}

			// Execute the command
			cmdRunner := getCommandRunnerFromCommandContext(cmd)
			cmdRunner.Command(pm, cmdArgs...)

			goEnv.ExecuteIfModeIsProduction(func() {
				log.Infof("Running: %s %s\n", pm, strings.Join(cmdArgs, " "))
			})

			return cmdRunner.Run()
		},
	}

	// Add flags
	cmd.Flags().Bool("if-present", false, "Run script only if it exists")

	return cmd
}

type PackageJSON struct {
	Scripts map[string]string `json:"scripts"`
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

type DenoJSON struct {
	Tasks map[string]string `json:"tasks"`
}

func readDenoJSON() (*DenoJSON, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	denoJSONPath := filepath.Join(cwd, "deno.json")
	data, err := os.ReadFile(denoJSONPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read deno.json: %w", err)
	}

	var pkg DenoJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("failed to parse deno.json: %w", err)
	}

	return &pkg, nil
}
