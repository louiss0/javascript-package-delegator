# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Project Overview

JavaScript Package Delegator (jpd) is a universal CLI that automatically detects JavaScript package managers (npm, yarn, pnpm, bun, deno) and delegates commands to the appropriate tool. Built with Go and inspired by @antfu/ni, it provides a unified interface across different package managers.

## Quick Start Commands

### Testing
```bash
# Run all tests (BDD-style with Ginkgo)
ginkgo run

# Watch mode for TDD development
ginkgo watch

# Run tests for specific package
ginkgo run ./cmd
ginkgo run ./detect

# Generate coverage report (80% minimum threshold)
go test ./... -race -coverprofile=coverage.out

# Run tests with CI build flags (relaxed --cwd validation)
go test ./... -race -coverprofile=coverage.out -ldflags "-X github.com/louiss0/javascript-package-delegator/build_info.rawCI=true"

# Check coverage threshold with automated cleanup
./scripts/check-coverage.sh

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html
```

### Building
```bash
# Local development build
go build -o jpd ./main.go

### Code Quality
```bash
# Lint with golangci-lint
golangci-lint run

# Format code (required before commits)
gofmt -s -w .
goimports -w .

# Verify Go modules are tidy
go mod tidy
go mod verify

# Run all checks manually
go vet ./...
```

### Cleanup
```bash
# Clean up test artifacts and binaries
./scripts/cleanup.sh
```

### Documentation Site
```bash
# Start documentation development server
cd docs && pnpm dev
# Or using jpd itself: jpd run --cwd ./docs/ dev

# Build documentation site
cd docs && pnpm build

# Preview built documentation
cd docs && pnpm preview
```

## Architecture Overview

### Core Design Principles
- **Modular Architecture**: Clear separation of concerns across packages
- **Dependency Injection**: Testable design using the `Dependencies` struct in `cmd/root.go`
- **Interface-Based Abstractions**: `FileSystem`, `PathLookup`, `CommandRunner` interfaces enable comprehensive mocking
- **Command Pattern**: Cobra framework with consistent command structure
- **Context-Based Configuration**: Configuration and dependencies passed through command context

### Key Architectural Components

**Command Execution Flow:**
1. Root command (`cmd/root.go`) initializes dependencies and detects package manager
2. Subcommands receive dependencies via context
3. Commands use injected interfaces for file operations and command execution
4. Real implementations use actual system calls, test implementations use mocks

**Package Manager Detection:**
1. Lock file detection (`detect/DetectLockfile`) scans for package manager lock files
2. Path-based detection (`detect/DetectJSPackageManager`) checks PATH for package managers
3. Fallback to interactive installation prompt if nothing is found

## Package Structure

| Package | Responsibility |
|---------|---------------|
| **cmd/** | Command implementations using Cobra framework. Each command (install, run, exec, etc.) has its own file with `New{Command}Cmd()` function |
| **detect/** | Package manager detection logic. Contains interfaces `FileSystem` and `PathLookup` for testability |
| **build_info/** | Build-time configuration and version management. Handles ldflags injection for CI/releases |
| **custom_flags/** | Custom Cobra flags including path validation for `--cwd` flag |
| **custom_errors/** | Named, reusable error types for consistent error handling |
| **env/** | Environment utilities (`GoEnv`) for development/production mode detection |
| **services/** | Business logic and external integrations (npm registry service) |
| **internal/integrations/** | Generators for Warp workflows and Carapace completion specs |
| **mock/** | Mock implementations for testing (excluded from coverage) |
| **testutil/** | Test utilities and helpers (excluded from coverage) |

## Testing Strategy

### BDD Approach with Ginkgo
- **Test Runner**: Ginkgo v2 for BDD-style testing
- **Assertions**: Testify assertions
- **Test Organization**: Each package has a `*_suite_test.go` file that sets up the Ginkgo test suite
- **Parallel Execution**: Ginkgo runs tests in parallel by default - be mindful of shared state

### Mocking Strategy
The codebase uses interface-based mocking for external dependencies:

```go
// Production implementations
type RealFileSystem struct{}
type RealPathLookup struct{}
type commandRunner struct{}

