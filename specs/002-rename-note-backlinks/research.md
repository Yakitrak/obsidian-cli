# Research: Rename Note with Backlink Updates

## Git-aware rename
- **Decision**: Use git-aware rename when vault root is a git repo and working tree is in a state that permits rename; fall back to filesystem rename otherwise and inform user.
- **Rationale**: Preserves history while avoiding failure in non-git environments.
- **Alternatives considered**: Always filesystem rename (loses history); force git rename even when repo absent or dirty (would fail or surprise users).

## Backlink rewriting strategy
- **Decision**: Rewrite all incoming links (wikilinks, markdown links, embeds) to the new path while preserving alias text and block/header fragments; skip ignored/system paths.
- **Rationale**: Maintains link fidelity and presentation while preventing corruption in non-content areas.
- **Alternatives considered**: Rewrite only wikilinks (would miss embeds/markdown links); rewrite everything including ignored/system paths (risking unwanted edits).

## Conflict and safety handling
- **Decision**: Prevent overwrite by default; require explicit opt-in to replace targets; abort with clear errors when source missing, target exists, or permissions/git state block operation.
- **Rationale**: Aligns with safety principle and avoids destructive actions.
- **Alternatives considered**: Silent overwrite (data loss risk); silent skip on conflict (ambiguous result).

## MCP/CLI parity
- **Decision**: Keep MCP tool inputs/outputs aligned with CLI flags and summaries; surface identical validation and result reporting.
- **Rationale**: Ensures consistent behavior across agents and terminal users.
- **Alternatives considered**: Divergent behaviors per surface (hard to reason about, breaks constitution parity).
