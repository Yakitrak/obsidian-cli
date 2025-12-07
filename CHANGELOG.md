# Changelog

## [v0.7.0] - 2025-12-07
- Added `graph vault-context` and `graph note-context` commands to emit JSON vault and per-note graph context (communities, hubs/authorities, neighbors, backlinks, recency).
- Replaced PageRank with HITS hub/authority scores across graph stats, communities, and MCP responses, with CLI output updated to show both hub and authority metrics.
- Turned on multi-hop recency cascading by default for graph analysis (CLI and MCP); use `--recency-cascade=false` or `recencyCascade:false` to opt out. **Default behavior change.**
- Enriched MCP graph tools (`community_list`, `community_detail`) and added `note_context` / `vault_context` tools with authority distributions, recency summaries, bridge strength, weak components, and key-note/MOC detection.
- Parallelized graph build and recency computation, reused cached note metadata (including derived content times), and fixed a Windows watcher deadlock for more reliable, faster runs on large vaults.
- Increased default graph listing limits (e.g., `graph --limit` now 100) and added optional timing output for graph commands and MCP to inspect analysis cost.


## [0.6.2] - 2025-12-06
- Refresh analysis cache providers before using cached backlinks/graph data for more accurate results.
- Prevent the files MCP tool from mutating base `SuppressedTags` when per-call overrides are supplied.
- Reload `.obsidianignore` patterns on crawl/resync so ignore changes apply without restarting.
- Clarify daily note MCP tool description (no longer claims to create missing notes).
- Update `move_notes` MCP tool comment to reflect default backlink rewriting behavior.
- Simplify release script to pass the generated release notes file directly to GoReleaser.


## [0.6.1] - 2025-12-06

- Fix Windows-specific issues (path separators, JSON escaping, permission/error handling) for more reliable behavior on Windows
- Harden file-watcher and cache behavior, including race-condition fixes and real fsnotify-based integration tests
- Improve analysis cache correctness by avoiding caching results when the vault version changes during computation
- Add GitHub Actions CI workflow for linting, unit/integration tests across Linux/macOS/Windows, and multi-OS builds
- Enhance developer tooling with new Makefile targets (`lint`, `integration`, `test_all`) and a more capable release helper script
- Rewrite README into a clearer project landing page with badges, feature overview, command reference, and MCP usage docs


## [v0.6.0] - 2025-12-06

- Add `file` command group so move/rename operations work for both notes and attachments, updating backlinks/embeds by default.
- Improve backlink/graph performance and reliability via analysis memoization and a hardened vault cache with watcher fallback.
- Extend `properties` CLI with `set`, `delete`, and `rename` operations for bulk frontmatter editing (YAML-aware, with dry-run and worker controls).
- Replace MCP tag/property tools with unified `mutate_tags` and `mutate_properties` operations, supporting scoped inputs and dry-run summaries.
- Enhance `prompt` command to optionally emit absolute paths in `<file path="...">` blocks for better downstream tooling.
- Breaking: flip default for `note move --update-backlinks` to `true`; pass `--update-backlinks=false` to skip backlink/embedding rewrites.

