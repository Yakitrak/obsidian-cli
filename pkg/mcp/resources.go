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

**Options:** ` + "`" + `includeContent` + "`" + ` (default true), ` + "`" + `includeFrontmatter` + "`" + `, ` + "`" + `includeBacklinks` + "`" + `, ` + "`" + `maxDepth` + "`" + ` (set >0 to follow links), ` + "`" + `absolutePaths` + "`" + `, ` + "`" + `suppressTags` + "`" + `/` + "`" + `noSuppress` + "`" + `.

### list_tags
List all tags with individual and hierarchical counts. Response: ` + "`" + `{tags:[{name, individualCount, aggregateCount}]}` + "`" + `

**Options:** ` + "`" + `match` + "`" + ` array to scope the scan.

### list_properties
Inspect frontmatter and inline properties. Response: ` + "`" + `{properties:[{name, noteCount, shape, valueType, enumValues?, enumValueCounts?}]}` + "`" + `

**Options:** ` + "`" + `source` + "`" + ` ("all", "frontmatter", or "inline"), ` + "`" + `match` + "`" + `, ` + "`" + `excludeTags` + "`" + `, ` + "`" + `only` + "`" + ` (limit to specific properties), ` + "`" + `valueLimit` + "`" + ` (default 25; maxValues-1 when ` + "`" + `only` + "`" + ` is used), ` + "`" + `verbose` + "`" + `, ` + "`" + `valueCounts` + "`" + ` (default true).

### community_list
List communities (label propagation) with anchors, top tags, and top authority notes (with hub/authority scores) plus authorityBuckets (adaptive quantile-style distribution with example paths) and authorityStats (mean/p50/p75/p90/p95/p99/max) and recency (latest note + count in last 30d). Response: ` + "`" + `{communities:[{id,size,fractionOfVault,anchor,topTags,topAuthority,density,bridges,bridgesDetailed,authorityBuckets,authorityStats,recency}], stats, orphanCount, orphans?, components:[{id,size,fractionOfVault}]}` + "`" + `. Use the ` + "`" + `id` + "`" + ` for community_detail.

**Options:** same filters as ` + "`" + `graph_stats` + "`" + `: ` + "`" + `skipAnchors` + "`" + `, ` + "`" + `skipEmbeds` + "`" + `, ` + "`" + `includeTags` + "`" + `, ` + "`" + `exclude` + "`" + `, ` + "`" + `include` + "`" + `, ` + "`" + `minDegree` + "`" + `, ` + "`" + `mutualOnly` + "`" + `, plus ` + "`" + `maxCommunities` + "`" + `, ` + "`" + `maxTopNotes` + "`" + `.

### community_detail
Show full detail for a community by ` + "`" + `id` + "`" + ` (from community_list/graph_stats) or by ` + "`" + `file` + "`" + ` (vault-relative/absolute path): anchor, density, bridges (with counts), fractionOfVault, top tags/authority (with hub/authority scores), authorityBuckets (adaptive quantile-style with examples), authorityStats (mean/p50/p75/p90/p95/p99/max), recency (latest note + count in last 30d), and members with hub/authority/in/out (optional tags/neighbors with linksIn/linksOut and bridge flags). Response: ` + "`" + `{id,anchor,size,fractionOfVault,density,bridges,bridgesDetailed,topTags,topAuthority,authorityBuckets,authorityStats,recency,members[{path,title,inbound,outbound,hub,authority,weakComponent,isBridge?,bridgeEdges?,tags?,neighbors?,linksOut?,linksIn?}],internalEdges}` + "`" + `.

**Options:** provide one of ` + "`" + `id` + "`" + ` or ` + "`" + `file` + "`" + ` plus: ` + "`" + `skipAnchors` + "`" + `, ` + "`" + `skipEmbeds` + "`" + `, ` + "`" + `includeTags` + "`" + `, ` + "`" + `includeNeighbors` + "`" + `, ` + "`" + `exclude` + "`" + `, ` + "`" + `include` + "`" + `, ` + "`" + `minDegree` + "`" + `, ` + "`" + `mutualOnly` + "`" + `, ` + "`" + `limit` + "`" + `.

### note_context
Return graph + community context for one or more notes (contexts follow input order). Response: ` + "`" + `{contexts:[...],count}` + "`" + `. Fields per note: ` + "`" + `{path,title?,error?,tags?,frontmatter?,graph:{inbound,outbound,hub,hubPercentile,authority,authorityPercentile,isOrphan,weakComponent,strongComponent},community:{id,size,fractionOfVault,density,anchor,topTags,topAuthority,authorityBuckets,authorityStats,recency,bridges,isBridge},neighbors:{linksOut:[{path,community}],linksIn:[...],sameCommunity,crossCommunity},neighborsTruncated?,neighborsLimit?,backlinks?,backlinksTruncated?,backlinksLimit?}` + "`" + `. If a file is not found in the graph, the response includes ` + "`" + `error` + "`" + ` instead of failing the batch. Neighbors and backlinks are sorted by Authority (descending) before truncation, so the most important are kept.

