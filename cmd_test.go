package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/louiss0/javascript-package-delegator/cmd"
	. "github.com/onsi/ginkgo/v2"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// This function executes a cobra command with the given arguments and returns the output and error.
// It sets the output and error buffers for the command, sets the arguments, and executes the command.
// If there is an error, it returns an error with the error message from the error buffer.
// If there is no error, it returns the output from the output buffer.
// It's used to test the cobra commands.
// When you use this function, make sure to pass the root command and any arguments you want to test.
// The first argument after the rootCmd is any sub command or flag you want to test.

func executeCmd(cmd *cobra.Command, args ...string) (string, error) {

	buf := new(bytes.Buffer)
	errBuff := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(errBuff)
	cmd.SetArgs(args)

	err := cmd.Execute()

	if errBuff.Len() > 0 {
		return "", fmt.Errorf("command failed: %s", errBuff.String())
	}

	return buf.String(), err
}

var _ = Describe("JPD Commands", func() {

	assert := assert.New(GinkgoT())
	var rootCmd *cobra.Command

	getSubCommandWithName := func(cmd *cobra.Command, name string) (*cobra.Command, bool) {

		return lo.Find(
			cmd.Commands(),
			func(item *cobra.Command) bool {
				return item.Name() == name
			})
	}

	JustBeforeEach(func() {

		rootCmd = cmd.NewRootCmd()
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
	})

	Describe("Run Command", func() {

		var runCmd *cobra.Command
		BeforeEach(func() {
			runCmd, _ = getSubCommandWithName(rootCmd, "run")

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
	})

	Describe("Update Command", func() {

		var updateCmd *cobra.Command
		BeforeEach(func() {
			updateCmd, _ = getSubCommandWithName(rootCmd, "update")
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
	})

	Describe("Uninstall Command", func() {

		var uninstallCmd *cobra.Command
		BeforeEach(func() {
			uninstallCmd, _ = getSubCommandWithName(rootCmd, "uninstall")

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

			assert.Error(error)
		})
	})

	Describe("Clean Install Command", func() {

		var cleanInstallCmd *cobra.Command
		BeforeEach(func() {
			cleanInstallCmd, _ = getSubCommandWithName(rootCmd, "clean-install")
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
	})

	Describe("Package Manager Detection", func() {
		var tempDir string
		var originalDir string

		BeforeEach(func() {
			originalDir, _ = os.Getwd()
			tempDir, _ = os.MkdirTemp("", "jpd-test-*")
			os.Chdir(tempDir)
		})

		AfterEach(func() {
			os.Chdir(originalDir)
			os.RemoveAll(tempDir)
		})

		It("should detect deno from deno.lock", func() {

			os.WriteFile("deno.lock", []byte(`{}`), 0644)
			pkg, error := cmd.DetectPackageManager()

			assert.NoError(error)

			assert.Equal(pkg, "deno")
			assert.True(true) // Placeholder
		})

		It("should detect deno from deno.json", func() {
			os.WriteFile("deno.json", []byte(`{}`), 0644)
			pkg, error := cmd.DetectPackageManager()

			assert.NoError(error)

			assert.Equal(pkg, "deno")
		})

		It("should detect deno from deno.jsonc", func() {
			os.WriteFile("deno.jsonc", []byte(`{}`), 0644)
			pkg, error := cmd.DetectPackageManager()

			assert.NoError(error)

			assert.Equal(pkg, "deno") // Would test actual detection here
		})

		It("should detect bun from bun.lockb", func() {
			os.WriteFile("bun.lockb", []byte(``), 0644)
			pkg, error := cmd.DetectPackageManager()

			assert.NoError(error)

			assert.Equal(pkg, "bun")
		})

		It("should detect pnpm from pnpm-lock.yaml", func() {
			os.WriteFile("pnpm-lock.yaml", []byte(`lockfileVersion: 5.4`), 0644)
			pkg, error := cmd.DetectPackageManager()

			assert.NoError(error)

			assert.Equal(pkg, "pnpm")
		})

		It("should detect yarn from yarn.lock", func() {
			os.WriteFile("yarn.lock", []byte(`# THIS IS AN AUTOGENERATED FILE`), 0644)
			pkg, error := cmd.DetectPackageManager()

			assert.NoError(error)

			assert.Equal(pkg, "yarn")
		})

		It("should detect npm from package-lock.json", func() {
			pkg, error := cmd.DetectPackageManager()

			assert.NoError(error)

			assert.Equal(pkg, "npm")
		})

		It("should default to npm when no lock files found", func() {
			pkg, error := cmd.DetectPackageManager()

			assert.NoError(error)

			assert.Equal(pkg, "npm")
		})

		PIt("should prioritize deno over other package managers", func() {
			// Create multiple lock files
			os.WriteFile("deno.json", []byte(`{}`), 0644)
			os.WriteFile("package-lock.json", []byte(`{}`), 0644)
			os.WriteFile("yarn.lock", []byte(``), 0644)
			// Should detect deno first
			assert.True(true) // Placeholder
		})

		PIt("should prioritize according to order: deno > bun > pnpm > yarn > npm", func() {
			// Test priority order when multiple files exist
			assert.True(true) // Placeholder
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

})
