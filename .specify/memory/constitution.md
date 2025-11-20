<!--
Sync Impact Report
Version change: N/A -> 1.0.0
Modified principles: Added I. Layered CLI & MCP Surfaces; II. Go Standards & Tooling; III. Test-First & Fixtures; IV. UX Stability & Documentation; V. Safety & Integrity
Added sections: Additional Constraints & Practices; Development Workflow & Review
Removed sections: None
Templates requiring updates: ✅ .specify/templates/plan-template.md (aligns with Constitution Check); ✅ .specify/templates/spec-template.md (mandatory sections unchanged); ✅ .specify/templates/tasks-template.md (task grouping guidance aligns); ✅ .specify/templates/agent-file-template.md (no conflicts)
Follow-up TODOs: None
-->

# Obsidian CLI Constitution

## Core Principles

### I. Layered CLI & MCP Surfaces
Commands stay thin: parse flags and delegate to `pkg/actions`, `pkg/obsidian`, and `pkg/config`. Business logic lives in packages, not `cmd/`. MCP surfaces mirror CLI capabilities to keep behavior consistent across agents and the terminal.

### II. Go Standards & Tooling
Use Go 1.23+. Run `go fmt ./...` and `go test ./...` before changes are submitted; keep imports/go formatting clean. Use Cobra for CLI wiring and keep dependencies minimal and idiomatic.

### III. Test-First & Fixtures
Favor table-driven tests with `stretchr/testify`; co-locate `_test.go` files with implementations. Use temp dirs or in-repo fixtures (no personal vault data). Add coverage for edge cases: vault paths, tag/link parsing, filesystem behavior.

### IV. UX Stability & Documentation
Preserve existing defaults and output shapes; new flags or behaviors are opt-in unless explicitly approved. Update README/docs for user-facing changes. Do not commit built artifacts in `bin/` or local config; keep diffs focused on source, docs, and necessary fixtures.

### V. Safety & Integrity
Respect vault data: avoid destructive actions without explicit user intent, handle unreadable files gracefully, and ensure link/tag operations are consistent and reversible when possible. Maintain clarity in errors and avoid silent failures.

## Additional Constraints & Practices

- Primary tooling commands: `go fmt ./...`, `go build ./...` for sanity checks, `go test ./...` or `make test`. Use `make test_coverage` when adding sizable features.
- Coding style: Go-idiomatic naming (CamelCase types, lowerCamel identifiers); keep exports documented only when required.
- Project layout: `cmd/` for flag parsing, `pkg/actions/` for orchestration, `pkg/obsidian/` for vault primitives, `pkg/config/` for paths/defaults, `pkg/mcp/` for MCP exposure, `mocks/` for shared fixtures/doubles, `docs/` for supporting assets. Do not commit generated binaries in `bin/`.

## Development Workflow & Review

- Specifications and plans should be kept in `specs/[feature]/` using the provided templates; ensure user-facing behavior changes are captured in spec/plan before implementation.
- PRs and reviews must verify constitution compliance: formatting, tests, layering, UX stability, and documentation updates.
- When adding CLI behavior, prefer opt-in flags and maintain backward-compatible outputs unless explicitly agreed otherwise.

## Governance

- This constitution governs development standards for Obsidian CLI. All changes must be reviewed for compliance with Core Principles and Practices.
- Amendments: propose edits via PR with rationale; update version per semantic rules (MAJOR if principles are removed or fundamentally altered, MINOR for new principles/sections, PATCH for clarifications). `LAST_AMENDED_DATE` updates with any change; `RATIFICATION_DATE` reflects original adoption.
- Compliance checks occur during planning and code review; violations require documented justification and, when possible, mitigation or follow-up tasks.

**Version**: 1.0.0 | **Ratified**: 2025-11-19 | **Last Amended**: 2025-11-19
