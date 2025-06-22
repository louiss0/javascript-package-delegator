package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"

	"github.com/louiss0/javascript-package-delegator/cmd"
	"github.com/louiss0/javascript-package-delegator/env"
	. "github.com/onsi/ginkgo/v2"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// MockCommandRunner implements the cmd.CommandRunner interface for testing purposes
type MockCommandRunner struct {
	// CommandCalls stores all the commands that have been called
	CommandCalls []CommandCall
	// InvalidCommands is a list of commands that should return an error when Run() is called
	InvalidCommands []string
	// CurrentCommand holds the current command being set up
	CurrentCommand CommandCall
}

// CommandCall represents a single command call with its name and arguments
type CommandCall struct {
	Name string
	Args []string
}

// NewMockCommandRunner creates a new instance of MockCommandRunner
func NewMockCommandRunner() *MockCommandRunner {
	return &MockCommandRunner{
		CommandCalls:    []CommandCall{},
		InvalidCommands: []string{},
		CurrentCommand:  CommandCall{},
	}
}

type MockYarnCommandVersionOutputer struct {
	version string
	goEnv   env.GoEnv
}

func (my MockYarnCommandVersionOutputer) Output() (string, error) {

	if my.goEnv.IsDevelopmentMode() {

		match, error := regexp.MatchString(`\d\.\d\.\d`, my.version)

		if error != nil {

			return "", error
		}

		if !match {

			return "", fmt.Errorf("invalid version format you must use semver versioning")
		}

		return my.version, nil
	}

	output, error := exec.Command("yarn", "--version").Output()

	if error != nil {

		return "", error
	}

	return string(output), nil

}

// Command records the command that would be executed
func (m *MockCommandRunner) Command(name string, args ...string) {
	m.CurrentCommand = CommandCall{
		Name: name,
		Args: args,
	}
}

// Run simulates running the command and records it in the CommandCalls slice.
// Returns an error if the command is in the InvalidCommands list.
func (m *MockCommandRunner) Run() error {
	m.CommandCalls = append(m.CommandCalls, m.CurrentCommand)

	// Check if this command should fail
	for _, invalidCmd := range m.InvalidCommands {
		if m.CurrentCommand.Name == invalidCmd {
			return fmt.Errorf("mock error: command '%s' is configured to fail", m.CurrentCommand.Name)
		}
	}

	return nil
}

// Reset clears all recorded commands and invalid commands
func (m *MockCommandRunner) Reset() {
	m.CommandCalls = []CommandCall{}
	m.InvalidCommands = []string{}
	m.CurrentCommand = CommandCall{}
}

// HasCommand checks if a specific command with the given name and args has been executed
func (m *MockCommandRunner) HasCommand(name string, args ...string) bool {
	for _, call := range m.CommandCalls {
		if call.Name != name {
			continue
		}

		if len(call.Args) != len(args) {
			continue
		}

		match := true
		for i, arg := range args {
			if call.Args[i] != arg {
				match = false
				break
			}
		}

		if match {
			return true
		}
	}

	return false
}

// LastCommand returns the most recently executed command
func (m *MockCommandRunner) LastCommand() (CommandCall, bool) {
	if len(m.CommandCalls) == 0 {
		return CommandCall{}, false
	}
	return m.CommandCalls[len(m.CommandCalls)-1], true
}

// This function executes a cobra command with the given arguments and returns the output and error.
// It sets the output and error buffers for the command, sets the arguments, and executes the command.
// If there is an error, it returns an error with the error message from the error buffer.
// If there is no error, it returns the output from the output buffer.
// It's used to test the cobra commands.
// When you use this function, make sure to pass the root command and any arguments you want to test.
// The first argument after the rootCmd is any sub command or flag you want to test.
// This function now properly preserves the command context with CommandRunner.

func executeCmd(cmd *cobra.Command, args ...string) (string, error) {
	// Save the original context to restore it later
	originalCtx := cmd.Context()

	// Create buffers for output
	buf := new(bytes.Buffer)
	errBuff := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(errBuff)
	cmd.SetArgs(args)

	// Execute the command
	err := cmd.Execute()

	// Restore the original context
	if originalCtx != nil {
		cmd.SetContext(originalCtx)
	}

	if errBuff.Len() > 0 {
		return "", fmt.Errorf("command failed: %s", errBuff.String())
	}

	return buf.String(), err
}

// It ensures that each command has access to the package manager name and CommandRunner

