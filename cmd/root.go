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
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
	"github.com/joho/godotenv"
	"github.com/louiss0/javascript-package-delegator/detect"
	"github.com/louiss0/javascript-package-delegator/env"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const _LOCKFILE = "lockfile"
const _PACKAGE_NAME = "package-name"

const _JS_PACKAGE_MANAGER_KEY = "js_pkm"

const _OS_PACKAGE_MANAGER_KEY = "os_pkm"

const _INTERACTIVE_FLAG = "interactive"

const _SUPPORTED_CONFIG_PATHS_KEY = "supported_paths"

const _VIPER_CONFIG_INSTANCE_KEY = "viper_config_instance"

const _COMMAND_RUNNER_KEY = "command_runner"
const _GO_ENV = "go_env"

type CommandRunner interface {
	Command(string, ...string)
	Run() error
}

type _ExecCommandFunc func(string, ...string) *exec.Cmd

type executor struct {
	execCommandFunc _ExecCommandFunc
	cmd             *exec.Cmd
}

func newExecutor(execCommandFunc _ExecCommandFunc) *executor {
	return &executor{
		execCommandFunc: execCommandFunc,
	}
}

func (e *executor) Command(name string, args ...string) {

	e.cmd = e.execCommandFunc(name, args...)

}

func (e executor) Run() error {
	return e.cmd.Run()
}

