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
	"regexp"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
	"github.com/joho/godotenv"
	"github.com/louiss0/javascript-package-delegator/detect"
	"github.com/louiss0/javascript-package-delegator/env"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Constants for context keys and configuration
const _LOCK_FILE_ENV = "JPD_LOCK_FILE"
const _JS_PACKAGE_MANAGER_KEY = "js_pkm"
const _OS_PACKAGE_MANAGER_KEY = "os_pkm"
const _INTERACTIVE_FLAG = "interactive"
const _PACKAGE_NAME = "package-name"                       // Used for storing detected package name in context
const _GO_ENV = "go_env"                                   // Used for storing GoEnv in context
const _SUPPORTED_CONFIG_PATHS_KEY = "supported_paths"      // Used for storing config paths in context
const _VIPER_CONFIG_INSTANCE_KEY = "viper_config_instance" // Used for storing Viper instance in context
const _YARN_VERSION_OUTPUTTER = "yarn_version_outputter"   // Key for YarnCommandVersionOutputter

const _COMMAND_RUNNER_KEY = "command_runner"

const JPD_DEVELOPMENT_CONFIG_NAME = "jpd.config_test"
const JPD_PRODUCTION_CONFIG_NAME = "jpd.config"
const JPD_CONFIG_EXTENSION = "toml"

// CommandRunner Interface and its implementation (executor)
// This interface allows for mocking command execution in tests.
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
	e.cmd.Stdout = os.Stdout // Ensure output goes to stdout
	e.cmd.Stderr = os.Stderr // Ensure errors go to stderr
}

func (e executor) Run() error {
	if e.cmd == nil {
		return fmt.Errorf("no command set to run")
	}
	return e.cmd.Run()
}

// Dependencies holds the external dependencies for testing and real execution
type Dependencies struct {
	CommandRunner               CommandRunner
	JS_PackageManagerDetector   func() (string, error)
	YarnCommandVersionOutputter detect.YarnCommandVersionOutputter
	CommandUITexter
}

type CommandUITexter interface {
	Run() error
	Value() string
}

const VALID_INSTALL_COMMAND_STRING_RE = `^(?:[^\s=]+)\s+(?:[^\s=]+)\s+(?:[^\s=]+)(?:[^\s=]+\s+[^\s]+)*$`

var INVALID_COMMAND_STRUCTURE_ERROR_MESSAGE_STRUCTURE = []string{
	"You wrote this as your string %s",
	"A command for installing a package is at least three words",
	"In the form write the command like you'd normally write a command like this",
	"[command] [subcommand or flag] [package]",
	"Place flags after the command",
}

