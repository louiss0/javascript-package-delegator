// Package cmd provides command-line interface implementations for the JavaScript package delegator.
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/louiss0/javascript-package-delegator/detect"
	"github.com/louiss0/javascript-package-delegator/internal/deps"
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

func (t *dependencyMultiSelectUI) Run() error {
	return t.selectUI.Value(&t.selectedValues).Run()
}

const _INTERACTIVE_FLAG = "interactive"

func NewUninstallCmd(newDependencySelectorUI func(options []string) DependencyUIMultiSelector) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall <packages...>",
		Short: "Uninstall packages using the detected package manager",
		Long: `Remove packages from your project using the appropriate package manager.
Equivalent to 'nun' command - detects npm, yarn, pnpm, or bun and runs the uninstall command.

Examples:
  javascript-package-delegator uninstall lodash       # Uninstall lodash
  javascript-package-delegator uninstall lodash react # Uninstall multiple packages
  javascript-package-delegator uninstall -g typescript # Uninstall global package`,
		Aliases: []string{"un", "remove", "rm"},
		RunE: func(cmd *cobra.Command, args []string) error {
			pm, _ := cmd.Flags().GetString(AGENT_FLAG)

			goEnv := getGoEnvFromCommandContext(cmd)
			de := getDebugExecutorFromCommandContext(cmd)

			targetDir, err := cmd.Flags().GetString(_CWD_FLAG)
			if err != nil {
				return fmt.Errorf("failed to parse --%s flag: %w", _CWD_FLAG, err)
			}
			if targetDir == "" {
				targetDir, err = os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to determine working directory: %w", err)
				}
			}

			goEnv.ExecuteIfModeIsProduction(func() {
				log.Info("Using package manager", "pm", pm)
			})

			// Get flags
			global, _ := cmd.Flags().GetBool("global")

			interactive, err := cmd.Flags().GetBool(_INTERACTIVE_FLAG)
			if err != nil {
				return err
			}

			// Validate args: require at least one unless interactive mode
			if !interactive && len(args) == 0 {
				return fmt.Errorf("requires at least 1 arg(s), only received 0")
			}

			var selectedPackages []string

			if interactive {

				packageIsDeno := pm == detect.DENO

				var (
					dependencies []string
					err          error
				)

				if packageIsDeno {

					dependencies, err = deps.ExtractImportsFromDenoJSON(targetDir)
					if err != nil {
						return err
					}
				} else {

					dependencies, err = deps.ExtractProdAndDevDependenciesFromPackageJSON()
					if err != nil {
						return err
					}
				}

				if len(dependencies) == 0 {
					return fmt.Errorf("no packages found for interactive uninstall")
				}

				dependencySelectorUI := newDependencySelectorUI(dependencies)

				if err := dependencySelectorUI.Run(); err != nil {
					return err
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
				log.Info("Running command", "pm", pm, "args", strings.Join(cmdArgs, " "))
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
