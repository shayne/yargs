# Repository Guidelines

## Project Structure & Module Organization
- The module is a single Go package in the repository root (`go.mod`).
- Core library code lives in `yargs.go`; package docs live in `doc.go`.
- Tests are co-located in root-level `*_test.go` files (for example, `yargs_test.go`).
- Supporting files include `README.md` (usage), `LICENSE`, and `mise.toml` (tooling/tasks).

## Build, Test, and Development Commands
- `mise install`: installs the pinned Go toolchain defined in `mise.toml`.
- `mise run fmt`: formats all Go files via `gofmt`.
- `mise run tidy`: runs `go mod tidy` to keep dependencies clean.
- `mise run test`: runs the full test suite (`go test ./...`).
- `mise run check`: runs `fmt`, `tidy`, and `test` in sequence.

## Coding Style & Naming Conventions
- Follow standard Go formatting; use `gofmt` and tabs for indentation.
- Exported identifiers use `UpperCamelCase`; unexported use `lowerCamelCase`.
- Test files must end with `_test.go`; keep test names descriptive and focused on behavior.
- Keep public API changes documented with GoDoc comments and update `README.md` when behavior changes.

## Testing Guidelines
- Tests use Go’s standard `testing` package.
- Co-locate tests with production code in the root package.
- No explicit coverage target is defined; add tests for new parsing paths, error cases, and help output changes.
- Run tests with `mise run test` or `go test ./...`.

## Commit & Pull Request Guidelines
- Git history is minimal and uses short, scoped subjects (for example, `all: yargs`); there is no formal convention yet.
- Prefer concise, imperative commit subjects; optional scopes like `parse:` or `help:` are welcome.
- PRs should include a brief summary, tests run, and notes on any public API or CLI behavior changes.

## Tooling & Configuration Notes
- Go version is pinned in `mise.toml` (currently `1.24.10`).
- Keep `go.mod`/`go.sum` tidy; avoid adding dependencies unless necessary.
