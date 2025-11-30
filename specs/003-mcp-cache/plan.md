# Implementation Plan: MCP Vault Cache & Filesystem Watcher

**Branch**: `003-mcp-cache` | **Date**: 2026-12-07 | **Spec**: `specs/003-mcp-cache/spec.md`  
**Input**: PRD for in-memory cache, lazy crawl, and watcher-backed dirty tracking for MCP fast responses.

## Summary

Deliver a cache service that lazily crawls the vault on first use, keeps file/tag metadata hot, and refreshes incrementally from filesystem events. MCP tools should answer from cache with per-request dirty draining; CLI remains compatible via existing direct-scan paths (no cache opt-in). Default watcher is `fsnotify`; watcher is modular for future backends, but no Watchman implementation yet.

## Constitution & Layering Check

- Keep flag parsing in `cmd/`; orchestration in `pkg/actions`; vault primitives and parsing in `pkg/obsidian`; cache service in new `pkg/cache`; MCP wiring/parity in `pkg/mcp`; config in `pkg/config`.  
- Go 1.23+, gofmt, tests.  
- Ignore rules respected (`.obsidianignore`, system paths).  
- No binaries committed.  
- MCP docs (register/resources) and README updated for any new flags/behaviors.

## Architecture Notes

- **CacheService** (new `pkg/cache`): lazy crawl on first use, watcher setup, dirty set maintenance, refresh-on-read entrypoints (`ListFiles`, `FilesByTag`, `GetFile` etc.).  
- **Data**: `fileIndex map[string]*CachedFileEntry`, `tagIndex map[string]map[string]struct{}`, optional folder index; `dirtySet map[string]DirtyMeta` + delete handling; version counter.  
- **Watcher**: `fsnotify` default (FSEvents on macOS) behind a small interface so other backends can be added later; no Watchman implementation yet.  
- **Change detection**: mtime + size only.  
- **Scope**: One cache instance per vault; volatile (no persistence).

## Story-by-Story Plan

### Story 1 (P1): Instant tag queries after first use
- Add `pkg/cache` with `CacheService` interface and concrete impl; define `CachedFileEntry`, indexes, dirty metadata, normalization helpers (reuse `pkg/obsidian` path utils).
- Implement lazy initial crawl: reuse vault walker/tag extraction to populate file/tag indexes; record crawl stats (duration, counts).
- Expose read APIs that trigger crawl if cold, then drain dirty set before responding.
- Wire MCP tools (list/search/tag/backlink-aware reads) to call cache APIs; ensure MCP startup doesn’t block on crawl.
- Add config for cache enablement (default on for MCP) and watcher backend selection (interface only; single `fsnotify` impl initially).
- Tests: unit tests for cold-start crawl populating indexes; MCP integration test that first query triggers crawl and second is cache-hot with timing assertion (allow generous threshold in CI).

### Story 2 (P1): Edits reflected on next query
- Integrate `fsnotify`: watch vault root (respect ignores); debounce/coalesce events.
- Map events to normalized paths; mark dirty with reason (create/modify/delete/rename).
- Implement `RefreshDirty`: remove deleted paths from file/tag indexes; re-parse modified/created; update tag index atomically.
- Ensure read entrypoints call `RefreshDirty` before serving and are concurrency-safe (mutex/RWMutex).
- Tests: temp-dir integration simulating add/remove tag, rename file, delete file; validate cache reflects changes after refresh cycle.

### Story 3 (P2): Graceful degradation on watcher failure
- Detect watcher failures or silent stoppage (error channels, inactivity threshold); mark cache as stale/degraded.
- On stale flag, force targeted recount/validation before serving responses; log/user-facing warning.
- Expose status/metrics API (counts, dirty size, last crawl, watcher backend/state) for MCP tool/CLI surface.
- Tests: simulate watcher error/closure; assert degraded flag and fallback recount path.

### Story 4 (P3): CLI operates independently without cache
- Keep existing direct-scan path as default and only path for CLI; no cache opt-in.
- Ensure MCP cache implementation does not alter CLI behavior or outputs.
- Tests: action-layer/CLI integration verifying direct scan path remains unchanged.

## Cross-Cutting Tasks

- Update `pkg/mcp/register.go` tool descriptions and `pkg/mcp/resources.go` agent guide to note cache-backed behavior and any status tool; no CLI cache flags.
- Update README for watcher modularity and cache-backed MCP behavior (CLI remains direct-scan).
- Add metrics/logging hooks for crawl duration, dirty refresh count, cache size.
- Benchmarks/smoke: synthetic 5k/20k vaults to validate time/memory targets (manual/optional in CI).

## Risks & Mitigations

- **Watcher overload/branch switches**: debounce and batch dirty processing; cap dirty set triggers mini-batch refresh.  
- **Ignores drift**: centralize ignore evaluation in one helper shared by walker and watcher filters.  
- **Concurrency**: guard indexes with RWMutex; process dirty set under write lock with minimal contention.  
- **Windows/paths**: normalize separators; ensure watcher handles drive letters/rooted paths.
