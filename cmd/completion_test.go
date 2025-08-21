package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompletionCommand_OutputFlag(t *testing.T) {
	// Use t.TempDir() for automatic cleanup
	tempDir := t.TempDir()

	// Set HOME environment variable to temp dir
	originalHome := os.Getenv("HOME")
	defer func() {
		if originalHome != "" {
			_ = os.Setenv("HOME", originalHome)
		} else {
			_ = os.Unsetenv("HOME")
		}
	}()
	t.Setenv("HOME", tempDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(originalDir) })

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create completion command
	completionCmd := NewCompletionCmd()

	// Set up the command with output flag
	outputFile := "test_completion.bash"
	completionCmd.SetArgs([]string{"bash", "--output", outputFile})

	// Capture the command output
	var buf bytes.Buffer
	completionCmd.SetOut(&buf)

	// Execute the command
	err = completionCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	// Check that the file exists
	fullPath := filepath.Join(tempDir, outputFile)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Fatalf("Output file does not exist at expected path: %s", fullPath)
	}

	// Check that the file is not empty
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		t.Fatalf("Failed to get file info: %v", err)
	}
	if fileInfo.Size() == 0 {
		t.Fatalf("Output file is empty: %s", fullPath)
	}

	// Verify the completion script was generated properly by checking file contents
	content, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("Failed to read completion file: %v", err)
	}

	// Check that the completion script contains expected bash completion content
	contentStr := string(content)
	if !strings.Contains(contentStr, "bash completion") && !strings.Contains(contentStr, "_jpd_completion") {
		t.Fatalf("Generated completion file does not contain expected bash completion content")
	}

	// Check that no output was written to stdout when using --output flag
	output := buf.String()
	if output != "" {
		t.Fatalf("Expected no output to stdout when using --output flag, but got: %s", output)
	}

	// The success condition is that the file exists at the full path and contains valid completion content
	// AND the command returns the success line with the full path
	t.Logf("SUCCESS: Completion script generated successfully at %s", fullPath)
}

func TestCompletionCommand_MultipleShells(t *testing.T) {
	// Use t.TempDir() for automatic cleanup
	tempDir := t.TempDir()

	// Set HOME environment variable to temp dir
	t.Setenv("HOME", tempDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(originalDir) })

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Test different shell types
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
		t.Run(fmt.Sprintf("shell_%s", tc.shell), func(t *testing.T) {
			// Create completion command
			completionCmd := NewCompletionCmd()

			// Set up the command with output flag
			completionCmd.SetArgs([]string{tc.shell, "--output", tc.outputFile})

			// Execute the command
			err = completionCmd.Execute()
			if err != nil {
				t.Fatalf("Command execution failed for %s: %v", tc.shell, err)
			}

			// Check that the file exists
			fullPath := filepath.Join(tempDir, tc.outputFile)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				t.Fatalf("Output file does not exist at expected path for %s: %s", tc.shell, fullPath)
			}

			// Check that the file is not empty
			fileInfo, err := os.Stat(fullPath)
			if err != nil {
				t.Fatalf("Failed to get file info for %s: %v", tc.shell, err)
			}
			if fileInfo.Size() == 0 {
				t.Fatalf("Output file is empty for %s: %s", tc.shell, fullPath)
			}

			// Verify the completion script was generated
			content, err := os.ReadFile(fullPath)
			if err != nil {
				t.Fatalf("Failed to read completion file for %s: %v", tc.shell, err)
			}

			// For nushell, check the embedded content
			if tc.shell == "nushell" {
				contentStr := string(content)
				if !strings.Contains(contentStr, "export extern \"jpd\"") {
					t.Fatalf("Generated nushell completion file does not contain expected content")
				}
			}

			// Success message with full path
			t.Logf("SUCCESS: %s completion script generated successfully at %s", tc.shell, fullPath)
		})
	}
}

