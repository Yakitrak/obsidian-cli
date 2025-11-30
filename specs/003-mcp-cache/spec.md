# PRD: MCP Vault Cache & Filesystem Watcher

**Feature Branch**: `003-mcp-cache`  
**Created**: 2025-11-30  
**Status**: Draft  
**Input**: Build an in-memory cache and filesystem monitoring layer so the MCP server crawls the vault once, then serves file/tag-aware requests instantly by keeping metadata hot and only reprocessing files that change.

## Problem & Goals

- MCP responses that require vault-wide file/tag scans are slow because each request re-walks the vault.
- We need an in-memory cache that holds file paths, metadata, and extracted tags, refreshed incrementally from filesystem events.
- When files change, maintain a dirty set to reprocess only what changed on the next command; stale entries must not be served.
- Prefer a third-party Go watcher (default) with optional Watchman support when present.
- CLI commands should continue to work without the cache but be able to reuse it when the MCP server is running in the same process.

## Clarifications

_(Record questions and answers during spec development)_

- Q: Should the cache persist across MCP server restarts? → A: _(TBD)_
- Q: Should multiple vaults share a single cache process or have isolated caches? → A: _(TBD)_

## User Scenarios & Testing _(mandatory)_

### User Story 1 – Instant tag queries after first use (Priority: P1)

An MCP client queries files by tag. The first query triggers a vault crawl; subsequent queries respond in sub-200ms from the warm cache.

**Why this priority**: Fast tag queries are the primary motivation for caching; this validates the core value proposition.

**Independent Test**: Start the MCP server (no blocking crawl), issue a `files` query with `tag:project` (triggers crawl), then issue a second query and verify response time is under 200ms on a 5k-file vault.

**Acceptance Scenarios**:

1. **Given** an MCP server that has not yet crawled the vault, **When** a client issues the first query requiring file/tag data, **Then** the crawl happens on-demand and results are returned.
2. **Given** a warm cache from a prior query, **When** a client requests files tagged `project`, **Then** the response returns in under 200ms without re-walking the vault.
3. **Given** a vault with 10k files and hierarchical tags, **When** the client queries `tag:work/meeting`, **Then** all matching files (including `work/meeting/daily`) are returned from the cache.

---

### User Story 2 – Edits are reflected on next query (Priority: P1)

A user edits a note (adding or removing tags), and the next MCP query reflects those changes without requiring a full crawl.

**Why this priority**: Stale data undermines trust; incremental refresh is essential for correctness.

**Independent Test**: With the MCP server running, modify a note's tags via an external editor, then issue a tag query and confirm the updated file appears/disappears as expected.

**Acceptance Scenarios**:

1. **Given** a running MCP server and a note without the `urgent` tag, **When** the user adds `#urgent` to the note and saves, **Then** a subsequent `tag:urgent` query includes that note.
2. **Given** a note that appears in `tag:todo` results, **When** the user removes the `#todo` tag and saves, **Then** a subsequent query no longer returns that note.

---

### User Story 3 – Graceful degradation on watcher failure (Priority: P2)

If the filesystem watcher stops emitting events (e.g., too many files, permissions issue), the system falls back to synchronous validation without crashing.

**Why this priority**: Reliability in edge cases builds user trust; the system should never silently serve stale data.

**Independent Test**: Simulate watcher saturation or disconnect, then issue a query and verify the response is correct (even if slower) and the user receives a warning.

**Acceptance Scenarios**:

1. **Given** a watcher that has stopped emitting events, **When** a query is issued, **Then** the system detects staleness and performs a targeted recount before responding.
2. **Given** a watcher failure, **When** the user checks cache status, **Then** the status indicates degraded mode with a timestamp of last successful watch event.

---

### User Story 4 – CLI operates independently without cache (Priority: P3)

CLI commands continue to work via direct filesystem scans when no MCP server (or cache) is running.

**Why this priority**: Backward compatibility ensures existing workflows are unaffected.

**Independent Test**: Run `obsidian-cli list --vault "Test"` without an MCP server and verify it returns results via direct scan.

**Acceptance Scenarios**:

1. **Given** no MCP server is running, **When** a CLI command is executed, **Then** it completes using direct filesystem scans as before.
2. **Given** a CLI session with cache explicitly disabled, **When** commands are run, **Then** no cache initialization occurs and behavior matches pre-cache versions.

---

### Edge Cases

- Vault contains tens of thousands of files; cache must remain responsive and memory-bounded.
- Rapid burst of filesystem events (e.g., git checkout switching branches); watcher debounces and coalesces events.
- File renamed or moved; cache updates both old and new paths atomically.
- Symlinked directories or files; watcher handles without infinite loops.
- Files ignored by `.obsidianignore` or system patterns; excluded from cache and index.
- Cache warm-up interrupted by shutdown; graceful restart without corruption.

