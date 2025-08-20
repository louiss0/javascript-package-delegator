package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestRunWarpIntegration_StdoutMode(t *testing.T) {
	t.Run("outputs to stdout when no output-dir flag", func(t *testing.T) {
		cmd := &cobra.Command{}
		// Don't define the flag, which will trigger the first branch (flag not set)
		
		output := captureStdout(t, func() {
			err := runWarpIntegration(cmd)
			assert.NoError(t, err)
		})

		assert.NotEmpty(t, output, "Should output workflows to stdout")
		// Basic validation that it looks like YAML output
		assert.Contains(t, output, "---", "Output should contain YAML document separator")
	})

	t.Run("outputs to stdout when output-dir flag is empty", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("output-dir", "", "output directory")
		
		output := captureStdout(t, func() {
			err := runWarpIntegration(cmd)
			assert.NoError(t, err)
		})

		assert.NotEmpty(t, output, "Should output workflows to stdout")
		assert.Contains(t, output, "---", "Output should contain YAML document separator")
	})
}

func TestRunWarpIntegration_DirectoryMode(t *testing.T) {
	t.Run("generates files in specified directory", func(t *testing.T) {
		makeTempDir(t, func(tempDir string) {
			cmd := &cobra.Command{}
			cmd.Flags().String("output-dir", tempDir, "output directory")

			// Capture any output
			var output bytes.Buffer
			originalStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			go func() {
				defer w.Close()
				err := runWarpIntegration(cmd)
				assert.NoError(t, err)
			}()

			// Read from pipe
			buf := make([]byte, 1024)
			n, _ := r.Read(buf)
			output.Write(buf[:n])
			r.Close()
			os.Stdout = originalStdout

			// Check that files were created
			files, err := os.ReadDir(tempDir)
			assert.NoError(t, err)
			assert.Greater(t, len(files), 0, "Should create at least one file")

			// Verify files have .yaml extension
			yamlFileCount := 0
			for _, file := range files {
				if strings.HasSuffix(file.Name(), ".yaml") {
					yamlFileCount++
				}
			}
			assert.Greater(t, yamlFileCount, 0, "Should create at least one .yaml file")

			// Check that success message was printed
			outputStr := output.String()
			assert.Contains(t, outputStr, "Generated Warp workflow files in:", "Should print success message")
		})
	})

	t.Run("returns error when output directory creation fails", func(t *testing.T) {
		makeTempDir(t, func(tempDir string) {
			// Create a file with the same name as our intended directory
			filePath := filepath.Join(tempDir, "notadir")
			err := os.WriteFile(filePath, []byte("content"), 0644)
			assert.NoError(t, err)

			cmd := &cobra.Command{}
			cmd.Flags().String("output-dir", filePath, "output directory")

			err = runWarpIntegration(cmd)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "failed to generate Warp workflow files")
		})
	})
}

func TestNewIntegrateWarpCmd(t *testing.T) {
	t.Run("creates command with correct flags", func(t *testing.T) {
		cmd := NewIntegrateWarpCmd()

		assert.Equal(t, "warp", cmd.Use)
		assert.Contains(t, cmd.Short, "Warp")
		assert.NotEmpty(t, cmd.Long)

		// Check that the output-dir flag exists
		flag := cmd.Flag("output-dir")
		assert.NotNil(t, flag, "Should have output-dir flag")
		assert.Equal(t, "o", flag.Shorthand, "Should have -o shorthand")
	})

	t.Run("help runs without error", func(t *testing.T) {
		cmd := NewIntegrateWarpCmd()
		
		// Set help flag and execute
		cmd.SetArgs([]string{"--help"})
		
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		
		_ = cmd.Execute()
		// Help should exit with error (this is normal cobra behavior)
		// but the output should contain help text
		output := buf.String()
		assert.Contains(t, output, "Warp", "Help output should mention Warp")
		assert.Contains(t, output, "output-dir", "Help should describe output-dir flag")
	})
}