func TestCompletionCommand_DefaultStdout(t *testing.T) {
	// Test that completion outputs to stdout by default (no --output flag)
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
		t.Run(fmt.Sprintf("stdout_%s", tc.shell), func(t *testing.T) {
			// Create completion command
			completionCmd := NewCompletionCmd()

			// Set up the command without output flag (should output to stdout)
			completionCmd.SetArgs([]string{tc.shell})

			// Capture the command output
			var buf bytes.Buffer
			completionCmd.SetOut(&buf)

			// Execute the command
			err := completionCmd.Execute()
			if err != nil {
				t.Fatalf("Command execution failed for %s: %v", tc.shell, err)
			}

			// Check that output was written to stdout
			output := buf.String()
			if output == "" {
				t.Fatalf("Expected completion script output to stdout for %s, but got empty output", tc.shell)
			}

			// For nushell, check for specific content since it uses embedded script
			if tc.shell == "nushell" {
				if !strings.Contains(output, tc.expectedContent) {
					t.Fatalf("Generated %s completion does not contain expected content. Expected to contain: %s", tc.shell, tc.expectedContent)
				}
			} else {
				// For other shells, just check it contains some completion-related content
				if !strings.Contains(output, "completion") {
					t.Fatalf("Generated %s completion does not contain 'completion' keyword", tc.shell)
				}
			}

			// Success message
			t.Logf("SUCCESS: %s completion script generated successfully to stdout", tc.shell)
		})
	}
}

func TestCompletionCommand_WithShorthandsFlag(t *testing.T) {
	testCases := []struct {
		shell             string
		expectedAliases   []string
		expectedFunctions []string
	}{
		{"bash", []string{"jpe", "jpx", "jpi"}, []string{"function jpe()", "function jpx()", "function jpi()"}},
		{"zsh", []string{"jpe", "jpx", "jpi"}, []string{"jpe() { jpd exec", "jpx() { jpd dlx", "jpi() { jpd install"}},
		{"fish", []string{"jpe", "jpx", "jpi"}, []string{"function jpe", "function jpx", "function jpi", "jpd exec $argv"}},
		{"nushell", []string{"jpe", "jpx", "jpi"}, []string{"export def jpe", "export def jpx", "export def jpi"}},
		{"powershell", []string{"jpe", "jpx", "jpi"}, []string{"function jpe {", "function jpx {", "function jpi {", "jpd exec @args"}},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("with_shorthands_%s", tc.shell), func(t *testing.T) {
			assert := assert.New(t)

			// Create completion command
			completionCmd := NewCompletionCmd()

			// Set up the command with --with-shorthands flag
			completionCmd.SetArgs([]string{tc.shell, "--with-shorthands"})

			// Capture the command output
			var buf bytes.Buffer
			completionCmd.SetOut(&buf)

			// Execute the command
			err := completionCmd.Execute()
			assert.NoError(err, "Command execution should succeed for %s", tc.shell)

			// Check that output was generated
			output := buf.String()
			assert.NotEmpty(output, "Expected output for %s with shorthands", tc.shell)

			// Check that shorthand functions are included
			for _, expectedFunc := range tc.expectedFunctions {
				assert.Contains(output, expectedFunc, "Expected %s to contain alias function: %s", tc.shell, expectedFunc)
			}

			t.Logf("SUCCESS: %s completion with shorthands generated successfully", tc.shell)
		})
	}
}

