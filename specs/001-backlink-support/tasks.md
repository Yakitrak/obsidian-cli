---
description: "Tasks for Backlink discovery across CLI and MCP"
---

# Tasks: Backlink discovery across CLI and MCP

**Input**: Design documents from `/specs/001-backlink-support/`  
**Prerequisites**: plan.md, spec.md (user stories), research.md, data-model.md, contracts/, quickstart.md

**Tests**: Include targeted tests per story for backlink resolution, CLI formatting, and MCP payload parity.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

## Phase 1: Setup (Shared Infrastructure)

- [X] T001 [P] Review existing link parsing/helpers to locate backlink extension points in `pkg/obsidian/wikilinks.go`.
- [X] T002 Create shared backlink fixture vault (referrers/targets with alias, heading, block, embed) under `mocks/vaults/backlinks/` for tests.

---

## Phase 2: Foundational (Blocking Prerequisites)

- [X] T003 Implement one-hop backlink resolver (dedupe referrers, ignore external URLs, capture link type) in `pkg/obsidian/wikilinks.go`.
- [X] T004 Add table-driven backlink resolver tests using `mocks/vaults/backlinks/` in `pkg/obsidian/wikilinks_test.go`.

**Checkpoint**: Backlink resolution is available for CLI and MCP callers.

---

## Phase 3: User Story 1 - See referencing notes in CLI results (Priority: P1) 🎯 MVP

**Goal**: CLI `prompt`/`list` can show first-degree backlinks per match.

**Independent Test**: Run a search with known inbound links and confirm matched note plus labeled backlinks appear.

### Implementation for User Story 1

- [X] T005 [US1] Add opt-in backlinks flag(s) to `cmd/prompt.go` and `cmd/list.go` (default false; preserve existing behavior).
- [X] T006 [US1] Wire backlinks retrieval into `pkg/actions/list.go` (or equivalent `ListFiles` flow) when flag is set, passing one-hop/dedupe option.
- [X] T007 [US1] Render backlinks grouped under each match in CLI outputs (prompt/list) in `cmd/prompt.go` and `cmd/list.go`, keeping primary match visible.

### Tests for User Story 1

- [X] T008 [P] [US1] Add CLI/action coverage for backlink flag and inclusion in `pkg/actions/list_test.go` (ensure matches plus backlinks returned).

**Checkpoint**: Backlink-enabled CLI responses work for single match with inbound links.

---

## Phase 4: User Story 2 - Review context before acting on a result (Priority: P2)

**Goal**: Backlinks are grouped and deduped per match; empty states are explicit.

**Independent Test**: Multi-match query shows backlinks grouped per match, deduped, and indicates when none exist.

### Implementation for User Story 2

- [X] T009 [US2] Ensure per-match grouping/deduplication and "no backlinks" messaging in CLI presenters `cmd/prompt.go` and `cmd/list.go`.

### Tests for User Story 2

- [X] T010 [P] [US2] Add multi-match and no-backlink scenarios to `pkg/actions/list_test.go` covering grouping/dedupe/empty state.

**Checkpoint**: CLI output handles multiple matches, dedupe, and empty-backlink cases clearly.

---

## Phase 5: User Story 3 - MCP clients obtain backlinks via files tool (Priority: P3)

**Goal**: MCP `files` tool returns backlink data aligned with CLI behavior.

**Independent Test**: MCP `files` response includes one-hop deduped backlinks matching CLI for the same query.

### Implementation for User Story 3

- [X] T011 [US3] Add backlinks opt-in handling to MCP files tool in `pkg/mcp/tools.go` (respect includeBacklinks field).
- [X] T012 [US3] Attach backlink payload (path + linkType) to MCP response struct and serialization in `pkg/mcp/tools.go`.

### Tests for User Story 3

- [X] T013 [P] [US3] Add MCP files tool test in `pkg/mcp/register_test.go` (or new `pkg/mcp/tools_test.go`) asserting backlink payload parity with CLI fixtures.

**Checkpoint**: MCP and CLI return matching backlink sets for the same fixture query.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [X] T014 Update user-facing docs to mention backlink flags/behavior in `README.md` and relevant `docs/` snippets.
- [X] T015 Run `go fmt ./...` and `go test ./...` to verify formatting and coverage before handoff.

---

## Dependencies & Execution Order

- Story order: US1 (P1) → US2 (P2) → US3 (P3); each depends on Foundational completion.
- Phase dependencies: Setup → Foundational → US1 → US2 → US3 → Polish.

## Parallel Opportunities

- T001 and T002 can run in parallel.  
- After T003, tests T004/T008/T010/T013 can run in parallel with story-specific implementations that do not touch the same files.  
- Within story phases, tasks marked [P] can be executed concurrently when file touchpoints do not conflict.

## Implementation Strategy

- Deliver MVP by completing US1 (backlink flag + CLI output) after Foundational checkpoint.  
- Add grouping/empty-state robustness (US2), then extend MCP parity (US3).  
- Finish with documentation and repo-wide fmt/test before handoff.
