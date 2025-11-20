# Implementation Plan: Backlink discovery across CLI and MCP

**Branch**: `001-backlink-support` | **Date**: 2025-11-19 | **Spec**: specs/001-backlink-support/spec.md  
**Input**: Feature specification from `/specs/001-backlink-support/spec.md`

**Note**: This plan follows the `/speckit.plan` workflow and will guide Phase 0 (research) and Phase 1 (design/contracts).

## Summary

Add optional one-hop backlinks to `prompt` and `list` CLI commands and the `files` MCP tool, showing files that link to matched results. Backlinks must dedupe referring files, include all Obsidian link variants (alias/header/block/embed), exclude external URLs, and remain opt-in so default outputs stay unchanged. Approach: reuse existing vault parsing/link detection in `pkg/obsidian`, extend actions in `pkg/actions` to fetch and format backlinks, update CLI/MCP outputs to group backlinks under each match, and guard with performance target of ~2s for ~1k notes.

## Technical Context

**Language/Version**: Go 1.23+  
**Primary Dependencies**: Cobra CLI, Viper/config where used, internal packages `pkg/obsidian`, `pkg/actions`, `pkg/mcp`, `pkg/config`  
**Storage**: Local filesystem vaults (no external DB)  
**Testing**: `go test ./...` with `stretchr/testify`; prefer temp dirs/mocks for vaults  
**Target Platform**: CLI on macOS/Linux; MCP server for agents  
**Project Type**: Go CLI + MCP server  
**Performance Goals**: Backlink-enriched responses within ~2s for vaults up to ~1,000 notes in 95% of runs  
**Constraints**: One-hop backlinks only; opt-in flag to avoid changing default output ordering; deduped backlinks; ignore external URLs  
**Scale/Scope**: Typical vaults O(1k) notes; must stay responsive without prebuilding global index

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

The constitution file is placeholder/undefined; no enforceable gates detected. Proceeding with standard quality gates (tests, clarity, avoidance of unnecessary complexity). Post-Phase 1 review: still no constitution-specific gates triggered.

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)
```text
cmd/                    # Cobra command definitions (prompt, list, etc.)
pkg/actions/            # Orchestration for CLI commands
pkg/obsidian/           # Vault parsing, link/tag/file helpers
pkg/config/             # Paths and defaults
pkg/mcp/                # MCP server/tools (files, etc.)
mocks/                  # Test doubles and fixtures
docs/, README.md        # User docs
```

**Structure Decision**: Use existing Go CLI layout; keep command parsing in `cmd/`, business logic in `pkg/actions`, vault primitives in `pkg/obsidian`, MCP surfaces in `pkg/mcp`, tests alongside packages with fixtures in `mocks/`.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | N/A | N/A |

## Phase 0: Outline & Research

- Extract unknowns: None flagged; confirm link variant handling and performance assumptions from spec (already clarified).
- Research tasks: validate Obsidian link patterns supported; confirm one-hop link traversal strategy and dedupe; survey existing `pkg/obsidian` link parsing to align output formatting.

## Phase 1: Design & Contracts

- Data model: vault file, backlink relation (referrer → target), search match.
- Contracts: describe CLI flags/outputs and MCP `files` response fields for backlink inclusion (one-hop, deduped, labeled).
- Quickstart: implementation steps, tests to add, and manual validation of backlink display for CLI/MCP.
