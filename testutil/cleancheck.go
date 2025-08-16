// Package testutil provides utilities for testing.
package testutil

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

// TestingT is an interface that matches the subset of testing.T methods we need.
// This allows for easier testing of the test helpers themselves.
type TestingT interface {
	Helper()
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Fatal(args ...interface{})
}

// WorkingTreeSnapshot represents the state of the git working tree at a point in time.
type WorkingTreeSnapshot struct {
	files map[string]bool
}

// SnapshotWorkingTree captures the current state of the git working tree.
// It runs 'git status --porcelain' and stores the list of untracked/modified files.
// Files matched by .gitignore are automatically excluded from the snapshot.
func SnapshotWorkingTree() (*WorkingTreeSnapshot, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	snapshot := &WorkingTreeSnapshot{
		files: make(map[string]bool),
	}

	// Parse git status output
	// Format: XY filename
	// Where X and Y are modification status codes
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		// Extract filename (skip status codes and space)
		if len(line) > 3 {
			filename := strings.TrimSpace(line[3:])
			snapshot.files[filename] = true
		}
	}

	return snapshot, nil
}

// AssertWorkingTreeClean checks if new files have appeared in the working tree
// since the snapshot was taken. It fails the test if any new untracked or
// modified files are detected. Files ignored by .gitignore are automatically
// excluded from the check.
func AssertWorkingTreeClean(t TestingT, snapshot *WorkingTreeSnapshot) {
	t.Helper()

	if snapshot == nil {
		t.Fatal("snapshot is nil")
	}

	currentSnapshot, err := SnapshotWorkingTree()
	if err != nil {
		t.Fatalf("failed to get current working tree state: %v", err)
	}

	// Check for new files that weren't in the original snapshot
	var newFiles []string
	for file := range currentSnapshot.files {
		if !snapshot.files[file] {
			newFiles = append(newFiles, file)
		}
	}

	if len(newFiles) > 0 {
		t.Errorf("test left artifacts in working tree:\n%s", strings.Join(newFiles, "\n"))
	}
}

// CleanupWorkingTree is a helper that can be used with t.Cleanup() to ensure
// the working tree stays clean after a test. Usage:
//
//	func TestSomething(t *testing.T) {
//	    CleanupWorkingTree(t)
//	    // ... test code that might create files ...
//	}
func CleanupWorkingTree(t *testing.T) {
	t.Helper()

	snapshot, err := SnapshotWorkingTree()
	if err != nil {
		t.Fatalf("failed to snapshot working tree: %v", err)
	}

	t.Cleanup(func() {
		AssertWorkingTreeClean(t, snapshot)
	})
}
