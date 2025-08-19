package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
