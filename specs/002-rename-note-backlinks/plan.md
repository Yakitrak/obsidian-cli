# Implementation Plan: Rename Note with Backlink Updates

**Branch**: `002-rename-note-backlinks` | **Date**: 2025-11-22 | **Spec**: specs/002-rename-note-backlinks/spec.md
**Input**: Feature specification from `/specs/002-rename-note-backlinks/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Provide a CLI command and MCP tool to rename notes while preserving backlinks, alias text, and block/header references; when the vault is git-controlled, perform renames through git to keep history. Approach: detect git presence/cleanliness, run git-aware rename when available, rewrite links across vault content while respecting ignore rules, and produce clear summaries of changes/errors consistent between CLI and MCP surfaces.

## Reuse & Refactor Plan (keep it DRY)

- Reuse existing link parsing/backlink detection in `pkg/obsidian/wikilinks.go` and its fixtures; extend only as needed to expose alias text and fragments for rewrites.
- Introduce a shared link-rewrite helper in `pkg/obsidian` that accepts backlink/link findings and rewrites targets to the new path while preserving alias/display text and block/header fragments.
- Centralize ignore/suppression handling so rename and existing list/prompt flows share the same include/exclude logic (extract common bits from `pkg/actions/list.go` if necessary).
- Keep orchestration in `pkg/actions`: validate source/target, decide git vs filesystem rename, call the shared rewrite helper, and produce a unified summary.
- Maintain CLI/MCP parity by wiring only in `cmd/` and `pkg/mcp/register.go`, with all heavy lifting in shared packages.
- Tests: reuse `mocks/vaults/backlinks` and add temp-vault cases for git and non-git rename; ensure rewrite helper is unit-tested separately and via end-to-end action tests.

## Technical Context
**Language/Version**: Go 1.23+  
**Primary Dependencies**: Cobra CLI, Viper/config where applicable, obsidian vault/link parsing helpers in `pkg/obsidian`, MCP tooling in `pkg/mcp`  
**Storage**: Local filesystem vaults (Markdown files); git repository when present  
**Testing**: `go test ./...` with `stretchr/testify`, temp-dir fixtures for vaults  
**Target Platform**: macOS/Linux/Windows terminals; MCP clients consuming the same behavior  
**Project Type**: CLI with MCP server surface  
**Performance Goals**: Rename with ~200 backlinks completes within ~5 seconds on typical vaults; link rewriting covers vault-scale without noticeable lag  
**Constraints**: Preserve git history when available; avoid destructive overwrites by default; keep CLI/MCP outputs backward compatible and explicit about skipped actions  
**Scale/Scope**: Vaults with thousands of Markdown notes and mixed link styles (wikilinks, markdown links, embeds, aliases, block/header refs)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- Layered surfaces: Keep flag parsing in `cmd/`, orchestrate in `pkg/actions`, vault primitives in `pkg/obsidian`, config in `pkg/config`, MCP parity in `pkg/mcp` → PASS
- Go standards/tooling: Go 1.23+, keep dependencies minimal, run fmt/test → PASS
- Test-first/fixtures: Table-driven tests with temp vaults; cover link parsing/rewrites and git/no-git paths → PASS
- UX stability/docs: Preserve existing output shapes; add docs for new CLI/MCP options; avoid committing binaries → PASS
- Safety/integrity: Avoid destructive overwrites; handle unreadable files gracefully; provide clear errors → PASS
- Post-Phase-1 re-eval: Design artifacts maintain layering, testing, and UX/safety expectations → PASS

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
cmd/                    # Cobra commands (flag parsing only)
pkg/
├── actions/            # Orchestration for CLI/MCP commands
├── obsidian/           # Vault/path/link primitives and parsing
├── config/             # Vault paths/defaults
└── mcp/                # MCP server and tool registration (register.go)
docs/                   # User-facing documentation
mocks/                  # Shared fixtures/doubles
specs/                  # Feature specs/plans/research
bin/                    # Build artifacts (not committed)
Makefile, main.go, go.mod/go.sum
```

**Structure Decision**: Use existing single-CLI layout; place new orchestration in `pkg/actions`, vault/link logic in `pkg/obsidian`, git detection/helpers in relevant packages, command wiring in `cmd/`, and MCP parity in `pkg/mcp` with updated registration.

## Complexity Tracking

No constitution violations identified; complexity tracking not required.
