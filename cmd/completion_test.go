package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing" // Required for Ginkgo's entry point TestCompletion(t *testing.T)

	// Main Ginkgo package
	. "github.com/onsi/ginkgo/v2" // For Describe, It, BeforeEach, AfterEach, etc.
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCompletion is the Ginkgo entry point.
func TestCompletion(t *testing.T) {
	RunSpecs(t, "Completion Command Suite")
}

// testContext holds the common test setup and assertion helpers for completion command tests.
type testContext struct {
	assert       *assert.Assertions
	require      *require.Assertions
	cmd          *cobra.Command
	outBuf       *bytes.Buffer
	errBuf       *bytes.Buffer
	tempDir      string
	originalDir  string
	originalHome string
}

var _ = Describe("Completion Command", func() {
	var tc *testContext

	// Helper map for expected aliases, covering all default commands
	aliasExpectations := map[string][]string{
		"bash": {
			"function jpi()", "function jpr()", "function jpe()", "function jpx()",
			"function jpu()", "function jpun()", "function jpci()", "function jpa()",
		},
		"zsh": {
			"jpe() { jpd exec", "jpx() { jpd dlx", "jpi() { jpd install", "jpr() { jpd run",
			"jpu() { jpd update", "jpun() { jpd uninstall", "jpci() { jpd clean-install", "jpa() { jpd agent",
		},
		"fish": {
			"function jpe", "function jpx", "function jpi", "function jpr",
			"function jpu", "function jpun", "function jpci", "function jpa",
			"jpd exec $argv", "jpd dlx $argv", "jpd install $argv", "jpd run $argv",
			"jpd update $argv", "jpd uninstall $argv", "jpd clean-install $argv", "jpd agent $argv",
		},
		"nushell": {
			"export def jpe", "export def jpx", "export def jpi", "export def jpr",
			"export def jpu", "export def jpun", "export def jpci", "export def jpa",
		},
		"powershell": {
			"function jpe {", "function jpx {", "function jpi {", "function jpr {",
			"function jpu {", "function jpun {", "function jpci {", "function jpa {",
			"jpd exec @args", "jpd dlx @args", "jpd install @args", "jpd run @args",
			"jpd update @args", "jpd uninstall @args", "jpd clean-install @args", "jpd agent @args",
		},
	}

	BeforeEach(func() {
		ginkgoT := GinkgoT() // Get *testing.T for testify/assert and require

		currentAssert := assert.New(ginkgoT)
		currentRequire := require.New(ginkgoT)

		// Setup temp directory
		var err error
		tempDir, err := os.MkdirTemp("", "completion-test-*")
		currentRequire.NoError(err, "Failed to create temporary directory")

		// Store original HOME and set new one
		originalHome := os.Getenv("HOME")
		// The original code was setting HOME to its current value here.
		// Keeping this behavior for faithfulness to the original tests.
		err = os.Setenv("HOME", originalHome)
		currentRequire.NoError(err, "Failed to set HOME env var")

		// Change to temp directory
		originalDir, err := os.Getwd()
		currentRequire.NoError(err, "Failed to get current directory")
		err = os.Chdir(tempDir)
		currentRequire.NoError(err, "Failed to change to temporary directory")

		tc = &testContext{
			assert:       currentAssert,
			require:      currentRequire,
			cmd:          NewCompletionCmd(),
			outBuf:       new(bytes.Buffer),
			errBuf:       new(bytes.Buffer),
			tempDir:      tempDir,
			originalDir:  originalDir,
			originalHome: originalHome,
		}
		tc.cmd.SetOut(tc.outBuf)
		tc.cmd.SetErr(tc.errBuf)
	})

	AfterEach(func() {
		// Cleanup temp directory
		if tc.tempDir != "" {
			tc.assert.NoError(os.RemoveAll(tc.tempDir), "Failed to remove temporary directory")
		}

		// Restore original HOME
		if tc.originalHome != "" {
			tc.assert.NoError(os.Setenv("HOME", tc.originalHome), "Failed to restore HOME env var")
		} else {
			// If originalHome was empty, unset it to ensure a clean state
			_ = os.Unsetenv("HOME")
		}

		// Restore original directory
		if tc.originalDir != "" {
			tc.assert.NoError(os.Chdir(tc.originalDir), "Failed to restore original directory")
		}
	})

	Context("Output Flag (--output) generates completion script", func() {
		// Test case: basic output to file without shorthands
		It("without shorthands", func() {
			outputFile := "test_completion.bash"
			tc.cmd.SetArgs([]string{"bash", "--output", outputFile})

			err := tc.cmd.Execute()
			tc.assert.NoError(err, "Command execution failed")

			fullPath := filepath.Join(tc.tempDir, outputFile)
			tc.assert.FileExists(fullPath, "Output file does not exist at expected path: %s", fullPath)

			fileInfo, err := os.Stat(fullPath)
			tc.assert.NoError(err, "Failed to get file info")
			tc.assert.True(fileInfo.Size() > 0, "Output file is empty: %s", fullPath)

			content, err := os.ReadFile(fullPath)
			tc.assert.NoError(err, "Failed to read completion file")

			contentStr := string(content)
			tc.assert.Condition(func() bool {
				return strings.Contains(contentStr, "bash completion") || strings.Contains(contentStr, "_jpd_completion")
			}, "Generated completion file does not contain expected bash completion content")

			// Ensure no aliases are present when --with-shorthands is not used
			if len(aliasExpectations["bash"]) > 0 {
				tc.assert.NotContains(contentStr, aliasExpectations["bash"][0], "Expected no alias function in output without --with-shorthands")
			}

			tc.assert.Empty(tc.outBuf.String(), "Expected no output to stdout when using --output flag, but got: %s", tc.outBuf.String())
			tc.assert.Empty(tc.errBuf.String(), "Expected no output to stderr when using --output flag, but got: %s", tc.errBuf.String())
		})

		// Test case: output to file with shorthands
		It("with shorthands", func() {
			outputFile := "test_completion_with_shorthands.bash"
			tc.cmd.SetArgs([]string{"bash", "--output", outputFile, "--with-shorthands"})

			err := tc.cmd.Execute()
			tc.assert.NoError(err, "Command execution failed")

			fullPath := filepath.Join(tc.tempDir, outputFile)
			tc.assert.FileExists(fullPath, "Output file does not exist at expected path: %s", fullPath)

			fileInfo, err := os.Stat(fullPath)
			tc.assert.NoError(err, "Failed to get file info")
			tc.assert.True(fileInfo.Size() > 0, "Output file is empty: %s", fullPath)

			content, err := os.ReadFile(fullPath)
			tc.assert.NoError(err, "Failed to read completion file")

			contentStr := string(content)
			tc.assert.Condition(func() bool {
				return strings.Contains(contentStr, "bash completion") || strings.Contains(contentStr, "_jpd_completion")
			}, "Generated completion file does not contain expected bash completion content")

			// Check for expected alias functions
			for _, alias := range aliasExpectations["bash"] {
				tc.assert.Contains(contentStr, alias, "Expected output file to contain bash alias: %s", alias)
			}

			tc.assert.Empty(tc.outBuf.String(), "Expected no output to stdout when using --output flag, but got: %s", tc.outBuf.String())
			tc.assert.Empty(tc.errBuf.String(), "Expected no output to stderr when using --output flag, but got: %s", tc.errBuf.String())
		})
	})

	Context("Multiple Shells with Output Flag", func() {
		type outputTestCase struct {
			shell           string
			outputFile      string
			withShorthands  bool
			expectedContent string
			expectedAliases []string // Aliases to check for if withShorthands is true
		}

		testCases := []outputTestCase{
			// Cases without shorthands
			{"bash", "test_bash.bash", false, "bash completion", nil},
			{"zsh", "test_zsh.zsh", false, "zsh completion", nil},
			{"fish", "test_fish.fish", false, "fish completion", nil},
			{"nushell", "test_nushell.nu", false, "export extern \"jpd\"", nil},
			{"powershell", "test_powershell.ps1", false, "PowerShell completion", nil},

			// Cases with shorthands
			{"bash", "test_bash_sh.bash", true, "bash completion", aliasExpectations["bash"]},
			{"zsh", "test_zsh_sh.zsh", true, "zsh completion", aliasExpectations["zsh"]},
			{"fish", "test_fish_sh.fish", true, "fish completion", aliasExpectations["fish"]},
			{"nushell", "test_nushell_sh.nu", true, "export extern \"jpd\"", aliasExpectations["nushell"]},
			{"powershell", "test_powershell_sh.ps1", true, "PowerShell completion", aliasExpectations["powershell"]},
		}

		for _, tcVal := range testCases {
			tcVal := tcVal // Capture range variable
			testName := fmt.Sprintf("%s shell %s shorthands", tcVal.shell, map[bool]string{true: "with", false: "without"}[tcVal.withShorthands])
			It(testName, func() {
				args := []string{tcVal.shell, "--output", tcVal.outputFile}
				if tcVal.withShorthands {
					args = append(args, "--with-shorthands")
				}
				tc.cmd.SetArgs(args)

				err := tc.cmd.Execute()
				tc.assert.NoError(err, "Command execution failed for %s", testName)

				fullPath := filepath.Join(tc.tempDir, tcVal.outputFile)
				tc.assert.FileExists(fullPath, "Output file does not exist at expected path for %s: %s", testName, fullPath)

				fileInfo, err := os.Stat(fullPath)
				tc.assert.NoError(err, "Failed to get file info for %s", testName)
				tc.assert.True(fileInfo.Size() > 0, "Output file is empty for %s: %s", testName, fullPath)

				content, err := os.ReadFile(fullPath)
				tc.assert.NoError(err, "Failed to read completion file for %s", testName)

				contentStr := string(content)
				if tcVal.shell == "nushell" {
					tc.assert.Contains(contentStr, tcVal.expectedContent, "Generated nushell completion file does not contain expected content")
				} else {
					tc.assert.Contains(contentStr, tcVal.expectedContent, "Generated %s completion file does not contain expected content", tcVal.shell)
				}

				if tcVal.withShorthands {
					for _, alias := range tcVal.expectedAliases {
						tc.assert.Contains(contentStr, alias, "Expected %s completion with shorthands to contain alias: %s", tcVal.shell, alias)
					}
				} else {
					// Also assert that aliases are NOT present when --with-shorthands is false
					// Checking for one common alias from the expectations should be sufficient
					if len(aliasExpectations[tcVal.shell]) > 0 {
						tc.assert.NotContains(contentStr, aliasExpectations[tcVal.shell][0], "Expected no alias function in output without --with-shorthands for %s", tcVal.shell)
					}
				}

				tc.assert.Empty(tc.outBuf.String(), "Expected no output to stdout for %s when using --output flag", testName)
				tc.assert.Empty(tc.errBuf.String(), "Expected no error output to stderr for %s when using --output flag", testName)
			})
		}
	})

	Context("Default Stdout Output", func() {
		testCases := []struct {
			shell           string
			expectedContent string
		}{
			{"bash", "bash completion"},
			{"zsh", "zsh completion"},
			{"fish", "fish completion"},
			{"nushell", "export extern \"jpd\""},
			{"powershell", "PowerShell completion"},
		}

		for _, tcVal := range testCases {
			tcVal := tcVal // Capture range variable
			It(fmt.Sprintf("for %s shell", tcVal.shell), func() {
				tc.cmd.SetArgs([]string{tcVal.shell}) // No --with-shorthands for these tests

				err := tc.cmd.Execute()
				tc.assert.NoError(err, "Command execution failed for %s", tcVal.shell)

				output := tc.outBuf.String()
				tc.assert.NotEmpty(output, "Expected completion script output to stdout for %s, but got empty output", tcVal.shell)

				if tcVal.shell == "nushell" {
					tc.assert.Contains(output, tcVal.expectedContent, "Generated %s completion does not contain expected content. Expected to contain: %s", tcVal.shell, tcVal.expectedContent)
				} else {
					tc.assert.Contains(output, tcVal.expectedContent, "Generated %s completion does not contain expected content", tcVal.shell)
				}
				// Ensure no aliases are present when --with-shorthands is not used
				if len(aliasExpectations[tcVal.shell]) > 0 {
					tc.assert.NotContains(output, aliasExpectations[tcVal.shell][0], "Expected no alias function in stdout without --with-shorthands for %s", tcVal.shell)
				}
				tc.assert.Empty(tc.errBuf.String(), "Expected no error output to stderr for %s", tcVal.shell)
			})
		}
	})

	Context("With Shorthands Flag (--with-shorthands)", func() {
		testCases := []struct {
			shell string
		}{
			{"bash"},
			{"zsh"},
			{"fish"},
			{"nushell"},
			{"powershell"},
		}

		for _, tcVal := range testCases {
			tcVal := tcVal // Capture range variable
			It(fmt.Sprintf("for %s shell", tcVal.shell), func() {
				tc.cmd.SetArgs([]string{tcVal.shell, "--with-shorthands"})

				err := tc.cmd.Execute()
				tc.assert.NoError(err, "Command execution should succeed for %s", tcVal.shell)

				output := tc.outBuf.String()
				tc.assert.NotEmpty(output, "Expected output for %s with shorthands", tcVal.shell)

				// Also check for base completion content
				expectedBaseContent := ""
				switch tcVal.shell {
				case "bash":
					expectedBaseContent = "bash completion"
				case "zsh":
					expectedBaseContent = "zsh completion"
				case "fish":
					expectedBaseContent = "fish completion"
				case "nushell":
					expectedBaseContent = "export extern \"jpd\""
				case "powershell":
					expectedBaseContent = "PowerShell completion"
				}
				tc.assert.Contains(output, expectedBaseContent, "Generated %s completion does not contain expected base content", tcVal.shell)

				for _, expectedFunc := range aliasExpectations[tcVal.shell] {
					tc.assert.Contains(output, expectedFunc, "Expected %s to contain alias function: %s", tcVal.shell, expectedFunc)
				}
				tc.assert.Empty(tc.errBuf.String(), "Expected no error output to stderr for %s with shorthands", tcVal.shell)
			})
		}
	})

	Context("Error Cases", func() {
		It("should return an error for unsupported shell", func() {
			tc.cmd.SetArgs([]string{"unsupported"})

			err := tc.cmd.Execute()
			tc.assert.Error(err, "Expected error for unsupported shell")
			tc.assert.Contains(err.Error(), "unsupported shell", "Error message should mention unsupported shell")
			tc.assert.Contains(tc.errBuf.String(), "Error: unknown shell type", "Expected error output to stderr")
		})

		It("should return an error when no arguments provided", func() {
			tc.cmd.SetArgs([]string{})

			err := tc.cmd.Execute()
			tc.assert.Error(err, "Expected error when no arguments provided")
			tc.assert.Contains(err.Error(), "requires exactly one argument", "Error message should mention required argument")
			tc.assert.Contains(tc.errBuf.String(), "Error: requires exactly one argument", "Expected error output to stderr")
		})

		It("should return an error when too many arguments provided", func() {
			tc.cmd.SetArgs([]string{"bash", "zsh"})

			err := tc.cmd.Execute()
			tc.assert.Error(err, "Expected error when too many arguments provided")
			tc.assert.Contains(err.Error(), "requires exactly one argument", "Error message should mention required argument")
			tc.assert.Contains(tc.errBuf.String(), "Error: requires exactly one argument", "Expected error output to stderr")
		})

		It("should return an error for invalid output path (permission denied)", func() {
			// This test attempts to write to a protected path like /root.
			// It might behave differently based on OS/user permissions.
			// It should be skipped if running with elevated permissions where /root is writable.
			invalidPath := "/root/cannot-create/test_completion.sh"
			tc.cmd.SetArgs([]string{"bash", "--output", invalidPath})

			err := tc.cmd.Execute()
			if err != nil {
				tc.assert.Error(err)
				tc.assert.Contains(err.Error(), "failed to create", "Error message should mention creation failure")
				tc.assert.Contains(tc.errBuf.String(), "failed to create", "Expected error output to stderr about creation failure")
			} else {
				// Use Ginkgo's Skip function
				Skip("Skipping invalid output path test - command executed successfully, likely running with elevated permissions.")
			}
		})
	})
})

