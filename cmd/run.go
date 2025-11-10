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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	// "github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

type taskSelectorUI struct {
	selectedValue string
	selectUI      huh.Select[string]
}

func newTaskSelectorUI(options []string) TaskUISelector {
	return &taskSelectorUI{
		selectUI: *huh.NewSelect[string]().
			Title("Select a task").
			Description("Pick a task from one of your selected tasks").
			Options(huh.NewOptions(options...)...),
	}
}

func (t taskSelectorUI) Value() string {
	return t.selectedValue
}

func (t *taskSelectorUI) Run() error {
	return t.selectUI.Value(&t.selectedValue).Run()
}

func NewRunCmd(newTaskSelectorUI func(options []string) TaskUISelector) *cobra.Command {
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
			pm, _ := cmd.Flags().GetString(AGENT_FLAG)
			cmdRunner := getCommandRunnerFromCommandContext(cmd)

			goEnv := getGoEnvFromCommandContext(cmd)
			de := getDebugExecutorFromCommandContext(cmd)

			// If no script name provided, list available scripts

			var selectedPackage string

			if pm == "deno" {
				if len(args) == 0 {
					pkg, err := readDenoJSON()
					if err != nil {
						return err
					}

					if len(pkg.Tasks) == 0 {
						return fmt.Errorf("no tasks found in deno.json")
					}

					if goEnv.IsDevelopmentMode() {
						_, _ = fmt.Fprintf(
							cmd.OutOrStdout(),
							"Here are the scripts %s",
							strings.Join(lo.Keys(pkg.Tasks), ","),
						)
					}

					taskSelectorUI := newTaskSelectorUI(lo.Keys(pkg.Tasks))

					if err := taskSelectorUI.Run(); err != nil {
						return err
					}

					selectedPackage = taskSelectorUI.Value()
				}
			} else {
				if len(args) == 0 {
					pkg, err := readPackageJSONAndUnmarshalScripts()
					if err != nil {
						return err
					}

					if len(pkg.Scripts) == 0 {
						return fmt.Errorf("no scripts found in package.json")
					}

					if goEnv.IsDevelopmentMode() {
						_, _ = fmt.Fprintf(
							cmd.OutOrStdout(),
							"Here are the scripts %s",
							strings.Join(lo.Keys(pkg.Scripts), ","),
						)
					}

					taskSelectorUI := newTaskSelectorUI(lo.Keys(pkg.Scripts))

					if err := taskSelectorUI.Run(); err != nil {
						return err
					}

					selectedPackage = taskSelectorUI.Value()

				}
			}

			scriptName := lo.TernaryF(
				len(args) == 0,
				func() string {
					return selectedPackage
				},
				func() string {
					return args[0]
				},
			)

			var scriptArgs []string

			if len(args) > 1 {
				scriptArgs = args[1:]
			}

			// Check if script exists when --if-present flag is used
			ifPresent, _ := cmd.Flags().GetBool("if-present")
			if ifPresent {
				pkg, err := readPackageJSONAndUnmarshalScripts()
				if err != nil {
					return err
				}
				if _, exists := pkg.Scripts[scriptName]; !exists {
					goEnv.ExecuteIfModeIsProduction(func() {
						log.Info("Script not found, skipping", "script", scriptName)
					})
					return nil
				}
			}

			goEnv.ExecuteIfModeIsProduction(func() {
				log.Info("Using package manager", "pm", pm)
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
					return fmt.Errorf("don't pass --eval here use the exec command instead")
				}

				cmdArgs = append(cmdArgs, scriptArgs...)

			default:
				return fmt.Errorf("unsupported package manager: %s", pm)
			}

			// Execute the command
			cmdRunner.Command(pm, cmdArgs...)
			de.LogJSCommandIfDebugIsTrue(pm, cmdArgs...)

			goEnv.ExecuteIfModeIsProduction(func() {
				log.Info("Running command", "pm", pm, "args", strings.Join(cmdArgs, " "))
			})

			return cmdRunner.Run()
		},
	}

	// Add flags
	cmd.Flags().Bool("if-present", false, "Run script only if it exists")

	return cmd
}

type PackageJSONScripts struct {
	Scripts map[string]string `json:"scripts"`
}

func readPackageJSONAndUnmarshalScripts() (*PackageJSONScripts, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	packageJSONPath := filepath.Join(cwd, "package.json")
	data, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read package.json: %w", err)
	}

	var pkg PackageJSONScripts
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

// ParsePackageNames extracts package names from "name@version" strings.
// Handles scoped packages correctly by splitting on the last '@' character.
func ParsePackageNames(depWithVersions []string) []string {
	names := make([]string, len(depWithVersions))
	for i, dep := range depWithVersions {
		// Split on last '@' to handle scoped packages like @types/node@1.0.0
		lastAtIndex := strings.LastIndex(dep, "@")
		if lastAtIndex == -1 {
			// No version specified, use the whole string
			names[i] = dep
		} else if strings.HasPrefix(dep, "@") && lastAtIndex == 0 {
			// Scoped package without version like @types/node
			names[i] = dep
		} else {
			// Extract name before the last '@'
			names[i] = dep[:lastAtIndex]
		}
	}
	return names
}

// IsYarnPnpProject checks if the current directory is a Yarn PnP project
// by looking for .pnp.cjs or .pnp.data.json files.
func IsYarnPnpProject(cwd string) bool {
	pnpCjsPath := filepath.Join(cwd, ".pnp.cjs")
	pnpDataPath := filepath.Join(cwd, ".pnp.data.json")

	_, cjsErr := os.Stat(pnpCjsPath)
	_, dataErr := os.Stat(pnpDataPath)

	return cjsErr == nil || dataErr == nil
}

// MissingNodePackages checks which packages are missing from node_modules.
// Returns up to maxMissing packages to avoid excessive checking and noisy logs.
func MissingNodePackages(cwd string, depNames []string) []string {
	const maxMissing = 10
	missing := make([]string, 0, maxMissing)

	nodeModulesPath := filepath.Join(cwd, "node_modules")

	for _, name := range depNames {
		if len(missing) >= maxMissing {
			// Stop checking after finding maxMissing packages to avoid performance issues
			break
		}

		// Handle scoped packages: @scope/name -> node_modules/@scope/name
		// Regular packages: name -> node_modules/name
		var packagePath string
		if strings.HasPrefix(name, "@") {
			// Scoped package
			packagePath = filepath.Join(nodeModulesPath, name)
		} else {
			// Regular package
			packagePath = filepath.Join(nodeModulesPath, name)
		}

		_, err := os.Stat(packagePath)
		if os.IsNotExist(err) {
			missing = append(missing, name)
		}
		// Ignore other errors (permissions, etc.) and assume package exists
	}

	return missing
}
