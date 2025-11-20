# Research: Backlink discovery across CLI and MCP

## Decision 1: Include all internal Obsidian link variants as backlinks
- **Decision**: Treat alias (`[[File|Text]]`), heading (`[[File#Heading]]`), block (`[[File#^block]]`), and embed (`![[File]]`) links as backlinks to the target file; ignore external URLs.
- **Rationale**: Aligns with user expectation of "backlinks" in Obsidian; avoids missing context when embeds or aliased references are used.
- **Alternatives considered**: (a) Only basic `[[File]]` links — rejected because it misses many real backlinks; (b) Include external markdown links — rejected to avoid noise and unintended web refs.

## Decision 2: Enforce one-hop backlink depth with deduplication
- **Decision**: Limit backlink traversal to direct referrers of matched files, with deduplication per target.
- **Rationale**: Matches spec requirement, keeps output concise, controls performance, avoids cycles.
- **Alternatives considered**: (a) Multi-hop traversal — rejected for complexity/perf and out-of-scope; (b) Configurable depth — rejected to preserve simplicity and consistent output.

## Decision 3: Performance target and dataset size
- **Decision**: Aim for backlink-enriched responses within ~2s for vaults up to ~1,000 notes in 95% of runs.
- **Rationale**: Mirrors success criteria; keeps CLI/MCP responsive without indexing.
- **Alternatives considered**: (a) Higher latency tolerance — rejected to maintain UX; (b) Prebuilt index — rejected for now due to added complexity not required by scope.

## Decision 4: Output grouping and parity across surfaces
- **Decision**: Group backlinks beneath each primary match in CLI and mirror the same grouping/fields in MCP `files`, preserving existing ordering of primary matches.
- **Rationale**: Reduces context switching, ensures consistent behavior across interfaces, and avoids breaking existing output expectations.
- **Alternatives considered**: (a) Separate backlink-only listings — rejected because it splits context; (b) Reordering primaries based on backlink count — rejected to avoid surprising current users.
