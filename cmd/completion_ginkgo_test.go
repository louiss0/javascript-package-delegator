package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	cmdPkg "github.com/louiss0/javascript-package-delegator/cmd"
)

var _ = Describe("Completion Command", Label("fast", "unit"), func() {
	var (
		tempDir     string
		originalDir string
	)

	BeforeEach(func() {
		tempDir = GinkgoT().TempDir()
		var err error
		originalDir, err = os.Getwd()
		assert.NoError(GinkgoT(), err, "Failed to get current directory")

		err = os.Chdir(tempDir)
		assert.NoError(GinkgoT(), err, "Failed to change to temp directory")
	})

	AfterEach(func() {
		_ = os.Chdir(originalDir)
	})

	Describe("Output Flag Behavior", func() {
		It("should create output file when --output flag is provided", func() {
			completionCmd := cmdPkg.NewCompletionCmd()
			outputFile := "test_completion.bash"
			completionCmd.SetArgs([]string{"bash", "--output", outputFile})

			var buf bytes.Buffer
			completionCmd.SetOut(&buf)

			err := completionCmd.Execute()
			assert.NoError(GinkgoT(), err, "Command execution should succeed")

			// Check that the file exists
			fullPath := filepath.Join(tempDir, outputFile)
			_, err = os.Stat(fullPath)
			assert.NoError(GinkgoT(), err, "Output file should exist")

			// Check that the file is not empty
			fileInfo, err := os.Stat(fullPath)
			assert.NoError(GinkgoT(), err, "Should be able to get file info")
			assert.Greater(GinkgoT(), fileInfo.Size(), int64(0), "Output file should not be empty")

			// Verify the completion script contains expected content
			content, err := os.ReadFile(fullPath)
			assert.NoError(GinkgoT(), err, "Should be able to read completion file")

			contentStr := string(content)
			assert.True(GinkgoT(),
				assert.Contains(GinkgoT(), contentStr, "bash completion") ||
					assert.Contains(GinkgoT(), contentStr, "_jpd_completion"),
				"Generated completion file should contain expected bash completion content")

			// Check that no output was written to stdout when using --output flag
			output := buf.String()
			assert.Empty(GinkgoT(), output, "Expected no output to stdout when using --output flag")
		})
	})

	Describe("Multiple Shell Support", func() {
		testCases := []struct {
			shell           string
			outputFile      string
			expectedContent string
		}{
			{"bash", "test_bash.bash", "bash completion"},
			{"zsh", "test_zsh.zsh", "zsh completion"},
			{"fish", "test_fish.fish", "fish completion"},
			{"nushell", "test_nushell.nu", "extern jpd"},
			{"powershell", "test_powershell.ps1", "PowerShell completion"},
		}

		for _, tc := range testCases {
			tc := tc // capture loop variable
			It("should generate completion for "+tc.shell, func() {
				completionCmd := cmdPkg.NewCompletionCmd()
				completionCmd.SetArgs([]string{tc.shell, "--output", tc.outputFile})

				err := completionCmd.Execute()
				assert.NoError(GinkgoT(), err, "Command execution should succeed for "+tc.shell)

				// Check that the file exists
				fullPath := filepath.Join(tempDir, tc.outputFile)
				_, err = os.Stat(fullPath)
				assert.NoError(GinkgoT(), err, "Output file should exist for "+tc.shell)

				// Check that the file is not empty
				fileInfo, err := os.Stat(fullPath)
				assert.NoError(GinkgoT(), err, "Should be able to get file info for "+tc.shell)
				assert.Greater(GinkgoT(), fileInfo.Size(), int64(0), "Output file should not be empty for "+tc.shell)

				// For nushell, check specific content
				if tc.shell == "nushell" {
					content, err := os.ReadFile(fullPath)
					assert.NoError(GinkgoT(), err, "Should be able to read completion file for "+tc.shell)
					contentStr := string(content)
					assert.Contains(GinkgoT(), contentStr, "export extern \"jpd\"", "Generated nushell completion should contain expected content")
				}
			})
		}
	})

	Describe("Stdout Output Behavior", func() {
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

		for _, tc := range testCases {
			tc := tc // capture loop variable
			It("should output to stdout for "+tc.shell+" when no output file specified", func() {
				completionCmd := cmdPkg.NewCompletionCmd()
				completionCmd.SetArgs([]string{tc.shell})

				var buf bytes.Buffer
				completionCmd.SetOut(&buf)

				err := completionCmd.Execute()
				assert.NoError(GinkgoT(), err, "Command execution should succeed for "+tc.shell)

				// Check that output went to stdout
				output := buf.String()
				assert.NotEmpty(GinkgoT(), output, "Expected completion output to stdout for "+tc.shell)

				// Verify completion content is in stdout
				if tc.shell == "nushell" {
					assert.Contains(GinkgoT(), output, tc.expectedContent, "Expected stdout to contain expected content for "+tc.shell)
				} else {
					// For other shells, check for some completion-like content
					assert.True(GinkgoT(),
						assert.Contains(GinkgoT(), output, "completion") ||
							assert.Contains(GinkgoT(), output, "jpd"),
						"Expected stdout to contain completion content for "+tc.shell)
				}

				// Verify NO files were created in the directory
				files, err := os.ReadDir(tempDir)
				assert.NoError(GinkgoT(), err, "Should be able to read temp directory")
				assert.Empty(GinkgoT(), files, "Expected no files to be created when using stdout")
			})
		}
	})

	Describe("With Shorthand Flag", func() {
		testCases := []struct {
			shell               string
			expectedBaseContent string
			expectedAliasMarker string
		}{
			{"bash", "bash completion", "# jpd shorthand aliases"},
			{"zsh", "zsh completion", "# jpd shorthand aliases"},
			{"fish", "fish completion", "# jpd shorthand aliases"},
			{"nushell", "export extern \"jpd\"", "# jpd shorthand aliases"},
			{"powershell", "PowerShell completion", "# jpd shorthand aliases"},
		}

		for _, tc := range testCases {
			tc := tc // capture loop variable
			It("should include alias block when --with-shorthand is used for "+tc.shell, func() {
				completionCmd := cmdPkg.NewCompletionCmd()

				// Test shorthand version for bash, long version for others
				if tc.shell == "bash" {
					completionCmd.SetArgs([]string{tc.shell, "-w"})
				} else {
					completionCmd.SetArgs([]string{tc.shell, "--with-shorthand"})
				}

				var buf bytes.Buffer
				completionCmd.SetOut(&buf)

				err := completionCmd.Execute()
				assert.NoError(GinkgoT(), err, "Command execution should succeed for "+tc.shell+" with --with-shorthand")

				output := buf.String()
				assert.NotEmpty(GinkgoT(), output, "Expected completion output to stdout for "+tc.shell+" with --with-shorthand")

				// Verify base completion content is still present
				if tc.shell == "nushell" {
					assert.Contains(GinkgoT(), output, tc.expectedBaseContent, "Expected stdout to contain base completion content for "+tc.shell)
				} else {
					assert.True(GinkgoT(),
						assert.Contains(GinkgoT(), output, "completion") ||
							assert.Contains(GinkgoT(), output, "jpd"),
						"Expected stdout to contain base completion content for "+tc.shell)
				}

				// Verify alias block marker is present
				assert.Contains(GinkgoT(), output, tc.expectedAliasMarker, "Expected stdout to contain alias block marker for "+tc.shell)

				// Verify some expected alias function signatures based on shell type
				switch tc.shell {
				case "bash":
					assert.Contains(GinkgoT(), output, "function jpi", "Expected bash output to contain 'function jpi' alias")
					assert.Contains(GinkgoT(), output, "complete -F __start_jpd jpi", "Expected bash output to contain completion wiring for jpi alias")
				case "zsh":
					assert.Contains(GinkgoT(), output, "jpi()", "Expected zsh output to contain 'jpi()' alias")
					assert.Contains(GinkgoT(), output, "compdef _jpd jpi", "Expected zsh output to contain completion wiring for jpi alias")
				case "fish":
					assert.Contains(GinkgoT(), output, "function jpi", "Expected fish output to contain 'function jpi' alias")
					assert.Contains(GinkgoT(), output, "complete -c jpi -w jpd", "Expected fish output to contain completion wiring for jpi alias")
				case "nushell":
					assert.Contains(GinkgoT(), output, "export extern \"jpi\"", "Expected nushell output to contain 'export extern \"jpi\"' alias")
					assert.Contains(GinkgoT(), output, "export def jpi", "Expected nushell output to contain 'export def jpi' alias")
				case "powershell":
					// PowerShell aliases not fully implemented yet, just check for the marker
					// Test passes with just the marker check above
				}

				// Verify NO files were created in the directory
				files, err := os.ReadDir(tempDir)
				assert.NoError(GinkgoT(), err, "Should be able to read temp directory")
				assert.Empty(GinkgoT(), files, "Expected no files to be created when using stdout with --with-shorthand")
			})
		}
	})

	Describe("Without Shorthand Flag", func() {
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

		for _, tc := range testCases {
			tc := tc // capture loop variable
			It("should NOT include alias block when --with-shorthand is NOT used for "+tc.shell, func() {
				completionCmd := cmdPkg.NewCompletionCmd()
				completionCmd.SetArgs([]string{tc.shell})

				var buf bytes.Buffer
				completionCmd.SetOut(&buf)

				err := completionCmd.Execute()
				assert.NoError(GinkgoT(), err, "Command execution should succeed for "+tc.shell+" without --with-shorthand")

				output := buf.String()
				assert.NotEmpty(GinkgoT(), output, "Expected completion output to stdout for "+tc.shell+" without --with-shorthand")

				// Verify base completion content is present
				if tc.shell == "nushell" {
					assert.Contains(GinkgoT(), output, tc.expectedContent, "Expected stdout to contain expected content for "+tc.shell)
				} else {
					assert.True(GinkgoT(),
						assert.Contains(GinkgoT(), output, "completion") ||
							assert.Contains(GinkgoT(), output, "jpd"),
						"Expected stdout to contain completion content for "+tc.shell)
				}

				// Verify alias block marker is NOT present
				assert.NotContains(GinkgoT(), output, "# jpd shorthand aliases", "Expected stdout to NOT contain alias block marker for "+tc.shell+" without --with-shorthand")

				// Verify alias functions are NOT present
				assert.NotContains(GinkgoT(), output, "function jpi", "Expected stdout to NOT contain 'function jpi' for "+tc.shell+" without --with-shorthand")
				assert.NotContains(GinkgoT(), output, "jpi()", "Expected stdout to NOT contain 'jpi()' for "+tc.shell+" without --with-shorthand")
				assert.NotContains(GinkgoT(), output, "export def jpi", "Expected stdout to NOT contain 'export def jpi' for "+tc.shell+" without --with-shorthand")
			})
		}
	})
})
