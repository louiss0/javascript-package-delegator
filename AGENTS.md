# Repository Guidelines

## Project Structure & Module Organization
The CLI targets Go 1.23 and keeps one Cobra command per file under `cmd/` (for example `cmd/run.go`, `cmd/clean-install.go`) with shared shell assets parked in `cmd/assets/`.
`detect/` owns package manager lookup logic, `env/` handles host capability checks, and cross cutting helpers live in `custom_flags/`, `custom_errors/`, and `build_info/`.
Integration generators live in `internal/integrations/`, reusable doubles sit in `testutil/`, docs land in `docs/`, and release or cleanup scripts such as `check-coverage.sh` live in `scripts/`.
Keep tests beside their packages and maintain the global suite wiring in `javascript-package-delegator_suite_test.go` so new specs stay discoverable.

## Build, Test & Development Commands
Use these commands while developing:
- `go build -ldflags "-X github.com/louiss0/javascript-package-delegator/build_info.rawCI=true" -o jpd ./...` builds the CLI with the CI specific feature flags baked in.
- `ginkgo run` executes the full behavior driven suite once, while `ginkgo watch` reruns impacted specs on file saves.
- `go test ./... -race -coverprofile=coverage.out -ldflags "-X github.com/louiss0/javascript-package-delegator/build_info.rawCI=true"` validates race safety, coverage, and CI only paths such as `--cwd` validation.
- `golangci-lint run ./...` (configured via `.golangci.yml`) aggregates `revive`, `staticcheck`, and other linters used in CI.
- `scripts/check-coverage.sh` enforces the coverage floor and scrubs stale binaries or profiles.

## Coding Style & Naming Conventions
Always run `gofmt` before committing; keep tabs for indentation, grouped imports, and describe exported items with concise doc comments.
Match file names to CLI verbs (`cmd/install.go` drives `jpd install`) and keep Cobra flags kebab case when declared through helpers in `custom_flags/`.
Prefer descriptive lower case package names (`detect`, `testutil`) and let `golangci-lint` be the source of truth for style breaks before you push.

## Testing Guidelines
Write new behavior using Ginkgo with Testify assertions; colocate `_test.go` files with the code they cover.
Lean on factories in `testutil/` instead of ad hoc mocks so CLI flows remain deterministic.
Run the race enabled `go test` command above plus `ginkgo watch` during TDD, and update `coverprofile.out` only when you intentionally regenerate it.
Make mocks using `testify/mock` when making a new struct base in on an interface that is written in the code that uses it.
Make sure that all dependencies are injected using the `Dependencies` struct.
Do a Red-Green-Refactor cycle to ensure your tests pass and your code is clean.


## Commit & Pull Request Guidelines
Commits follow Conventional Commit prefixes already in the log (`fix(ci): allow docs lockfile for CI caching`, `style: format code with gofmt`) and must bundle tests with the behavior change.
Pull requests should summarize the CLI scenario addressed, link related issues, include output from `go test` or `ginkgo run`, and call out doc or `goreleaser` updates so reviewers can reproduce them.
Attach terminal captures when UX changes, double check `docs/` and `README.adoc`, and note any toggles to `build_info.rawCI` semantics so release automation stays predictable.
