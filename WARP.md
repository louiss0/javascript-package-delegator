# WARP Guide

This repository uses a consistent, test-first workflow and clear commit practices across all tools.

## Project Overview
JavaScript Package Delegator (jpd) is a universal CLI that detects the active JS package manager (npm, yarn, pnpm, bun, deno) and delegates commands accordingly. It’s written in Go and inspired by @antfu/ni.

## Development Principles (Common Strengths)
- **TDD by default**: write/adjust a test, watch it fail, make it pass, then refactor.
- **Small, meaningful commits**: use Conventional Commits; **always make a `git commit` before any `git push`**.
- **Consistency over cleverness**: formatting, imports, and linting are mandatory.
- **No scripts or makefiles required**: every step has a plain command.

## Daily Flow
1) **Watch tests (TDD loop)**
```bash
ginkgo watch
```
2) **Run focused/one-off tests**
```bash
ginkgo -r -p -race -cover --skip-package build_info,test_util,mock
```
3) **Format, fix imports, lint (mandatory order)**
```bash
gofmt -w .
goimports -w .
golangci-lint run
```
4) **Commit before pushing**
```bash
git add -A
git commit -m "feat: explain TDD loop in WARP guide"
git push
```
5) **Clean up artifacts (manual)**
- POSIX: run the bash snippet in **Cleanup (no scripts)**
- Windows: run the PowerShell snippet in **Cleanup (no scripts)**

## Functional Utilities (prefer `samber/lo` where it clarifies intent)
When it improves readability and reduces boilerplate, prefer functional helpers from `github.com/samber/lo` over ad-hoc imperative loops.

```go
import "github.com/samber/lo"

names := lo.Map(users, func(u User, _ int) string { return u.Name })
adults := lo.Filter(users, func(u User, _ int) bool { return u.Age >= 18 })
first, ok := lo.Find(users, func(u User) bool { return u.ID == targetID })
```

**Guidelines**
- Use `lo` when it makes the code *clearer* (mapping/filtering/transforming).
- Keep hot paths readable; avoid chaining when a small loop is clearer.
- Always keep types explicit enough for reviewers to follow.

## Cleanup (no scripts)

Run these after tests to remove stray binaries and coverage artifacts—no scripts, no makefiles.

### POSIX shells (macOS/Linux)
```bash
# from repo root
[ -f javascript-package-delegator ] && rm -f javascript-package-delegator
[ -f jpd ] && rm -f jpd
[ -f coverage.out ] && rm -f coverage.out
[ -f coverage.html ] && rm -f coverage.html
find . -name "*.test" -type f -delete 2>/dev/null || true
go clean -testcache
echo "✅ Cleanup complete"
```

### Windows PowerShell
```powershell
# from repo root
if (Test-Path javascript-package-delegator) { Remove-Item javascript-package-delegator -Force }
if (Test-Path jpd) { Remove-Item jpd -Force }
if (Test-Path coverage.out) { Remove-Item coverage.out -Force }
if (Test-Path coverage.html) { Remove-Item coverage.html -Force }
Get-ChildItem -Recurse -Filter *.test | Remove-Item -Force
go clean -testcache
Write-Output "✅ Cleanup complete"
```

> These commands replace the old `cleanup.sh` entirely.

## Code Review Checklist (Go + CLI + Google-style subset)
Reviewers should verify:
- **Go Coding**
  - Idiomatic naming (`mixedCase` for exported, short receivers, no stutter).
  - Errors wrapped with `%w`, context used properly, no panic for flow control.
  - Deterministic tests, no hidden global state, clear logging.
- **CLI Creation**
  - One command per file, descriptive `Use`, `Short`, and `Long` fields.
  - Flags validated; consistent help output and stable exit codes.
- **Google Style (subset)**
  - Comments explain *why*, doc comments full sentences.
  - Explicitness preferred over cleverness.

**Exceptions & Overrides**
If a Go source file explicitly documents a local style (e.g., `// style:allow-lo-chaining`), that directive takes precedence.
