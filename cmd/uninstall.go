// Package cmd provides command-line interface implementations for the JavaScript package delegator.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/louiss0/javascript-package-delegator/detect"
)

type dependencyMultiSelectUI struct {
	selectedValues []string
	selectUI       huh.MultiSelect[string]
}

func newDependencySelectorUI(options []string) DependencyUIMultiSelector {
	return &dependencyMultiSelectUI{
		selectUI: *huh.NewMultiSelect[string]().
			Title("Select a dependency to uninstall").
			Description("Pick a dependency to uninstall").
			Options(huh.NewOptions(options...)...),
	}
}

func (t dependencyMultiSelectUI) Values() []string {
	return t.selectedValues
}

func (t dependencyMultiSelectUI) Run() error {
	return t.selectUI.Value(&t.selectedValues).Run()
}

func extractProdAndDevDependenciesFromPackageJSON() ([]string, error) {
	type PackageJSONDependencies struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	packageJSONPath := filepath.Join(cwd, "package.json")
	data, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read package.json: %w", err)
	}

	var pkg PackageJSONDependencies

	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("failed to parse package.json: %w", err)
	}

	prodAndDevDependenciesMerged := lo.Map(
		lo.Entries(lo.Assign(pkg.Dependencies, pkg.DevDependencies)),
		func(item lo.Entry[string, string], index int) string {
			return fmt.Sprintf("%s@%s", item.Key, item.Value)
		},
	)

	return prodAndDevDependenciesMerged, nil
}

func extractImportsFromDenoJSON() ([]string, error) {
	type DenoJSONDependencies struct {
		Imports map[string]string `json:"imports"`
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	denoJSONPath := filepath.Join(cwd, "deno.json")
	data, err := os.ReadFile(denoJSONPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read deno.json: %w", err)
	}

	var pkg DenoJSONDependencies

	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("failed to parse deno.json: %w", err)
	}

	importValues := lo.Values(pkg.Imports)

	return importValues, nil
}

const _INTERACTIVE_FLAG = "interactive"

func NewUninstallCmd(newDependencySelectorUI func(options []string) DependencyUIMultiSelector) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall <packages...>",
		Short: "Uninstall packages using the detected package manager",
		Long: `Uninstall packages using thme appropriate package manager.
Equivalent to 'nun' command - detects npm, yarn, pnpm, or bun and runs the uninstall command.

Examples:
  javascript-package-delegator uninstall lodash       # Uninstall lodash
  javascript-package-delegator uninstall lodash react # Uninstall multiple packages
  javascript-package-delegator uninstall -g typescript # Uninstall global package`,
		Aliases: []string{"un", "remove", "rm"},
		Args: func(cmd *cobra.Command, args []string) error {
			interactive, err := cmd.Flags().GetBool(_INTERACTIVE_FLAG)
			if err != nil {
				return err
			}

			return lo.Ternary(
				!interactive,
				cobra.MinimumNArgs(1)(cmd, args),
				nil,
			)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			pm, _ := cmd.Flags().GetString(AGENT_FLAG)

			goEnv := getGoEnvFromCommandContext(cmd)
			de := getDebugExecutorFromCommandContext(cmd)

			goEnv.ExecuteIfModeIsProduction(func() {
				log.Infof("Using %s\n", pm)
			})

			// Get flags
			global, _ := cmd.Flags().GetBool("global")

			interactive, err := cmd.Flags().GetBool(_INTERACTIVE_FLAG)
			if err != nil {
				return err
			}

			var selectedPackages []string

			if interactive {

				packageIsDeno := pm == detect.DENO

				var (
					dependencies []string
					err          error
				)

				if packageIsDeno {

					dependencies, err = extractImportsFromDenoJSON()
					if err != nil {
						return err
					}
				} else {

					dependencies, err = extractProdAndDevDependenciesFromPackageJSON()
					if err != nil {
						return err
					}
				}

				if len(dependencies) == 0 {
					return fmt.Errorf("no packages found for interactive uninstall")
				}

				dependencySelectorUI := newDependencySelectorUI(dependencies)

				if error := dependencySelectorUI.Run(); error != nil {
					return error
				}

				selectedPackages = dependencySelectorUI.Values()

			}

			// Build command based on package manager and flags
			var cmdArgs []string
			switch pm {
			case detect.NPM:
				cmdArgs = []string{"uninstall"}
				cmdArgs = lo.Flatten([][]string{cmdArgs, selectedPackages, args})

				if global {
					cmdArgs = append(cmdArgs, "--global")
				}

			case detect.YARN:
				cmdArgs = []string{"remove"}
				cmdArgs = lo.Flatten([][]string{cmdArgs, selectedPackages, args})

				if global {
					cmdArgs = append(cmdArgs, "--global")
				}

			case detect.PNPM:
				cmdArgs = []string{"remove"}
				cmdArgs = lo.Flatten([][]string{cmdArgs, selectedPackages, args})

				if global {
					cmdArgs = append(cmdArgs, "--global")
				}

			case detect.BUN:
				cmdArgs = []string{"remove"}
				cmdArgs = lo.Flatten([][]string{cmdArgs, selectedPackages, args})

				if global {
					cmdArgs = append(cmdArgs, "--global")
				}

			case detect.DENO:
				if global {
					cmdArgs = []string{"uninstall"}
				} else {
					cmdArgs = []string{"remove"}
				}
				cmdArgs = lo.Flatten([][]string{cmdArgs, selectedPackages, args})

			default:
				return fmt.Errorf("unsupported package manager: %s", pm)
			}

			// Execute the command
			cmdRunner := getCommandRunnerFromCommandContext(cmd)
			cmdRunner.Command(pm, cmdArgs...)
			de.LogJSCommandIfDebugIsTrue(pm, cmdArgs...)
			goEnv.ExecuteIfModeIsProduction(func() {
				log.Infof("Running: %s %s\n", pm, strings.Join(cmdArgs, " "))
			})

			return cmdRunner.Run()
		},
	}

	// Add flags
	cmd.Flags().BoolP(_GLOBAL_FLAG, "g", false, "Uninstall global packages")
	cmd.Flags().BoolP(_INTERACTIVE_FLAG, "i", false, "Uninstall packages interactively")

	cmd.MarkFlagsMutuallyExclusive(_GLOBAL_FLAG, _INTERACTIVE_FLAG)

	return cmd
}
