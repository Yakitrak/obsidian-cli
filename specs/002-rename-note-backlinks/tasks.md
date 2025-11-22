# Tasks: Rename Note with Backlink Updates

**Input**: Design documents from `/specs/002-rename-note-backlinks/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Included per quickstart (table-driven/action/MCP coverage).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Baseline checks and branch readiness

- [X] T001 Run baseline formatting/tests in `./` (`go fmt ./...`, `go test ./...`) to confirm a clean starting point

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared helpers needed by all user stories

- [X] T002 [P] Review backlink parsing capabilities in `pkg/obsidian/wikilinks.go` to confirm alias/header/block data exposed for rewrite helper inputs
- [X] T003 Add shared ignore/suppression helper in `pkg/obsidian/ignore.go` with unit coverage in `pkg/obsidian/ignore_test.go` for vault/system path skips reused by rename flows
- [X] T004 Add link rewrite helper in `pkg/obsidian/rewrite.go` with table-driven tests in `pkg/obsidian/rewrite_test.go` covering wikilinks/markdown/embeds, alias preservation, header/block fragments, and ignored paths

**Checkpoint**: Link parsing inputs, ignore rules, and rewrite helper are ready for story work.

---

## Phase 3: User Story 1 - Rename a note safely via CLI (Priority: P1) 🎯 MVP

**Goal**: CLI rename preserves git history when available and rewrites all backlinks while keeping alias/block fragments intact.

**Independent Test**: Rename a note with backlinks in a git-tracked temp vault and verify backlinks resolve and git status shows a rename (not delete/add).

### Tests for User Story 1

- [X] T005 [P] [US1] Add action-level tests in `pkg/actions/rename_test.go` covering git repo rename with backlink rewrite and overwrite-prevention behavior

### Implementation for User Story 1

- [X] T006 [US1] Implement rename orchestration in `pkg/actions/rename.go` (validate source/target, choose git vs filesystem rename, invoke rewrite helper, emit summary)
- [X] T007 [P] [US1] Wire CLI command in `cmd/rename.go` with flags (source, target, overwrite default false, backlinks toggle) delegating to `pkg/actions` rename
- [X] T008 [P] [US1] Add CLI flow test in `cmd/rename_test.go` using a temp git vault to verify rename, backlink update, and clear summary/error messaging

**Checkpoint**: CLI rename delivers end-to-end behavior with history preservation and backlink rewrites.

---

## Phase 4: User Story 2 - Trigger rename from MCP client (Priority: P2)

**Goal**: MCP clients invoke rename with the same behavior and validations as CLI.

**Independent Test**: Call the rename MCP tool against a fixture vault and confirm identical outcomes to the CLI (renamed path, link updates, overwrite handling).

### Tests for User Story 2

- [ ] T009 [P] [US2] Add MCP rename tool tests in `pkg/mcp/tools_test.go` asserting request/response shape, summary fields, and parity with CLI behavior

### Implementation for User Story 2

- [ ] T010 [P] [US2] Implement MCP rename tool handler in `pkg/mcp/tools.go` (or dedicated file) delegating to `pkg/actions` rename with shared parameters
- [ ] T011 [P] [US2] Update tool registration/schema in `pkg/mcp/register.go` to expose rename parameters (source, target, overwrite, backlinks toggle) aligned with CLI flags

**Checkpoint**: MCP rename matches CLI behavior and schema.

---

## Phase 5: User Story 3 - Handle vaults without version control (Priority: P3)

**Goal**: Non-git vaults (or git vaults in dirty state) still rename notes and update backlinks with explicit feedback about history handling.

**Independent Test**: Rename a note in a non-git temp vault and confirm backlink updates plus messaging that history preservation was not applied.

### Tests for User Story 3

- [ ] T012 [P] [US3] Add non-git and dirty-git cases to `pkg/actions/rename_test.go` verifying filesystem rename path, backlink rewrites, and user-facing messaging

### Implementation for User Story 3

- [ ] T013 [US3] Extend `pkg/actions/rename.go` fallback logic to handle non-git vaults and dirty git states with clear summaries and safe aborts on conflicts

**Checkpoint**: Rename works safely without git and surfaces clear feedback.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, consistency, and final verification

- [ ] T014 [P] Update user-facing docs in `README.md` and `docs/` for the rename command/MCP tool usage, flags, and history/backlink behavior
- [ ] T015 [P] Run final `go fmt ./...` and `go test ./...` in `./` and address any failures before delivery

---

## Dependencies & Execution Order

- Phase 1 → Phase 2 → User stories (P1 → P2 → P3) → Polish.
- All user stories depend on foundational helpers (T002–T004) and benefit from action implementation in US1 before MCP wiring (US2) and non-git refinements (US3).
- Story dependencies: US2 depends on US1’s action; US3 depends on rename action but can parallelize testing work once action exists.

## Parallel Opportunities

- Foundational: T002–T004 can be split (parser review, ignore helper, rewrite helper) if coordination on interfaces is clear.
- US1: T007 and T008 can proceed after T006 API is defined; T005 can run in parallel once helper signatures are set.
- US2: T010 and T011 can proceed in parallel after T006 defines action API; T009 follows once handler is available.
- US3: T012 can start once action supports non-git hooks; T013 integrates learnings.
- Polish tasks (T014, T015) can run in parallel after core stories are done.

## Implementation Strategy

- MVP = User Story 1 (CLI rename with git-aware history and backlink rewrites). Deliver after Phase 1–3.
- Incrementally add US2 (MCP parity) then US3 (non-git/diff git feedback) to broaden coverage.
- Keep helpers in `pkg/obsidian` reusable; avoid duplicating parsing or ignore logic in actions/CLI/MCP layers.
