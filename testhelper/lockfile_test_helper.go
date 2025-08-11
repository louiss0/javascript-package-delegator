package testhelper

import (
	"os"
	"path/filepath"
	"testing"
)

// LockfileTestContext provides a clean test environment for lockfile-related tests
type LockfileTestContext struct {
	t           testing.TB
	tempDir     string
	originalCwd string
	cleanup     []func()
}

// NewLockfileTestContext creates a new test context with automatic cleanup
func NewLockfileTestContext(t testing.TB) *LockfileTestContext {
	ctx := &LockfileTestContext{
		t:       t,
		cleanup: []func(){},
	}

	// Store original working directory
	originalCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	ctx.originalCwd = originalCwd

	// Create temp directory
	ctx.tempDir = t.TempDir()

	// Register cleanup to restore working directory
	t.Cleanup(func() {
		// Execute all cleanup functions in reverse order
		for i := len(ctx.cleanup) - 1; i >= 0; i-- {
			ctx.cleanup[i]()
		}

		// Always try to restore original working directory
		if ctx.originalCwd != "" {
			if err := os.Chdir(ctx.originalCwd); err != nil {
				t.Logf("Warning: Failed to restore original working directory: %v", err)
			}
		}
	})

	return ctx
}

// TempDir returns the temporary directory path
func (ctx *LockfileTestContext) TempDir() string {
	return ctx.tempDir
}

// ChangeToTempDir changes the working directory to the temp directory
func (ctx *LockfileTestContext) ChangeToTempDir() {
	if err := os.Chdir(ctx.tempDir); err != nil {
		ctx.t.Fatalf("Failed to change to temp directory: %v", err)
	}
}

// CreateLockfile creates a lockfile in the temp directory
func (ctx *LockfileTestContext) CreateLockfile(filename string, content []byte) string {
	filepath := filepath.Join(ctx.tempDir, filename)
	if err := os.WriteFile(filepath, content, 0644); err != nil {
		ctx.t.Fatalf("Failed to create lockfile %s: %v", filename, err)
	}
	return filepath
}

// CreatePackageJSON creates a package.json file in the temp directory
func (ctx *LockfileTestContext) CreatePackageJSON(content string) string {
	return ctx.CreateLockfile("package.json", []byte(content))
}

// CreateDenoJSON creates a deno.json file in the temp directory
func (ctx *LockfileTestContext) CreateDenoJSON(content string) string {
	return ctx.CreateLockfile("deno.json", []byte(content))
}

// CreateYarnLock creates a yarn.lock file in the temp directory
func (ctx *LockfileTestContext) CreateYarnLock(content string) string {
	return ctx.CreateLockfile("yarn.lock", []byte(content))
}

// CreatePnpmLock creates a pnpm-lock.yaml file in the temp directory
func (ctx *LockfileTestContext) CreatePnpmLock(content string) string {
	return ctx.CreateLockfile("pnpm-lock.yaml", []byte(content))
}

// CreatePackageLock creates a package-lock.json file in the temp directory
func (ctx *LockfileTestContext) CreatePackageLock(content string) string {
	return ctx.CreateLockfile("package-lock.json", []byte(content))
}

// AddCleanup adds a cleanup function to be executed at the end of the test
func (ctx *LockfileTestContext) AddCleanup(fn func()) {
	ctx.cleanup = append(ctx.cleanup, fn)
}

// RestoreWorkingDir restores the original working directory
func (ctx *LockfileTestContext) RestoreWorkingDir() {
	if ctx.originalCwd != "" {
		if err := os.Chdir(ctx.originalCwd); err != nil {
			ctx.t.Logf("Warning: Failed to restore working directory: %v", err)
		}
	}
}
