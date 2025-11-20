# Quickstart: Backlink discovery across CLI and MCP

## Implement
1. Add an opt-in flag for backlinks to `cmd/prompt` and `cmd/list` parsing (default false).
2. Extend `pkg/actions` flows to request backlinks when flag is set; keep existing ordering of primary matches.
3. Update `pkg/obsidian` link discovery to surface all Obsidian variants (alias, heading, block, embed) as backlinks; dedupe referrers and enforce one-hop depth.
4. Format output grouping backlinks under each match in CLI presenters; clearly label backlinks vs primary matches.
5. Extend `pkg/mcp/files` tool response to include backlinks field mirroring CLI behavior and one-hop/dedupe rules.

## Test
1. Add table-driven tests for backlink gathering (single-hop, dedupe, cycles, no backlinks) using temp vault fixtures or mocks.
2. Add CLI result formatting tests to ensure primary matches remain visible and backlinks are labeled/grouped; include no-backlink case.
3. Add MCP `files` tool test to assert backlink payload matches CLI for the same fixture query.
4. Validate performance manually on a ~1k-note fixture (or simulated) ensuring backlink responses remain ~2s or better.

## Validate
1. Run `go test ./...`.
2. Run `go run . list --vault <vault> --backlinks "<query>"` against a test vault with known links; compare outputs to expectations.
3. Trigger `files` MCP tool with backlinks enabled and confirm parity with CLI results.
