# CreateAppSelector Mock Usage

The `CreateAppSelector` in the `cmd` package has been enhanced with test mode capabilities to avoid interactive UI during testing.

## Overview

The mock works by detecting when tests are running and switching to non-interactive mode. Instead of showing a `huh.Select` UI, it returns pre-configured values.

## Test Setup

The mock is automatically enabled in tests through the `BeforeEach/AfterEach` setup in `cmd_test.go`:

```go
BeforeEach(func() {
    // Enable test mode to avoid interactive UI
    cmd.SetTestMode(true)
    // Reset test behavior to defaults
    cmd.ResetCreateAppSelectorTestBehavior()
    // ... other setup
})

AfterEach(func() {
    // Disable test mode after each test
    cmd.SetTestMode(false)
    // ... other cleanup
})
```

## Controlling Mock Behavior

### Default Behavior
By default, the mock will return the first package name from the provided package list.

### Custom Behavior
You can control what the mock returns using these functions:

```go
// Make the mock return a specific value
cmd.SetCreateAppSelectorTestBehavior("my-selected-package", false, "")

// Make the mock return an error
cmd.SetCreateAppSelectorTestBehavior("", true, "custom error message")

// Reset to default behavior
cmd.ResetCreateAppSelectorTestBehavior()
```

## Example Usage in Tests

### Test with Default Behavior
```go
It("should use first package by default", func() {
    // The mock will automatically return the first package from packageInfo
    _, err := executeCmd(rootCmd, "create", "--search")
    assert.NoError(err)
})
```

### Test with Custom Selection
```go
It("should use configured package selection", func() {
    // Configure mock to return specific package
    cmd.SetCreateAppSelectorTestBehavior("react-app", false, "")
    
    _, err := executeCmd(rootCmd, "create", "--search") 
    assert.NoError(err)
    // Test will use "react-app" as the selected package
})
```

### Test with Error Behavior
```go
It("should handle selection errors", func() {
    // Configure mock to return error
    cmd.SetCreateAppSelectorTestBehavior("", true, "package selection failed")
    
    _, err := executeCmd(rootCmd, "create", "--search")
    assert.Error(err)
    assert.Contains(err.Error(), "package selection failed")
})
```

## Technical Details

### Type Constraints
The mock satisfies Go's strict `~struct{packageInfo []services.PackageInfo}` type constraint by using the exact same underlying struct type as the real implementation.

### Test Detection
Test mode is enabled/disabled through a global variable `testModeEnabled` in the `cmd` package. This approach:
- ✅ Works with Go's strict type constraints
- ✅ Avoids interactive UI during testing
- ✅ Allows fine-grained control of mock behavior
- ✅ Maintains type compatibility with generics

### Production Safety
The test mode flag is only enabled during test execution. Production code will always use the interactive UI unless explicitly disabled.

## Files Modified
- `cmd/create.go`: Added test mode detection and behavior controls
- `cmd/cmd_test.go`: Added test mode setup in BeforeEach/AfterEach
- `mock/pkg.go`: Contains legacy mock code (can be cleaned up)
