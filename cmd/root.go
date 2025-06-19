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

type jpdConfig struct {
	js_package_manager string
	os_package_manager string
}

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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	rootCmd.ExecuteContext(context.Background())
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
