package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/louiss0/javascript-package-delegator/internal/integrations"
)

func TestWriteToFile(t *testing.T) {
	t.Run("writes content to file", func(t *testing.T) {
		makeTempDir(t, func(tempDir string) {
			filePath := filepath.Join(tempDir, "test.txt")
			content := "test content\nline 2"

			err := writeToFile(filePath, content)
			assert.NoError(t, err)

			// Verify file exists and has correct content
			fileContent, err := os.ReadFile(filePath)
			assert.NoError(t, err)
			assert.Equal(t, content, string(fileContent))
		})
	})

	t.Run("returns error when trying to write to directory", func(t *testing.T) {
		makeTempDir(t, func(tempDir string) {
			// Try to write to the directory path itself
			err := writeToFile(tempDir, "content")
			assert.Error(t, err)
			// Error should be about permissions or that it's a directory
		})
	})

	t.Run("overwrites existing file", func(t *testing.T) {
		makeTempDir(t, func(tempDir string) {
			filePath := filepath.Join(tempDir, "test.txt")

			// Write initial content
			err := writeToFile(filePath, "initial")
			assert.NoError(t, err)

			// Overwrite with new content
			newContent := "overwritten content"
			err = writeToFile(filePath, newContent)
			assert.NoError(t, err)

			// Verify new content
			fileContent, err := os.ReadFile(filePath)
			assert.NoError(t, err)
			assert.Equal(t, newContent, string(fileContent))
		})
	})
}

func TestRunCarapaceIntegration_OutputFileMode(t *testing.T) {
	t.Run("writes spec to custom output file", func(t *testing.T) {
		makeTempDir(t, func(tempDir string) {
			outputFile := filepath.Join(tempDir, "custom-spec.yaml")

			// Create a minimal root command for spec generation
			rootCmd := &cobra.Command{
				Use:   "jpd",
				Short: "Test command",
			}

			cmd := &cobra.Command{}
			rootCmd.AddCommand(cmd)
			cmd.Flags().String("output", outputFile, "output file")

			// Capture stdout to verify success message
			var output bytes.Buffer
			originalStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			go func() {
				defer func() { _ = w.Close() }()
				err := runCarapaceIntegration(cmd)
				assert.NoError(t, err)
			}()

			// Read from pipe
			buf := make([]byte, 1024)
			n, _ := r.Read(buf)
			output.Write(buf[:n])
			_ = r.Close()
			os.Stdout = originalStdout

			// Verify file was created
			assert.FileExists(t, outputFile)

			// Verify file is not empty
			stat, err := os.Stat(outputFile)
			assert.NoError(t, err)
			assert.Greater(t, stat.Size(), int64(0), "File should not be empty")

			// Verify success message was printed
			outputStr := output.String()
			assert.Contains(t, outputStr, "Generated Carapace spec file:", "Should print success message")
		})
	})
}

func TestRunCarapaceIntegration_StdoutMode(t *testing.T) {
	t.Run("outputs spec to stdout", func(t *testing.T) {
		// Create a minimal root command for spec generation
		rootCmd := &cobra.Command{
			Use:   "jpd",
			Short: "Test command",
		}

		cmd := &cobra.Command{}
		rootCmd.AddCommand(cmd)
		cmd.Flags().Bool("stdout", true, "output to stdout")

		output := captureStdout(t, func() {
			err := runCarapaceIntegration(cmd)
			assert.NoError(t, err)
		})

		assert.NotEmpty(t, output, "Should output spec to stdout")
		// Basic validation that it looks like YAML
		assert.Contains(t, output, "name:", "Output should contain YAML spec")
	})
}

func TestRunCarapaceIntegration_DefaultGlobalMode(t *testing.T) {
	t.Run("installs spec to global location", func(t *testing.T) {
		makeTempDir(t, func(tempDir string) {
			// Set XDG_DATA_HOME to control where the spec gets installed
			withEnv(t, "XDG_DATA_HOME", tempDir, func() {
				// Create a minimal root command for spec generation
				rootCmd := &cobra.Command{
					Use:   "jpd",
					Short: "Test command",
				}

				cmd := &cobra.Command{}
				rootCmd.AddCommand(cmd)
				// No flags set, should trigger default global install

				// Capture stdout to verify success message
				var output bytes.Buffer
				originalStdout := os.Stdout
				r, w, _ := os.Pipe()
				os.Stdout = w

				go func() {
					defer func() { _ = w.Close() }()
					err := runCarapaceIntegration(cmd)
					assert.NoError(t, err)
				}()

				// Read from pipe
				buf := make([]byte, 1024)
				n, _ := r.Read(buf)
				output.Write(buf[:n])
				_ = r.Close()
				os.Stdout = originalStdout

				// Verify the spec file was created in the expected location
				expectedPath, err := integrations.DefaultCarapaceSpecPath()
				assert.NoError(t, err)
				assert.FileExists(t, expectedPath)

				// Verify file is not empty
				stat, err := os.Stat(expectedPath)
				assert.NoError(t, err)
				assert.Greater(t, stat.Size(), int64(0), "File should not be empty")

				// Verify success message was printed
				outputStr := output.String()
				assert.Contains(t, outputStr, "Installed Carapace spec:", "Should print success message")
			})
		})
	})
}

func TestNewIntegrateCarapaceCmd(t *testing.T) {
	t.Run("creates command with correct flags", func(t *testing.T) {
		cmd := NewIntegrateCarapaceCmd()

		assert.Equal(t, "carapace", cmd.Use)
		assert.Contains(t, cmd.Short, "Carapace")
		assert.NotEmpty(t, cmd.Long)

		// Check that the output flag exists
		outputFlag := cmd.Flag("output")
		assert.NotNil(t, outputFlag, "Should have output flag")
		assert.Equal(t, "o", outputFlag.Shorthand, "Should have -o shorthand")

		// Check that the stdout flag exists
		stdoutFlag := cmd.Flag("stdout")
		assert.NotNil(t, stdoutFlag, "Should have stdout flag")
	})

	t.Run("help runs without error", func(t *testing.T) {
		cmd := NewIntegrateCarapaceCmd()

		// Set help flag and execute
		cmd.SetArgs([]string{"--help"})

		var buf bytes.Buffer
		cmd.SetOut(&buf)

		_ = cmd.Execute()
		// Help should exit with error (this is normal cobra behavior)
		// but the output should contain help text
		output := buf.String()
		assert.Contains(t, output, "Carapace", "Help output should mention Carapace")
		assert.Contains(t, output, "output", "Help should describe output flag")
		assert.Contains(t, output, "stdout", "Help should describe stdout flag")
	})
}
