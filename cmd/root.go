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
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
)

const _LOCKFILE = "lockfile"
const _PACKAGE_NAME = "package-name"

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

			var LOCKFILES = [7][2]string{
				{"deno.lock", "deno"},
				{"deno.json", "deno"},
				{"deno.jsonc", "deno"},
				{"bun.lockb", "bun"},
				{"pnpm-lock.yaml", "pnpm"},
				{"yarn.lock", "yarn"},
				{"package-lock.json", "npm"},
			}

			cwd, err := os.Getwd()

			if err != nil {
				return err
			}

			// Check for lock files and config files in order of preference
			cmdContext := cmd.Context()
			for _, lockFileAndPakageName := range LOCKFILES {

				lockFile := lockFileAndPakageName[0]
				packageName := lockFileAndPakageName[1]

				if _, err := os.Stat(filepath.Join(cwd, lockFile)); err == nil {

					slog.Info(fmt.Sprintf("Found lock file %s", lockFile))

					cmdContext = context.WithValue(
						cmdContext,
						_PACKAGE_NAME,
						packageName,
					)

					cmd.SetContext(cmdContext)

					return nil

				}
			}

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

func getPackageNameFromCommandContext(cmd cobra.Command) string {

	ctx := cmd.Context()

	value, ok := ctx.Value(_PACKAGE_NAME).(string)

	if !ok {
		return "npm"

	}

	return value

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