func newCommandTextUI() *CommandTextUI {

	return &CommandTextUI{
		textUI: huh.NewText().
			Title("Command").
			Description("The command you want to use to install your js package manager").
			Validate(func(s string) error {

				match, error := regexp.MatchString(VALID_INSTALL_COMMAND_STRING_RE, s)

				if error != nil {

					return error
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

	cmd := &cobra.Command{
		Use:     "jpd",
		Version: "0.0.0", // Default version or set via build process
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

		PersistentPreRunE: func(c *cobra.Command, args []string) error {
			// Load .env file
			err := godotenv.Load()
			if err != nil && !os.IsNotExist(err) {
				log.Error(err.Error()) // Log error, but don't stop execution unless critical
			}

			goEnv, err := env.NewGoEnv() // Instantiate GoEnv

			homeDir, err := os.UserHomeDir()
			if err != nil {
				return err
			}

			// Setup Viper for configuration management
			var supportedConfigPaths []string
			vf := viper.New()
			supportedConfigPaths = []string{
				filepath.Join(homeDir, ".config/"),
				filepath.Join(homeDir, ".local/share/"),
			}

			lo.ForEach(supportedConfigPaths, func(path string, index int) {
				vf.AddConfigPath(path)
			})

			configFileName := JPD_PRODUCTION_CONFIG_NAME

			if goEnv.IsDevelopmentMode() {
				configFileName = JPD_DEVELOPMENT_CONFIG_NAME
			}

			vf.SetConfigName(configFileName)
			vf.SetConfigType(JPD_CONFIG_EXTENSION)

			// Read config, but don't fail if file not found
			if err := vf.ReadInConfig(); err != nil {
				if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
					log.Warnf("Error reading config file: %v", err)
				}
			}

			// Store dependencies and other derived values in the command context
			c_ctx := c.Context() // Capture the current context to pass into lo.ForEach

			commandRunner := deps.CommandRunner
			lo.ForEach([][2]any{
				{_GO_ENV, goEnv},
				{_SUPPORTED_CONFIG_PATHS_KEY, supportedConfigPaths},
				{_VIPER_CONFIG_INSTANCE_KEY, vf},
				{_COMMAND_RUNNER_KEY, commandRunner},
				{_YARN_VERSION_OUTPUTTER, deps.YarnCommandVersionOutputter},
			}, func(item [2]any, index int) {
				c_ctx = context.WithValue(
					c_ctx,
					item[0],
					item[1],
				)
			})

			// c.SetContext(c_ctx) // Update the command's context with the new values

			// Package manager detection and potential installation logic
			pm, err := deps.JS_PackageManagerDetector() // Use injected detector

			if err != nil {

				goEnv.ExecuteIfModeIsProduction(func() {
					log.Warn("The package manager wasn't detected:")
					log.Warn("You be asked to fill in which command you'd like to use to install it")

				})

				commandTextUI := deps.CommandUITexter

				if err := commandTextUI.Run(); err != nil {

					return err
				}

				goEnv.ExecuteIfModeIsProduction(func() {

					log.Info("Installing the package manager using ", "command", commandTextUI.Value())

				})

				splitCommandString := strings.Split(commandTextUI.Value(), `\s`)

				commandRunner.Command(splitCommandString[0], splitCommandString[1:]...)

				if err := commandRunner.Run(); err != nil {

					return err
				}

			}

			// If PM detected successfully, set it in context
			c_ctx = context.WithValue(c_ctx, _PACKAGE_NAME, pm)
			c.SetContext(c_ctx)
			return nil
		},
		RunE: func(c *cobra.Command, args []string) error {
			// If no subcommand is provided (i.e., just 'jpd' or 'jpd --interactive'),
			// handle the interactive setup or default 'agent' command.
			if len(args) == 0 {
				interactiveFlag, err := c.Flags().GetBool(_INTERACTIVE_FLAG)
				if err != nil {
					return err
				}

				if interactiveFlag {
					// Retrieve necessary dependencies from context for interactive setup
					goEnv := getGoEnvFromCommandContext(c)
					supportedConfigPaths := getSupportedPathsFromCommandContext(c)
					vf := getViperInstanceFromCommandContext(c)

					var Form struct {
						OS_PackageManager string
						JS_PackageManager string
						ConfigPath        string
					}

					err := huh.NewForm(
						huh.NewGroup(
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

					if err != nil {
						log.Error(err.Error())
						return err
					}

					vf.Set(_OS_PACKAGE_MANAGER_KEY, Form.OS_PackageManager)
					vf.Set(_JS_PACKAGE_MANAGER_KEY, Form.JS_PackageManager)

					configFileName := JPD_PRODUCTION_CONFIG_NAME

					if goEnv.IsDevelopmentMode() {
						configFileName = JPD_DEVELOPMENT_CONFIG_NAME
					}
					configFilePath := filepath.Join(Form.ConfigPath, fmt.Sprintf("%s.%s", configFileName, JPD_CONFIG_EXTENSION))

					if err := vf.WriteConfigAs(configFilePath); err != nil {
						return fmt.Errorf("failed to write config file: %w", err)
					}

					log.Infof("Your config file was created at this path %s", configFilePath)
					log.Info(
						"It has these values",
						_OS_PACKAGE_MANAGER_KEY, Form.OS_PackageManager,
						_JS_PACKAGE_MANAGER_KEY, Form.JS_PackageManager,
					)
					return nil // Interactive setup completed
				}
			}
			return nil // Let subcommands handle their own RunE
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

	// Add persistent flags (like interactive)
	cmd.PersistentFlags().BoolP(
		"interactive",
		"i",
		false,
		"Allows the user to setup the config file for jpd",
	)

	return cmd
}

// Global variable for the root command, initialized in init()
var rootCmd *cobra.Command

func init() {
	// Initialize the global rootCmd with real implementations of its dependencies
	rootCmd = NewRootCmd(
		Dependencies{
			CommandRunner:               newExecutor(exec.Command), // Use the newExecutor constructor
			JS_PackageManagerDetector:   detect.DetectJSPacakgeManager,
			YarnCommandVersionOutputter: detect.NewRealYarnCommandVersionRunner(),
			CommandUITexter:             newCommandTextUI(),
		},
	)
	cobra.OnInitialize(initConfig) // Register initConfig to be run by Cobra
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.ExecuteContext(context.Background())
	if err != nil {
		os.Exit(1)
	}
}

// initConfig reads in config file and ENV variables if set.
// This function is intended to be called by cobra.OnInitialize.
func initConfig() {
	// Load .env file first
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Errorf("Error loading .env file: %s", err)
	}

	// Set environment variables for Viper to read
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) == 2 {
			viper.SetDefault(pair[0], pair[1])
		}
	}
}

// Helper functions to retrieve dependencies and other values from the command context.
// These functions are used by subcommands to get their required dependencies.

func getCommandRunnerFromCommandContext(cmd *cobra.Command) CommandRunner {

	return cmd.Context().Value(_COMMAND_RUNNER_KEY).(CommandRunner)
}

func getJS_PackageManagerNameFromCommandContext(cmd *cobra.Command) string {
	return cmd.Context().Value(_PACKAGE_NAME).(string)
}

func getYarnVersionRunnerCommandContext(cmd *cobra.Command) detect.YarnCommandVersionOutputter {

	return cmd.Context().Value(_YARN_VERSION_OUTPUTTER).(detect.YarnCommandVersionOutputter)
}

func getGoEnvFromCommandContext(cmd *cobra.Command) env.GoEnv {
	goEnv := cmd.Context().Value(_GO_ENV).(env.GoEnv)
	return goEnv
}

func getSupportedPathsFromCommandContext(cmd *cobra.Command) []string {
	paths := cmd.Context().Value(_SUPPORTED_CONFIG_PATHS_KEY).([]string)
	return paths
}

func getViperInstanceFromCommandContext(cmd *cobra.Command) *viper.Viper {
	vf := cmd.Context().Value(_VIPER_CONFIG_INSTANCE_KEY).(*viper.Viper)
	return vf
}