**HITS Scores:** ` + "`" + `hub` + "`" + ` measures how well a note curates/aggregates links to important notes (high for MOCs/indexes). ` + "`" + `authority` + "`" + ` measures how often a note is referenced by good hubs (high for cornerstone concepts).

**Options:** ` + "`" + `files` + "`" + ` (required) plus ` + "`" + `skipAnchors` + "`" + `, ` + "`" + `skipEmbeds` + "`" + `, ` + "`" + `includeFrontmatter` + "`" + `, ` + "`" + `includeBacklinks` + "`" + ` (default true), ` + "`" + `includeNeighbors` + "`" + `, ` + "`" + `includeTags` + "`" + `, ` + "`" + `neighborLimit` + "`" + ` (default 50), ` + "`" + `backlinksLimit` + "`" + ` (default 50), ` + "`" + `exclude` + "`" + `, ` + "`" + `include` + "`" + `, ` + "`" + `minDegree` + "`" + `, ` + "`" + `mutualOnly` + "`" + `.

### vault_context
Compact vault summary to orient an agent: graph stats, orphan counts, weak components, top communities (size/fraction/density/anchor/top tags & notes/bridges with hub/authority scores + adaptive authorityBuckets with examples + authorityStats mean/p50/p75/p90/p95/p99/max + recency: latest note and recent count in last 30d), key notes (anchors/bridges/top authority), optional MOC/key-note list from patterns (with pattern info), and optional embedded note_context payloads. Response: ` + "`" + `{stats, orphanCount, topOrphans?, components?[{id,size,fractionOfVault}], communities:[{id,size,fractionOfVault,anchor,density,topTags,topAuthority,authorityBuckets,authorityStats,recency,bridgesDetailed}], keyNotes?, mocs?[{path,pattern?}], keyPatterns?, noteContexts?}` + "`" + `.

**Options:** ` + "`" + `skipAnchors` + "`" + `, ` + "`" + `skipEmbeds` + "`" + `, ` + "`" + `includeTags` + "`" + `, ` + "`" + `exclude` + "`" + `, ` + "`" + `include` + "`" + `, ` + "`" + `minDegree` + "`" + `, ` + "`" + `mutualOnly` + "`" + `, ` + "`" + `maxCommunities` + "`" + `, ` + "`" + `communityTopNotes` + "`" + `, ` + "`" + `communityTopTags` + "`" + `, ` + "`" + `bridgeLimit` + "`" + `, ` + "`" + `topOrphans` + "`" + `, ` + "`" + `topComponents` + "`" + `, ` + "`" + `topNotes` + "`" + `, ` + "`" + `keyPatterns` + "`" + ` (defaults to vault config keyNotePatterns), ` + "`" + `contextFiles` + "`" + `, contextIncludeBacklinks/frontmatter/neighbors/tags, ` + "`" + `contextNeighborLimit` + "`" + `, ` + "`" + `contextBacklinksLimit` + "`" + `.

### daily_note / daily_note_path
Get or locate the daily note. Defaults to today; pass ` + "`" + `date` + "`" + ` as YYYY-MM-DD.

### rename_note
Rename a note or attachment and update backlinks. Requires ` + "`" + `source` + "`" + ` and ` + "`" + `target` + "`" + `. Options: ` + "`" + `updateBacklinks` + "`" + ` (default true), ` + "`" + `overwrite` + "`" + `. Non-markdown files keep their extensions; embeds/links are rewritten when enabled.

### move_notes
Move one or more notes or attachments. Pass ` + "`" + `moves` + "`" + ` array of ` + "`" + `{source, target}` + "`" + ` or single ` + "`" + `source` + "`" + `/` + "`" + `target` + "`" + `. Options: ` + "`" + `updateBacklinks` + "`" + ` (default true), ` + "`" + `overwrite` + "`" + `, ` + "`" + `open` + "`" + `. Backlinks/embeds update by default.

### Tag & Property Management (read-write mode only)

- **mutate_tags**: ` + "`" + `op` + "`" + ` = add|delete|rename. add requires ` + "`" + `tags` + "`" + ` + ` + "`" + `inputs` + "`" + `; delete requires ` + "`" + `tags` + "`" + ` (optional ` + "`" + `inputs` + "`" + `); rename requires ` + "`" + `fromTags` + "`" + ` + ` + "`" + `toTag` + "`" + ` (optional ` + "`" + `inputs` + "`" + `). All support ` + "`" + `dryRun` + "`" + `.
- **mutate_properties**: ` + "`" + `op` + "`" + ` = set|delete|rename. set requires ` + "`" + `property` + "`" + ` + ` + "`" + `value` + "`" + ` (YAML) + ` + "`" + `inputs` + "`" + `; delete requires ` + "`" + `properties` + "`" + ` (optional ` + "`" + `inputs` + "`" + `); rename requires ` + "`" + `fromProperties` + "`" + ` + ` + "`" + `toProperty` + "`" + ` (optional ` + "`" + `inputs` + "`" + `). ` + "`" + `dryRun` + "`" + ` supported.

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
- **Add vaults not registered in Obsidian** via the CLI command ` + "`" + `add-vault <name> <path>` + "`" + `; defaults and MCP tools honor these preferences alongside Obsidian's own config.
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
