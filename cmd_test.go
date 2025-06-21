package main

import (
	"bytes"
	"context"
	"fmt"

	"github.com/louiss0/javascript-package-delegator/cmd"
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

// setupCommandContext sets up the command context with necessary values for testing
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

	JustBeforeEach(func() {
		mockRunner = NewMockCommandRunner()
		rootCmd = cmd.NewRootCmd(mockRunner)
		// This needs to be set because Ginkgo will pass a --test.timeout flag to the root command
		// The test.timeout flag will get in the way
		// If the args are empty before they are set by executeCommand the right args can be passed
		rootCmd.SetArgs([]string{})

		// Set default context values that all commands will inherit
		ctx := context.Background()
		ctx = context.WithValue(ctx, "go_mode", "development")
		ctx = context.WithValue(ctx, "command_runner", mockRunner)
		rootCmd.SetContext(ctx)
	})

	// Helper function to set up command context with the necessary values for testing
	// This preserves any existing context values and adds or updates the ones we need
	setupCommandContext := func(cmd *cobra.Command, packageManager string) {
		var ctx context.Context
		if cmd.Context() != nil {
			ctx = cmd.Context()
		} else {
			ctx = context.Background()
		}
		ctx = context.WithValue(ctx, "package-name", packageManager)
		ctx = context.WithValue(ctx, "go_mode", "development")
		ctx = context.WithValue(ctx, "command_runner", mockRunner)
		cmd.SetContext(ctx)
	}

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
			setupCommandContext(rootCmd, "npm")
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
			mockRunner.Reset()
			_, err := executeCmd(rootCmd, "install")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npm", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"install"}, mockRunner.CommandCalls[0].Args)
		})

		It("should run npm install with package names", func() {
			mockRunner.Reset()
			_, err := executeCmd(rootCmd, "install", "lodash", "express")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npm", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"install", "lodash", "express"}, mockRunner.CommandCalls[0].Args)
		})

		It("should run npm install with dev flag", func() {
			mockRunner.Reset()
			_, err := executeCmd(rootCmd, "install", "--dev", "typescript")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npm", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "--save-dev")
		})

		It("should run npm install with global flag", func() {
			mockRunner.Reset()
			_, err := executeCmd(rootCmd, "install", "--global", "typescript")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npm", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "--global")
		})

		It("should run npm install with production flag", func() {
			mockRunner.Reset()
			_, err := executeCmd(rootCmd, "install", "--production")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npm", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "--omit=dev")
		})

		It("should run yarn add with dev flag", func() {
			mockRunner.Reset()
			setupCommandContext(rootCmd, "yarn")
			_, err := executeCmd(rootCmd, "install", "--dev", "typescript")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("yarn", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "--dev")
			// Reset to npm for other tests
			setupCommandContext(rootCmd, "npm")
		})

		It("should run pnpm add with dev flag", func() {
			mockRunner.Reset()
			setupCommandContext(rootCmd, "pnpm")
			_, err := executeCmd(rootCmd, "install", "--dev", "typescript")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("pnpm", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "--save-dev")
			// Reset to npm for other tests
			setupCommandContext(rootCmd, "npm")
		})
	})

	Describe("Run Command", func() {

		var runCmd *cobra.Command
		BeforeEach(func() {
			runCmd, _ = getSubCommandWithName(rootCmd, "run")
			setupCommandContext(rootCmd, "npm")
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
			mockRunner.Reset()
			_, err := executeCmd(rootCmd, "run", "test")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npm", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"run", "test"}, mockRunner.CommandCalls[0].Args)
		})

		It("should run npm run with script args", func() {
			mockRunner.Reset()
			_, err := executeCmd(rootCmd, "run", "test", "--", "--watch")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npm", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"run", "test", "--", "--watch"}, mockRunner.CommandCalls[0].Args)
		})

		It("should run npm run with if-present flag", func() {
			mockRunner.Reset()
			_, err := executeCmd(rootCmd, "run", "--if-present", "test")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npm", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "--if-present")
		})

		It("should run yarn run with script name", func() {
			mockRunner.Reset()
			setupCommandContext(rootCmd, "yarn")
			_, err := executeCmd(rootCmd, "run", "test")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("yarn", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"run", "test"}, mockRunner.CommandCalls[0].Args)
			// Reset to npm for other tests
			setupCommandContext(rootCmd, "npm")
		})

		It("should run pnpm run with script args", func() {
			mockRunner.Reset()
			setupCommandContext(rootCmd, "pnpm")
			_, err := executeCmd(rootCmd, "run", "test", "--", "--watch")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("pnpm", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"run", "test", "--", "--watch"}, mockRunner.CommandCalls[0].Args)
			// Reset to npm for other tests
			setupCommandContext(rootCmd, "npm")
		})

		It("should run deno task with script name", func() {
			mockRunner.Reset()
			setupCommandContext(rootCmd, "deno")
			_, err := executeCmd(rootCmd, "run", "test")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("deno", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"task", "test"}, mockRunner.CommandCalls[0].Args)
			// Reset to npm for other tests
			setupCommandContext(rootCmd, "npm")
		})
	})

	Describe("Exec Command", func() {
		var execCmd *cobra.Command
		BeforeEach(func() {
			execCmd, _ = getSubCommandWithName(rootCmd, "exec")
			setupCommandContext(rootCmd, "npm")
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

		It("should execute npx with package name", func() {
			mockRunner.Reset()
			_, err := executeCmd(rootCmd, "exec", "create-react-app")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npx", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"create-react-app"}, mockRunner.CommandCalls[0].Args)
		})

		It("should execute npx with package name and args", func() {
			mockRunner.Reset()
			_, err := executeCmd(rootCmd, "exec", "create-react-app", "my-app")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npx", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"create-react-app", "my-app"}, mockRunner.CommandCalls[0].Args)
		})

		It("should fail when command execution fails", func() {
			mockRunner.Reset()
			mockRunner.InvalidCommands = []string{"npx"}
			_, err := executeCmd(rootCmd, "exec", "create-react-app", "my-app")
			assert.Error(err)
			mockRunner.InvalidCommands = []string{}
		})

		It("should execute yarn with package name", func() {
			mockRunner.Reset()
			setupCommandContext(rootCmd, "yarn")
			_, err := executeCmd(rootCmd, "exec", "create-react-app")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("yarn", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"create-react-app"}, mockRunner.CommandCalls[0].Args)
			// Reset to npm for other tests
			setupCommandContext(rootCmd, "npm")
		})

		It("should execute pnpm dlx with package name", func() {
			mockRunner.Reset()
			setupCommandContext(rootCmd, "pnpm")
			_, err := executeCmd(rootCmd, "exec", "create-react-app", "my-app")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("pnpm", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"dlx", "create-react-app", "my-app"}, mockRunner.CommandCalls[0].Args)
			// Reset to npm for other tests
			setupCommandContext(rootCmd, "npm")
		})

		It("should execute bunx with package name", func() {
			mockRunner.Reset()
			setupCommandContext(rootCmd, "bun")
			_, err := executeCmd(rootCmd, "exec", "create-react-app", "my-app")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("bunx", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"create-react-app", "my-app"}, mockRunner.CommandCalls[0].Args)
			// Reset to npm for other tests
			setupCommandContext(rootCmd, "npm")
		})

		It("should return error for deno", func() {
			mockRunner.Reset()
			setupCommandContext(rootCmd, "deno")
			_, err := executeCmd(rootCmd, "exec", "create-react-app", "my-app")
			assert.Error(err)
			assert.Contains(err.Error(), "Deno doesn't have a dlx")
			// Reset to npm for other tests
			setupCommandContext(rootCmd, "npm")
		})
	})

	Describe("Update Command", func() {

		var updateCmd *cobra.Command
		BeforeEach(func() {
			updateCmd, _ = getSubCommandWithName(rootCmd, "update")
			setupCommandContext(rootCmd, "npm")
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
			mockRunner.Reset()
			_, err := executeCmd(rootCmd)
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npm", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"update"}, mockRunner.CommandCalls[0].Args)
		})

		It("should run npm update with package names", func() {
			mockRunner.Reset()
			_, err := executeCmd(rootCmd, "update", "lodash", "express")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npm", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"update", "lodash", "express"}, mockRunner.CommandCalls[0].Args)
		})

		It("should run npm update with global flag", func() {
			mockRunner.Reset()
			_, err := executeCmd(rootCmd, "update", "--global", "typescript")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npm", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "--global")
		})

		It("should handle latest flag for npm", func() {
			mockRunner.Reset()
			_, err := executeCmd(rootCmd, "update", "--latest", "lodash")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npm", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "lodash@latest")
		})

		It("should handle interactive flag for yarn", func() {
			mockRunner.Reset()
			setupCommandContext(rootCmd, "yarn")
			_, err := executeCmd(rootCmd, "update", "--interactive")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("yarn", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"upgrade-interactive"}, mockRunner.CommandCalls[0].Args)
			// Reset to npm for other tests
			setupCommandContext(rootCmd, "npm")
		})

		It("should handle interactive flag for pnpm", func() {
			mockRunner.Reset()
			setupCommandContext(rootCmd, "pnpm")
			_, err := executeCmd(rootCmd, "update", "--interactive")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("pnpm", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "--interactive")
			// Reset to npm for other tests
			setupCommandContext(rootCmd, "npm")
		})

		It("should handle latest flag for yarn", func() {
			mockRunner.Reset()
			setupCommandContext(rootCmd, "yarn")
			_, err := executeCmd(rootCmd, "update", "--latest", "lodash")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("yarn", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "--latest")
			// Reset to npm for other tests
			setupCommandContext(rootCmd, "npm")
		})

		It("should error on npm with interactive flag", func() {
			mockRunner.Reset()
			_, err := executeCmd(rootCmd, "update", "--interactive")
			assert.Error(err)
			assert.Contains(err.Error(), "npm does not support interactive updates")
		})
	})

	Describe("Uninstall Command", func() {

		var uninstallCmd *cobra.Command
		BeforeEach(func() {
			uninstallCmd, _ = getSubCommandWithName(rootCmd, "uninstall")
			setupCommandContext(rootCmd, "npm")
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
			mockRunner.Reset()
			_, err := executeCmd(rootCmd, "uninstall", "lodash")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npm", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"uninstall", "lodash"}, mockRunner.CommandCalls[0].Args)
		})

		It("should run npm uninstall with multiple package names", func() {
			mockRunner.Reset()
			_, err := executeCmd(rootCmd, "uninstall", "lodash", "express")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npm", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"uninstall", "lodash", "express"}, mockRunner.CommandCalls[0].Args)
		})

		It("should run npm uninstall with global flag", func() {
			mockRunner.Reset()
			_, err := executeCmd(rootCmd, "uninstall", "--global", "typescript")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npm", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "--global")
		})

		It("should run yarn remove with package name", func() {
			mockRunner.Reset()
			setupCommandContext(rootCmd, "yarn")
			_, err := executeCmd(rootCmd, "uninstall", "lodash")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("yarn", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"remove", "lodash"}, mockRunner.CommandCalls[0].Args)
			setupCommandContext(rootCmd, "npm")
		})

		It("should run pnpm remove with global flag", func() {
			mockRunner.Reset()
			setupCommandContext(rootCmd, "pnpm")
			_, err := executeCmd(rootCmd, "uninstall", "--global", "typescript")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("pnpm", mockRunner.CommandCalls[0].Name)
			assert.Contains(mockRunner.CommandCalls[0].Args, "--global")
			setupCommandContext(rootCmd, "npm")
		})

		It("should run bun remove with multiple packages", func() {
			mockRunner.Reset()
			setupCommandContext(rootCmd, "bun")
			_, err := executeCmd(rootCmd, "uninstall", "react", "react-dom")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("bun", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"remove", "react", "react-dom"}, mockRunner.CommandCalls[0].Args)
			setupCommandContext(rootCmd, "npm")
		})

		It("should return error for unsupported package manager", func() {
			mockRunner.Reset()
			setupCommandContext(rootCmd, "unsupported")
			_, err := executeCmd(rootCmd, "uninstall", "lodash")
			assert.Error(err)
			assert.Contains(err.Error(), "unsupported package manager")
			setupCommandContext(rootCmd, "npm")
		})
	})

	Describe("Clean Install Command", func() {

		var cleanInstallCmd *cobra.Command
		BeforeEach(func() {
			cleanInstallCmd, _ = getSubCommandWithName(rootCmd, "clean-install")
			setupCommandContext(rootCmd, "npm")
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
			mockRunner.Reset()
			_, err := executeCmd(rootCmd, "ci")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npm", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"ci"}, mockRunner.CommandCalls[0].Args)
		})

		It("should run yarn install with frozen lockfile", func() {
			setupCommandContext(rootCmd, "yarn")
			mockRunner.Reset()
			_, err := executeCmd(rootCmd, "clean-install")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("yarn", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"install", "--frozen-lockfile"}, mockRunner.CommandCalls[0].Args)
		})

		It("should run pnpm install with frozen lockfile", func() {
			setupCommandContext(rootCmd, "pnpm")
			mockRunner.Reset()
			_, err := executeCmd(rootCmd, "clean-install")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("pnpm", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"install", "--frozen-lockfile"}, mockRunner.CommandCalls[0].Args)
		})

		It("should run bun install with frozen lockfile", func() {
			setupCommandContext(rootCmd, "bun")
			mockRunner.Reset()
			_, err := executeCmd(rootCmd, "clean-install")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("bun", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"install", "--frozen-lockfile"}, mockRunner.CommandCalls[0].Args)
		})

		It("should return error for deno", func() {
			setupCommandContext(rootCmd, "deno")
			mockRunner.Reset()
			_, err := executeCmd(rootCmd, "clean-install")
			assert.Error(err)
			assert.Contains(err.Error(), "deno doesn't support this command")
		})
	})

	Describe("Agent Command", func() {

		var agentCmd *cobra.Command
		BeforeEach(func() {
			agentCmd, _ = getSubCommandWithName(rootCmd, "agent")
			setupCommandContext(rootCmd, "npm")
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
			mockRunner.Reset()
			_, err := executeCmd(rootCmd, "agent")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npm", mockRunner.CommandCalls[0].Name)
		})

		It("should pass arguments to package manager", func() {
			mockRunner.Reset()
			_, err := executeCmd(rootCmd, "agent", "--", "--version")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("npm", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"--version"}, mockRunner.CommandCalls[0].Args)
		})

		It("should execute yarn with arguments", func() {
			mockRunner.Reset()
			setupCommandContext(rootCmd, "yarn")
			_, err := executeCmd(rootCmd, "agent", "--version")
			assert.NoError(err)
			assert.Equal(2, len(mockRunner.CommandCalls))
			assert.Equal("yarn", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"--version"}, mockRunner.CommandCalls[0].Args)
			setupCommandContext(rootCmd, "npm")
		})

		It("should execute pnpm with arguments", func() {
			mockRunner.Reset()
			setupCommandContext(rootCmd, "pnpm")
			_, err := executeCmd(rootCmd, "agent", "info")
			assert.NoError(err)
			assert.Equal(1, len(mockRunner.CommandCalls))
			assert.Equal("pnpm", mockRunner.CommandCalls[0].Name)
			assert.Equal([]string{"info"}, mockRunner.CommandCalls[0].Args)
		})

		It("should fail when command execution fails", func() {
			mockRunner.Reset()
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