func NewRootCmd(commandRunner CommandRunner, jsPackageManagerDetector func() (string, error)) *cobra.Command {

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

			error := godotenv.Load()

			goEnv, error := env.NewGoEnv()

			if error != nil {

				return error
			}

			homeDir, error := os.UserHomeDir()

			if error != nil {

				return error
			}

			var supportedConfigPaths []string

			vf := viper.New()

			supportedConfigPaths = []string{
				fmt.Sprintf("%s/.config/", homeDir),
				fmt.Sprintf("%s/.local/share/", homeDir),
			}

			lo.ForEach(supportedConfigPaths, func(path string, index int) {

				vf.AddConfigPath(path)
			})

			vf.SetConfigName("jpd.config")

			vf.SetConfigType("toml")

			packageName, error := jsPackageManagerDetector()

			if error != nil {

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

			lo.ForEach([][2]any{
				{_GO_ENV, goEnv},
				{_PACKAGE_NAME, packageName},
				{_SUPPORTED_CONFIG_PATHS_KEY, supportedConfigPaths},
				{_VIPER_CONFIG_INSTANCE_KEY, vf},
				{_COMMAND_RUNNER_KEY, commandRunner},
			}, func(item [2]any, index int) {

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
		Run: func(cmd *cobra.Command, args []string) {

			interactiveFlag, error := cmd.Flags().GetBool(_INTERACTIVE_FLAG)

			if error != nil {

				log.Error(error.Error())

			}

			var Form struct {
				OS_PackageManager string
				JS_PackageManager string
				ConfigPath        string
			}

			if interactiveFlag {

				supportedConfigPaths := getSupportedPathsFromCommandContext(cmd)

				error := huh.NewForm(
					huh.NewGroup(
						huh.NewSelect[string]().
							Title("OS Package Manager").
							Description("Pick from the selected OS package managers").
							Options(huh.NewOptions(detect.SupportedOperatingSystemPackageManagers[:]...)...).
							Value(&Form.OS_PackageManager),
						huh.NewSelect[string]().
							Title("JS Package Manager").
							Description("Pick from the selected JS package managers").
							Options(huh.NewOptions(detect.SupportedJSPackageManagers[:]...)...).
							Value(&Form.JS_PackageManager),
						huh.NewSelect[string]().
							Title("Config File Path").
							Description("Pick the path you want to use for the config").
							Options(huh.NewOptions(supportedConfigPaths...)...).
							Value(&Form.ConfigPath),
					).
						Title("Javascript Package Delegator Setup").
						Description("This form is supposed to help you setup JPD according to your preferences ").
						WithTheme(huh.ThemeDracula()),
				).Run()

				if error != nil {

					log.Error(error.Error())

				}

				vf := getViperInstanceFronCommandContext(cmd)

				vf.Set(_OS_PACKAGE_MANAGER_KEY, Form.OS_PackageManager)
				vf.Set(_JS_PACKAGE_MANAGER_KEY, Form.JS_PackageManager)

				configFilePath := filepath.Join(Form.ConfigPath, "jpd.config.toml")

				vf.WriteConfigAs(configFilePath)

				log.Infof("Your config file was created at this path %s", configFilePath)

				log.Info(
					"It has these values",
					_OS_PACKAGE_MANAGER_KEY, Form.OS_PackageManager,
					_JS_PACKAGE_MANAGER_KEY, Form.JS_PackageManager,
				)

			}

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

	cmd.Flags().BoolP(
		"interactive",
		"i",
		false,
		"Allows the user to setup the config file for jpd",
	)

	// Cobra also supports local flags, which will only run
	// when this action is called directly.

	return cmd
}

var rootCmd = NewRootCmd(newExecutor(exec.Command), detect.DetectJSPacakgeManager)

func getPackageNameFromCommandContext(cmd *cobra.Command) string {

	ctx := cmd.Context()

	return ctx.Value(_PACKAGE_NAME).(string)

}

func getGoEnvFromCommandContext(cmd *cobra.Command) env.GoEnv {

	ctx := cmd.Context()
	return ctx.Value(_GO_ENV).(env.GoEnv)

}

func getSupportedPathsFromCommandContext(cmd *cobra.Command) []string {

	ctx := cmd.Context()

	return ctx.Value(_SUPPORTED_CONFIG_PATHS_KEY).([]string)

}

func getViperInstanceFronCommandContext(cmd *cobra.Command) *viper.Viper {
	ctx := cmd.Context()

	return ctx.Value(_VIPER_CONFIG_INSTANCE_KEY).(*viper.Viper)
}

func getCommandRunnerFromCommandContext(cmd *cobra.Command) CommandRunner {
	ctx := cmd.Context()

	return ctx.Value(_COMMAND_RUNNER_KEY).(CommandRunner)
}

func installJSManager(jsPkgMgr, osPkgMgr string) error {
	cmdRunner := newExecutor(exec.Command)
	var cmdName string
	var cmdArgs []string

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

	constructNixProfileCommand := func(jsPkgMgr string) (string, []string) {
		var profile string

		error := huh.NewText().
			Value(&profile).
			Run()

		if profile == "" || error != nil {
			return "nix profile", []string{"install", fmt.Sprintf("nixpkgs#%s", jsPkgMgr)}
		}

		return "nix profile", []string{"install", profile, fmt.Sprintf("nixpkgs#%s", jsPkgMgr)}
	}

	switch jsPkgMgr {
	case "deno":
		switch osPkgMgr {
		case "brew":
			cmdName = "brew"
			cmdArgs = []string{"install", jsPkgMgr}
		case "winget":
			cmdName = "winget"
			cmdArgs = []string{"install", "--id", "DenoLand.Deno", "-e"}
		case "scoop", "choco":
			cmdName = osPkgMgr
			cmdArgs = []string{"install", jsPkgMgr}
		case "nix":
			answer, error := promptUserToSelectNixEnvORProfile()

			if error != nil {
				return error
			}

			if answer == supportedNixInstallationChoices[0] {
				cmdName, cmdArgs = constructNixProfileCommand(jsPkgMgr)
				break
			}

			cmdName = "nix-env"
			cmdArgs = []string{"-iA", fmt.Sprintf("nixpkgs.%s", jsPkgMgr)}

		default:
			return fmt.Errorf("unsupported OS package manager: %s", osPkgMgr)
		}
	case "bun":
		switch osPkgMgr {
		case "brew":
			cmdName = "sh"
			cmdArgs = []string{"-c", "curl -fsSL https://bun.sh/install | bash"}
		case "scoop":
			cmdName = "scoop"
			cmdArgs = []string{"install", jsPkgMgr}
		default:
			return fmt.Errorf("bun not supported on %s", osPkgMgr)
		}
	case "npm":
		switch osPkgMgr {
		case "brew":
			cmdName = "brew"
			cmdArgs = []string{"install", "node"}
		case "winget":
			cmdName = "winget"
			cmdArgs = []string{"install", "Node.js"}
		case "scoop", "choco":
			cmdName = osPkgMgr
			cmdArgs = []string{"install", "nodejs-lts"}
		case "nix":
			answer, error := promptUserToSelectNixEnvORProfile()

			if error != nil {
				return error
			}

			if answer == supportedNixInstallationChoices[0] {
				cmdName, cmdArgs = constructNixProfileCommand("node")
				break
			}

			cmdName = "nix-env"
			cmdArgs = []string{"-iA", fmt.Sprintf("nixpkgs.%s", jsPkgMgr)}
		default:
			return fmt.Errorf("unsupported OS package manager: %s", osPkgMgr)
		}
	case "pnpm":
		switch osPkgMgr {
		case "brew":
			cmdName = "brew"
			cmdArgs = []string{"install", jsPkgMgr}
		case "winget":
			cmdName = "winget"
			cmdArgs = []string{"install", "-e", "--id", "pnpm.pnpm"}
		case "scoop", "choco":
			cmdName = osPkgMgr
			cmdArgs = []string{"install", jsPkgMgr}
		case "nix":
			answer, error := promptUserToSelectNixEnvORProfile()

			if error != nil {
				return error
			}

			if answer == supportedNixInstallationChoices[0] {
				cmdName, cmdArgs = constructNixProfileCommand(jsPkgMgr)
				break
			}

			cmdName = "nix-env"
			cmdArgs = []string{"-iA", fmt.Sprintf("nixpkgs.%s", jsPkgMgr)}
		default:
			return fmt.Errorf("unsupported OS package manager: %s", osPkgMgr)
		}
	case "yarn":
		switch osPkgMgr {
		case "brew":
			cmdName = "brew"
			cmdArgs = []string{"install", jsPkgMgr}
		case "winget":
			cmdName = "winget"
			cmdArgs = []string{"install", "--id", "Yarn.Yarn", "-e"}
		case "scoop", "choco":
			cmdName = osPkgMgr
			cmdArgs = []string{"install", jsPkgMgr}
		case "nix":
			answer, error := promptUserToSelectNixEnvORProfile()

			if error != nil {
				return error
			}

			if answer == supportedNixInstallationChoices[0] {
				cmdName, cmdArgs = constructNixProfileCommand(jsPkgMgr)
				break
			}

			cmdName = "nix-env"
			cmdArgs = []string{"-iA", fmt.Sprintf("nixpkgs.%s", jsPkgMgr)}
		default:
			return fmt.Errorf("unsupported OS package manager: %s", osPkgMgr)
		}
	default:
		return fmt.Errorf("unsupported JS package manager: %s", jsPkgMgr)
	}

	cmdRunner.Command(cmdName, cmdArgs...)
	err := cmdRunner.Run()

	if err != nil {
		return err
	}

	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.ExecuteContext(context.Background())
	if err != nil {
		os.Exit(1)
	}
}
