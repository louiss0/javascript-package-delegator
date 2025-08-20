package completion_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"  // For general assertions
	"github.com/stretchr/testify/require" // For critical assertions (fail fast, like file operations)

	. "github.com/onsi/ginkgo/v2" // Import Ginkgo
	// Gomega is explicitly not used as per the prompt.

	"github.com/louiss0/javascript-package-delegator/internal/completion"
)

// Helper to capture output when `filename` is empty (output to cmd.OutOrStdout())
func captureOutput(cmd *cobra.Command, f func() error) (string, error) {
	oldOut := cmd.OutOrStdout()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	defer cmd.SetOut(oldOut) // Restore original output writer
	err := f()
	return buf.String(), err
}

// Helper for tests that write to a file via the `filename` argument
// Now accepts FullGinkgoTInterface to integrate with Ginkgo's test context.
func withTempFile(ginkgoT FullGinkgoTInterface, testFunc func(filename string) error) (string, error) {
	// Use testify/require with the GinkgoT() for critical assertions within this helper
	requires := require.New(ginkgoT)
	tmpfile, err := os.CreateTemp("", "jpd-completion-test-*.out")
	requires.NoError(err, "failed to create temporary file for test")

	filename := tmpfile.Name()
	requires.NoError(tmpfile.Close()) // Close the temporary file handle; GenerateCompletion will re-open/create it
	defer func() {
		if rErr := os.Remove(filename); rErr != nil {
			// Log error using GinkgoT's logging, but don't fail the test
			ginkgoT.Logf("warning: failed to remove temporary file %s: %v", filename, rErr)
		}
	}() // Clean up the file after the test

	err = testFunc(filename)
	if err != nil {
		return "", err
	}

	content, readErr := os.ReadFile(filename)
	requires.NoError(readErr, "failed to read content from temporary file")
	return string(content), nil
}

// TestCompletionGenerator is the entry point for running Ginkgo specs.
func TestCompletionGenerator(t *testing.T) {
	RunSpecs(t, "Completion Generator Suite")
}

