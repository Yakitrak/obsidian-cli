# Feature Specification: Backlink discovery across CLI and MCP

**Feature Branch**: `001-backlink-support`  
**Created**: 2025-11-19  
**Status**: Draft  
**Input**: User description: "I want to add a back-link feature to our prompt and list commands, as well as the `files` mcp tool. My thinking is it can probably be limited to a depth of 1 (include files that link to the files matching the search criteria), however if"

## Clarifications

### Session 2025-11-19

- Q: Should backlinks include Obsidian link variants such as aliases, headers/blocks, and embeds? → A: All internal Obsidian link variants (including alias, header, block, and embed forms) count as backlinks to the target file.
- Q: Should backlinks apply to link-followed results or only the original matches? → A: Backlinks apply only to the original files matched by the filters; files pulled in via `--depth` do not receive their own backlinks list.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - See referencing notes in CLI results (Priority: P1)

CLI users searching notes via `prompt` or `list` want to immediately see which other notes link to each matched note so they can jump into the most relevant context without manual cross-referencing.

**Why this priority**: Surfacing incoming links alongside search results is the core value of backlinks and unblocks users who rely on related context to decide what to open or summarize.

**Independent Test**: Run a search that matches a note with known inbound links and confirm the output presents the matched note plus its first-degree backlinks with clear labels.

**Acceptance Scenarios**:

1. **Given** a vault where Note B is linked by Note A, **When** the user searches and Note B matches while requesting backlinks, **Then** the output lists Note B and clearly indicates Note A as a backlink targetting Note B.
2. **Given** a vault where Note C only links to Note B which links to Note A, **When** the user requests backlinks, **Then** only direct backlinks (Note B) appear and deeper references (Note C) are excluded.

---

### User Story 2 - Review context before acting on a result (Priority: P2)

Users reviewing search results want backlinks grouped with each matched note so they can decide whether to open, summarize, or copy content without running additional queries.

**Why this priority**: Reduces context switching and repeat searches by providing adjacent information at the decision point.

**Independent Test**: Execute a query returning multiple matches, each with backlinks, and verify backlinks are grouped under the correct match and deduplicated so each referring note appears once.

**Acceptance Scenarios**:

1. **Given** two matched notes each with multiple inbound links, **When** the user views results with backlinks, **Then** backlinks are grouped beneath their corresponding match and duplicates are collapsed.
2. **Given** a matched note with no inbound links, **When** the user requests backlinks, **Then** the output explicitly shows that no backlinks exist while preserving the primary result.

---

### User Story 3 - MCP clients obtain backlinks via files tool (Priority: P3)

MCP clients (e.g., external agents) using the `files` tool want backlink information returned with file matches so they can supply richer context to downstream prompts without additional calls.

**Why this priority**: Keeps external clients aligned with CLI capabilities and reduces back-and-forth to assemble context.

**Independent Test**: Invoke the `files` tool for a query that returns at least one matched file with inbound links and confirm backlinks are included, labeled, and limited to one hop.

**Acceptance Scenarios**:

1. **Given** a file that is matched by the `files` tool and linked by two other files, **When** the client requests results with backlinks, **Then** both linking files are returned once each with metadata indicating they are backlinks.
2. **Given** the same search repeated through CLI and MCP, **When** backlinks are requested, **Then** both channels return the same set of backlink files for the same matched targets.

---

### Edge Cases

- When no backlinks exist for any matched note, the commands return primary results without error and explicitly indicate the absence of backlinks.
- When backlinks form cycles or self-references, only a single backlink entry per referring file is shown to avoid loops.
- When a backlink points to a file that has been deleted or is unreadable, the system skips or flags it without blocking other results.
- When backlink counts are large, the output still applies deduplication and respects the one-hop limit without overfetching deeper links.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The `prompt` command MUST offer an option to include first-degree backlinks for each matched note while keeping the primary matches visible.
- **FR-002**: The `list` command MUST offer the same backlink option and present backlinks grouped with their originating matched note.
- **FR-003**: Backlink discovery MUST be limited to a single hop (files directly linking to matched files) and MUST NOT traverse deeper chains.
- **FR-004**: Each backlink entry MUST clearly distinguish itself from primary matches (e.g., label or grouping) and include the file identifier users expect in normal command output.
- **FR-005**: Backlink entries MUST be deduplicated so each referring file appears once per target even if multiple links exist.
- **FR-006**: Users MUST be able to request results without backlinks (default) and with backlinks (opt-in per command/tool invocation) without altering existing non-backlink output ordering.
- **FR-007**: The `files` MCP tool MUST return backlink information consistent with CLI behavior for the same query filters, including one-hop depth and deduplication.
- **FR-008**: When no backlinks are found for a matched file, the response MUST explicitly indicate that status while still returning the primary match.
- **FR-009**: Backlink detection MUST treat all internal Obsidian link variants (e.g., `[[File]]`, `[[File|Alias]]`, `[[File#Heading]]`, `[[File#^block]]`, `![[File]]`) as backlinks to the matched file path, and MUST ignore external URLs.
- **FR-010**: Backlinks MUST be returned only for the original filter-matched files; files added solely via link traversal (`--depth`/follow) MUST NOT have their own backlinks listed.

Assumptions: Backlinks are opt-in to avoid changing existing default outputs; backlink display follows the current result ordering rules unless a grouped layout is needed to keep backlinks adjacent to their matched note.

### Key Entities *(include if feature involves data)*

- **Vault file**: A note or document within the vault, identified by its path or title as shown in current command outputs.
- **Search match**: A file returned because it meets the query/filter criteria for `prompt`, `list`, or `files`.
- **Backlink**: A vault file that contains a link pointing to a matched file; constrained to direct, single-hop references and deduplicated per target.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users receive backlink-enriched CLI responses within 2 seconds for vaults with up to 1,000 notes in 95% of runs when backlinks are requested.
- **SC-002**: For test vaults with known link graphs, 100% of first-degree backlinks appear exactly once per matched note and zero second-degree links are present.
- **SC-003**: In usability checks, at least 90% of users correctly distinguish primary results from backlinks without needing documentation.
- **SC-004**: For identical queries across CLI and the `files` MCP tool, the backlink sets match with no discrepancies in 5 out of 5 sampled test cases.
