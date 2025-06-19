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
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/charmbracelet/huh"
	"github.com/louiss0/javascript-package-delegator/detect"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const _LOCKFILE = "lockfile"
const _PACKAGE_NAME = "package-name"
const _APP_ENV_KEY = "app_env"

type AppEnv string

const _DEV = AppEnv("development")
const _PROD = AppEnv("production")

const _JS_PACKAGE_MANAGER_KEY = "js_pkm"

const _OS_PACKAGE_MANAGER_KEY = "os_pkm"

func NewRootCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:     "jpd",
		Version: "0.0.0",
		Short:   "JavaScript Package Delegator - A universal package manager interface",
		Long: `JavaScript Package Delegator (jpd) - A universal package manager interface that detects
and delegates to the appropriate package manager (npm, yarn, pnpm, bun, or deno) based on
lock files and config files in your project.

Inspired by @antfu/ni, this tool provides a unified CLI experience across different
JavaScript runtimes and package managers, making it easy to work in teams with different
preferences or switch between Node.js and Deno projects.

Available commands:
  install    - Install packages (equivalent to 'ni')
  run        - Run package.json scripts (equivalent to 'nr')
  exec       - Execute packages (equivalent to 'nlx')
  update     - Update packages (equivalent to 'nup')
  uninstall  - Uninstall packages (equivalent to 'nun')
  clean-install - Clean install with frozen lockfile (equivalent to 'nci')
  agent      - Show detected package manager (equivalent to 'na')`,

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {

			ve := viper.New()
			ve.SetEnvPrefix("APP")
			ve.AutomaticEnv()
			ve.BindEnv(_APP_ENV_KEY)

			appEnv := ve.GetString(_APP_ENV_KEY)

			allowedAppEnvValues := []string{string(_DEV), string(_PROD)}

			if appEnv != "" && !lo.Contains(allowedAppEnvValues, appEnv) {

				return fmt.Errorf(
					"The APP_ENV variable can only be %v",
					allowedAppEnvValues,
				)

			}

			packageName, error := detect.DetectJSPacakgeManager()

			if error != nil {

				vf := viper.New()

				homeDir, error := os.UserHomeDir()

				if error != nil {

					return error
				}

				vf.AddConfigPath(
					fmt.Sprintf("%s/.config/", homeDir),
				)

				vf.AddConfigPath(
					fmt.Sprintf("%s/.local/share/", homeDir),
				)

				vf.SetConfigName("jpd.config")

				vf.SetConfigType("toml")

				jsPackageManagerFromConfig := vf.GetString(_JS_PACKAGE_MANAGER_KEY)
				osPackageManagerFromConfig := vf.GetString(_OS_PACKAGE_MANAGER_KEY)

				if jsPackageManagerFromConfig != "" && osPackageManagerFromConfig == "" {

					detectedOSManager, error := detect.SupportedOperatingSystemPackageManager()

					if error != nil {

						return error
					}

					error = installJSManager(jsPackageManagerFromConfig, detectedOSManager)

					if error != nil {
						return fmt.Errorf("Something went wrong with the command from what you have chosen %v", error)
					}

					return nil

				}

				if osPackageManagerFromConfig != "" && jsPackageManagerFromConfig == "" {

					choices := detect.SupportedJSPackageManagers

					var selectedJSPkgManager string

					error := huh.NewSelect[string]().
						Title("Choose JS package manager").
						Options(huh.NewOptions(choices[:]...)...).
						Value(&selectedJSPkgManager).
						Run()

					if error != nil {

						return fmt.Errorf("Well there's nothing else to do If you have a JS Package Manager you'd like to use please use it")
					}

					error = installJSManager(selectedJSPkgManager, osPackageManagerFromConfig)

					if error != nil {
						return fmt.Errorf("Something went wrong with the command from what you have chosen %v", error)
					}

					return nil

				}

				if osPackageManagerFromConfig != "" && jsPackageManagerFromConfig != "" {
					// This assumes that both are filled!
					error := installJSManager(jsPackageManagerFromConfig, osPackageManagerFromConfig)

					if error != nil {

						return error
					}

					return nil

				}

				return error
			}

			cmdContext := cmd.Context()

			lo.ForEach([][2]string{
				{_APP_ENV_KEY, appEnv},
				{_PACKAGE_NAME, packageName},
			}, func(item [2]string, index int) {

				key := item[0]
				value := item[1]

				cmdContext = context.WithValue(
					cmdContext,
					key,
					value,
				)

			})

			cmd.SetContext(cmdContext)

			return nil

		},
	}

	// Add all subcommands
	cmd.AddCommand(NewInstallCmd())
	cmd.AddCommand(NewRunCmd())
	cmd.AddCommand(NewExecCmd())
	cmd.AddCommand(NewUpdateCmd())
	cmd.AddCommand(NewUninstallCmd())
	cmd.AddCommand(NewCleanInstallCmd())
	cmd.AddCommand(NewAgentCmd())

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mini-clis.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.

	return cmd
}

var rootCmd = NewRootCmd()

func getPackageNameFromCommandContext(cmd *cobra.Command) string {

	ctx := cmd.Context()

	return ctx.Value(_PACKAGE_NAME).(string)

}

