package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// AddBuiltinResources registers static resources useful to MCP agents.
func AddBuiltinResources(s *server.MCPServer) {
	const uri = "obsidian-cli/docs/agent-guide"
	const name = "Obsidian CLI Agent Guide"
	const mime = "text/markdown"

	body := `# Obsidian CLI – Agent Guide

This MCP server exposes Obsidian vault operations to AI agents. Tools are preferred for dynamic queries; resources are for static reference.

## Core Tools

- **files**: List files and optionally include content/frontmatter.
  - Inputs: pass an array of patterns (` + "`" + `find:` + "`" + ` for filename globs, ` + "`" + `tag:` + "`" + ` for tags, or literal paths/folders). Patterns are ORed.
  - Options: ` + "`" + `includeContent` + "`" + ` (default true), ` + "`" + `includeFrontmatter` + "`" + ` (default false), ` + "`" + `includeBacklinks` + "`" + `, ` + "`" + `followLinks/maxDepth` + "`" + `, ` + "`" + `absolutePaths` + "`" + `.
  - Good for: grabbing note bodies, titles, or frontmatter without a second call.

- **list_tags**: Aggregate tags with per-note and hierarchical counts.
  - Options: ` + "`" + `match` + "`" + ` to scope scan using the same patterns as ` + "`" + `files` + "`" + ` (find:/tag:/paths).
  - Good for: discovering tag vocabularies or pruning unused tags.

- **list_properties**: Inspect properties (YAML frontmatter + inline ` + "`" + `Key:: Value` + "`" + `) across the vault.
  - Defaults: enumThreshold=25, includeEnumCounts=true, inline parsing enabled, tags included.
  - Options: ` + "`" + `match` + "`" + ` (find:/tag:/paths), ` + "`" + `excludeTags` + "`" + `, ` + "`" + `disableInline` + "`" + `, ` + "`" + `enumThreshold` + "`" + `, ` + "`" + `maxValues` + "`" + `, ` + "`" + `verbose` + "`" + ` (forces enums for mixed types, raises threshold to 50), ` + "`" + `includeEnumCounts` + "`" + `.
  - Good for: discovering metadata schemas, enums like Office or Archetype, spotting mixed/dirty fields.

- **daily_note / daily_note_path**: Retrieve or locate the daily note for a date (defaults to today).

- **move_notes / rename_note**: Move/rename notes. ` + "`" + `updateBacklinks` + "`" + ` controls backlink rewriting. Use cautiously; destructive operations require read-write mode.

## How to Choose a Tool

- Need bodies or frontmatter from specific notes? Use ` + "`" + `files` + "`" + ` with ` + "`" + `includeContent` + "`" + ` or ` + "`" + `includeFrontmatter` + "`" + `.
- Need tag vocab/counts? Use ` + "`" + `list_tags` + "`" + ` (optionally scoped with ` + "`" + `match` + "`" + `).
- Need property schemas/enums? Use ` + "`" + `list_properties` + "`" + ` (optionally scoped with ` + "`" + `match` + "`" + `).
- Need today’s note? Use ` + "`" + `daily_note` + "`" + ` or ` + "`" + `daily_note_path` + "`" + `.
- Need to move/rename? Use ` + "`" + `move_notes` + "`" + ` or ` + "`" + `rename_note` + "`" + `; prefer dry runs if available in client UX.

## Patterns and Matching

- ` + "`" + `find:` + "`" + ` uses filename globbing (` + "`" + `*` + "`" + `, ` + "`" + `?` + "`" + `).
- ` + "`" + `tag:` + "`" + ` matches frontmatter tags and inline hashtags (hierarchical: ` + "`" + `tag:project` + "`" + ` matches ` + "`" + `project/work` + "`" + `).
- Literal paths/folders are relative to vault root.

## Safe Usage Notes

- Avoid destructive tools unless explicitly requested; prefer read-only tools (` + "`" + `files` + "`" + `, ` + "`" + `list_tags` + "`" + `, ` + "`" + `list_properties` + "`" + `).
- Large scans: scope with ` + "`" + `match` + "`" + ` when possible to reduce payloads.
- Inline properties: ` + "`" + `list_properties` + "`" + ` treats ` + "`" + `Key:: Value` + "`" + ` lines as properties by default; disable with ` + "`" + `disableInline` + "`" + ` if that’s noisy.
`

	res := mcp.Resource{
		URI:      uri,
		Name:     name,
		MIMEType: mime,
	}

	handler := func(_ context.Context, _ mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		return []mcp.ResourceContents{mcp.TextResourceContents{
			URI:      uri,
			MIMEType: mime,
			Text:     body,
		}}, nil
	}

	s.AddResource(res, handler)
}
