# Changelog

## [v0.6.0] - 2025-12-06

- Add `file` command group so move/rename operations work for both notes and attachments, updating backlinks/embeds by default.
- Improve backlink/graph performance and reliability via analysis memoization and a hardened vault cache with watcher fallback.
- Extend `properties` CLI with `set`, `delete`, and `rename` operations for bulk frontmatter editing (YAML-aware, with dry-run and worker controls).
- Replace MCP tag/property tools with unified `mutate_tags` and `mutate_properties` operations, supporting scoped inputs and dry-run summaries.
- Enhance `prompt` command to optionally emit absolute paths in `<file path="...">` blocks for better downstream tooling.
- Breaking: flip default for `note move --update-backlinks` to `true`; pass `--update-backlinks=false` to skip backlink/embedding rewrites.

