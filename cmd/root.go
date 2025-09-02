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
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	// external
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
	"github.com/joho/godotenv"
	"github.com/samber/lo"
	"github.com/spf13/cobra"

	// internal
	"github.com/louiss0/javascript-package-delegator/build_info"
	"github.com/louiss0/javascript-package-delegator/custom_flags"
	"github.com/louiss0/javascript-package-delegator/detect"
	"github.com/louiss0/javascript-package-delegator/env"
	"github.com/louiss0/javascript-package-delegator/services"
)

// Constants for context keys and configuration
const (
	PACKAGE_NAME            = "package-name"           // Used for storing detected package name in context
	_GO_ENV                 = "go_env"                 // Used for storing GoEnv in context
	_YARN_VERSION_OUTPUTTER = "yarn_version_outputter" // Key for YarnCommandVersionOutputter
	_DEBUG_EXECUTOR         = "debug_executor"
)

const (
	COMMAND_RUNNER_KEY = "command_runner"
	JPD_AGENT_ENV_VAR  = "JPD_AGENT"
	AGENT_FLAG         = "agent"
	_CWD_FLAG          = "cwd"
	_DEBUG_FLAG        = "debug"
)

// CommandRunner Interface and its implementation
// This interface allows for mocking command execution in tests.
// **Remember:** always use the `Command()` before using the `Run()`
type CommandRunner interface {
	// A method that is used to determine whether run should be in debug mode or not
	// Use a boolean state in struct that implements this inter face and return that
	// This method uses `exec.Cmd` struct to execute commands behind the scenes
	// Make this method useful by creating a field that will hold `exec.Command`.
	// Then make a second field that will hold the
	Command(string, ...string)
	// This method calls the underlying `exec.Run()` to execute the command from `exec.Cmd`!
	Run() error
	SetTargetDir(string) error
}

type _ExecCommandFunc func(string, ...string) *exec.Cmd

type commandRunner struct {
	execCommandFunc _ExecCommandFunc
	cmd             *exec.Cmd
	targetDir       string
}

func newCommandRunner(execCommandFunc _ExecCommandFunc) CommandRunner {
	return &commandRunner{
		execCommandFunc: execCommandFunc,
	}
}

func (e *commandRunner) Command(name string, args ...string) {
	e.cmd = e.execCommandFunc(name, args...)
	e.cmd.Stdin = os.Stdin   // Ensure stdin is connected for interactive commands
	e.cmd.Stdout = os.Stdout // Ensure output goes to stdout
	e.cmd.Stderr = os.Stderr // Ensure errors go to stderr

	// Apply any previously set target directory
	if e.targetDir != "" {
		e.cmd.Dir = e.targetDir
	}
}

