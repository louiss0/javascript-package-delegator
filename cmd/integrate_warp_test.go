package cmd_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"github.com/louiss0/javascript-package-delegator/cmd"
	"github.com/louiss0/javascript-package-delegator/env"
)

// Define a custom type for context key to avoid collisions
type contextKey string

const goEnvKey contextKey = "go_env"

var _ = Describe("Integrate Warp Command", func() {
	Describe("NewIntegrateWarpCmd", func() {
		Context("default behavior (no --output-dir)", func() {
			It("should install workflows to default directory and not print to stdout", func() {
				// Save original env and create temp XDG dir
				originalXDG := os.Getenv("XDG_DATA_HOME")
				defer func() { _ = os.Setenv("XDG_DATA_HOME", originalXDG) }()

				tmpXdg, err := os.MkdirTemp("", "xdg-warp-test-*")
				assert.NoError(GinkgoT(), err)
				defer func() { _ = os.RemoveAll(tmpXdg) }()

				// Set XDG_DATA_HOME
				err = os.Setenv("XDG_DATA_HOME", tmpXdg)
				assert.NoError(GinkgoT(), err)

				// Create warp command
				warpCmd := cmd.NewIntegrateWarpCmd()

				// Set up buffers to capture output
				outBuf := new(bytes.Buffer)
				errBuf := new(bytes.Buffer)
				warpCmd.SetOut(outBuf)
				warpCmd.SetErr(errBuf)

				// Add GoEnv to context (required by the command)
				goEnv := env.NewGoEnv()
				ctx := context.WithValue(context.Background(), goEnvKey, goEnv)
				// Also set the untyped string key used by the command context
				ctx = context.WithValue(ctx, "go_env", goEnv) // nolint:staticcheck // command expects string key
				warpCmd.SetContext(ctx)

				// Execute command with no args (default behavior)
				err = warpCmd.Execute()
				assert.NoError(GinkgoT(), err)

				// Assert no stdout output
				assert.Equal(GinkgoT(), 0, outBuf.Len(), "Should not print to stdout")
				assert.Equal(GinkgoT(), 0, errBuf.Len(), "Should not print to stderr")

				// Check that files were created in the default directory
				workflowsDir := filepath.Join(tmpXdg, "warp-terminal", "workflows")
				expectedFiles := []string{
					"jpd-install.yaml",
					"jpd-run.yaml",
					"jpd-exec.yaml",
					"jpd-dlx.yaml",
					"jpd-update.yaml",
					"jpd-uninstall.yaml",
					"jpd-clean-install.yaml",
					"jpd-agent.yaml",
				}

				for _, filename := range expectedFiles {
					filePath := filepath.Join(workflowsDir, filename)
					_, err := os.Stat(filePath)
					assert.NoError(GinkgoT(), err, "File should exist: %s", filename)
				}
			})
		})

		Context("with --output-dir flag", func() {
			It("should install workflows to custom directory and not print to stdout", func() {
				// Create temp output directory
				tmpOut, err := os.MkdirTemp("", "warp-output-test-*")
				assert.NoError(GinkgoT(), err)
				defer func() { _ = os.RemoveAll(tmpOut) }()

				// Create warp command
				warpCmd := cmd.NewIntegrateWarpCmd()

				// Set up buffers to capture output
				outBuf := new(bytes.Buffer)
				errBuf := new(bytes.Buffer)
				warpCmd.SetOut(outBuf)
				warpCmd.SetErr(errBuf)

				// Add GoEnv to context
				goEnv := env.NewGoEnv()
				ctx := context.WithValue(context.Background(), goEnvKey, goEnv)
				// Also set the untyped string key used by the command context
				ctx = context.WithValue(ctx, "go_env", goEnv) // nolint:staticcheck // command expects string key
				warpCmd.SetContext(ctx)

				// Set arguments for custom output directory (add trailing slash for validation)
				warpCmd.SetArgs([]string{"--output-dir", tmpOut + "/"})

				// Execute command
				err = warpCmd.Execute()
				assert.NoError(GinkgoT(), err)

				// Assert no stdout output
				assert.Equal(GinkgoT(), 0, outBuf.Len(), "Should not print to stdout")
				assert.Equal(GinkgoT(), 0, errBuf.Len(), "Should not print to stderr")

				// Check that files were created in the custom directory
				expectedFiles := []string{
					"jpd-install.yaml",
					"jpd-run.yaml",
					"jpd-exec.yaml",
					"jpd-dlx.yaml",
					"jpd-update.yaml",
					"jpd-uninstall.yaml",
					"jpd-clean-install.yaml",
					"jpd-agent.yaml",
				}

				for _, filename := range expectedFiles {
					filePath := filepath.Join(tmpOut, filename)
					_, err := os.Stat(filePath)
					assert.NoError(GinkgoT(), err, "File should exist: %s", filename)
				}
			})
		})

		Context("with nested custom directory", func() {
			It("should create nested directories and install workflows", func() {
				// Create temp base directory
				tmpBase, err := os.MkdirTemp("", "warp-nested-test-*")
				assert.NoError(GinkgoT(), err)
				defer func() { _ = os.RemoveAll(tmpBase) }()

				// Specify nested directory that doesn't exist yet
				nestedDir := filepath.Join(tmpBase, "nested", "path", "workflows")

				// Create warp command
				warpCmd := cmd.NewIntegrateWarpCmd()

				// Set up buffers
				outBuf := new(bytes.Buffer)
				errBuf := new(bytes.Buffer)
				warpCmd.SetOut(outBuf)
				warpCmd.SetErr(errBuf)

				// Add GoEnv to context
				goEnv := env.NewGoEnv()
				ctx := context.WithValue(context.Background(), goEnvKey, goEnv)
				// Also set the untyped string key used by the command context
				ctx = context.WithValue(ctx, "go_env", goEnv) // nolint:staticcheck // command expects string key
				warpCmd.SetContext(ctx)

				// Set arguments for nested directory (add trailing slash for validation)
				warpCmd.SetArgs([]string{"--output-dir", nestedDir + "/"})

				// Execute command
				err = warpCmd.Execute()
				assert.NoError(GinkgoT(), err)

				// Assert no stdout output
				assert.Equal(GinkgoT(), 0, outBuf.Len(), "Should not print to stdout")

				// Check that nested directory was created
				stat, err := os.Stat(nestedDir)
				assert.NoError(GinkgoT(), err)
				assert.True(GinkgoT(), stat.IsDir())

				// Check that files exist
				files, err := os.ReadDir(nestedDir)
				assert.NoError(GinkgoT(), err)
				assert.Len(GinkgoT(), files, 8, "Should have 8 workflow files")
			})
		})
	})
})