## Requirements _(mandatory)_

### Functional Requirements

- **FR-001**: Perform a vault crawl lazily on first command that requires file/tag data; populate caches for paths, metadata (size, mtime, frontmatter summary), and tags.
- **FR-002**: Start a filesystem watcher after the first crawl completes; enqueue path-level dirty entries on create/update/delete/rename.
- **FR-003**: Maintain a dirty set so the next request that needs metadata/tags reprocesses only dirty paths before answering.
- **FR-004**: Expose cache-aware MCP tools (list/search/tag queries/backlink-aware operations) with consistent behavior vs. existing commands.
- **FR-005**: Provide a cache warm status/metrics endpoint for observability (counts of cached files, dirty size, last crawl duration).
- **FR-006**: Allow graceful fallback to a synchronous crawl when watcher setup fails or watchman is unavailable.
- **FR-007**: Ensure CLI commands can either instantiate the cache service (opt-in flag/config) or continue with direct filesystem scans.
- **FR-008**: Keep MCP registrations and agent guide updated to describe cache-backed behavior and any new flags.
- **FR-009**: Respect existing ignore patterns (`.obsidianignore`, system directories) when populating and updating the cache.
- **FR-010**: Provide configuration options for cache behavior: enable/disable, watcher backend selection, debounce timing, memory limits.

### Non-Functional Requirements

- **NFR-001**: MCP server startup is instantaneous (no blocking crawl). First query that requires vault data completes within 5 seconds on a 5k-file vault; incremental refresh for dirty files must be sub-100ms per file in steady state.
- **NFR-002**: Memory overhead under 150 MB for 20k files (paths, metadata, tag indexes).
- **NFR-003**: Thread-safe reads/writes; no data races under concurrent MCP requests.
- **NFR-004**: Degraded mode when watcher stops emitting events: mark cache stale and force a targeted recount before serving responses.

### Key Entities

- **CacheService**: Singleton per vault session responsible for lazy crawl on first use, watch setup, dirty bookkeeping, and refresh-on-demand. Exposes read APIs that trigger crawl if cold, then ensure dirty set is drained before serving.
- **CachedFileEntry**: `{ Path, ModTime, Size, Hash?, Tags[], FrontmatterSummary, LastIndexed }` – represents cached metadata for a single vault file. Named distinctly from `mcp.FileEntry` to avoid confusion.
- **FileIndex**: `map[string]*CachedFileEntry` keyed by normalized vault-relative path.
- **TagIndex**: `map[string]map[string]struct{}` mapping tag name to set of file paths for O(1) tag lookups.
- **DirtySet**: `map[string]DirtyMeta` where `DirtyMeta` records event reason (`Created|Modified|Deleted|Renamed`) and timestamp.
- **Watcher**: Abstraction over filesystem event sources (`fsnotify` default, optional `watchman`).

## Data Model & Indexes

- **CachedFileEntry**: `{ Path, ModTime, Size, Hash (optional), Tags [], FrontmatterSummary map[string]string, LastIndexed }`.
- **Caches**:
  - `fileIndex map[string]*CachedFileEntry` keyed by normalized vault-relative path.
  - `tagIndex map[string]map[string]struct{}` mapping tag → set of paths (use small-set backing structs to reduce allocations).
  - Optional `folderIndex map[string][]string` for quick folder listings without full scans.
- **Dirty Tracking**:
  - `dirtySet map[string]DirtyMeta` where `DirtyMeta` records reason (`Created|Modified|Deleted|Renamed`) and last event time.
  - `deleteSet []string` for removed paths to purge indexes before answering.
  - A monotonic `version` counter so readers can detect when refresh is needed.

## Architecture

- **CacheService** (new package `pkg/cache`):
  - Responsible for lazy crawl on first use, watch setup, dirty bookkeeping, and refresh-on-demand.
  - Exposes read APIs (`ListFiles`, `FilesByTag`, `GetFile`) that trigger crawl if cache is cold, then ensure `dirtySet` is drained before serving.
- **Crawler**: Uses existing vault traversal utilities (`pkg/obsidian`) to hydrate `fileIndex` and `tagIndex`.
- **Watcher**:
  - Default: `github.com/fsnotify/fsnotify` (add to `go.mod`).
  - Optional: Watchman client (shell out to `watchman` CLI) when available and configured, to handle deep tree watching efficiently.
- **Refresh Pipeline**:
  - Watch events enqueue paths into `dirtySet`.
  - On next cache read, process `dirtySet`: re-parse files (respect ignores), update `fileIndex`, update `tagIndex`, remove deleted paths.
