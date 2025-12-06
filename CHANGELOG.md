# Changelog

## [v0.6.0] - 2025-12-06

- Add `file` command so move/rename operations work uniformly for notes and attachments.
- Change `note move` default to `--update-backlinks=true`, updating backlinks/embeds for moves and renames (including attachments).
- Add `properties set|delete|rename` commands for bulk frontmatter edits with YAML values, dry-run, and worker tuning.
- Replace MCP tag tools with unified `mutate_tags` (add/delete/rename with optional scoped inputs and dry-run).
- Add MCP `mutate_properties` tool for setting, deleting, and renaming frontmatter properties from MCP clients.
- Improve MCP performance and robustness with analysis caching and hardened cache watcher/polling behavior.
- Allow `prompt` command to emit absolute file paths when requested.

