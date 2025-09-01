# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
