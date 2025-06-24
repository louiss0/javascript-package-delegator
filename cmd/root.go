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

// Nix UI Interfaces and Implementations
// These interfaces allow for mocking user interaction for Nix installation prompts.
type NixUISelector interface {
	Selection() string
	Choices() [2]string // Changed from huh.OptionValue to string
	Run() error
}

type nixInstallationOption string

// Constants for Nix installation options
const NIX_INSTALLATION_OPTION_PROFILE = "Nix Profile (recommended for most users)" // Changed to string
const NIX_INSTALLATION_OPTION_ENV = "nix-env (legacy)"                             // Changed to string

// RealNixSelectUI implements NixUISelector for real user interaction.
type RealNixSelectUI struct {
	selection string // Changed from huh.OptionValue to string
}

func (s *RealNixSelectUI) Selection() string {
	return s.selection
}

func (s *RealNixSelectUI) Choices() [2]string { // Changed from huh.OptionValue to string
	return [2]string{NIX_INSTALLATION_OPTION_PROFILE, NIX_INSTALLATION_OPTION_ENV}
}

func (s *RealNixSelectUI) Run() error {
	form := huh.NewSelect[string](). // Changed from huh.OptionValue to string
						Title("How would you like to install it?").
						Options(
			huh.NewOption("Nix Profile (recommended for most users)", NIX_INSTALLATION_OPTION_PROFILE),
			huh.NewOption("nix-env (legacy)", NIX_INSTALLATION_OPTION_ENV),
		).
		Value(&s.selection)
	return form.Run()
}

// NixProfileNameInputer defines the interface for prompting a Nix profile name.
type NixProfileNameInputer interface {
	Value() string
	Run() error
}

// RealNixProfileNameInput implements NixProfileNameInputer for real user interaction.
type RealNixProfileNameInput struct {
	value string
}

func (n *RealNixProfileNameInput) Value() string {
	return n.value
}

func (n *RealNixProfileNameInput) Run() error {
	form := huh.NewInput().
		Title("Enter a name for the Nix profile (e.g., nodejs)").
		Value(&n.value)
	return form.Run()
}

