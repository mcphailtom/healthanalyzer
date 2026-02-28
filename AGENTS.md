# Agent Guidelines

Conventions and standards for working in this codebase.

## Stack

- **Language**: Go
- **Persistence**: SQLite (with sqlite-vec for vector search)
- **Web UI**: HTMX with Go `html/template`
- **TUI**: Bubbletea / Lipgloss
- **Build**: CGo required (SQLite driver)

## Code Style

- Write succinct, idiomatic Go. Favour clarity over cleverness.
- Minimise external dependencies. Prefer the standard library where practical.
- Use documentation comments where they add value, not as a substitute for readable code. If a comment restates what the code already says, remove it.
- Keep functions short and focused. If a function needs a comment explaining its flow, it should probably be split.

## Project Structure

- All application code lives under `internal/`.
- CLI entrypoints live under `cmd/`.
- Documentation lives in `docs/` and should be short, well-scoped to a single feature or concern, and kept up to date with implementation.

## Testing

- Use `testify` for assertions and suite constructs.
- Use `mockery` for generating mocks where needed.
- Prefer table-driven tests for functions with multiple input/output cases.
- Tests exist to verify requirements. If a test fails, fix the code to satisfy the requirement -- do not weaken the test to accommodate an edge case.
- All code should have test coverage where practical.

## Workflow

- All feature work occurs on a dedicated branch with a pull request to `main`.
- Commits should be atomic and descriptive. Favour smaller, well-scoped PRs over large changesets.
- Ensure the project compiles and passes `go vet` before committing.