// Define the Ginkgo BDD test suite
var _ = Describe("CompletionGenerator", func() {
	// Common setup variables accessible to all It blocks within this Describe
	var generator completion.Generator
	var cmd *cobra.Command

	// BeforeEach runs before each It block in this Describe context
	BeforeEach(func() {
		generator = completion.NewGenerator()
		cmd = &cobra.Command{Use: "test"}
	})

	Context("GetSupportedShells", func() {
		It("should return the correct list of supported shells", func() {
			// Use testify/assert with GinkgoT() for assertions
			asserts := assert.New(GinkgoT())
			shells := generator.GetSupportedShells()

			asserts.Len(shells, 7)
			// Use ElementsMatch for content regardless of order, then Equal for exact order
			asserts.ElementsMatch([]string{"bash", "carapace", "fish", "nushell", "powershell", "warp", "zsh"}, shells)
			asserts.Equal([]string{"bash", "carapace", "fish", "nushell", "powershell", "warp", "zsh"}, shells)
		})
	})

	Context("GetDefaultAliasMapping", func() {
		It("should return the correct default alias mapping", func() {
			asserts := assert.New(GinkgoT())
			aliasMap := generator.GetDefaultAliasMapping()

			asserts.Contains(aliasMap, "install")
			asserts.Contains(aliasMap, "run")
			asserts.Contains(aliasMap, "exec")
			asserts.Contains(aliasMap, "dlx")
			asserts.Contains(aliasMap, "update")
			asserts.Contains(aliasMap, "uninstall")
			asserts.Contains(aliasMap, "clean-install")
			asserts.Contains(aliasMap, "agent")

			asserts.Contains(aliasMap["exec"], "jpe")
			asserts.Contains(aliasMap["dlx"], "jpx")

			asserts.Contains(aliasMap["install"], "jpd-install")
			asserts.Contains(aliasMap["run"], "jpd-run")
			asserts.Contains(aliasMap["exec"], "jpd-exec")
			asserts.Contains(aliasMap["dlx"], "jpd-dlx")
		})
	})

	Context("GenerateCompletion", func() {
		Context("when generating base scripts", func() {
			shellTestCases := []struct {
				name         string
				shell        string
				substrings   []string
				nonEmptyOnly bool
			}{
				{name: "bash", shell: "bash", substrings: []string{"bash completion"}},
				{name: "zsh", shell: "zsh", substrings: []string{"zsh completion"}},
				{name: "fish", shell: "fish", substrings: []string{"fish completion"}},
				{name: "nushell", shell: "nushell", nonEmptyOnly: true},
				{name: "powershell", shell: "powershell", substrings: []string{"PowerShell"}},
				{name: "carapace", shell: "carapace", substrings: []string{"# Carapace completion spec for jpd", "name: jpd"}},
				{name: "warp", shell: "warp", nonEmptyOnly: true},
			}

			for _, tc := range shellTestCases {
				tc := tc // Capture range variable for closure
				It(fmt.Sprintf("should generate %s completion script to stdout when filename is empty", tc.name), func() {
					asserts := assert.New(GinkgoT())
					output, err := captureOutput(cmd, func() error {
						return generator.GenerateCompletion(cmd, tc.shell, "", false)
					})
					asserts.NoError(err)

					if tc.nonEmptyOnly {
						asserts.NotEmpty(output)
					} else {
						for _, s := range tc.substrings {
							asserts.Contains(output, s)
						}
					}
				})

				It(fmt.Sprintf("should generate %s completion script to a file when filename is provided", tc.name), func() {
					asserts := assert.New(GinkgoT())

					// Skip warp since it requires directory not file
					if tc.shell == "warp" {
						asserts.True(true, "warp requires directory output, not file - skipping file test")
						return
					}

					output, err := withTempFile(GinkgoT(), func(filename string) error {
						return generator.GenerateCompletion(cmd, tc.shell, filename, false)
					})
					asserts.NoError(err)

					if tc.nonEmptyOnly {
						asserts.NotEmpty(output)
					} else {
						for _, s := range tc.substrings {
							asserts.Contains(output, s)
						}
					}
				})
			}
		})

		Context("when generating with shorthand aliases", func() {
			shellTestCases := []struct {
				name       string
				shell      string
				substrings []string
			}{
				{name: "bash", shell: "bash", substrings: []string{"function jpe()", "function jpx()", "function jpi()"}},
				{name: "zsh", shell: "zsh", substrings: []string{"jpe() { jpd exec", "jpx() { jpd dlx"}},
				{name: "fish", shell: "fish", substrings: []string{"function jpe", "function jpx", "jpd exec $argv"}},
				{name: "nushell", shell: "nushell", substrings: []string{"export def jpe", "export def jpx"}},
				{name: "powershell", shell: "powershell", substrings: []string{"function jpe {", "function jpx {", "jpd exec @args", "Register-ArgumentCompleter"}},
			}

			for _, tc := range shellTestCases {
				tc := tc // Capture range variable for closure
				It(fmt.Sprintf("should generate %s completion script with shorthands to stdout when filename is empty", tc.name), func() {
					asserts := assert.New(GinkgoT())
					output, err := captureOutput(cmd, func() error {
						return generator.GenerateCompletion(cmd, tc.shell, "", true)
					})
					asserts.NoError(err)

					for _, s := range tc.substrings {
						asserts.Contains(output, s)
					}
				})

				It(fmt.Sprintf("should generate %s completion script with shorthands to a file when filename is provided", tc.name), func() {
					asserts := assert.New(GinkgoT())
					output, err := withTempFile(GinkgoT(), func(filename string) error {
						return generator.GenerateCompletion(cmd, tc.shell, filename, true)
					})
					asserts.NoError(err)

					for _, s := range tc.substrings {
						asserts.Contains(output, s)
					}
				})
			}
		})

		Context("error handling", func() {
			It("should return an error for an unsupported shell", func() {
				asserts := assert.New(GinkgoT())
				output, err := captureOutput(cmd, func() error {
					return generator.GenerateCompletion(cmd, "unsupported", "", false)
				})
				asserts.Empty(output)
				asserts.Error(err)
				asserts.Contains(err.Error(), "unsupported shell: unsupported")
			})

			It("should return an error if file creation fails when filename is provided", func() {
				asserts := assert.New(GinkgoT())
				// Attempt to create a file in a non-existent and uncreatable directory
				invalidDir := filepath.Join("non-existent-dir", "invalid-path", "test_completion.sh")
				err := generator.GenerateCompletion(cmd, "bash", invalidDir, false)
				asserts.Error(err)
				asserts.Contains(err.Error(), "failed to create output file")
			})
		})
	})

	// Note: GetNushellCompletionScript and GenerateCarapaceBridge tests have been moved to integrations package
	// since the functions have been refactored into the AliasGenerator interface
})
