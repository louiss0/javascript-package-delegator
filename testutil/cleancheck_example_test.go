package testutil_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/louiss0/javascript-package-delegator/testutil"
)

// Example of using cleancheck in TestMain to enforce clean working tree
// for all tests in a package.
func TestMain(m *testing.M) {
	// Take snapshot before running tests
	snapshot, err := testutil.SnapshotWorkingTree()
	if err != nil {
		panic("failed to snapshot working tree: " + err.Error())
	}

	// Run all tests
	code := m.Run()

	// Check that no artifacts were left behind
	// This will fail if any test left files in the working tree
	t := &testing.T{}
	testutil.AssertWorkingTreeClean(t, snapshot)

	os.Exit(code)
}

// Example of using cleancheck in an individual test
func TestExample_WithCleanupHelper(t *testing.T) {
	// This automatically checks for artifacts when the test completes
	testutil.CleanupWorkingTree(t)

	// Test code here...
	// Any files created during the test that aren't cleaned up
	// will cause the test to fail
}

// Example of manual snapshot/assert pattern for more control
func TestExample_ManualSnapshot(t *testing.T) {
	// Take snapshot at the beginning
	snapshot, err := testutil.SnapshotWorkingTree()
	if err != nil {
		t.Fatalf("failed to snapshot working tree: %v", err)
	}

	// Defer the cleanup check
	defer func() {
		testutil.AssertWorkingTreeClean(t, snapshot)
	}()

	// Test code here...
	// Any files created during the test that aren't cleaned up
	// will cause the test to fail
}

// Test that verifies the artifact detection actually works
func TestCleanCheck_DetectsArtifacts(t *testing.T) {
	// Save current working directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	// Change to parent directory (project root)
	if err := os.Chdir(".."); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer os.Chdir(origDir)

	// Take initial snapshot
	snapshot, err := testutil.SnapshotWorkingTree()
	if err != nil {
		t.Fatalf("failed to snapshot working tree: %v", err)
	}

	// Create a temporary test file (artifact) in the project root
	// Using .artifact extension which is not in .gitignore
	testFile := "test_artifact.artifact"
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Clean up the file after test
	defer os.Remove(testFile)

	// Create a custom testing.T to capture the error without failing this test
	mockT := &mockTestingT{}
	testutil.AssertWorkingTreeClean(mockT, snapshot)

	// Verify that the check detected the artifact
	if !mockT.errored {
		t.Error("AssertWorkingTreeClean should have detected the test artifact")
	}
	if !contains(mockT.errors, "test_artifact.artifact") {
		t.Errorf("expected error message to contain %q, got: %v", "test_artifact.artifact", mockT.errors)
	}
}

// mockTestingT is a mock implementation of testing.T for testing purposes
type mockTestingT struct {
	failed  bool
	errored bool
	errors  []string
}

func (m *mockTestingT) Helper() {}

func (m *mockTestingT) Fatal(args ...interface{}) {
	m.failed = true
	m.errors = append(m.errors, fmt.Sprint(args...))
}

func (m *mockTestingT) Fatalf(format string, args ...interface{}) {
	m.failed = true
	m.errors = append(m.errors, fmt.Sprintf(format, args...))
}

func (m *mockTestingT) Errorf(format string, args ...interface{}) {
	m.errored = true
	m.errors = append(m.errors, fmt.Sprintf(format, args...))
}

// contains checks if any string in the slice contains the substring
func contains(strs []string, substr string) bool {
	for _, s := range strs {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}
