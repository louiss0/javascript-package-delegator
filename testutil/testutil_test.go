package testutil_test

import (
	"testing"

	"github.com/louiss0/javascript-package-delegator/testutil"
)

// TestSnapshotWorkingTree tests the SnapshotWorkingTree function.
func TestSnapshotWorkingTree(t *testing.T) {
	snapshot, err := testutil.SnapshotWorkingTree()
	if err != nil {
		t.Fatalf("SnapshotWorkingTree() error = %v", err)
	}

	if snapshot == nil {
		t.Fatal("Expected SnapshotWorkingTree() to return a non-nil snapshot")
	}
}

// TestAssertWorkingTreeClean tests the AssertWorkingTreeClean function.
func TestAssertWorkingTreeClean(t *testing.T) {
	snapshot, err := testutil.SnapshotWorkingTree()
	if err != nil {
		t.Fatalf("SnapshotWorkingTree() error = %v", err)
	}

	testutil.AssertWorkingTreeClean(t, snapshot)
}

// TestCleanupWorkingTree tests the CleanupWorkingTree helper function.
func TestCleanupWorkingTree(t *testing.T) {
	testutil.CleanupWorkingTree(t)
	// Normally, additional test code would go here that might create files.
}