func getAppEnvFromCommandContext(cmd *cobra.Command) AppEnv {

	ctx := cmd.Context()

	return ctx.Value(_APP_ENV_KEY).(AppEnv)

}

func installJSManager(jsPkgMgr, osPkgMgr string) error {
	var cmd *exec.Cmd

	supportedNixInstallationChoices := [2]string{"profiles", "env"}
	promptUserToSelectNixEnvORProfile := func() (string, error) {

		var selection string

		error := huh.NewSelect[string]().
			Options(huh.NewOptions(supportedNixInstallationChoices[:]...)...).
			Value(&selection).
			Run()

		if error != nil {

			return "", error
		}

		return selection, nil
	}

	constructNixProfileCommandBasedOnProfileChoice := func(jsPkgMgr string) *exec.Cmd {
		var profile string

		error := huh.NewText().
			Value(&profile).
			Run()

		if profile == "" || error != nil {

			return exec.Command("nix profile", "install", fmt.Sprintf("nixpkgs#%s", jsPkgMgr))
		}

		return exec.Command("nix profile", "install", profile, fmt.Sprintf("nixpkgs#%s", jsPkgMgr))
	}

	switch jsPkgMgr {
	case "deno":
		switch osPkgMgr {
		case "brew":
			cmd = exec.Command("brew", "install", jsPkgMgr)
		case "winget":
			cmd = exec.Command("winget", "install", "--id", "DenoLand.Deno", "-e")
		case "scoop", "choco":
			cmd = exec.Command(osPkgMgr, "install", jsPkgMgr)
		case "nix":
			answer, error := promptUserToSelectNixEnvORProfile()

			if error != nil {

				return error
			}

			if answer == supportedNixInstallationChoices[0] {

				cmd = constructNixProfileCommandBasedOnProfileChoice(jsPkgMgr)
				break
			}

			cmd = exec.Command("nix-env", "-iA", fmt.Sprintf("nixpkgs.%s", jsPkgMgr))

		default:
			return fmt.Errorf("unsupported OS package manager: %s", osPkgMgr)
		}
	case "bun":
		switch osPkgMgr {
		case "brew":
			cmd = exec.Command("sh", "-c", "curl -fsSL https://bun.sh/install | bash")
		case "scoop":
			cmd = exec.Command("scoop", "install", jsPkgMgr)
		default:
			return fmt.Errorf("bun not supported on %s", osPkgMgr)
		}
	case "npm":
		switch osPkgMgr {
		case "brew":
			cmd = exec.Command("brew", "install", "node")
		case "winget":
			cmd = exec.Command("winget", "install", "Node.js")
		case "scoop", "choco":
			cmd = exec.Command(osPkgMgr, "install", "nodejs-lts")
		case "nix":
			answer, error := promptUserToSelectNixEnvORProfile()

			if error != nil {

				return error
			}

			if answer == supportedNixInstallationChoices[0] {

				cmd = constructNixProfileCommandBasedOnProfileChoice("node")
				break
			}

			cmd = exec.Command("nix-env", "-iA", fmt.Sprintf("nixpkgs.%s", jsPkgMgr))
		default:
			return fmt.Errorf("unsupported OS package manager: %s", osPkgMgr)
		}
	case "pnpm":
		switch osPkgMgr {
		case "brew":
			cmd = exec.Command("brew", "install", jsPkgMgr)
		case "winget":
			cmd = exec.Command("winget", "install", "-e", "--id", "pnpm.pnpm")
		case "scoop", "choco":
			cmd = exec.Command(osPkgMgr, "install", jsPkgMgr)
		case "nix":
			answer, error := promptUserToSelectNixEnvORProfile()

			if error != nil {

				return error
			}

			if answer == supportedNixInstallationChoices[0] {

				cmd = constructNixProfileCommandBasedOnProfileChoice(jsPkgMgr)
				break
			}

			cmd = exec.Command("nix-env", "-iA", fmt.Sprintf("nixpkgs.%s", jsPkgMgr))
		default:
			return fmt.Errorf("unsupported OS package manager: %s", osPkgMgr)
		}
	case "yarn":
		switch osPkgMgr {
		case "brew":
			cmd = exec.Command("brew", "install", jsPkgMgr)
		case "winget":
			cmd = exec.Command("winget", "install", "--id", "Yarn.Yarn", "-e")
		case "scoop", "choco":
			cmd = exec.Command(osPkgMgr, "install", jsPkgMgr)
		case "nix":
			answer, error := promptUserToSelectNixEnvORProfile()

			if error != nil {

				return error
			}

			if answer == supportedNixInstallationChoices[0] {

				cmd = constructNixProfileCommandBasedOnProfileChoice(jsPkgMgr)
				break
			}

			cmd = exec.Command("nix-env", "-iA", fmt.Sprintf("nixpkgs.%s", jsPkgMgr))
		default:
			return fmt.Errorf("unsupported OS package manager: %s", osPkgMgr)
		}
	default:
		return fmt.Errorf("unsupported JS package manager: %s", jsPkgMgr)
	}

	err := cmd.Run()

	if err != nil {
		return err
	}

	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	rootCmd.ExecuteContext(context.Background())
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
