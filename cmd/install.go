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
	// standard library
	"fmt"

	// external
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
	"github.com/samber/lo"
	"github.com/spf13/cobra"

	// internal
	"github.com/louiss0/javascript-package-delegator/custom_errors"
	"github.com/louiss0/javascript-package-delegator/custom_flags"
	"github.com/louiss0/javascript-package-delegator/detect"
	"github.com/louiss0/javascript-package-delegator/services"
)

// Add flags
const (
	_DEV_FLAG        = "dev"
	_GLOBAL_FLAG     = "global"
	_PRODUCTION_FLAG = "production"
	_FROZEN_FLAG     = "frozen"
	_SEARCH_FLAG     = "search"
	_NO_VOLTA_FLAG   = "no-volta"
)


type packageMultiSelectUI struct {
	value         []string
	multiSelectUI *huh.MultiSelect[string]
}

func newPackageMultiSelectUI(packageInfo []services.PackageInfo) MultiUISelecter {
	return &packageMultiSelectUI{
		multiSelectUI: huh.NewMultiSelect[string]().
			Title("What packages do you want to install?").
			Options(
				lo.Map(
					packageInfo,
					func(packageInfo services.PackageInfo, index int) huh.Option[string] {
						return huh.NewOption(
							packageInfo.Name,
							fmt.Sprintf(
								"%s@%s",
								packageInfo.Name, packageInfo.Version,
							),
						)
					})...,
			),
	}
}

func (p packageMultiSelectUI) Values() []string {
	return p.value
}

func (p *packageMultiSelectUI) Run() error {
	return p.multiSelectUI.Value(&p.value).Run()
}

