# Repository Guidelines

These rules mirror the WARP guide so every tool and editor behaves the same.

## Core Practices (Common Strengths)
- **TDD first**: Red → Green → Refactor using Ginkgo + Testify.
- **Commit before push**: make a Conventional Commit for each logical change; **never push without a commit**.
- **Formatting & linting are mandatory**: format → imports → lint (in that order).
- **No scripts or makefiles**: use the plain commands below.

## Functional Utilities (prefer `samber/lo` when it clarifies intent)
Agents should consider `github.com/samber/lo` to reduce boilerplate and make transformations obvious.

```go
import "github.com/samber/lo"

ids := lo.Map(pkgs, func(p Package, _ int) string { return p.ID })
valid := lo.Filter(pkgs, func(p Package, _ int) bool { return p.Valid })
```

Use `lo` where it **improves clarity**; avoid over-chaining if a short loop reads better.

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
Reviewers should ensure:
- **Go Coding**
  - Idiomatic names, small focused funcs, clear `%w` error wrapping.
  - Deterministic tests, explicit types, no unnecessary globals.
  - Encourage readable use of `samber/lo` where it clarifies intent.
- **CLI Creation**
  - Cobra commands modular, clear help text, stable exit codes.
- **Google Style (subset)**
  - Package comments, full-sentence doc comments.
  - Comments focus on intent, not restating code.

**Exceptions & Overrides**
If a file declares its own local style directive (e.g., `// style:override <reason>`), follow it for that scope.
