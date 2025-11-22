# Feature Specification: Rename Note with Backlink Updates

**Feature Branch**: `002-rename-note-backlinks`  
**Created**: 2025-11-22  
**Status**: Draft  
**Input**: User description: "Add a rename note command/mcp tool that updates all backlinks and uses git to do the rename to preserve history (iff the vault is a git repo). Preserve link aliases, block references, etc."

## User Scenarios & Testing *(mandatory)*

<!--
  IMPORTANT: User stories should be PRIORITIZED as user journeys ordered by importance.
  Each user story/journey must be INDEPENDENTLY TESTABLE - meaning if you implement just ONE of them,
  you should still have a viable MVP (Minimum Viable Product) that delivers value.
  
  Assign priorities (P1, P2, P3, etc.) to each story, where P1 is the most critical.
  Think of each story as a standalone slice of functionality that can be:
  - Developed independently
  - Tested independently
  - Deployed independently
  - Demonstrated to users independently
-->

### User Story 1 - Rename a note safely via CLI (Priority: P1)

A vault user runs the rename command to move or retitle a note and expects every backlink to continue working while history is preserved when the vault uses version control.

**Why this priority**: Preventing broken links and maintaining history is critical to trust the tool for day-to-day note maintenance.

**Independent Test**: Rename a note with existing backlinks in a version-controlled vault and verify backlinks resolve and the history shows a rename instead of delete/add.

**Acceptance Scenarios**:

1. **Given** a vault tracked by version control and a note referenced by wikilinks and embedded links, **When** the user renames the note via the CLI command, **Then** the note is renamed with version history preserved and all backlinks now point to the new path.
2. **Given** a note that includes alias-style links and block reference links, **When** the note is renamed via CLI, **Then** alias labels and block reference fragments remain intact while pointing to the new location.

---

### User Story 2 - Trigger rename from MCP client (Priority: P2)

An MCP client (e.g., agent or editor plugin) requests a note rename through the exposed tool and expects identical behavior to the CLI command without needing local filesystem knowledge.

**Why this priority**: Ensures automation and integrations can manage notes with the same safety guarantees as manual commands.

**Independent Test**: Invoke the MCP rename tool for a note with backlinks and confirm the resulting vault state mirrors the CLI path, link updates, and history preservation rules.

**Acceptance Scenarios**:

1. **Given** an MCP client connected to the vault, **When** it calls the rename tool with source and target note identifiers, **Then** the note is renamed, backlinks are updated, and any alias or block reference formatting is preserved.

---

### User Story 3 - Handle vaults without version control (Priority: P3)

A user renames a note in a vault that is not under version control and expects the rename and backlink updates to still complete safely with clear feedback.

**Why this priority**: Users without git still need reliable renames; explicit handling prevents silent failures or skipped history expectations.

**Independent Test**: Perform a rename in a non-version-controlled vault and verify the note moves, backlinks update, and the user is informed that history preservation was not applied.

**Acceptance Scenarios**:

1. **Given** a vault not tracked by version control, **When** a note is renamed, **Then** the rename completes via filesystem operations, backlinks update, and the user is told that version-controlled history was not applied.

---

### Edge Cases

- Rename target path already exists (file or folder).
- Source note does not exist or path points to a directory instead of a note file.
- Backlinks include embeds, alias syntax, headers, or block reference fragments.
- Vault contains ignored paths (e.g., hidden/system folders) that should not be rewritten.
- Version-controlled vault has unstaged changes that could block or conflict with the rename.
- Links pointing to the note from templates or non-Markdown assets (e.g., canvas/attachment references).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Provide a CLI rename command that accepts a source note path/name and a target path/name within the vault.
- **FR-002**: Expose an MCP tool that performs the same rename behavior and validations as the CLI command.
- **FR-003**: Validate that the source note exists and the target path does not overwrite another note unless explicitly allowed by the command/tool options.
- **FR-004**: When the vault is under version control, perform the rename through version-control-aware operations so file history is preserved; otherwise use filesystem rename.
- **FR-005**: Update all backlinks across the vault to point to the new path while preserving alias display text, embedded link formatting, and block reference fragments.
- **FR-006**: Ensure link text (aliases) and block reference identifiers remain unchanged while their targets update to the new note location.
- **FR-007**: Report a clear, single summary of changed files and links after the rename, including any skipped or failed updates.
- **FR-008**: Abort with a descriptive error when preconditions fail (missing source, target conflict, insufficient permissions, or version-control operation issues).
- **FR-009**: Provide an option or default behavior to skip rewriting links in ignored/system paths so only vault content is updated.
- **FR-010**: Keep CLI and MCP usage consistent in required/optional parameters, supported paths, and user-facing outcomes.

### Key Entities *(include if feature involves data)*

- **Vault**: Root directory of notes, tracks whether it is under version control and what paths should be ignored during link updates.
- **Note**: Markdown file within the vault identified by path/title and optional frontmatter metadata.
- **Backlink**: Incoming reference to a note, including wikilinks, markdown links, embeds, and links with alias text or block/header fragments.
- **Rename Request**: Parameters specifying source note, target note path/title, and invocation context (CLI or MCP).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Renaming a note with up to 200 backlinks completes in under 5 seconds on a typical vault and reports 0 broken links post-operation.
- **SC-002**: In version-controlled vaults, rename operations retain file history as a single rename event in version control status output in 95% of cases, with deviations reported to the user.
- **SC-003**: MCP-initiated renames match CLI behavior, confirmed by identical post-rename link resolution and summaries in 100% of tested scenarios.
- **SC-004**: At least 95% of user test cases involving alias or block reference links retain original display text/fragments while pointing to the new location.
- **SC-005**: Error messaging identifies the blocking condition (e.g., missing source, target conflict, unclean working tree) in 100% of failed rename attempts.

## Assumptions

- Vaults use Obsidian-compatible Markdown linking formats (wikilinks, markdown links, embeds, header and block references).
- Users prefer preventing overwrites by default; explicit options are required to replace existing targets.
- Version control awareness refers to git-based workflows commonly used with vaults.
