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
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
	"github.com/joho/godotenv"
	"github.com/louiss0/javascript-package-delegator/build_info"
	"github.com/louiss0/javascript-package-delegator/custom_flags"
	"github.com/louiss0/javascript-package-delegator/detect"
	"github.com/louiss0/javascript-package-delegator/env"
	"github.com/louiss0/javascript-package-delegator/services"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

// Constants for context keys and configuration
const PACKAGE_NAME = "package-name"                      // Used for storing detected package name in context
const _GO_ENV = "go_env"                                 // Used for storing GoEnv in context
const _YARN_VERSION_OUTPUTTER = "yarn_version_outputter" // Key for YarnCommandVersionOutputter

const COMMAND_RUNNER_KEY = "command_runner"
const JPD_AGENT_ENV_VAR = "JPD_AGENT"
const DEBUG_FLAG = "debug"
const AGENT_FLAG = "agent"
const _CWD_FLAG = "cwd"

// CommandRunner Interface and its implementation
// This interface allows for mocking command execution in tests.
// **Remember:** always use the `Command()` before using the `Run()`
type CommandRunner interface {
	// A method that is used to determine whether run should be in debug mode or not
	// Use a boolean state in struct that implements this inter face and return that
	IsDebug() bool
	// This method uses `exec.Cmd` struct to execute commands behind the scenes
	// Make this method useful by creating a field that will hold `exec.Command`.
	// Then make a second field that will hold the
	Command(string, ...string)
	// This method calls the underlying `exec.Run()` to execute the command from `exec.Cmd`!
	Run() error
	SetTargetDir(string) error
}

type _ExecCommandFunc func(string, ...string) *exec.Cmd

type executor struct {
	execCommandFunc _ExecCommandFunc
	cmd             *exec.Cmd
	debug           bool
	targetDir       string
}

func newExecutor(execCommandFunc _ExecCommandFunc, debug bool) *executor {
	return &executor{
		execCommandFunc: execCommandFunc,
		debug:           debug,
	}
}

func (e executor) IsDebug() bool {

	return e.debug
}

func (e *executor) Command(name string, args ...string) {
	e.cmd = e.execCommandFunc(name, args...)
	e.cmd.Stdin = os.Stdin   // Ensure stdin is connected for interactive commands
	e.cmd.Stdout = os.Stdout // Ensure output goes to stdout
	e.cmd.Stderr = os.Stderr // Ensure errors go to stderr

	// Apply any previously set target directory
	if e.targetDir != "" {
		e.cmd.Dir = e.targetDir
	}
}

func (e *executor) SetTargetDir(dir string) error {
	fileInfo, err := os.Stat(dir) // Get file information
	if err != nil {
		return err
	}

	// Check if it's a directory
	if !fileInfo.IsDir() {
		return fmt.Errorf("target directory %s is not a directory", dir)
	}

	// Persist the target directory regardless of command initialization state
	e.targetDir = dir

	// If a command has already been created, update it immediately
	if e.cmd != nil {
		e.cmd.Dir = dir
	}
	return nil
}

func (e executor) Run() error {
	if e.cmd == nil {
		return fmt.Errorf("no command set to run")
	}

	if e.debug {
		log.Debug("Using this command: %s \n", strings.Join(e.cmd.Args, " "))
		if e.cmd.Dir != "" {
			log.Debug("Target directory: %s", e.cmd.Dir)
		}
	}

	return e.cmd.Run()
}

// Dependencies holds the external dependencies for testing and real execution

type Dependencies struct {
	CommandRunnerGetter                   func(bool) CommandRunner
	DetectJSPacakgeManagerBasedOnLockFile func(detectedLockFile string) (packageManager string, error error)
	YarnCommandVersionOutputter           detect.YarnCommandVersionOutputter
	NewCommandTextUI                      func(lockfile string) CommandUITexter
	DetectLockfile                        func() (lockfile string, error error)
	DetectJSPacakgeManager                func() (string, error)
	DetectVolta                           func() bool
	NewPackageMultiSelectUI               func([]services.PackageInfo) MultiUISelecter
	NewTaskSelectorUI                     func(options []string) TaskUISelector
	NewDependencyMultiSelectUI            func(options []string) DependencyUIMultiSelector
}

type CommandUITexter interface {
	Run() error
	Value() string
}

const VALID_INSTALL_COMMAND_STRING_RE = `^(?:[^\s=]+)\s+(?:[^\s=]+)\s+(?:[^\s=]+)(?:[^\s=]+\s+[^\s]+)*$`

var INVALID_COMMAND_STRUCTURE_ERROR_MESSAGE_STRUCTURE = []string{
	"You wrote this as your string %s ",
	"A command for installing a package is at least three words",
	"In the form write the command like you'd normally write a command like this",
	"[command] [subcommand or flag] [package]",
	"Place flags after the command",
}