// NewInstallCmd creates a new Cobra command for the "install" functionality.
// This command delegates to the appropriate JavaScript package manager (npm, Yarn, pnpm, Bun, or Deno)
// to install project dependencies or specific packages.
// It also includes optional Volta integration to ensure consistent toolchain usage.
func NewInstallCmd(detectVolta func() bool, newPackageMultiSelectUI func([]services.PackageInfo) MultiUISelecter) *cobra.Command {
	searchFlag := custom_flags.NewEmptyStringFlag(_SEARCH_FLAG)

	cmd := &cobra.Command{
		Use:   "install [packages...]",
		Short: "Install packages using the detected package manager",
		Long: `Install packages using the appropriate package manager based on lock files.
Equivalent to 'ni' command - detects npm, yarn, pnpm, bun, or deno and runs the appropriate install command.

Examples:
  jpd install           # Install all dependencies
  jpd install lodash    # Install lodash
  jpd install -D vitest # Install vitest as dev dependency
  jpd install -g typescript # Install globally
  jpd install --no-volta # Install packages bypassing Volta, even if installed
`,
		Aliases: []string{"i", "add"},
		Args: func(cmd *cobra.Command, args []string) error {
			return lo.Ternary(
				searchFlag.String() != "" && len(args) > 0,
				custom_errors.CreateInvalidArgumentErrorWithMessage(
					"No arguments must be passed while the search flag is used",
				),
				nil,
			)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			pm, _ := cmd.Flags().GetString(AGENT_FLAG)
			goEnv := getGoEnvFromCommandContext(cmd)
			cmdRunner := getCommandRunnerFromCommandContext(cmd)
			de := getDebugExecutorFromCommandContext(cmd)
			if dbg, _ := cmd.Flags().GetBool(_DEBUG_FLAG); dbg {
				de.LogDebugMessageIfDebugIsTrue("Command start", "name", "install", "pm", pm)
			}

			// Build command based on package manager and flags
			var cmdArgs []string
			var selectedPackages []string

			if searchFlag.String() != "" {

				npmRegistryService := services.NewNpmRegistryService()

				packageInfo, err := npmRegistryService.SearchPackages(searchFlag.String())
				if err != nil {
					return err
				}

				if len(packageInfo) == 0 {
				return fmt.Errorf("query failed: %s", searchFlag.String())
				}

				installMultiSelect := newPackageMultiSelectUI(packageInfo)

				if err := installMultiSelect.Run(); err != nil {
					return err
				}

				selectedPackages = installMultiSelect.Values()

			}

			switch pm {
			case "npm":
				if len(args) == 0 {
					cmdArgs = lo.Flatten([][]string{{"install"}, selectedPackages})
				} else {
					cmdArgs = lo.Flatten([][]string{
						{"install"},
						args,
					})
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
				if len(args) == 0 {
					cmdArgs = lo.Flatten([][]string{{"install"}, selectedPackages})
				} else {
					cmdArgs = lo.Flatten([][]string{
						{"add"},
						args,
					})
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
				if len(args) == 0 {
					cmdArgs = lo.Flatten([][]string{{"install"}, selectedPackages})
				} else {
					cmdArgs = lo.Flatten([][]string{
						{"add"},
						selectedPackages,
						args,
					})
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
				if len(args) == 0 {
					cmdArgs = lo.Flatten([][]string{{"install"}, selectedPackages})
				} else {
					cmdArgs = lo.Flatten([][]string{
						{"add"},
						args,
					})
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

			case "deno":

				if len(args) == 0 {
				return fmt.Errorf("for deno one or more packages is required")
				}

				if production, _ := cmd.Flags().GetBool("production"); production {
				return fmt.Errorf("deno doesn't support prod")
				}

				if global, _ := cmd.Flags().GetBool("global"); global {

					cmdArgs = append([]string{"install"}, args...)
					break
				}

				cmdArgs = append([]string{"add"}, args...)

				if dev, _ := cmd.Flags().GetBool("dev"); dev {
					cmdArgs = append(cmdArgs, "--dev")
				}

			default:
				return fmt.Errorf("unsupported package manager: %s", pm)
			}

			noVolta, err := cmd.Flags().GetBool(_NO_VOLTA_FLAG)
			if err != nil {
				return err
			}

			// shouldUseVoltaWithPackageManager is true if:
			// 1. Volta is detected on the system (detectVolta())
			// 2. The detected package manager (pm) is one of npm, pnpm, or yarn (lo.Contains checks this)
			// 3. The --no-volta flag was NOT provided (!noVolta)
			shouldUseVoltaWithPackageManager := detectVolta() &&
				lo.Contains([]string{detect.NPM, detect.PNPM, detect.YARN}, pm) &&
				!noVolta

			if shouldUseVoltaWithPackageManager {

				completeVoltaCommand := lo.Flatten([][]string{
					detect.VOLTA_RUN_COMMAND,
					{pm},
					cmdArgs,
				})
				cmdRunner.Command(completeVoltaCommand[0], completeVoltaCommand[1:]...)

				goEnv.ExecuteIfModeIsProduction(func() {
					log.Info("Executing this ", "command", completeVoltaCommand)
				})
				de.LogJSCommandIfDebugIsTrue(completeVoltaCommand[0], completeVoltaCommand[1:]...)
			} else {

				cmdRunner.Command(pm, cmdArgs...)

				goEnv.ExecuteIfModeIsProduction(func() {
					log.Info("Executing this ", "command", append([]string{pm}, cmdArgs...))
				})
				de.LogJSCommandIfDebugIsTrue(pm, cmdArgs...)
			}

			// Execute the command
			return cmdRunner.Run()
		},
	}

	cmd.Flags().BoolP(_DEV_FLAG, "D", false, "Install as dev dependency")
	cmd.Flags().BoolP(_GLOBAL_FLAG, "g", false, "Install globally")
	cmd.Flags().BoolP(_PRODUCTION_FLAG, "P", false, "Install production dependencies only")
	cmd.Flags().Bool(_FROZEN_FLAG, false, "Install with frozen lockfile")
	cmd.Flags().VarP(&searchFlag, _SEARCH_FLAG, "s", "Interactive package search selection")
	cmd.Flags().Bool(_NO_VOLTA_FLAG, false, "Disable Volta integration for this command") // New flag for Volta opt-out

	return cmd
}
