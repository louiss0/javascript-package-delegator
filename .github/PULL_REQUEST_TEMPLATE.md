# Pull Request

Provide a clear, concise summary of the change. Keep titles imperative and short.

## Why
Explain the problem this PR solves or the motivation behind the change.

## What
Summarize the user-visible behavior or developer-facing changes.

## How
Briefly describe the approach taken, key decisions, and any alternatives considered.

## Breaking changes
- [ ] None
- If there are breaking changes, describe them and migration steps here.

## Screenshots / CLI output (optional)
Paste relevant terminal output or screenshots that help reviewers.

## Tests
- Describe what you tested and how to reproduce.
- Include any new/updated unit or integration tests.

## Checklist
- [ ] Commits follow the project's Conventional Commit rules (type(scope): subject; imperative, <= 64 chars)
- [ ] Changes are atomic, buildable, and tests pass locally (`go test ./...`)
- [ ] New/updated tests are included for new behavior
- [ ] Documentation updated where applicable
- [ ] CI passes after pushing
- [ ] Linked issue referenced, if applicable (e.g., Closes #123)
- [ ] Git Flow: feature branches target `develop` (use `git flow feature start <name>`)
