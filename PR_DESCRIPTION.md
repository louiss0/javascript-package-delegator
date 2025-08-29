# Fix Warp Integration Default Directory Installation

## Summary
This PR fixes the `jpd integrate warp` command to properly write workflow files to the default Warp terminal directory when no `--output-dir` flag is provided, and removes the stdout printing behavior entirely.

## Problem
Previously, the `integrate warp` command had several issues:
1. Did not write to the default Warp workflows directory when no output flag was provided
2. Incorrectly printed multi-doc YAML to stdout
3. Generated workflow files lacked the `command` field

## Solution
### Core Changes
- **Default Directory Installation**: Command now writes to `${XDG_DATA_HOME:-$HOME/.local/share}/warp-terminal/workflows` by default
- **Removed Stdout Printing**: All workflow files are now written to disk only (never to stdout)
- **Proper Flag Validation**: Integrated `NewFolderPathFlag` for proper directory path validation

### Implementation Details

#### Files Modified
- `internal/warp.go`:
  - Added `DefaultWarpWorkflowsDir()` helper function to resolve default Warp workflows directory
  - Properly handles XDG_DATA_HOME environment variable with fallback

- `cmd/integrate.go`:
  - Modified `runWarpIntegration()` to accept `output_dir` parameter directly
  - Uses `FolderPathFlag` for automatic path validation
  - Fixed logic to handle both default and custom directories properly
  - Removed duplicate `GenerateJPDWorkflows` call

- `cmd/cmd_test.go`:
  - Updated test to expect file creation instead of stdout output
  - Test now verifies files are created in the default directory

#### Tests Added
- `cmd/integrate_warp_test.go`: Comprehensive test suite covering:
  - Default directory installation
  - Custom directory installation
  - Nested directory creation
  - Verification that no stdout output occurs

- `internal/internal_test.go`: Tests for `DefaultWarpWorkflowsDir()` function:
  - With XDG_DATA_HOME set
  - Without XDG_DATA_HOME (fallback to home directory)

## Testing
- ✅ All unit tests pass
- ✅ Integration tests verified with manual testing
- ✅ Coverage maintained at 77% for cmd package
- ✅ No linting issues (golangci-lint)
- ✅ Code formatted with `go fmt`

## Breaking Changes
⚠️ **Breaking Change**: The `integrate warp` command no longer prints to stdout. It always writes files to disk.
- Previous behavior: Without `--output-dir`, printed multi-doc YAML to stdout
- New behavior: Without `--output-dir`, writes to default Warp workflows directory

## Usage Examples

### Install to default directory
```bash
jpd integrate warp
# Files written to: ${XDG_DATA_HOME:-$HOME/.local/share}/warp-terminal/workflows/
```

### Install to custom directory
```bash
jpd integrate warp --output-dir ./my-workflows/
# Files written to: ./my-workflows/
```

## Checklist
- [x] Code follows project conventions
- [x] Tests added/updated
- [x] Documentation updated (help text)
- [x] No lint warnings
- [x] Tests pass locally
- [x] Breaking changes documented

## Related Issues
Fixes issues with Warp integration not installing to the proper default directory.
