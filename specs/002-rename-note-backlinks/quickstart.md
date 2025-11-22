# Quickstart: Rename Note with Backlink Updates

1) Sync & branch  
- Ensure branch `002-rename-note-backlinks` is checked out.  
- `go fmt ./...` and `go test ./...` for baseline.

2) CLI command wiring  
- Add a Cobra command under `cmd/` that delegates to a new action in `pkg/actions` for rename.  
- Define flags/args: source, target, overwrite (default false), maybe `--no-backlinks` if needed; keep output consistent with existing commands.

3) Core logic  
- In `pkg/actions`, orchestrate: validate vault paths, detect git repo/cleanliness, perform git-aware rename when available otherwise filesystem rename.  
- Call/link utilities in `pkg/obsidian` to rewrite backlinks across vault content while preserving alias text and block/header refs; honor ignored paths.

4) MCP parity  
- Mirror inputs/outputs in `pkg/mcp` with a new tool registration in `register.go`; ensure schema reflects CLI flags/options.

5) Safety & reporting  
- Default to non-overwrite; abort on conflicts with clear errors.  
- Emit a summary of renamed path, updated links, skips, and whether git history was preserved.

6) Testing  
- Add table-driven tests with temp vaults (git and non-git) covering: successful git rename, fallback rename, overwrite prevention, alias/block ref preservation, ignored path skip.  
- Run `go test ./...` before submission.

7) Docs  
- Update README/docs for new command flags and MCP tool description if user-facing.