func newCommandTextUI(lockfile string) CommandUITexter {

	return &CommandTextUI{
		textUI: huh.NewText().
			Title("Command").
			Description(
				lo.Ternary(
					lockfile != "",
					fmt.Sprintf(
						"We dectcted a lock file but there is no %s",
						detect.LockFileToPackageManagerMap[lockfile],
					),
					"The command you want to use to install your js package manager",
				),
			).
			Validate(func(s string) error {

				match, error := regexp.MatchString(VALID_INSTALL_COMMAND_STRING_RE, s)

				if error != nil {

					return error
				}

				if lockfile != "" && !strings.Contains(s, detect.LockFileToPackageManagerMap[lockfile]) {

					return fmt.Errorf("The command you entered does not contain the package manager command for %s", detect.LockFileToPackageManagerMap[lockfile])

				}

				if match {

					return nil

				}

				return fmt.Errorf(strings.Join(INVALID_COMMAND_STRUCTURE_ERROR_MESSAGE_STRUCTURE, "\n"), s)

			}),
	}
}

type CommandTextUI struct {
	value  string
	textUI *huh.Text
}

func (ui CommandTextUI) Value() string {

	return ui.value

}

func (ui *CommandTextUI) Run() error {

	return ui.textUI.Value(&ui.value).Run()

}

