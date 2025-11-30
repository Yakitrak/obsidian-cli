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

This MCP server exposes Obsidian vault operations to AI agents. Use tools for dynamic queries; this resource is for reference.

## Tools Reference

### files
List files and optionally include content/frontmatter. Response: ` + "`" + `{vault, count, files:[{path, tags, frontmatter?, content?}]}` + "`" + `

**Inputs** (required array): ` + "`" + `find:pattern` + "`" + `, ` + "`" + `tag:name` + "`" + `, ` + "`" + `key:value` + "`" + `, or literal paths. Combine with AND/OR/NOT.

**Options:** ` + "`" + `includeContent` + "`" + ` (default true), ` + "`" + `includeFrontmatter` + "`" + `, ` + "`" + `includeBacklinks` + "`" + `, ` + "`" + `followLinks` + "`" + `/` + "`" + `maxDepth` + "`" + `, ` + "`" + `absolutePaths` + "`" + `, ` + "`" + `suppressTags` + "`" + `/` + "`" + `noSuppress` + "`" + `.

### list_tags
List all tags with individual and hierarchical counts. Response: ` + "`" + `{tags:[{name, individualCount, aggregateCount}]}` + "`" + `

**Options:** ` + "`" + `match` + "`" + ` array to scope the scan.

### list_properties
Inspect frontmatter and inline properties. Response: ` + "`" + `{properties:[{name, noteCount, shape, valueType, enumValues?, enumValueCounts?}]}` + "`" + `

**Options:** ` + "`" + `source` + "`" + ` ("all", "frontmatter", or "inline"), ` + "`" + `match` + "`" + `, ` + "`" + `excludeTags` + "`" + `, ` + "`" + `enumThreshold` + "`" + ` (default 25), ` + "`" + `verbose` + "`" + `, ` + "`" + `includeEnumCounts` + "`" + ` (default true).

### daily_note / daily_note_path
Get or locate the daily note. Defaults to today; pass ` + "`" + `date` + "`" + ` as YYYY-MM-DD.

### rename_note
Rename a note and update backlinks. Requires ` + "`" + `source` + "`" + ` and ` + "`" + `target` + "`" + `. Options: ` + "`" + `updateBacklinks` + "`" + ` (default true), ` + "`" + `overwrite` + "`" + `.

### move_notes
Move one or more notes. Pass ` + "`" + `moves` + "`" + ` array of ` + "`" + `{source, target}` + "`" + ` or single ` + "`" + `source` + "`" + `/` + "`" + `target` + "`" + `. Options: ` + "`" + `updateBacklinks` + "`" + `, ` + "`" + `overwrite` + "`" + `, ` + "`" + `open` + "`" + `.

### Tag Management (read-write mode only)

- **add_tags**: Add tags to notes matching input criteria. Requires ` + "`" + `tags` + "`" + ` and ` + "`" + `inputs` + "`" + ` arrays.
- **delete_tags**: Remove tags vault-wide. Requires ` + "`" + `tags` + "`" + ` array.
- **rename_tag**: Rename tags vault-wide. Requires ` + "`" + `fromTags` + "`" + ` array and ` + "`" + `toTag` + "`" + ` string.

All tag tools support ` + "`" + `dryRun` + "`" + ` to preview changes.

## Input Patterns

- ` + "`" + `find:*.md` + "`" + ` – filename glob (` + "`" + `*` + "`" + `, ` + "`" + `?` + "`" + ` wildcards)
- ` + "`" + `tag:project` + "`" + ` – matches ` + "`" + `#project` + "`" + ` and children like ` + "`" + `#project/work` + "`" + `
- ` + "`" + `Status:active` + "`" + ` – match property value (frontmatter or inline ` + "`" + `Key:: Value` + "`" + `)
- ` + "`" + `folder/path` + "`" + ` – literal path relative to vault root
- Boolean: ` + "`" + `["tag:meeting", "AND", "NOT", "tag:archived"]` + "`" + `

## Example Use Cases

**Find all meeting notes from Q4:**
` + "```" + `json
{"inputs": ["find:*meeting*", "AND", "find:2024-1*"]}
` + "```" + `

**Get project notes with their backlinks:**
` + "```" + `json
{"inputs": ["tag:project"], "includeBacklinks": true, "includeContent": false}
` + "```" + `

**Discover what properties exist on person notes:**
` + "```" + `json
{"match": ["tag:person"], "source": "frontmatter"}
` + "```" + `

**Find notes by a specific author:**
` + "```" + `json
{"inputs": ["Author:[[Jane Smith]]"]}
` + "```" + `

**Archive old project tags:**
` + "```" + `json
{"fromTags": ["project/2023"], "toTag": "archive/project/2023", "dryRun": true}
` + "```" + `

**Add review tag to untagged notes in a folder:**
` + "```" + `json
{"tags": ["needs-review"], "inputs": ["Inbox/"], "dryRun": true}
` + "```" + `

## Best Practices

- **Scope queries** with ` + "`" + `match` + "`" + ` to reduce payload size on large vaults.
- **Use dryRun** before any destructive tag operation.
- **Batch operations**: pass multiple tags/patterns in one call rather than looping.
- **Prefer read-only tools** (` + "`" + `files` + "`" + `, ` + "`" + `list_tags` + "`" + `, ` + "`" + `list_properties` + "`" + `) unless writes are requested.
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