func TestCompletionCommand_ErrorCases(t *testing.T) {
	t.Run("unsupported_shell", func(t *testing.T) {
		assert := assert.New(t)

		// Create completion command
		completionCmd := NewCompletionCmd()

		// Set up the command with unsupported shell
		completionCmd.SetArgs([]string{"unsupported"})

		// Suppress error output to avoid logs in tests
		var errBuf bytes.Buffer
		completionCmd.SetErr(&errBuf)
		completionCmd.SetOut(&errBuf)

		// Execute the command
		err := completionCmd.Execute()
		assert.Error(err, "Expected error for unsupported shell")
		assert.Contains(err.Error(), "unsupported shell", "Error message should mention unsupported shell")
	})

	t.Run("no_arguments", func(t *testing.T) {
		assert := assert.New(t)

		// Create completion command
		completionCmd := NewCompletionCmd()

		// Set up the command with no arguments
		completionCmd.SetArgs([]string{})

		// Suppress error output to avoid logs in tests
		var errBuf bytes.Buffer
		completionCmd.SetErr(&errBuf)
		completionCmd.SetOut(&errBuf)

		// Execute the command
		err := completionCmd.Execute()
		assert.Error(err, "Expected error when no arguments provided")
		assert.Contains(err.Error(), "requires exactly one argument", "Error message should mention required argument")
	})

	t.Run("too_many_arguments", func(t *testing.T) {
		assert := assert.New(t)

		// Create completion command
		completionCmd := NewCompletionCmd()

		// Set up the command with too many arguments
		completionCmd.SetArgs([]string{"bash", "zsh"})

		// Suppress error output to avoid logs in tests
		var errBuf bytes.Buffer
		completionCmd.SetErr(&errBuf)
		completionCmd.SetOut(&errBuf)

		// Execute the command
		err := completionCmd.Execute()
		assert.Error(err, "Expected error when too many arguments provided")
		assert.Contains(err.Error(), "requires exactly one argument", "Error message should mention required argument")
	})

	t.Run("invalid_output_path", func(t *testing.T) {
		assert := assert.New(t)

		// Create completion command
		completionCmd := NewCompletionCmd()

		// Try to write to a path that cannot be created (permission denied scenario)
		// Use a path like /root/test which should fail due to permissions
		invalidPath := "/root/cannot-create/test_completion.sh"
		completionCmd.SetArgs([]string{"bash", "--output", invalidPath})

		// Suppress all command output including usage help
		var errBuf bytes.Buffer
		completionCmd.SetErr(&errBuf)
		completionCmd.SetOut(&errBuf)

		// Execute the command
		err := completionCmd.Execute()
		if err != nil {
			// This should fail due to permission denied, but the exact error depends on the system
			assert.Contains(err.Error(), "failed to create", "Error message should mention creation failure")
		} else {
			// If we're running as root or have unusual permissions, skip this test
			t.Skip("Skipping invalid output path test - running with elevated permissions")
		}
	})
}

func TestGetSupportedShells(t *testing.T) {
	assert := assert.New(t)

	supportedShells := getSupportedShells()

	// Check that we have the expected number of shells
	assert.Len(supportedShells, 5, "Should support exactly 5 shells")

	// Check that all expected shells are present
	expectedShells := []string{"bash", "fish", "nushell", "powershell", "zsh"}
	for _, expectedShell := range expectedShells {
		assert.Contains(supportedShells, expectedShell, "Should support %s", expectedShell)
	}

	// Verify the exact order and content
	assert.Equal(expectedShells, supportedShells, "Shells should be in the expected order")
}

func TestGetDefaultAliasMapping(t *testing.T) {
	assert := assert.New(t)

	aliasMap := getDefaultAliasMapping()

	// Check that all expected commands have aliases
	expectedCommands := []string{"install", "run", "exec", "dlx", "update", "uninstall", "clean-install", "agent"}
	for _, cmd := range expectedCommands {
		assert.Contains(aliasMap, cmd, "Should have aliases for %s command", cmd)
		assert.NotEmpty(aliasMap[cmd], "Should have at least one alias for %s command", cmd)
	}

	// Check specific aliases
	assert.Contains(aliasMap["exec"], "jpe", "exec command should have jpe alias")
	assert.Contains(aliasMap["dlx"], "jpx", "dlx command should have jpx alias")
	assert.Contains(aliasMap["install"], "jpi", "install command should have jpi alias")
	assert.Contains(aliasMap["run"], "jpr", "run command should have jpr alias")

	// Check that all aliases contain jpd- prefixed versions
	for cmd, aliases := range aliasMap {
		expectedJpdAlias := fmt.Sprintf("jpd-%s", cmd)
		assert.Contains(aliases, expectedJpdAlias, "Command %s should have jpd- prefixed alias %s", cmd, expectedJpdAlias)
	}
}
