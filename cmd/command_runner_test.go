package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandRunner_Run(t *testing.T) {
	t.Run("returns error when no command is set", func(t *testing.T) {
		runner := newCommandRunner(exec.Command)

		err := runner.Run()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no command set to run")
	})
}

func TestCommandRunner_Command(t *testing.T) {
	t.Run("sets stdio and persists target directory", func(t *testing.T) {
		var capturedCmd *exec.Cmd
		stubExec := func(name string, args ...string) *exec.Cmd {
			capturedCmd = exec.Command(name, args...)
			return capturedCmd
		}

		makeTempDir(t, func(tempDir string) {
			runner := newCommandRunner(stubExec)

			// Set target directory first
			err := runner.SetTargetDir(tempDir)
			assert.NoError(t, err)

			// Execute Command
			runner.Command("echo", "hello")

			// Verify stdio is set correctly
			assert.Equal(t, os.Stdin, capturedCmd.Stdin, "Stdin should be set to os.Stdin")
			assert.Equal(t, os.Stdout, capturedCmd.Stdout, "Stdout should be set to os.Stdout")
			assert.Equal(t, os.Stderr, capturedCmd.Stderr, "Stderr should be set to os.Stderr")

			// Verify target directory is persisted
			assert.Equal(t, tempDir, capturedCmd.Dir, "Directory should be set to target directory")
		})
	})

	t.Run("works without target directory set", func(t *testing.T) {
		var capturedCmd *exec.Cmd
		stubExec := func(name string, args ...string) *exec.Cmd {
			capturedCmd = exec.Command(name, args...)
			return capturedCmd
		}

		runner := newCommandRunner(stubExec)
		runner.Command("echo", "hello")

		// Verify stdio is set correctly
		assert.Equal(t, os.Stdin, capturedCmd.Stdin)
		assert.Equal(t, os.Stdout, capturedCmd.Stdout)
		assert.Equal(t, os.Stderr, capturedCmd.Stderr)

		// Directory should be empty (default behavior)
		assert.Empty(t, capturedCmd.Dir)
	})
}

func TestCommandRunner_SetTargetDir(t *testing.T) {
	t.Run("validates directory exists", func(t *testing.T) {
		runner := newCommandRunner(exec.Command)

		err := runner.SetTargetDir("/non/existent/path")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no such file or directory")
	})

	t.Run("validates path is a directory", func(t *testing.T) {
		makeTempDir(t, func(tempDir string) {
			// Create a temporary file
			tempFile := filepath.Join(tempDir, "testfile.txt")
			err := os.WriteFile(tempFile, []byte("test"), 0644)
			assert.NoError(t, err)

			runner := newCommandRunner(exec.Command)

			err = runner.SetTargetDir(tempFile)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "is not a directory")
		})
	})

	t.Run("updates existing command directory", func(t *testing.T) {
		var capturedCmd *exec.Cmd
		stubExec := func(name string, args ...string) *exec.Cmd {
			capturedCmd = exec.Command(name, args...)
			return capturedCmd
		}

		makeTempDir(t, func(tempDir string) {
			runner := newCommandRunner(stubExec)

			// First create a command
			runner.Command("echo", "hello")
			assert.Empty(t, capturedCmd.Dir, "Directory should be empty initially")

			// Then set target directory
			err := runner.SetTargetDir(tempDir)
			assert.NoError(t, err)

			// Verify the command's directory was updated
			assert.Equal(t, tempDir, capturedCmd.Dir, "Existing command directory should be updated")
		})
	})

	t.Run("accepts valid directory", func(t *testing.T) {
		makeTempDir(t, func(tempDir string) {
			runner := newCommandRunner(exec.Command)

			err := runner.SetTargetDir(tempDir)

			assert.NoError(t, err)
		})
	})
}

func TestCommandRunner_RunSuccess(t *testing.T) {
	t.Run("executes successful command", func(t *testing.T) {
		if !execAvailable("true") {
			t.Skip("'true' command not available")
		}

		runner := newCommandRunner(exec.Command)
		runner.Command("true") // 'true' always succeeds

		err := runner.Run()

		assert.NoError(t, err)
	})

	t.Run("executes failing command", func(t *testing.T) {
		if !execAvailable("false") {
			t.Skip("'false' command not available")
		}

		runner := newCommandRunner(exec.Command)
		runner.Command("false") // 'false' always fails

		err := runner.Run()

		assert.Error(t, err)
	})
}
