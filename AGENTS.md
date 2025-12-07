# Repository Guidelines

## Project Structure & Module Organization
- `main.go` wires the Cobra CLI; command entrypoints live in `cmd/` (one file per command) and should stay thin.
- `pkg/actions/` holds orchestration for user-facing commands, while `pkg/obsidian/` contains vault/file/tag primitives and `pkg/config/` manages paths and defaults. Keep new logic in these layers rather than `cmd/`.
- `pkg/mcp/` exposes the CLI as an MCP server; mirror any new capabilities here when relevant.
- Keep MCP tool registrations/descriptions in `pkg/mcp/register.go` in sync with CLI flags (e.g., new options must be declared so clients see them). Bias toward minimal, high-signal options for MCP; prefer sensible defaults over adding knobs.
- Keep the agent guide resource in `pkg/mcp/resources.go` up to date when adding or changing tools; this is the primary reference for AI agents using the MCP server. Document only the pared-down MCP options (skipAnchors/skipEmbeds, includeTags, recencyCascade, and a handful of limits), not the full CLI flag set.
- When adding user-facing behavior, update both README.md and MCP docs (`pkg/mcp/register.go` descriptions, `pkg/mcp/resources.go` agent guide) so agents and humans stay aligned.
- Supporting assets live in `docs/` (images, manual excerpts); build outputs land in `bin/` (do not commit generated binaries). Tests sit next to sources with `_test.go`; shared fixtures and doubles can be placed under `mocks/`.

## Build, Test, and Development Commands
- `go fmt ./...` before changes to keep gofmt-consistent output.
- `go build ./...` for a quick sanity compilation; `make build-all` cross-compiles into `bin/{darwin,linux,windows}/obsidian-cli`.
- `make test` (or `go test ./...`) runs the full suite; `make test_coverage` emits `coverage.out` for `go tool cover -func=coverage.out`.
- Try CLI changes locally with `go run . --help` or `go run . <command> --vault "<vault-name>"` against a test vault.

## Coding Style & Naming Conventions
- Go 1.23+ code; rely on `gofmt` defaults (tabs, grouped imports, trailing newlines). Keep exports documented and prefer clear, Go-idiomatic names (`CamelCase` types, lowerCamel vars, package-scope funcs only when needed).
- Keep `cmd/` focused on flag parsing; push business logic into `pkg/actions/` and `pkg/obsidian/` for reuse and testability.
- Follow existing file naming (`create.go`, `tags.go`, etc.) and align with Cobra command verbs.

## Testing Guidelines
- Favor table-driven tests and assertions via `stretchr/testify`. Add `_test.go` beside the implementation and target packages with `go test ./pkg/<area> -run TestName`.
- Cover new edge cases for vault paths, tag parsing, and file operations; avoid depending on personal vault data and use in-repo fixtures or temp dirs.
- Generate coverage with `make test_coverage` when adding sizable features; keep flaky filesystem timing to a minimum.

## Commit & Pull Request Guidelines
- Write imperative, concise commits similar to recent history (`Optimize tag search and support wildcards…`). Squash locally if you create noisy WIP commits.
- PRs should describe behavior changes, reference issues when applicable, and note any new flags or defaults. Include `go test ./...` results and update `README.md`/`docs/` when user-facing behavior shifts.
- Avoid committing built artifacts in `bin/` or local config files; keep diffs limited to source, docs, and necessary test data.

## Recent Changes
- 002-rename-note-backlinks: Added Go 1.23+ + Cobra CLI, Viper/config where applicable, obsidian vault/link parsing helpers in `pkg/obsidian`, MCP tooling in `pkg/mcp`
- 001-backlink-support: Added Go 1.23+ + Cobra CLI, Viper/config where used, internal packages `pkg/obsidian`, `pkg/actions`, `pkg/mcp`, `pkg/config`
- 001-backlink-support: Added [if applicable, e.g., PostgreSQL, CoreData, files or N/A]

## Active Technologies
- Go 1.23+ + Cobra CLI, Viper/config where applicable, obsidian vault/link parsing helpers in `pkg/obsidian`, MCP tooling in `pkg/mcp` (002-rename-note-backlinks)
- Local filesystem vaults (Markdown files); git repository when presen (002-rename-note-backlinks)