// NewRootCmd creates a new root command with injectable dependencies.
func NewRootCmd(deps Dependencies) *cobra.Command {
	cwdFlag := custom_flags.NewFolderPathFlag(_CWD_FLAG)
	cmd := &cobra.Command{
		Use:     "jpd",
		Version: build_info.CLI_VERSION.String(), // Default version or set via build process
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
  dlx        - Execute packages with package runner (dedicated package-runner command)
  update     - Update packages (equivalent to 'nup')
  uninstall  - Uninstall packages (equivalent to 'nun')
  clean-install - Clean install with frozen lockfile (equivalent to 'nci')
  agent      - Show detected package manager (equivalent to 'na')`,
		SilenceUsage: true,

		PersistentPreRunE: func(c *cobra.Command, args []string) error {
			// Load .env file
			err := godotenv.Load()

			if err != nil && !os.IsNotExist(err) {
				log.Error(err.Error()) // Log error, but don't stop execution unless critical
			}

			goEnv := env.NewGoEnv() // Instantiate GoEnv

			// Store dependencies and other derived values in the command context
			c_ctx := c.Context() // Capture the current context to pass into lo.ForEach

			isDebug, err := c.Flags().GetBool(DEBUG_FLAG)

			if err != nil {
				return err
			}

			commandRunner := deps.CommandRunnerGetter(isDebug)

			if cwd := cwdFlag.String(); cwd != "" {

				err := commandRunner.SetTargetDir(cwd)

				if err != nil {
					return err
				}

			}

			lo.ForEach([][2]any{
				{_GO_ENV, goEnv},
				{COMMAND_RUNNER_KEY, commandRunner},
				{_YARN_VERSION_OUTPUTTER, deps.YarnCommandVersionOutputter},
			}, func(item [2]any, index int) {
				c_ctx = context.WithValue(
					c_ctx,
					item[0],
					item[1],
				)
			})

			persistentFlags := c.Flags()
			agent, error := persistentFlags.GetString(AGENT_FLAG)

			if error != nil {

				return error
			}

			if agent != "" {

				persistentFlags.Set(AGENT_FLAG, agent)
				c.SetContext(c_ctx)
				return nil
			}

			agent, ok := os.LookupEnv(JPD_AGENT_ENV_VAR)

			if ok {

				if !lo.Contains(detect.SupportedJSPackageManagers[:], agent) {

					return fmt.Errorf(
						"The %s variable is set the wrong way use one of these values instead %v",
						JPD_AGENT_ENV_VAR,
						detect.SupportedJSPackageManagers,
					)

				}

				goEnv.ExecuteIfModeIsProduction(func() {
					log.Info("Using package manager", "agent", agent)

				})

				persistentFlags.Set(AGENT_FLAG, agent)
				c.SetContext(c_ctx)
				return nil

			}

			lockFile, err := deps.DetectLockfile()

			if err != nil {

				pm, err := deps.DetectJSPacakgeManager()

				if err != nil {

					commandTextUI := deps.NewCommandTextUI("")

					if err := commandTextUI.Run(); err != nil {

						return err
					}

					goEnv.ExecuteIfModeIsProduction(func() {

						log.Info("Installing the package manager using ", "command", commandTextUI.Value())

					})

					re := regexp.MustCompile(`\s+`)
					splitCommandString := re.Split(commandTextUI.Value(), -1)

					commandRunner.Command(splitCommandString[0], splitCommandString[1:]...)

					if err := commandRunner.Run(); err != nil {

						return err
					}

					return nil
				}

				persistentFlags.Set(AGENT_FLAG, pm)
				c.SetContext(c_ctx)
				return nil
			}

			// Package manager detection and potential installation logic
			pm, err := deps.DetectJSPacakgeManagerBasedOnLockFile(lockFile) // Use injected detector

			if err != nil {

				if errors.Is(detect.ErrNoPackageManager, err) {
					// The package manager indicated by the lock file is not installed
					// Let's check if any other package manager is available in PATH
					goEnv.ExecuteIfModeIsProduction(func() {
						log.Warn("Package manager indicated by lock file is not installed")
						log.Info("Checking for other available package managers...")
					})

					// Try to detect any available package manager from PATH
					pm, err := deps.DetectJSPacakgeManager()
					if err == nil {
						// Found an alternative package manager!
						goEnv.ExecuteIfModeIsProduction(func() {
							log.Infof("Found %s as an alternative package manager\n", pm)
						})
						persistentFlags.Set(AGENT_FLAG, pm)
						c.SetContext(c_ctx)
						return nil
					}

					// No package manager found at all, prompt for installation
					goEnv.ExecuteIfModeIsProduction(func() {
						log.Warn("No package manager found on the system")
						log.Warn("You'll be asked to provide a command to install one")
					})

					commandTextUI := deps.NewCommandTextUI(lockFile)

					if err := commandTextUI.Run(); err != nil {

						return err
					}

					goEnv.ExecuteIfModeIsProduction(func() {

						log.Info("Installing the package manager using ", "command", commandTextUI.Value())

					})

					re := regexp.MustCompile(`\s+`)
					splitCommandString := re.Split(commandTextUI.Value(), -1)

					commandRunner.Command(splitCommandString[0], splitCommandString[1:]...)

					if err := commandRunner.Run(); err != nil {

						return err
					}

					return nil
				}

				// Return any other errors as-is
				return err
			}

			persistentFlags.Set(AGENT_FLAG, pm)
			c.SetContext(c_ctx)
			return nil
		},
	}

	// Add all subcommands
	cmd.AddCommand(NewInstallCmd(deps.DetectVolta, deps.NewPackageMultiSelectUI))
	cmd.AddCommand(NewRunCmd(deps.NewTaskSelectorUI))
	cmd.AddCommand(NewExecCmd())
	cmd.AddCommand(NewDlxCmd())
	cmd.AddCommand(NewUpdateCmd())
	cmd.AddCommand(NewUninstallCmd(deps.NewDependencyMultiSelectUI))
	cmd.AddCommand(NewCleanInstallCmd(deps.DetectVolta))
	cmd.AddCommand(NewAgentCmd())
	cmd.AddCommand(NewCompletionCmd())

	cmd.PersistentFlags().BoolP(DEBUG_FLAG, "d", false, "Make commands run in debug mode")

	cmd.PersistentFlags().StringP(AGENT_FLAG, "a", "", "Select the JS package manager you want to use")

	cmd.PersistentFlags().VarP(&cwdFlag, _CWD_FLAG, "C", "Set the working directory for commands")

	cmd.RegisterFlagCompletionFunc(
		AGENT_FLAG,
		cobra.FixedCompletions(detect.SupportedJSPackageManagers[:], cobra.ShellCompDirectiveNoFileComp),
	)

	return cmd
}

// Global variable for the root command, initialized in init()
var rootCmd *cobra.Command

func init() {
	// Initialize the global rootCmd with real implementations of its dependencies
	rootCmd = NewRootCmd(
		Dependencies{
			CommandRunnerGetter: func(b bool) CommandRunner {
				return newExecutor(exec.Command, b)
			}, // Use the newExecutor constructor
			DetectJSPacakgeManagerBasedOnLockFile: func(detectedLockFile string) (packageManager string, error error) {

				return detect.DetectJSPacakgeManagerBasedOnLockFile(detectedLockFile, detect.RealPathLookup{})
			},
			YarnCommandVersionOutputter: detect.NewRealYarnCommandVersionRunner(),
			NewCommandTextUI:            newCommandTextUI,
			DetectVolta: func() bool {

				return detect.DetectVolta(detect.RealPathLookup{})
			},
			DetectLockfile: func() (lockfile string, error error) {

				return detect.DetectLockfile(detect.RealFileSystem{})
			},
			DetectJSPacakgeManager: func() (string, error) {
				return detect.DetectJSPackageManager(detect.RealPathLookup{})
			},
			NewPackageMultiSelectUI: newPackageMultiSelectUI,
		},
	)

}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.ExecuteContext(context.Background())
	if err != nil {
		os.Exit(1)
	}
}

// Helper functions to retrieve dependencies and other values from the command context.
// These functions are used by subcommands to get their required dependencies.

func getCommandRunnerFromCommandContext(cmd *cobra.Command) CommandRunner {

	return cmd.Context().Value(COMMAND_RUNNER_KEY).(CommandRunner)
}

func getYarnVersionRunnerCommandContext(cmd *cobra.Command) detect.YarnCommandVersionOutputter {

	return cmd.Context().Value(_YARN_VERSION_OUTPUTTER).(detect.YarnCommandVersionOutputter)
}

func getGoEnvFromCommandContext(cmd *cobra.Command) env.GoEnv {
	goEnv := cmd.Context().Value(_GO_ENV).(env.GoEnv)
	return goEnv
}