var _ = Describe("GetSupportedShells", func() {
	It("should return the list of supported shells in expected order", func() {
		ginkgoT := GinkgoT()
		currentAssert := assert.New(ginkgoT)

		supportedShells := getSupportedShells()

		expectedShells := []string{"bash", "fish", "nushell", "powershell", "zsh"}
		currentAssert.Equal(expectedShells, supportedShells, "Shells should be in the expected order")
	})
})

var _ = Describe("GetDefaultAliasMapping", func() {
	It("should return the correct default alias mapping", func() {
		ginkgoT := GinkgoT()
		currentAssert := assert.New(ginkgoT)

		aliasMap := getDefaultAliasMapping()

		expectedCommands := []string{"install", "run", "exec", "dlx", "update", "uninstall", "clean-install", "agent"}
		for _, cmd := range expectedCommands {
			currentAssert.Contains(aliasMap, cmd, "Should have aliases for %s command", cmd)
			currentAssert.NotEmpty(aliasMap[cmd], "Should have at least one alias for %s command", cmd)
		}

		currentAssert.Contains(aliasMap["exec"], "jpe", "exec command should have jpe alias")
		currentAssert.Contains(aliasMap["dlx"], "jpx", "dlx command should have jpx alias")
		currentAssert.Contains(aliasMap["install"], "jpi", "install command should have jpi alias")
		currentAssert.Contains(aliasMap["run"], "jpr", "run command should have jpr alias")

		for cmd, aliases := range aliasMap {
			expectedJpdAlias := fmt.Sprintf("jpd-%s", cmd)
			currentAssert.Contains(aliases, expectedJpdAlias, "Command %s should have jpd- prefixed alias %s", cmd, expectedJpdAlias)
		}
	})
})