- **Integration Points**:
  - MCP server holds a singleton `CacheService` per vault session (`pkg/mcp`).
  - CLI commands can request a cache handle via config flag/env; otherwise they call existing filesystem code paths.
  - Config (`pkg/config`) gains optional cache settings (enable, watcher choice, max memory, debounce interval).

## Implementation Plan

### Phase 1: Foundations (P1)

1. Add `pkg/cache` with `CacheService` interface and concrete implementation.
2. Define `CachedFileEntry`, indexes, and `dirtySet` types; include normalization helpers shared with `pkg/obsidian`.
3. Build a crawler that reuses existing vault walkers and tag extraction to populate `fileIndex` and `tagIndex`.
4. Persist crawl stats (duration, counts) for observability.

### Phase 2: Watcher Layer (P1)

5. Integrate `fsnotify` as the default watcher; wrap in an abstraction so watchman can be plugged in.
6. Add config to select watcher backend: `fsnotify` (default) or `watchman` (if binary present or env enables).
7. Handle create/modify/delete/rename events; debounce bursts and collapse duplicate events.

### Phase 3: Dirty Processing (P1)

8. Implement `MarkDirty` handlers to update `dirtySet` from watcher events.
9. On cache read entrypoints, call `RefreshDirty()` that:
   - Removes deleted paths from all indexes.
   - Re-reads/retags modified paths.
   - Updates `tagIndex` and `fileIndex` atomically.

### Phase 4: MCP Integration (P1)

10. Wire `CacheService` into MCP server; tools call cache APIs which trigger lazy crawl on first use.
11. Update MCP tools (list/search/tag/backlink) to use cache reads and ensure `dirtySet` is flushed before responding.
12. Update `pkg/mcp/register.go` tool descriptions and `pkg/mcp/resources.go` agent guide to document cache-backed behavior and flags.

### Phase 5: CLI Integration (P2)

13. Add config/flag to enable cache for CLI sessions; share the same `CacheService` when running alongside MCP or start a transient instance per command.
14. Maintain backward compatibility for users who prefer direct scans.

### Phase 6: Observability & Control (P2)

15. Add status endpoint/command/tool to surface cache warm state, counts, dirty queue size, last refresh time, and watcher backend.
16. Emit logs when falling back to synchronous scans.

### Phase 7: Resilience (P2)

17. Detect watcher failures; mark cache stale and trigger a partial recount on next request.
18. Add backpressure when dirty set grows beyond threshold (e.g., trigger mini-batch refresh in background).

### Phase 8: Testing (P1-P2)

19. Unit tests for cache mutation, tag index updates, and dirty processing.
20. Integration tests with temp dirs simulating create/update/delete/rename events.
21. Performance smoke tests on synthetic 5k/20k file trees to validate memory/time budgets.

### Phase 9: Docs & Rollout (P2)

22. Update README for new flags/options; add MCP agent guide notes for instant responses.
23. Ship feature behind config defaulting to enabled for MCP, disabled for standalone CLI until validated.

## Success Criteria _(mandatory)_

### Measurable Outcomes

- **SC-001**: First MCP request requiring file/tag data triggers a single full crawl; subsequent requests respond from cache with only dirty files reprocessed.
- **SC-002**: Filesystem changes are reflected in cache queries after the next command invocation without a full vault scan.
- **SC-003**: Cache-backed MCP tools return results matching the existing behavior/filters in 100% of tested scenarios.
- **SC-004**: On a 5k-file vault, first query (triggering lazy crawl) completes in under 5 seconds in 95% of runs.
- **SC-005**: On a 20k-file vault, steady-state memory usage remains under 150 MB.
- **SC-006**: Watcher failures trigger graceful fallback and clear user-facing messaging; no silent stale data.
- **SC-007**: Configurable cache usage for CLI does not regress existing command behavior when disabled.

## Assumptions

- Vaults use Obsidian-compatible Markdown files; non-Markdown assets are indexed by path only.
- The default watcher (`fsnotify`) handles typical vault sizes (up to 50k files); Watchman is recommended for larger vaults.
- MCP sessions are typically single-vault; multi-vault support uses separate `CacheService` instances.
- Users accept a brief delay on the first query that requires vault data in exchange for instant subsequent queries.
- Existing ignore patterns (`.obsidianignore`) are respected and do not require additional configuration.

## Design Decisions

- **Watcher**: Use `fsnotify` (leveraging FSEvents on macOS) as the primary watcher. Watchman is deprioritized unless native events prove insufficient; if needed later, use a Go client.
- **Change Detection**: Rely solely on `mtime` and `size`.
- **Memory Strategy**: Full-cache model; no eviction or partial caching.
- **Multi-vault**: One `CacheService` instance per vault.
- **Persistence**: No persistence; cache is volatile and rebuilds on first use.