// Interfaces for mocking
type FileSystem interface {
    Stat(name string) (os.FileInfo, error)
    Getwd() (string, error)
}

type PathLookup interface {
    LookPath(file string) (string, error)
}

type CommandRunner interface {
    Command(string, ...string)
    Run() error
    SetTargetDir(string) error
}
```
They are implemented using the `githhub.com/testify/mock` package.

### Coverage Requirements
- **Minimum**: 80% test coverage enforced in CI
- **Excluded Packages**: `build_info`, `mock`, `testutil` (infrastructure code)
- **Enforcement**: CI fails automatically if coverage drops below threshold

## Build Configuration

### Build Flags and Version Injection
The project uses Go's `-ldflags` to inject build-time information:

```bash
# CI build flag (relaxes --cwd path validation)
-X github.com/louiss0/javascript-package-delegator/build_info.rawCI=true

# Version injection (set by GoReleaser)
-X github.com/louiss0/javascript-package-delegator/build_info.rawCLI_VERSION=v1.0.0

# Environment mode (production for releases)
-X github.com/louiss0/javascript-package-delegator/build_info.rawGO_MODE=production

# Build date (RFC3339 format, converted to YYYY-MM-DD)
-X github.com/louiss0/javascript-package-delegator/build_info.rawBUILD_DATE=2024-01-01T00:00:00Z
```

### CI vs Local Behavior
- **CI Mode**: `--cwd` flag accepts both `path` and `path/` (relaxed validation)
- **Local Mode**: `--cwd` flag requires trailing `/` except for root `/`
- **Control**: Set via build flag only, no environment detection

### GoReleaser Integration
- **Cross-platform builds**: Linux, macOS, Windows for amd64 and arm64
- **Package managers**: Homebrew, Scoop, Winget, Nix
- **Automatic releases**: Triggered on Git tags with conventional commit changelog

## Common Workflows

### Adding a New Command
1. Create new file in `cmd/` directory: `cmd/mycommand.go`
2. Implement `NewMyCommandCmd() *cobra.Command` function
3. Register command in `cmd/root.go`: `cmd.AddCommand(NewMyCommandCmd())`
4. Add tests in `cmd/mycommand_test.go`
5. Update completion if needed

### Development Workflow
```bash
# Start TDD workflow
ginkgo watch

# Make changes, tests run automatically
# When tests pass, check coverage
./scripts/check-coverage.sh

# Format code before committing
gofmt -s -w .
goimports -w .

# Clean up artifacts
./scripts/cleanup.sh
```

### Debugging
```bash
# Run with debug output
jpd --debug install
jpd --debug --cwd ./subproject/ run dev

# Test CI behavior locally
go test ./... -ldflags "-X github.com/louiss0/javascript-package-delegator/build_info.rawCI=true"
```


## Code Coverage

### Requirements
- **Minimum Threshold**: 80% coverage enforced in CI
- **Excluded Packages**: `build_info`, `mock`, `testutil`
- **Failure Behavior**: CI automatically fails if below threshold

### Coverage Commands
```bash
# Generate coverage report
go test $(go list ./... | grep -v -E '/(build_info|mock|testutil)$') -coverprofile=coverage.out

# Check threshold (automated script)
./scripts/check-coverage.sh 80

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# View coverage by function
go tool cover -func=coverage.out
```

## Project Conventions

### File Naming
- **Command files**: `cmd/{command}.go` with `New{Command}Cmd()` function
- **Test files**: `{package}_test.go` or `{specific}_test.go`
- **Suite files**: `{package}_suite_test.go` for Ginkgo setup

### Error Handling
- Use custom error types from `custom_errors` package
- Wrap errors with context using `fmt.Errorf`
- Return meaningful error messages for CLI users

### Logging
- Use `github.com/charmbracelet/log` for structured logging
- Debug level controlled by `--debug` flag
- Production mode logging controlled by `GO_MODE` build variable

### Dependencies
- **CLI Framework**: `spf13/cobra`
- **Testing**: `onsi/ginkgo` + `testify/assert`
- **UI**: `charmbracelet/huh` for interactive prompts
- **Utilities**: `samber/lo` for functional helpers

This document should be updated as the project evolves to reflect current practices and architecture.
