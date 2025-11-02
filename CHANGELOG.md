# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [3.0.0] - 2025-10-03

### ⚠️ BREAKING CHANGES

- **exec/dlx commands**: The `exec` and `dlx` commands now produce different outputs matching the correct package manager patterns. This affects the command structure for all package managers:
  - **exec (local dependencies)**: `npm exec <bin> -- <args>`, `pnpm exec <bin> <args>`, `yarn <bin> <args>`, `bun x <bin> <args>`, `deno run <script> <args>`
  - **dlx (temporary packages)**: `npm dlx <package> <args>`, `pnpm dlx <package> <args>`, `yarn dlx <package> <args>` (v2+), `bunx <package> <args>`, `deno run <url> <args>`
- **create command**: Deno create now expects URL as the first argument instead of using the `--url` flag

### Added

#### New Commands
- **create command**: Add new `create` command for scaffolding projects with package manager delegation
  - Support for npm, pnpm, yarn (v1/v2+), bun, and deno
  - Automatic `create-` prefix handling for npm/pnpm/yarn/bun
  - Direct URL support for deno scaffolding
  - `--search` flag for interactive package selection via npm registry
  - `--size` flag to control search results (default: 10)
  - `-s` alias for `--search` flag
  - Support for scoped packages without auto-prefixing `create-`
  - Optimized pnpm mapping using `pnpm dlx` for scaffolding

#### Run Command Enhancements
- **auto-install feature**: Add comprehensive auto-install functionality for `run` command
  - Automatic dependency installation for `dev` and `start` scripts
  - `--auto-install` flag with smart detection
  - `--if-present` flag for conditional script execution
  - `--no-volta` flag to disable Volta integration
  - Hash-based dependency change detection using SHA256
  - Missing package detection for Node.js package managers
  - Yarn PnP project detection to skip node_modules checks
  - Deno import accessibility checking via `deno info`
  - Dependency hash storage in `node_modules/.jpd-deps-hash`
  - Triggers on: missing node_modules, missing packages, unresolvable imports, dependency changes
  - Enhanced debug logging for all dependency checks

#### Testing & Coverage
- Raise test coverage to 80% across the project
  - Add comprehensive test coverage for `cmd` package
  - Add cross-platform completion tests (bash/zsh/fish/powershell/nushell)
  - Add integration tests for Carapace
  - Add create command tests with error handling
  - Add auto-install flow tests with debug log assertions
  - Add dependency management unit tests
  - Platform-aware tests for Windows (cmd lookup, home dir, invalid paths)
  - OS-aware CI path validation tests

#### Integrations
- Add `create` command to Warp workflows integration
- Add `create` command to Carapace completion spec
- Add `create` and `dlx` commands to Nushell extern completions
- Add run command flags (`--if-present`, `--auto-install`, `--no-volta`) to Carapace spec

#### Services
- Add npm registry service for package search
  - `SearchPackages` method for general package search
  - `SearchCreateApps` method for create-app specific searches
  - Configurable HTTP client and base URL for testing
  - Comprehensive error handling and timeout configuration

### Changed

#### Refactoring
- Consolidate command mapping tests into main test file
- Convert standalone alias tests to Ginkgo format
- Move dependency extraction helpers to `internal/deps/extract.go`
- Add SHA256 hashing utilities for Node.js and Deno dependencies
- Use pointer receivers for mutating `Run` methods in UI components
- Consolidate integration tests into main test files
- Remove test-mode path from create.go for cleaner architecture

#### CI/CD Improvements
- Run CI on `develop` branch and PRs to develop
- Add test matrix for Ubuntu, Windows, and macOS
- Add docs build job using Astro with pnpm and Node 20
- Upload per-OS coverage artifacts
- Set coverage threshold to 80% (later adjusted to 70%)
- Install Ginkgo CLI at module version to avoid mismatch
- Update golangci-lint config to v2 structure
- Ensure pnpm availability via Corepack in docs job

#### Documentation
- Add create command to README and API documentation
- Document create command in mental model and getting started guide
- Add create flags guidance and npm separator note
- Document run command auto-install feature and flags
- Update PR template with repo-specific checklist
- Improve formatting and spacing in WARP.md
- Update examples to reflect pnpm dlx usage

### Fixed

#### Bug Fixes
- **exec/dlx**: Fix command execution and delegation logic
- **create**: Enable search without arguments using `--search` flag
- **create**: Resolve generic type references and argument passing
- **create**: Fix npm `--` separator normalization
- **deps**: Stop `WriteStoredDepsHash` from creating node_modules directory
- **custom-flags**: Accept POSIX folder paths on Windows in CI mode
- **lint**: Address staticcheck SA1006 and remove unused test helper
- **testing**: Update expected workflow file count in integrate test
- **testing**: Resolve generic type issues in cmd tests

#### Test Improvements
- Make detect/internal tests platform-aware on Windows
- Fix broken merge in standard cmd tests
- Convert standard tests to Ginkgo and remove empty stubs
- Add fish shorthands and arg-count error tests
- Replace broad auto-install expectations with specific assertions
- Route auto-install debug logs via MockDebugExecutor in tests

### Removed
- Delete unused create selector files
- Remove duplicate create_test.go file (consolidated into cmd_test suite)
- Remove complex test-mode conditional handling from createAppSelector
- Remove obsolete standalone test files

## [2.0.1] - 2025-09-01

### Bug Fixes
- **test(cmd)**: Fix all test structure and import issues
  - Remove Gomega import from cmd_suite_test.go (violates project policy)
  - Fix undefined goEnvKey in integrate_warp_test.go by defining the type
  - Remove unused contextKey type
  - Remove unnecessary RegisterFailHandler call (not required for Ginkgo v2)
  - Ensure all test files follow project standards: Ginkgo v2 + Testify only
  - All 18 identified bugs are now fixed
  - 0 syntax/typecheck errors from golangci-lint
  - 458 specs passing across all test suites
  - Full compliance with project test policy

### Documentation
- **docs**: Update documentation to reflect jpd 2.0.0
  - JPD was released in Winget! Updated docs for v2.0.0 release

### Build System
- **build(goreleaser)**: winget base branch to master
  - Configure goreleaser for proper winget integration

### Chores
- **chore**: Add pull request template and configure goreleaser
  - Add pull request template to guide contributors
  - Configure goreleaser to create pull requests to update winget-pkgs repository
  - Enable automated manifest updates

## [2.0.0] - Previous Release
- Major version release with significant improvements and features
- Full CLI functionality for JavaScript package manager delegation
