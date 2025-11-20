# Data Model: Backlink discovery across CLI and MCP

## Entities

### VaultFile
- **Identifier**: vault-relative path (string)
- **Attributes**: title/name (derived), content (for parsing), tags/links (already modeled in `pkg/obsidian`)
- **Constraints**: Must be readable; skips deleted/unreadable files without blocking results.

### SearchMatch
- **Identifier**: VaultFile path
- **Attributes**: match reason (query-driven), ordering (existing sort rules from command), optional backlinks list
- **Relationships**: References one `VaultFile`; may include zero or more `Backlink` entries.

### Backlink
- **Identifier**: referrer VaultFile path (deduped per target)
- **Attributes**: link type (alias, heading, block, embed, basic), display label (as needed per surface)
- **Constraints**: Only first-degree links; ignore external URLs; one entry per referrer per target even if multiple links exist.
- **Relationships**: Points from referrer `VaultFile` to target `SearchMatch` VaultFile.

## Relationships
- `SearchMatch` 1 → N `Backlink` (zero or more backlinks grouped under each match).
- `Backlink` references exactly one referrer `VaultFile` and one target `VaultFile` (the match).

## Validation Rules
- Backlinks must be drawn only from direct referrers (one-hop).
- Backlink list must be deduped per matched target.
- If no backlinks exist, explicitly mark absence while returning the primary match.
- Link recognition must support Obsidian variants: `[[File]]`, `[[File|Alias]]`, `[[File#Heading]]`, `[[File#^block]]`, `![[File]]`; external URLs are excluded.
