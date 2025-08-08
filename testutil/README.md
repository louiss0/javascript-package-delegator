# Test Utilities

This package provides utilities for testing, including helpers to detect when tests leave artifacts in the working tree.

## Clean Check Helper

The `cleancheck.go` file provides functions to ensure tests don't leave artifacts in the git working tree. This helps maintain a clean repository and catch tests that don't properly clean up after themselves.

### Functions

#### `SnapshotWorkingTree() (*WorkingTreeSnapshot, error)`
Captures the current state of the git working tree by running `git status --porcelain`. Files matched by `.gitignore` are automatically excluded.

#### `AssertWorkingTreeClean(t TestingT, snapshot *WorkingTreeSnapshot)`
Compares the current working tree state against a previously taken snapshot and fails the test if any new files have appeared.

#### `CleanupWorkingTree(t *testing.T)`
A convenience helper that automatically takes a snapshot and sets up a cleanup function to check for artifacts when the test completes.

### Usage Examples

#### 1. Using in TestMain (Package-level enforcement)

```go
func TestMain(m *testing.M) {
    // Take snapshot before running tests
    snapshot, err := testutil.SnapshotWorkingTree()
    if err != nil {
        panic("failed to snapshot working tree: " + err.Error())
    }
    
    // Run all tests
    code := m.Run()
    
    // Check that no artifacts were left behind
    t := &testing.T{}
    testutil.AssertWorkingTreeClean(t, snapshot)
    
    os.Exit(code)
}
```

#### 2. Using CleanupWorkingTree helper (Recommended for individual tests)

```go
func TestSomething(t *testing.T) {
    // This automatically checks for artifacts when the test completes
    testutil.CleanupWorkingTree(t)
    
    // Your test code here...
    // Any files created that aren't cleaned up will fail the test
}
```

#### 3. Manual snapshot and assert (For more control)

```go
func TestSomething(t *testing.T) {
    snapshot, err := testutil.SnapshotWorkingTree()
    if err != nil {
        t.Fatalf("failed to snapshot working tree: %v", err)
    }
    
    defer func() {
        testutil.AssertWorkingTreeClean(t, snapshot)
    }()
    
    // Your test code here...
}
```

### Benefits

- **Early Detection**: Catches tests that create files but don't clean them up
- **Clean Repository**: Prevents test artifacts from accumulating in the repository
- **Git-aware**: Automatically ignores files that are in `.gitignore`
- **Flexible**: Can be used at package level or individual test level

### Notes

- The helper uses `git status --porcelain` to detect changes
- Files matched by `.gitignore` patterns are automatically excluded from checks
- Go build artifacts (*.test, *.out, etc.) should be added to `.gitignore` to avoid false positives