// Dependencies holds the external dependencies for testing and real execution
type Dependencies struct {
	CommandRunner               CommandRunner
	JS_PackageManagerDetector   func() (string, error)
	OS_PackageManagerDetector   func() (string, error)
	YarnCommandVersionOutputter detect.YarnCommandVersionOutputter
	NixUISelector               NixUISelector
	NixProfileNameInputer       NixProfileNameInputer
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

			lo.ForEach([][2]any{
				{_GO_ENV, goEnv},
				{_SUPPORTED_CONFIG_PATHS_KEY, supportedConfigPaths},
				{_VIPER_CONFIG_INSTANCE_KEY, vf},
				{_COMMAND_RUNNER_KEY, deps.CommandRunner},
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
				// No JS package manager detected by lockfile, check config or prompt
				jsPackageManagerFromConfig := vf.GetString(_JS_PACKAGE_MANAGER_KEY)
				osPackageManagerFromConfig := vf.GetString(_OS_PACKAGE_MANAGER_KEY)

				installJSPackageManagerFunc := createInstallJSPackageManager(
					deps.CommandRunner,
					deps.NixUISelector,
					deps.NixProfileNameInputer,
				)

				// Scenario 1: JS package manager defined in config, OS package manager needs detection/prompt
				if jsPackageManagerFromConfig != "" && osPackageManagerFromConfig == "" {
					detectedOSManager, err := deps.OS_PackageManagerDetector() // Use injected OS detector
					if err != nil {
						// If OS package manager not detected, prompt user to select one
						selectedOSPackageManager, promptErr := promptUserToSelectOSPackageManager()
						if promptErr != nil {
							return fmt.Errorf("failed to select OS package manager: %w", promptErr)
						}
						detectedOSManager = selectedOSPackageManager
					}

					err = installJSPackageManagerFunc(jsPackageManagerFromConfig, detectedOSManager)
					if err != nil {
						return fmt.Errorf("failed to install %s via %s: %w", jsPackageManagerFromConfig, detectedOSManager, err)
					}
					c_ctx = context.WithValue(c_ctx, _PACKAGE_NAME, jsPackageManagerFromConfig)
					c.SetContext(c_ctx)
					return nil
				}

				// Scenario 2: OS package manager defined in config, JS package manager needs selection
				if osPackageManagerFromConfig != "" && jsPackageManagerFromConfig == "" {
					selectedJSPkgManager, promptErr := promptUserToSelectJSPackageManager()
					if promptErr != nil {
						return fmt.Errorf("no JS package manager chosen for installation: %w", promptErr)
					}
					err = installJSPackageManagerFunc(selectedJSPkgManager, osPackageManagerFromConfig)
					if err != nil {
						return fmt.Errorf("failed to install %s via %s: %w", selectedJSPkgManager, osPackageManagerFromConfig, err)
					}
					c_ctx = context.WithValue(c_ctx, _PACKAGE_NAME, selectedJSPkgManager)
					c.SetContext(c_ctx)
					return nil
				}

				// Scenario 3: Both JS and OS package managers are defined in config
				if osPackageManagerFromConfig != "" && jsPackageManagerFromConfig != "" {
					err = installJSPackageManagerFunc(jsPackageManagerFromConfig, osPackageManagerFromConfig)
					if err != nil {
						return fmt.Errorf("failed to install %s via %s from config: %w", jsPackageManagerFromConfig, osPackageManagerFromConfig, err)
					}
					c_ctx = context.WithValue(c_ctx, _PACKAGE_NAME, jsPackageManagerFromConfig)
					c.SetContext(c_ctx)
					return nil
				}

				// If no config and no detection, and it's the root command without subcommands
				if len(args) == 0 {
					log.Warn("No JavaScript package manager detected. You might need to install one or use the --interactive flag for setup.")
				}
				return fmt.Errorf("no JavaScript package manager detected or configured: %w", err)
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
			OS_PackageManagerDetector:   detect.SupportedOperatingSystemPackageManager,
			YarnCommandVersionOutputter: detect.NewRealYarnCommandVersionRunner(),
			NixUISelector:               &RealNixSelectUI{},
			NixProfileNameInputer:       &RealNixProfileNameInput{},
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

// createInstallJSPackageManager orchestrates the installation of a JS package manager
// using a detected or selected OS package manager.
func createInstallJSPackageManager(commandRunner CommandRunner, selectUI NixUISelector, input NixProfileNameInputer) func(string, string) error {
	// Local constants for Nix installation choices to avoid shadowing global ones.
	localSupportedNixInstallationChoices := [2]string{NIX_INSTALLATION_OPTION_PROFILE, NIX_INSTALLATION_OPTION_ENV} // Changed to string

	// Helper function for Nix installation option selection.
	promptUserToSelectNixEnvORProfile := func() (string, error) { // Changed return type to string
		err := selectUI.Run()
		if err != nil {
			return "", err
		}
		// Directly use selectUI.Selection() as it now returns string
		return selectUI.Selection(), nil
	}

	// Helper function to construct Nix profile installation command.
	constructNixProfileCommand := func(jsPkgMgr string) (string, []string) {
		err := input.Run()
		profileName := input.Value()
		if profileName == "" || err != nil {
			log.Warnf("Failed to get Nix profile name, using default. Error: %v", err)
			return "nix", []string{"profile", "install", "--impure", "--expr", fmt.Sprintf(`with import <nixpkgs> {}; pkgs.%s`, jsPkgMgr)}
		}
		return "nix", []string{
			"profile",
			"install",
			"--impure",
			"--expr",
			fmt.Sprintf(`with import <nixpkgs> {}; pkgs.%s`, jsPkgMgr),
			"--as",
			profileName,
		}
	}

	// The returned function performs the actual installation logic.
	return func(jsPkgMgr, osPkgMgr string) error {
		var cmdName string
		var cmdArgs []string

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
				answer, err := promptUserToSelectNixEnvORProfile()
				if err != nil {
					return err
				}
				if answer == localSupportedNixInstallationChoices[0] {
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
			case "winget": // Added winget support for bun
				cmdName = "winget"
				cmdArgs = []string{"install", "--id", "OwenKelly.Bun", "-e"}
			case "nix": // Added nix support for bun
				answer, err := promptUserToSelectNixEnvORProfile()
				if err != nil {
					return err
				}
				if answer == localSupportedNixInstallationChoices[0] {
					cmdName, cmdArgs = constructNixProfileCommand(jsPkgMgr)
					break
				}
				cmdName = "nix-env"
				cmdArgs = []string{"-iA", fmt.Sprintf("nixpkgs.%s", jsPkgMgr)}
			default:
				return fmt.Errorf("bun not supported on %s", osPkgMgr)
			}
		case "npm":
			switch osPkgMgr {
			case "brew":
				cmdName = "brew"
				cmdArgs = []string{"install", "node"} // npm usually comes with node
			case "winget":
				cmdName = "winget"
				cmdArgs = []string{"install", "Node.js"}
			case "scoop", "choco":
				cmdName = osPkgMgr
				cmdArgs = []string{"install", "nodejs-lts"}
			case "nix":
				answer, err := promptUserToSelectNixEnvORProfile()
				if err != nil {
					return err
				}
				if answer == localSupportedNixInstallationChoices[0] {
					cmdName, cmdArgs = constructNixProfileCommand("nodejs") // npm is part of nodejs on Nix
					break
				}
				cmdName = "nix-env"
				cmdArgs = []string{"-iA", "nixpkgs.nodejs"} // nixpkgs.nodejs usually includes npm
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
				answer, err := promptUserToSelectNixEnvORProfile()
				if err != nil {
					return err
				}
				if answer == localSupportedNixInstallationChoices[0] {
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
				answer, err := promptUserToSelectNixEnvORProfile()
				if err != nil {
					return err
				}
				if answer == localSupportedNixInstallationChoices[0] {
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

		commandRunner.Command(cmdName, cmdArgs...)
		err := commandRunner.Run()
		if err != nil {
			return err
		}
		return nil
	}
}

// Helper to get selected JS package manager from prompt, using injected UI.
func promptUserToSelectJSPackageManager() (string, error) {
	selectedPackageManager := ""
	options := lo.Map(detect.SupportedJSPackageManagers[:], func(item string, index int) huh.Option[string] {
		return huh.NewOption(item, item)
	})

	form := huh.NewSelect[string]().
		Title("No JavaScript package manager found. Please select one to install:").
		Options(options...).
		Value(&selectedPackageManager)

	err := form.Run()
	if err != nil {
		return "", fmt.Errorf("failed to get JS package manager selection: %w", err)
	}
	return selectedPackageManager, nil
}

// Helper to get selected OS package manager from prompt, using injected UI.
func promptUserToSelectOSPackageManager() (string, error) {
	selectedPackageManager := ""
	options := lo.Map(detect.SupportedOperatingSystemPackageManagers[:], func(item string, index int) huh.Option[string] {
		return huh.NewOption(item, item)
	})

	form := huh.NewSelect[string]().
		Title("No supported OS package manager found. Please select one to install the JavaScript package manager:").
		Options(options...).
		Value(&selectedPackageManager)

	err := form.Run()
	if err != nil {
		return "", fmt.Errorf("failed to get OS package manager selection: %w", err)
	}
	return selectedPackageManager, nil
}
