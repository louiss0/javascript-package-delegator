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

	// Check that the command returns the success line with the full path
	output := buf.String()
	// Account for macOS adding /private prefix to TMP paths when rendering Abs paths
	prefix := "Completion script created at: "
	expected1 := fmt.Sprintf("%s%s", prefix, fullPath)
	expected2 := fmt.Sprintf("%s%s%s", prefix, "/private", fullPath)
	if !strings.Contains(output, expected1) && !strings.Contains(output, expected2) {
		t.Fatalf("Success message missing. Expected one of: %s OR %s, Got: %s", expected1, expected2, output)
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

func TestCompletionCommand_DefaultFilenames(t *testing.T) {
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

	// Test default filename generation (no --output flag)
	testCases := []struct {
		shell    string
		expected string
	}{
		{"bash", "jpd_bash_completion.bash"},
		{"zsh", "jpd_zsh_completion.zsh"},
		{"fish", "jpd_fish_completion.fish"},
		{"nushell", "jpd_nushell_completion.nu"},
		{"powershell", "jpd_powershell_completion.ps1"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("default_%s", tc.shell), func(t *testing.T) {
			// Create completion command
			completionCmd := NewCompletionCmd()

			// Set up the command without output flag (should use default filename)
			completionCmd.SetArgs([]string{tc.shell})

			// Execute the command
			err = completionCmd.Execute()
			if err != nil {
				t.Fatalf("Command execution failed for %s: %v", tc.shell, err)
			}

			// Check that the file exists with default name
			fullPath := filepath.Join(tempDir, tc.expected)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				t.Fatalf("Default output file does not exist at expected path for %s: %s", tc.shell, fullPath)
			}

			// Success message with full path
			t.Logf("SUCCESS: %s completion script generated successfully at %s with default filename", tc.shell, fullPath)
		})
	}
}