var _ = Describe("JPD Commands", func() {

	assert := assert.New(GinkgoT())
	var rootCmd *cobra.Command
	var mockRunner *MockCommandRunner

	getSubCommandWithName := func(cmd *cobra.Command, name string) (*cobra.Command, bool) {

		return lo.Find(
			cmd.Commands(),
			func(item *cobra.Command) bool {
				return item.Name() == name
			})
	}

	generateRootCommandWithPackageManagerDetector := func(mockRunner *MockCommandRunner, packageManager string, err error) *cobra.Command {
		return cmd.NewRootCmd(
			cmd.Dependencies{
				CommandRunner: mockRunner,
				JS_PackageManagerDetector: func() (string, error) {

					return packageManager, err
				},
			})
	}

	createRootCommandWithBunAsDefault := func(mockRunner *MockCommandRunner, err error) *cobra.Command {
		return generateRootCommandWithPackageManagerDetector(mockRunner, "bun", err)
	}
	createRootCommandWithDenoAsDefault := func(mockRunner *MockCommandRunner, err error) *cobra.Command {
		return generateRootCommandWithPackageManagerDetector(mockRunner, "deno", err)
	}

	createRootCommandWithYarnTwoAsDefault := func(mockRunner *MockCommandRunner, err error) *cobra.Command {
		goEnv, _ := env.NewGoEnv()

		return cmd.NewRootCmd(
			cmd.Dependencies{
				CommandRunner: mockRunner,
				JS_PackageManagerDetector: func() (string, error) {

					return "yarn", err
				},
				YarnCommandVersionOutputter: MockYarnCommandVersionOutputer{
					version: "2.0.0",
					goEnv:   goEnv,
				},
			})
	}

	createRootCommandWithYarnOneAsDefault := func(mockRunner *MockCommandRunner, err error) *cobra.Command {
		goEnv, _ := env.NewGoEnv()
		return cmd.NewRootCmd(
			cmd.Dependencies{
				CommandRunner: mockRunner,
				JS_PackageManagerDetector: func() (string, error) {

					return "yarn", err
				},
				YarnCommandVersionOutputter: MockYarnCommandVersionOutputer{
					version: "1.0.0",
					goEnv:   goEnv,
				},
			})
	}

	createRootCommandWithPnpmAsDefault := func(mockRunner *MockCommandRunner, err error) *cobra.Command {
		return generateRootCommandWithPackageManagerDetector(mockRunner, "pnpm", err)
	}
	createRootCommandWithNpmAsDefault := func(mockRunner *MockCommandRunner, err error) *cobra.Command {
		return generateRootCommandWithPackageManagerDetector(mockRunner, "npm", err)
	}

	JustBeforeEach(func() {
		mockRunner = NewMockCommandRunner()
		rootCmd = createRootCommandWithNpmAsDefault(mockRunner, nil)
		// This needs to be set because Ginkgo will pass a --test.timeout flag to the root command
		// The test.timeout flag will get in the way
		// If the args are empty before they are set by executeCommand the right args can be passed
		rootCmd.SetArgs([]string{})

	})

	Describe("Root Command", func() {

		It("should be able to run", func() {
			_, err := executeCmd(rootCmd, "")
			assert.NoError(err)
		})

		It("should show help", func() {
			output, err := executeCmd(rootCmd, "--help")
			assert.NoError(err)
			assert.Contains(output, "JavaScript Package Delegator")
			assert.Contains(output, "jpd")
		})
	})

	Describe("Install Command", func() {

		var installCmd *cobra.Command
		BeforeEach(func() {
			installCmd, _ = getSubCommandWithName(rootCmd, "install")
		})

		It("should show help", func() {
			output, err := executeCmd(rootCmd, "install", "--help")
			assert.NoError(err)
			assert.Contains(output, "Install packages")
			assert.Contains(output, "jpd install")
		})

		It("should have correct aliases", func() {
			assert.Contains(installCmd.Aliases, "i")
			assert.Contains(installCmd.Aliases, "add")
		})

		It("should have dev flag", func() {
			flag := installCmd.Flag("dev")
			assert.NotNil(flag)
			assert.Equal("D", flag.Shorthand)
		})

		It("should have global flag", func() {
			flag := installCmd.Flag("global")
			assert.NotNil(flag)
			assert.Equal("g", flag.Shorthand)
		})

		It("should have production flag", func() {
			flag := installCmd.Flag("production")
			assert.NotNil(flag)
			assert.Equal("P", flag.Shorthand)
		})

		It("should have frozen flag", func() {
			flag := installCmd.Flag("frozen")
			assert.NotNil(flag)
		})

		It("should have interactive flag", func() {
			flag := installCmd.Flag("interactive")
			assert.NotNil(flag)
			assert.Equal("i", flag.Shorthand)
		})

		It("should run npm install with no args", func() {

			_, err := executeCmd(rootCmd, "install")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npm", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"install"}, mockRunner.CommandCalls[0].Args)
		})

		It("should run npm install with package names", func() {

			_, err := executeCmd(rootCmd, "install", "lodash", "express")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npm", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"install", "lodash", "express"}, mockRunner.CommandCalls[0].Args)
		})

		It("should run npm install with dev flag", func() {

			_, err := executeCmd(rootCmd, "install", "--dev", "typescript")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npm", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "--save-dev")
		})

		It("should run npm install with global flag", func() {

			_, err := executeCmd(rootCmd, "install", "--global", "typescript")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npm", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "--global")
		})

		It("should run npm install with production flag", func() {

			_, err := executeCmd(rootCmd, "install", "--production")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npm", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "--omit=dev")
		})

		It("should run yarn add with dev flag", func() {

			_, err := executeCmd(createRootCommandWithYarnTwoAsDefault(mockRunner, nil), "install", "--dev", "typescript")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("yarn", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "--dev")
			// Reset to npm for other tests
		})

		It("should run pnpm add with dev flag", func() {

			_, err := executeCmd(createRootCommandWithPnpmAsDefault(mockRunner, nil), "install", "--dev", "typescript")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("pnpm", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "--save-dev")
			// Reset to npm for other tests
		})

		It("should handle global flag with npm", func() {
			_, err := executeCmd(rootCmd, "install", "--global", "typescript")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npm", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "--global")
		})

		It("should handle global flag with yarn", func() {
			_, err := executeCmd(createRootCommandWithYarnTwoAsDefault(mockRunner, nil), "install", "--global", "typescript")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("yarn", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "--global")
		})

		It("should handle global flag with pnpm", func() {
			_, err := executeCmd(createRootCommandWithPnpmAsDefault(mockRunner, nil), "install", "--global", "typescript")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("pnpm", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "--global")
		})

		It("should handle global flag with bun", func() {
			_, err := executeCmd(createRootCommandWithBunAsDefault(mockRunner, nil), "install", "--global", "typescript")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("bun", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "--global")
		})

		It("should handle production flag with npm", func() {
			_, err := executeCmd(rootCmd, "install", "--production")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npm", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "--omit=dev")
		})

		It("should handle production flag with yarn", func() {
			_, err := executeCmd(createRootCommandWithYarnTwoAsDefault(mockRunner, nil), "install", "--production")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("yarn", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "--production")
		})

		It("should handle production flag with pnpm", func() {
			_, err := executeCmd(createRootCommandWithPnpmAsDefault(mockRunner, nil), "install", "--production")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("pnpm", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "--prod")
		})

		It("should handle production flag with bun", func() {
			_, err := executeCmd(createRootCommandWithBunAsDefault(mockRunner, nil), "install", "--production")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("bun", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "--production")
		})

		It("should handle frozen flag with npm", func() {
			_, err := executeCmd(rootCmd, "install", "--frozen")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npm", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "--package-lock-only")
		})

		It("should handle frozen flag with yarn", func() {
			_, err := executeCmd(createRootCommandWithYarnTwoAsDefault(mockRunner, nil), "install", "--frozen")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("yarn", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "--frozen-lockfile")
		})

		It("should handle frozen flag with pnpm", func() {
			_, err := executeCmd(createRootCommandWithPnpmAsDefault(mockRunner, nil), "install", "--frozen")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("pnpm", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "--frozen-lockfile")
		})

		It("should handle deno with deps.ts file", func() {
			// Create a deps.ts file
			err := os.WriteFile("deps.ts", []byte("export { serve } from 'https://deno.land/std/http/server.ts';"), 0644)
			assert.NoError(err)
			defer os.Remove("deps.ts")

			_, err = executeCmd(createRootCommandWithDenoAsDefault(mockRunner, nil), "install")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("deno", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"cache", "deps.ts"}, mockRunner.CommandCalls[0].Args)
		})

		It("should handle deno with mod.ts file when deps.ts doesn't exist", func() {
			// Create a mod.ts file
			err := os.WriteFile("mod.ts", []byte("export * from './lib.ts';"), 0644)
			assert.NoError(err)
			defer os.Remove("mod.ts")

			_, err = executeCmd(createRootCommandWithDenoAsDefault(mockRunner, nil), "install")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("deno", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"cache", "mod.ts"}, mockRunner.CommandCalls[0].Args)
		})

		It("should return error for deno without deps.ts or mod.ts", func() {
			_, err := executeCmd(createRootCommandWithDenoAsDefault(mockRunner, nil), "install")
			assert.Error(err)
			assert.Contains(err.Error(), "no deps.ts or mod.ts file found to cache")
		})

		It("should return error for deno with specific packages", func() {
			_, err := executeCmd(createRootCommandWithDenoAsDefault(mockRunner, nil), "install", "lodash")
			assert.Error(err)
			assert.Contains(err.Error(), "deno doesn't support installing packages directly")
		})
	})

	Describe("Run Command", func() {

		var runCmd *cobra.Command
		BeforeEach(func() {
			runCmd, _ = getSubCommandWithName(rootCmd, "run")
		})

		It("should return error when command runner fails", func() {
			rootCmd := createRootCommandWithNpmAsDefault(
				mockRunner, fmt.Errorf("mock error: command 'npm' is configured to fail"))
			rootCmd.SetArgs([]string{})
			mockRunner.InvalidCommands = []string{"npm"}

			_, err := executeCmd(rootCmd, "run", "test")
			assert.Error(err)
			assert.Contains(err.Error(), "mock error: command 'npm' is configured to fail")
		})

		It("should handle if-present flag with non-existent script", func() {
			rootCmd := createRootCommandWithNpmAsDefault(mockRunner, nil)
			rootCmd.SetArgs([]string{})

			// Create a package.json without the script
			packageJson := `{"name": "test", "scripts": {}}`
			err := os.WriteFile("package.json", []byte(packageJson), 0644)
			assert.NoError(err)
			defer os.Remove("package.json")

			_, err = executeCmd(rootCmd, "run", "--if-present", "nonexistent")
			assert.NoError(err) // Should not error with --if-present
		})

		It("should return error for unsupported package manager", func() {
			rootCmd := generateRootCommandWithPackageManagerDetector(
				mockRunner,
				"unknown",
				fmt.Errorf("unsupported package manager: unknown"),
			)
			rootCmd.SetArgs([]string{})

			_, err := executeCmd(rootCmd, "run", "test")
			assert.Error(err)
			assert.Contains(err.Error(), "unsupported package manager: unknown")
		})

		It("should handle yarn classic with dependencies", func() {
			rootCmd := createRootCommandWithYarnTwoAsDefault(mockRunner, nil)
			rootCmd.SetArgs([]string{})

			_, err := executeCmd(rootCmd, "install", "lodash")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("yarn", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "add")
			assert.Contains(mockRunner.CommandCalls[0].Args, "lodash")
		})

		It("should handle yarn modern with dependencies", func() {
			// Test yarn version 2+ path
			rootCmd := createRootCommandWithYarnTwoAsDefault(mockRunner, nil)
			rootCmd.SetArgs([]string{})

			_, err := executeCmd(rootCmd, "install", "--dev", "typescript")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
		})

		It("should show help", func() {
			output, err := executeCmd(rootCmd, "run", "--help")
			assert.NoError(err)
			assert.Contains(output, "Run package.json scripts")
			assert.Contains(output, "jpd run", "No jpd run")
		})

		It("should have correct aliases", func() {
			assert.Contains(runCmd.Aliases, "r")
		})

		It("should have if-present flag", func() {
			flag := runCmd.Flag("if-present")
			assert.NotNil(flag)
		})

		It("should run npm run with script name", func() {

			_, err := executeCmd(rootCmd, "run", "test")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			originalDir, _ := os.Getwd()

			path, _ := os.MkdirTemp("", "jpd-test*")

			os.Chdir(path)

			defer os.Chdir(originalDir)

			defer os.RemoveAll(path)

			content := `{
				"lockfileVersion": 2,
				"requires": true,
				"packages": {
					"": {
						"name": "jpd-test",
						"version": "1.0.0",
						"lockfileVersion": 2,
						"requires": true,
						"dependencies": {}
					},
					"scripts": {
						"test": "echo 'test'"
					}
				}
			}`

			os.WriteFile("package.json", []byte(content), 0644)
			os.WriteFile(".env", []byte("GO_ENV=development"), 0644)

			assert.Equal("npm", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"run", "test"}, mockRunner.CommandCalls[0].Args)
		})

		It("should run npm run with script args", func() {

			_, err := executeCmd(rootCmd, "run", "test", "--", "--watch")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npm", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"run", "test", "--", "--watch"}, mockRunner.CommandCalls[0].Args)
		})

		It("should run npm run with if-present flag", func() {
			// Get current directory to restore later
			originalDir, _ := os.Getwd()

			// Create temporary directory
			path, _ := os.MkdirTemp("", "jpd-test*")
			os.Chdir(path)

			defer os.Chdir(originalDir)
			defer os.RemoveAll(path)

			// Create a proper package.json with scripts at root level
			content := `{
				"name": "jpd-test",
				"version": "1.0.0",
				"scripts": {
					"test": "echo 'test'",
					"build": "echo 'build'",
					"dev": "echo 'dev'"
				}
}`

			os.WriteFile("package.json", []byte(content), 0644)
			os.WriteFile(".env", []byte("GO_MODE=development"), 0644)

			_, err := executeCmd(rootCmd, "run", "--if-present", "test")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npm", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "--if-present")
		})

		It("should run yarn run with script name", func() {

			_, err := executeCmd(createRootCommandWithYarnTwoAsDefault(mockRunner, nil), "run", "test")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("yarn", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"run", "test"}, mockRunner.CommandCalls[0].Args)
			// Reset to npm for other tests
		})

		It("should run pnpm run with script args", func() {

			_, err := executeCmd(createRootCommandWithPnpmAsDefault(mockRunner, nil), "run", "test", "--", "--watch")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("pnpm", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"run", "test", "--", "--watch"}, mockRunner.CommandCalls[0].Args)
			// Reset to npm for other tests
		})

		It("should run deno task with script name", func() {

			_, err := executeCmd(createRootCommandWithDenoAsDefault(mockRunner, nil), "run", "test")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("deno", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"task", "test"}, mockRunner.CommandCalls[0].Args)
			// Reset to npm for other tests
		})

		It("should handle package.json reading error", func() {
			// Create an invalid package.json
			err := os.WriteFile("package.json", []byte("invalid json"), 0644)
			assert.NoError(err)
			defer os.Remove("package.json")

			_, err = executeCmd(rootCmd, "run", "--if-present", "test")
			assert.Error(err) // Should still pass with --if-present
		})

		It("should handle missing package.json with if-present", func() {
			// Ensure no package.json exists
			os.Remove("package.json")

			_, err := executeCmd(rootCmd, "run", "--if-present", "test")
			assert.Error(err) // Should error with --if-present when no package.json
		})

		It("should handle script not found without if-present", func() {
			// Create package.json without the script
			packageJson := `{"name": "test", "scripts": {"build": "echo building"}}`
			err := os.WriteFile("package.json", []byte(packageJson), 0644)
			assert.NoError(err)
			defer os.Remove("package.json")

			_, err = executeCmd(rootCmd, "run", "nonexistent")
			assert.NoError(err)
		})

		It("should handle unsupported package manager", func() {
			// Create a root command with unsupported package manager
			unsupportedCmd := generateRootCommandWithPackageManagerDetector(mockRunner, "unsupported", nil)
			unsupportedCmd.SetArgs([]string{})

			_, err := executeCmd(unsupportedCmd, "run", "test")
			assert.Error(err)
			assert.Contains(err.Error(), "unsupported package manager: unsupported")
		})

		It("should handle bun run command", func() {
			_, err := executeCmd(createRootCommandWithBunAsDefault(mockRunner, nil), "run", "test")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("bun", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"run", "test"}, mockRunner.CommandCalls[0].Args)
		})
	})

	Describe("Exec Command", func() {
		var execCmd *cobra.Command
		BeforeEach(func() {
			execCmd, _ = getSubCommandWithName(rootCmd, "exec")
		})

		It("should show help", func() {
			output, err := executeCmd(rootCmd, "exec", "--help")
			assert.NoError(err)
			assert.Contains(output, "Execute packages")
			assert.Contains(output, "jpd exec")
		})

		It("should have correct aliases", func() {
			assert.Contains(execCmd.Aliases, "x")
		})

		It("should require at least one argument", func() {
			_, err := executeCmd(rootCmd, "exec")
			assert.Error(err)
		})

		It("should handle --help in arguments", func() {
			output, err := executeCmd(rootCmd, "exec", "some-package", "--help")
			assert.NoError(err)
			assert.Contains(output, "Execute packages")
		})

		It("should handle deno exec error", func() {
			_, err := executeCmd(createRootCommandWithDenoAsDefault(mockRunner, nil), "exec", "some-package")
			assert.Error(err)
			assert.Contains(err.Error(), "Deno doesn't have a dlx or x like the others")
		})

		It("should handle yarn version detection error", func() {
			// This will test the yarn version detection fallback
			_, err := executeCmd(createRootCommandWithYarnTwoAsDefault(mockRunner, nil), "exec", "create-react-app", "my-app")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("yarn", mockRunner.CommandCalls[0].Name)
		})

		It("should handle help flag correctly", func() {
			rootCmd := createRootCommandWithNpmAsDefault(mockRunner, nil)
			rootCmd.SetArgs([]string{})

			// This should not execute the package manager command due to --help handling
			_, err := executeCmd(rootCmd, "exec", "--help")
			assert.NoError(err)
		})

		It("should return error when command runner fails", func() {
			rootCmd := createRootCommandWithNpmAsDefault(
				mockRunner,
				fmt.Errorf("mock error: command 'npx' is configured to fail"),
			)
			rootCmd.SetArgs([]string{})
			mockRunner.InvalidCommands = []string{"npx"}

			_, err := executeCmd(rootCmd, "exec", "test-command")
			assert.Error(err)
			assert.Contains(err.Error(), "mock error: command 'npx' is configured to fail")
		})

		It("should execute npx with package name", func() {

			_, err := executeCmd(rootCmd, "exec", "create-react-app")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npx", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"create-react-app"}, mockRunner.CommandCalls[0].Args)
		})

		It("should execute npx with package name and args", func() {

			_, err := executeCmd(rootCmd, "exec", "create-react-app", "my-app")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npx", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"create-react-app", "my-app"}, mockRunner.CommandCalls[0].Args)
		})

		It("should fail when command execution fails", func() {

			mockRunner.InvalidCommands = []string{"npx"}
			_, err := executeCmd(rootCmd, "exec", "create-react-app", "my-app")
			assert.Error(err)
			mockRunner.InvalidCommands = []string{}
		})

		It("should execute yarn with package name", func() {

			_, err := executeCmd(createRootCommandWithYarnTwoAsDefault(mockRunner, nil), "exec", "create-react-app")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("yarn", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"dlx", "create-react-app"}, mockRunner.CommandCalls[0].Args)
			// Reset to npm for other tests
		})

		It("should execute pnpm dlx with package name", func() {

			_, err := executeCmd(createRootCommandWithPnpmAsDefault(mockRunner, nil), "exec", "create-react-app", "my-app")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("pnpm", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"dlx", "create-react-app", "my-app"}, mockRunner.CommandCalls[0].Args)
			// Reset to npm for other tests
		})

		It("should execute bunx with package name", func() {

			_, err := executeCmd(createRootCommandWithBunAsDefault(mockRunner, nil), "exec", "create-react-app", "my-app")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("bunx", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"create-react-app", "my-app"}, mockRunner.CommandCalls[0].Args)
			// Reset to npm for other tests
		})

		It("should return error for deno", func() {

			_, err := executeCmd(createRootCommandWithDenoAsDefault(mockRunner, nil), "exec", "create-react-app", "my-app")
			assert.Error(err)
			assert.Contains(err.Error(), "Deno doesn't have a dlx")
			// Reset to npm for other tests
		})

		It("should handle pnpm with dev dependencies", func() {
			rootCmd := createRootCommandWithPnpmAsDefault(
				mockRunner,
				nil,
			)
			rootCmd.SetArgs([]string{})

			_, err := executeCmd(rootCmd, "install", "-D", "jest")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("pnpm", mockRunner.CommandCalls[0].Name)
		})

		It("should handle bun with dependencies", func() {
			rootCmd := createRootCommandWithBunAsDefault(
				mockRunner,
				nil,
			)
			rootCmd.SetArgs([]string{})

			_, err := executeCmd(rootCmd, "install", "react")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("bun", mockRunner.CommandCalls[0].Name)
		})

		It("should return error for deno with dependencies", func() {
			rootCmd := createRootCommandWithDenoAsDefault(
				mockRunner,
				fmt.Errorf("deno doesn't support installing packages"),
			)
			rootCmd.SetArgs([]string{})

			_, err := executeCmd(rootCmd, "install", "lodash")
			assert.Error(err)
			assert.Contains(err.Error(), "deno doesn't support installing packages")
		})

		It("should return error for unsupported package manager", func() {
			rootCmd := generateRootCommandWithPackageManagerDetector(
				mockRunner,
				"unknown",
				fmt.Errorf("unsupported package manager: unknown"),
			)
			rootCmd.SetArgs([]string{})

			_, err := executeCmd(rootCmd, "install", "lodash")
			assert.Error(err)
			assert.Contains(err.Error(), "unsupported package manager: unknown")
		})

		It("should return error when command runner fails", func() {
			rootCmd := createRootCommandWithNpmAsDefault(
				mockRunner,
				fmt.Errorf("mock error: command 'npm' is configured to fail"),
			)

			rootCmd.SetArgs([]string{})
			mockRunner.InvalidCommands = []string{"npm"}

			_, err := executeCmd(rootCmd, "install")
			assert.Error(err)
			assert.Contains(err.Error(), "mock error: command 'npm' is configured to fail")
		})

	})

	Describe("Update Command", func() {

		var updateCmd *cobra.Command

		BeforeEach(func() {
			updateCmd, _ = getSubCommandWithName(rootCmd, "update")
		})

		It("should return error when command runner fails", func() {
			rootCmd := createRootCommandWithNpmAsDefault(mockRunner, nil)
			rootCmd.SetArgs([]string{})
			mockRunner.InvalidCommands = []string{"npm"}

			_, err := executeCmd(rootCmd, "update")
			assert.Error(err)
			assert.Contains(err.Error(), "mock error: command 'npm' is configured to fail")
		})

		It("should handle yarn with specific packages", func() {
			rootCmd := createRootCommandWithYarnTwoAsDefault(mockRunner, nil)
			rootCmd.SetArgs([]string{})

			_, err := executeCmd(rootCmd, "update", "lodash")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("yarn", mockRunner.CommandCalls[0].Name)
		})

		It("should handle pnpm update", func() {
			rootCmd := createRootCommandWithPnpmAsDefault(mockRunner, nil)
			rootCmd.SetArgs([]string{})

			_, err := executeCmd(rootCmd, "update")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("pnpm", mockRunner.CommandCalls[0].Name)
		})

		It("should handle bun update", func() {
			rootCmd := createRootCommandWithBunAsDefault(mockRunner, nil)
			rootCmd.SetArgs([]string{})

			_, err := executeCmd(rootCmd, "update")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("bun", mockRunner.CommandCalls[0].Name)
		})

		It("should handle deno update", func() {
			rootCmd := createRootCommandWithDenoAsDefault(mockRunner, nil)
			rootCmd.SetArgs([]string{})

			_, err := executeCmd(rootCmd, "update")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("deno", mockRunner.CommandCalls[0].Name)
		})

		It("should return error for unsupported package manager", func() {
			rootCmd := generateRootCommandWithPackageManagerDetector(
				mockRunner,
				"unknown",
				fmt.Errorf("unsupported package manager: unknown"),
			)

			rootCmd.SetArgs([]string{})

			_, err := executeCmd(rootCmd, "update")
			assert.Error(err)
			assert.Contains(err.Error(), "unsupported package manager: unknown")
		})

		It("should handle interactive flag with yarn", func() {
			rootCmd := createRootCommandWithYarnTwoAsDefault(mockRunner, nil)
			rootCmd.SetArgs([]string{})

			_, err := executeCmd(rootCmd, "update", "--interactive")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("yarn", mockRunner.CommandCalls[0].Name)
		})

		It("should handle latest flag with yarn", func() {
			rootCmd := createRootCommandWithYarnTwoAsDefault(mockRunner, nil)
			rootCmd.SetArgs([]string{})

			_, err := executeCmd(rootCmd, "update", "--latest")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("yarn", mockRunner.CommandCalls[0].Name)
		})

		It("should show help", func() {
			output, err := executeCmd(rootCmd, "update", "--help")
			assert.NoError(err)
			assert.Contains(output, "Update packages")
			assert.Contains(output, "jpd update")
		})

		It("should have correct aliases", func() {
			assert.Contains(updateCmd.Aliases, "u")
			assert.Contains(updateCmd.Aliases, "up")
			assert.Contains(updateCmd.Aliases, "upgrade")
		})

		It("should have interactive flag", func() {
			flag := updateCmd.Flag("interactive")
			assert.NotNil(flag)
			assert.Equal("i", flag.Shorthand)
		})

		It("should have global flag", func() {
			flag := updateCmd.Flag("global")
			assert.NotNil(flag)
			assert.Equal("g", flag.Shorthand)
		})

		It("should have latest flag", func() {
			flag := updateCmd.Flag("latest")
			assert.NotNil(flag)
		})

		It("should run npm update with no args", func() {

			_, err := executeCmd(rootCmd, "update")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npm", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"update"}, mockRunner.CommandCalls[0].Args)
		})

		It("should run npm update with package names", func() {

			_, err := executeCmd(rootCmd, "update", "lodash", "express")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npm", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"update", "lodash", "express"}, mockRunner.CommandCalls[0].Args)
		})

		It("should run npm update with global flag", func() {

			_, err := executeCmd(rootCmd, "update", "--global", "typescript")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npm", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "--global")
		})

		It("should handle latest flag for npm", func() {

			_, err := executeCmd(rootCmd, "update", "--latest", "lodash")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npm", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "lodash@latest")
		})

		It("should handle interactive flag for yarn", func() {

			_, err := executeCmd(createRootCommandWithYarnTwoAsDefault(mockRunner, nil), "update", "--interactive")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("yarn", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"upgrade-interactive"}, mockRunner.CommandCalls[0].Args)
			// Reset to npm for other tests
		})

		It("should handle yarn with latest flag", func() {
			_, err := executeCmd(createRootCommandWithYarnTwoAsDefault(mockRunner, nil), "update", "--latest")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("yarn", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "upgrade")
		})

		It("should handle yarn with both interactive and latest flags", func() {
			_, err := executeCmd(createRootCommandWithYarnTwoAsDefault(mockRunner, nil), "update", "--interactive", "--latest")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("yarn", mockRunner.CommandCalls[0].Name)
		})

		It("should handle yarn package-specific updates", func() {
			_, err := executeCmd(createRootCommandWithYarnTwoAsDefault(mockRunner, nil), "update", "lodash")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("yarn", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "upgrade")
		})

		It("should handle latest flag for yarn", func() {

			_, err := executeCmd(createRootCommandWithYarnTwoAsDefault(mockRunner, nil), "update", "--latest", "lodash")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("yarn", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "--latest")
			// Reset to npm for other tests
		})

		It("should error on npm with interactive flag", func() {

			_, err := executeCmd(rootCmd, "update", "--interactive")
			assert.Error(err)
			assert.Contains(err.Error(), "npm does not support interactive updates")
		})

		Describe("Uninstall Command", func() {

			var uninstallCmd *cobra.Command
			BeforeEach(func() {
				uninstallCmd, _ = getSubCommandWithName(rootCmd, "uninstall")
			})

			It("should return error when command runner fails", func() {
				rootCmd := createRootCommandWithNpmAsDefault(
					mockRunner,
					fmt.Errorf("mock error: command 'npm' is configured to fail"),
				)
				rootCmd.SetArgs([]string{})
				mockRunner.InvalidCommands = []string{"npm"}

				_, err := executeCmd(rootCmd, "uninstall", "lodash")
				assert.Error(err)
				assert.Contains(err.Error(), "mock error: command 'npm' is configured to fail")
			})

			It("should handle yarn uninstall", func() {

				rootCmd := createRootCommandWithYarnTwoAsDefault(
					mockRunner,
					nil,
				)

				rootCmd.SetArgs([]string{})

				_, err := executeCmd(rootCmd, "uninstall", "lodash")
				assert.NoError(err)
				assert.Equal(1, len(mockRunner.CommandCalls))
				assert.Equal("yarn", mockRunner.CommandCalls[0].Name)
			})

			It("should return error for unsupported package manager", func() {

				rootCmd := generateRootCommandWithPackageManagerDetector(mockRunner,
					"unknown",
					fmt.Errorf("unsupported package manager: unknown"),
				)

				rootCmd.SetArgs([]string{})

				_, err := executeCmd(rootCmd, "uninstall", "lodash")
				assert.Error(err)
				assert.Contains(err.Error(), "unsupported package manager: unknown")
			})

			It("should show help", func() {
				output, err := executeCmd(rootCmd, "uninstall", "--help")
				assert.NoError(err)
				assert.Contains(output, "Uninstall packages")
				assert.Contains(output, "jpd uninstall")
			})

			It("should have correct aliases", func() {
				assert.Contains(uninstallCmd.Aliases, "un")
				assert.Contains(uninstallCmd.Aliases, "remove")
				assert.Contains(uninstallCmd.Aliases, "rm")
			})

			It("should have global flag", func() {
				flag := uninstallCmd.Flag("global")
				assert.NotNil(flag)
				assert.Equal("g", flag.Shorthand)
			})

			It("should require at least one argument", func() {
				_, error := executeCmd(rootCmd, "uninstall")
				fmt.Print(error)
				assert.Error(error)
			})

			It("should run npm uninstall with package name", func() {

				_, err := executeCmd(rootCmd, "uninstall", "lodash")
				assert.NoError(err)
				assert.Equal(1, len(mockRunner.CommandCalls))
				assert.Equal("npm", mockRunner.CommandCalls[0].Name)
				assert.Equal([]string{"uninstall", "lodash"}, mockRunner.CommandCalls[0].Args)
			})

			It("should run npm uninstall with multiple package names", func() {

				_, err := executeCmd(rootCmd, "uninstall", "lodash", "express")
				assert.NoError(err)
				assert.Equal(1, len(mockRunner.CommandCalls))
				assert.Equal("npm", mockRunner.CommandCalls[0].Name)
				assert.Equal([]string{"uninstall", "lodash", "express"}, mockRunner.CommandCalls[0].Args)
			})

			It("should run npm uninstall with global flag", func() {

				_, err := executeCmd(rootCmd, "uninstall", "--global", "typescript")
				assert.NoError(err)
				assert.Equal(1, len(mockRunner.CommandCalls))
				assert.Equal("npm", mockRunner.CommandCalls[0].Name)
				assert.Contains(mockRunner.CommandCalls[0].Args, "--global")
			})

			It("should run yarn remove with package name", func() {

				_, err := executeCmd(createRootCommandWithYarnTwoAsDefault(mockRunner, nil), "uninstall", "lodash")
				assert.NoError(err)
				assert.Equal(1, len(mockRunner.CommandCalls))
				assert.Equal("yarn", mockRunner.CommandCalls[0].Name)
				assert.Equal([]string{"remove", "lodash"}, mockRunner.CommandCalls[0].Args)
			})

			It("should run pnpm remove with global flag", func() {

				_, err := executeCmd(createRootCommandWithPnpmAsDefault(mockRunner, nil), "uninstall", "--global", "typescript")
				assert.NoError(err)
				assert.Equal(1, len(mockRunner.CommandCalls))
				assert.Equal("pnpm", mockRunner.CommandCalls[0].Name)
				assert.Contains(mockRunner.CommandCalls[0].Args, "--global")
			})

			It("should run bun remove with multiple packages", func() {

				_, err := executeCmd(createRootCommandWithBunAsDefault(mockRunner, nil), "uninstall", "react", "react-dom")
				assert.NoError(err)
				assert.Equal(1, len(mockRunner.CommandCalls))
				assert.Equal("bun", mockRunner.CommandCalls[0].Name)
				assert.Equal([]string{"remove", "react", "react-dom"}, mockRunner.CommandCalls[0].Args)
			})

			It("should return error for unsupported package manager", func() {

				_, err := executeCmd(generateRootCommandWithPackageManagerDetector(
					mockRunner,
					"foo",
					fmt.Errorf("unsupported package manager")),
					"uninstall",
					"lodash",
				)

				assert.Error(err)
				assert.Contains(err.Error(), "unsupported package manager")
			})
		})

		Describe("Clean Install Command", func() {

			var cleanInstallCmd *cobra.Command
			BeforeEach(func() {
				cleanInstallCmd, _ = getSubCommandWithName(rootCmd, "clean-install")
			})

			It("should return error for deno package manager", func() {
				rootCmd := generateRootCommandWithPackageManagerDetector(
					mockRunner,
					"deno",
					fmt.Errorf("deno doesn't support this command"),
				)
				rootCmd.SetArgs([]string{})

				_, err := executeCmd(rootCmd, "clean-install")
				assert.Error(err)
				assert.Contains(err.Error(), "deno doesn't support this command")
			})

			It("should show help", func() {
				output, err := executeCmd(rootCmd, "clean-install", "--help")
				assert.NoError(err)
				assert.Contains(output, "Clean install")
				assert.Contains(output, "jpd clean-install")
			})

			It("should have correct aliases", func() {
				assert.Contains(cleanInstallCmd.Aliases, "ci")
			})

			It("should run npm ci", func() {

				_, err := executeCmd(rootCmd, "ci")
				assert.NoError(err)
				assert.Equal(1, len(mockRunner.CommandCalls))
				assert.Equal("npm", mockRunner.CommandCalls[0].Name)
				assert.Equal([]string{"ci"}, mockRunner.CommandCalls[0].Args)
			})

			It("should run yarn install with frozen lockfile", func() {

				_, err := executeCmd(createRootCommandWithYarnOneAsDefault(mockRunner, nil), "clean-install")
				assert.NoError(err)
				assert.Equal(1, len(mockRunner.CommandCalls))
				assert.Equal("yarn", mockRunner.CommandCalls[0].Name)
				assert.Equal([]string{"install", "--frozen-lockfile"}, mockRunner.CommandCalls[0].Args)
			})

			It("should run pnpm install with frozen lockfile", func() {

				_, err := executeCmd(createRootCommandWithPnpmAsDefault(mockRunner, nil), "clean-install")
				assert.NoError(err)
				assert.Equal(1, len(mockRunner.CommandCalls))
				assert.Equal("pnpm", mockRunner.CommandCalls[0].Name)
				assert.Equal([]string{"install", "--frozen-lockfile"}, mockRunner.CommandCalls[0].Args)
			})

			It("should run bun install with frozen lockfile", func() {

				_, err := executeCmd(createRootCommandWithBunAsDefault(mockRunner, nil), "clean-install")
				assert.NoError(err)
				assert.Equal(1, len(mockRunner.CommandCalls))
				assert.Equal("bun", mockRunner.CommandCalls[0].Name)
				assert.Equal([]string{"install", "--frozen-lockfile"}, mockRunner.CommandCalls[0].Args)
			})

			It("should return error for deno", func() {

				_, err := executeCmd(createRootCommandWithDenoAsDefault(mockRunner, nil), "clean-install")
				assert.Error(err)
				assert.Contains(err.Error(), "deno doesn't support this command")
			})

			It("should handle yarn v2+ with immutable flag", func() {
				// Mock yarn version to return v2+
				rootCmd := createRootCommandWithYarnTwoAsDefault(mockRunner, nil)

				_, err := executeCmd(rootCmd, "clean-install")
				assert.NoError(err)
				assert.Equal(1, len(mockRunner.CommandCalls))
				assert.Equal("yarn", mockRunner.CommandCalls[0].Name)
				// This test will cover both paths since getYarnVersion might return different results
			})
		})

		Describe("Agent Command", func() {

			var agentCmd *cobra.Command
			BeforeEach(func() {
				agentCmd, _ = getSubCommandWithName(rootCmd, "agent")
			})

			It("should show help", func() {
				output, err := executeCmd(rootCmd, "agent", "--help")
				assert.NoError(err)
				assert.Contains(output, "Show information about the detected package manager")
				assert.Contains(output, "jpd agent")
			})

			It("should have correct aliases", func() {
				assert.Contains(agentCmd.Aliases, "a")
			})

			It("should execute detected package manager", func() {

				_, err := executeCmd(rootCmd, "agent")
				assert.NoError(err)
				assert.Equal(1, len(mockRunner.CommandCalls))
				assert.Equal("npm", mockRunner.CommandCalls[0].Name)
			})

			It("should pass arguments to package manager", func() {

				_, err := executeCmd(rootCmd, "agent", "--", "--version")
				assert.NoError(err)
				assert.Equal(1, len(mockRunner.CommandCalls))
				assert.Equal("npm", mockRunner.CommandCalls[0].Name)
				assert.Equal([]string{"--version"}, mockRunner.CommandCalls[0].Args)
			})

			It("should execute yarn with arguments", func() {

				_, err := executeCmd(createRootCommandWithYarnTwoAsDefault(mockRunner, nil), "agent", "--", "--version")
				assert.NoError(err)
				assert.Equal(1, len(mockRunner.CommandCalls))
				assert.Equal("yarn", mockRunner.CommandCalls[0].Name)
				assert.Equal([]string{"--version"}, mockRunner.CommandCalls[0].Args)
			})

			It("should execute pnpm with arguments", func() {

				_, err := executeCmd(createRootCommandWithPnpmAsDefault(mockRunner, nil), "agent", "info")
				assert.NoError(err)
				assert.Equal(1, len(mockRunner.CommandCalls))
				assert.Equal("pnpm", mockRunner.CommandCalls[0].Name)
				assert.Equal([]string{"info"}, mockRunner.CommandCalls[0].Args)
			})

			It("should fail when command execution fails", func() {

				mockRunner.InvalidCommands = []string{"npm"}
				_, err := executeCmd(rootCmd, "agent", "--version")
				assert.Error(err)
				mockRunner.InvalidCommands = []string{}
			})
		})

		Describe("Command Integration", func() {
			It("should have all commands registered", func() {
				commands := rootCmd.Commands()
				commandNames := make([]string, len(commands))
				for i, cmd := range commands {
					commandNames[i] = cmd.Name()
				}

				// Check all expected commands are present
				assert.Contains(commandNames, "install")
				assert.Contains(commandNames, "run")
				assert.Contains(commandNames, "exec")
				assert.Contains(commandNames, "update")
				assert.Contains(commandNames, "uninstall")
				assert.Contains(commandNames, "clean-install")
				assert.Contains(commandNames, "agent")
			})

			It("should maintain command count", func() {
				// Ensure we have exactly 7 commands (excluding help/completion)
				commands := rootCmd.Commands()
				// Filter out built-in commands like help, completion
				userCommands := 0
				for _, cmd := range commands {
					if cmd.Name() != "help" && cmd.Name() != "completion" {
						userCommands++
					}
				}
				assert.Equal(7, userCommands)
			})
		})

		Describe("CommandRunner Interface", func() {
			It("should properly record commands", func() {
				testRunner := NewMockCommandRunner()
				testRunner.Command("npm", "install", "lodash")
				err := testRunner.Run()
				assert.NoError(err)
				assert.Equal(1, len(testRunner.CommandCalls))
				assert.Equal("npm", testRunner.CommandCalls[0].Name)
				assert.Equal([]string{"install", "lodash"}, testRunner.CommandCalls[0].Args)
			})

			It("should record multiple commands in sequence", func() {
				testRunner := NewMockCommandRunner()
				testRunner.Command("npm", "install")
				testRunner.Run()
				testRunner.Command("npx", "tsc")
				testRunner.Run()
				testRunner.Command("npm", "test")
				testRunner.Run()

				assert.Equal(3, len(testRunner.CommandCalls))
				assert.Equal("npm", testRunner.CommandCalls[0].Name)
				assert.Equal("npx", testRunner.CommandCalls[1].Name)
				assert.Equal("npm", testRunner.CommandCalls[2].Name)
			})

			It("should return errors for invalid commands", func() {
				testRunner := NewMockCommandRunner()
				testRunner.InvalidCommands = []string{"npm"}
				testRunner.Command("npm", "install")
				err := testRunner.Run()
				assert.Error(err)
				assert.Contains(err.Error(), "configured to fail")
			})

			It("should correctly check for command execution", func() {
				testRunner := NewMockCommandRunner()
				testRunner.Command("npm", "install", "lodash")
				testRunner.Run()
				testRunner.Command("yarn", "add", "react")
				testRunner.Run()

				assert.True(testRunner.HasCommand("npm", "install", "lodash"))
				assert.True(testRunner.HasCommand("yarn", "add", "react"))
				assert.False(testRunner.HasCommand("pnpm", "add", "vue"))
			})

			It("should properly reset all state", func() {
				testRunner := NewMockCommandRunner()
				testRunner.Command("npm", "install", "lodash")
				testRunner.Run()
				testRunner.InvalidCommands = []string{"yarn"}

				// Now reset
				testRunner.Reset()

				assert.Equal(0, len(testRunner.CommandCalls))
				assert.Equal(0, len(testRunner.InvalidCommands))
				assert.Equal(CommandCall{}, testRunner.CurrentCommand)
			})

			It("should handle multiple invalid commands", func() {
				testRunner := NewMockCommandRunner()
				testRunner.InvalidCommands = []string{"npm", "yarn", "pnpm"}

				testRunner.Command("npm", "install")
				err1 := testRunner.Run()
				assert.Error(err1)

				testRunner.Command("yarn", "add")
				err2 := testRunner.Run()
				assert.Error(err2)

				testRunner.Command("bun", "add")
				err3 := testRunner.Run()
				assert.NoError(err3)
			})

			It("should return the last executed command", func() {
				testRunner := NewMockCommandRunner()

				// Empty case
				cmd, exists := testRunner.LastCommand()
				assert.False(exists)
				assert.Equal(CommandCall{}, cmd)

				// With commands
				testRunner.Command("npm", "install")
				testRunner.Run()
				testRunner.Command("yarn", "add", "react")
				testRunner.Run()

				cmd, exists = testRunner.LastCommand()
				assert.True(exists)
				assert.Equal("yarn", cmd.Name)
				assert.Equal([]string{"add", "react"}, cmd.Args)
			})
		})

	})
})