func (e *commandRunner) SetTargetDir(dir string) error {
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

func (e *commandRunner) Run() error {
	if e.cmd == nil {
		return fmt.Errorf("no command set to run")
	}
	return e.cmd.Run()
}

// Dependencies holds the external dependencies for testing and real execution

type MultiUISelecter interface {
	Values() []string
	Run() error
}

type TaskUISelector interface {
	Value() string
	Run() error
}

type DependencyUIMultiSelector interface {
	Values() []string
	Run() error
}

// PackageMultiSelectUI is a type alias for backwards compatibility with test code
type PackageMultiSelectUI = MultiUISelecter
type TaskSelectorUI = TaskUISelector
type DependencyMultiSelectUI = DependencyUIMultiSelector

type Dependencies struct {
	CommandRunnerGetter                   func() CommandRunner
	DetectJSPackageManagerBasedOnLockFile func(detectedLockFile string) (packageManager string, err error)
	YarnCommandVersionOutputter           detect.YarnCommandVersionOutputter
	NewCommandTextUI                      func(lockfile string) CommandUITexter
	DetectLockfile                        func(targetDir string) (lockfile string, err error)
	DetectJSPackageManager                func() (string, error)
	DetectVolta                           func() bool
	NewPackageMultiSelectUI               func([]services.PackageInfo) MultiUISelecter
	NewTaskSelectorUI                     func(options []string) TaskUISelector
	NewDependencyMultiSelectUI            func(options []string) DependencyUIMultiSelector
	NewDebugExecutor                      func(bool) DebugExecutor
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
						"We detected a lock file but there is no %s",
						detect.LockFileToPackageManagerMap[lockfile],
					),
					"The command you want to use to install your js package manager",
				),
			).
			Validate(func(s string) error {
				match, err := regexp.MatchString(VALID_INSTALL_COMMAND_STRING_RE, s)

				if err != nil {
					return err
				}

				if lockfile != "" && !strings.Contains(s, detect.LockFileToPackageManagerMap[lockfile]) {
					return fmt.Errorf("the command you entered does not contain the package manager command for %s", detect.LockFileToPackageManagerMap[lockfile])
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

type DebugExecutor interface {
	ExecuteIfDebugIsTrue(cb func())
	LogDebugMessageIfDebugIsTrue(msg string, keyvals ...interface{})
	LogJSCommandIfDebugIsTrue(command string, args ...string)
}

type debugExecutor struct {
	debugFlag bool
}

func newDebugExecutor(debugFlag bool) DebugExecutor {
	return debugExecutor{debugFlag}
}

func (d debugExecutor) ExecuteIfDebugIsTrue(cb func()) {
	if d.debugFlag {
		cb()
	}
}

func (d debugExecutor) LogDebugMessageIfDebugIsTrue(msg string, keyvals ...interface{}) {
	if d.debugFlag {
		log.Debug(msg, keyvals...)
	}
}

func (d debugExecutor) LogJSCommandIfDebugIsTrue(command string, args ...string) {
	if d.debugFlag {
		log.Debug("Executing command:", "command", strings.Join(append([]string{command}, args...), " "))
	}
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

			commandRunner := deps.CommandRunnerGetter()

			if cwd := cwdFlag.String(); cwd != "" {

				err := commandRunner.SetTargetDir(cwd)
				if err != nil {
					return err
				}

			}

			debug, err := c.Flags().GetBool(_DEBUG_FLAG)
			if err != nil {
				return err
			}

			if debug {
				log.SetLevel(log.DebugLevel)
			}

			debugExecutor := deps.NewDebugExecutor(debug)

			lo.ForEach([][2]any{
				{_GO_ENV, goEnv},
				{COMMAND_RUNNER_KEY, commandRunner},
				{_YARN_VERSION_OUTPUTTER, deps.YarnCommandVersionOutputter},
				{_DEBUG_EXECUTOR, debugExecutor},
			}, func(item [2]any, index int) {
				c_ctx = context.WithValue(
					c_ctx,
					item[0],
					item[1],
				)
			})

			persistentFlags := c.Flags()

			// Determine the target directory from cwd flag or use current working directory
			targetDir := cwdFlag.String()
			if targetDir == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				targetDir = cwd
			}

			// Always run detection logic first (for --cwd support)
			var detectedPM string
			lockFile, err := deps.DetectLockfile(targetDir)
			if err != nil {
				debugExecutor.LogDebugMessageIfDebugIsTrue("Lock file is not detected")

				pm, err := deps.DetectJSPackageManager()
				if err != nil {

					debugExecutor.LogDebugMessageIfDebugIsTrue("Package manager is not detected from path")

					// Check if agent flag or env var is set before prompting for install
					agent, err := persistentFlags.GetString(AGENT_FLAG)
					if err != nil {
						return err
					}
					if agent == "" {
						agent, _ = os.LookupEnv(JPD_AGENT_ENV_VAR)
					}

					if agent != "" {
						// Agent specified, skip installation prompt
						detectedPM = agent
					} else {
						// No agent specified and no PM detected, prompt for installation
						commandTextUI := deps.NewCommandTextUI("")

						if err := commandTextUI.Run(); err != nil {
							return err
						}

						goEnv.ExecuteIfModeIsProduction(func() {
							log.Info("Installing the package manager using ", "command", commandTextUI.Value())
						})
						// Split the command string into name and args consistently with how tests parse it
						splitCommandString := strings.Fields(commandTextUI.Value())
						if len(splitCommandString) == 0 {
							return fmt.Errorf(strings.Join(INVALID_COMMAND_STRUCTURE_ERROR_MESSAGE_STRUCTURE, "\n"), commandTextUI.Value())
						}

						debugExecutor.LogJSCommandIfDebugIsTrue(splitCommandString[0], splitCommandString[1:]...)
						commandRunner.Command(splitCommandString[0], splitCommandString[1:]...)

						if err := commandRunner.Run(); err != nil {
							return err
						}
						return nil
					}
				} else {
					debugExecutor.LogDebugMessageIfDebugIsTrue("Package manager detected from path", "pm", pm)
					detectedPM = pm
				}
			} else {
				debugExecutor.LogDebugMessageIfDebugIsTrue("Lock file is detected", "lockfile", lockFile)

				// Package manager detection and potential installation logic
				pm, err := deps.DetectJSPackageManagerBasedOnLockFile(lockFile) // Use injected detector
				if err != nil {

					if errors.Is(err, detect.ErrNoPackageManager) {
						// The package manager indicated by the lock file is not installed
						// Let's check if any other package manager is available in PATH
						goEnv.ExecuteIfModeIsProduction(func() {
							log.Warn("Package manager indicated by lock file is not installed")
							log.Info("Checking for other available package managers...")
						})

						// Try to detect any available package manager from PATH
						pm, err := deps.DetectJSPackageManager()
						if err == nil {
							// Found an alternative package manager!
							goEnv.ExecuteIfModeIsProduction(func() {
								log.Info("Found alternative package manager", "pm", pm)
							})
							detectedPM = pm
						} else {
							// Check if agent flag or env var is set before prompting for install
							agent, err := persistentFlags.GetString(AGENT_FLAG)
							if err != nil {
								return err
							}
							if agent == "" {
								agent, _ = os.LookupEnv(JPD_AGENT_ENV_VAR)
							}

							if agent != "" {
								// Agent specified, use it
								detectedPM = agent
							} else {
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

								// Split the command string into name and args consistently with how tests parse it
								splitCommandString := strings.Fields(commandTextUI.Value())
								if len(splitCommandString) == 0 {
									return fmt.Errorf(strings.Join(INVALID_COMMAND_STRUCTURE_ERROR_MESSAGE_STRUCTURE, "\n"), commandTextUI.Value())
								}

								debugExecutor.LogJSCommandIfDebugIsTrue(splitCommandString[0], splitCommandString[1:]...)
								commandRunner.Command(splitCommandString[0], splitCommandString[1:]...)

								if err := commandRunner.Run(); err != nil {
									return err
								}

								return nil
							}
						}
					} else {
						// Return any other errors as-is
						return err
					}
				} else {
					debugExecutor.LogDebugMessageIfDebugIsTrue("Package manager is detected based on lock file", "pm", pm)
					detectedPM = pm
				}
			}

			// Now check for agent override after detection
			agent, err := persistentFlags.GetString(AGENT_FLAG)
			if err != nil {
				return err
			}

			if agent != "" {
				debugExecutor.LogDebugMessageIfDebugIsTrue(
					"Agent flag is set",
					"agent", agent,
				)
				_ = persistentFlags.Set(AGENT_FLAG, agent)
				c.SetContext(c_ctx)
				return nil
			}

			agent, ok := os.LookupEnv(JPD_AGENT_ENV_VAR)

			if ok {

				if !lo.Contains(detect.SupportedJSPackageManagers[:], agent) {
					return fmt.Errorf(
						"the %s variable is set the wrong way use one of these values instead %v",
						JPD_AGENT_ENV_VAR,
						detect.SupportedJSPackageManagers,
					)
				}

				goEnv.ExecuteIfModeIsProduction(func() {
					log.Info("Using package manager", "agent", agent)
				})

				debugExecutor.LogDebugMessageIfDebugIsTrue(
					"JPD_AGENT environment variable detected setting agent",
					"agent", agent,
				)
				_ = persistentFlags.Set(AGENT_FLAG, agent)
				c.SetContext(c_ctx)
				return nil
			}

			// Use detected package manager if no agent override
			if detectedPM != "" {
				_ = persistentFlags.Set(AGENT_FLAG, detectedPM)
			}
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
	cmd.AddCommand(NewIntegrateCmd())

	cmd.PersistentFlags().BoolP(_DEBUG_FLAG, "d", false, "Make commands run in debug mode")

	cmd.PersistentFlags().StringP(AGENT_FLAG, "a", "", "Select the JS package manager you want to use")

	cmd.PersistentFlags().VarP(cwdFlag, _CWD_FLAG, "C", "Set the working directory for commands (must end with '/' unless it's just '/')")

	_ = cmd.RegisterFlagCompletionFunc(
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
			CommandRunnerGetter: func() CommandRunner {
				return newCommandRunner(exec.Command)
			}, // Use the newExecutor constructor
			DetectJSPackageManagerBasedOnLockFile: func(detectedLockFile string) (packageManager string, err error) {
				return detect.DetectJSPackageManagerBasedOnLockFile(detectedLockFile, detect.RealPathLookup{})
			},
			YarnCommandVersionOutputter: detect.NewRealYarnCommandVersionRunner(),
			NewCommandTextUI:            newCommandTextUI,
			DetectVolta: func() bool {
				return detect.DetectVolta(detect.RealPathLookup{})
			},
			DetectLockfile: func(targetDir string) (lockfile string, err error) {
				return detect.DetectLockfileIn(targetDir, detect.RealFileSystem{})
			},
			DetectJSPackageManager: func() (string, error) {
				return detect.DetectJSPackageManager(detect.RealPathLookup{})
			},
			NewPackageMultiSelectUI:    newPackageMultiSelectUI,
			NewTaskSelectorUI:          newTaskSelectorUI,
			NewDependencyMultiSelectUI: newDependencySelectorUI,
			NewDebugExecutor:           newDebugExecutor,
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

func getDebugExecutorFromCommandContext(cmd *cobra.Command) DebugExecutor {
	return cmd.Context().Value(_DEBUG_EXECUTOR).(DebugExecutor)
}

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
