package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterAll registers all MCP tools with the given server
func RegisterAll(s *server.MCPServer, config Config) error {
	// Register files tool (list files and optionally return content/frontmatter as JSON)
	promptFilesTool := mcp.NewTool("files",
		mcp.WithDescription(`List matching files and optionally return their contents/frontmatter as JSON. Response: {vault,count,files:[{path,absolutePath?,tags,frontmatter?,content?}]} Designed for agents to fetch paths or bulk-load note bodies in one call. Uses a hot in-memory cache with filesystem watching after the first crawl for fast responses.

**Input Patterns:**
- **find:pattern** - Filename pattern (supports * and ? wildcards)
- **tag:tagname** - Files containing the tag (hierarchy supported, e.g., tag:project matches project/work)
- **key:value** - Match frontmatter/inline properties (wikilinks normalized; supports Dataview Key:: Value)
- **literal paths** - File or folder paths relative to vault root
- **Boolean filters** - Combine inputs with AND/OR/NOT and parentheses. Pass operators as separate inputs (e.g., ["tag:foo","AND","tag:bar"]). Terms without operators are OR'd.

**Key Options:**
- **includeContent** (default: true) - Include full note content
- **includeFrontmatter** (default: false) - Include parsed frontmatter map
- **maxDepth** - Traverse wikilinks to include neighbors (set >0 to follow)
- **includeBacklinks** (default: false) - Include first-degree backlinks for matched notes
- **suppressTags/noSuppress** - Control tag-based suppression (defaults come from server config)
- **absolutePaths** - Add absolute paths alongside relative paths`),
		mcp.WithArray("inputs",
			mcp.Required(),
			mcp.Description("List of input patterns (find:pattern, tag:tagname, or literal folder/file paths). Supports AND/OR/NOT with parentheses; otherwise patterns are OR'd."),
			mcp.WithStringItems(mcp.Description("Input pattern - use find:*, tag:name, or literal paths")),
		),
		mcp.WithNumber("maxDepth", mcp.Description("Maximum depth for following wikilinks (0 = don't follow, 1 = direct links only)"), mcp.Min(0)),
		mcp.WithBoolean("skipAnchors", mcp.Description("Skip wikilinks containing anchors (e.g. [[Note#Section]])")),
		mcp.WithBoolean("skipEmbeds", mcp.Description("Skip embedded wikilinks (e.g. ![[Embedded Note]])")),
		mcp.WithBoolean("includeContent", mcp.Description("Include note content in the response (default true)")),
		mcp.WithBoolean("includeFrontmatter", mcp.Description("Include parsed frontmatter in the response")),
		mcp.WithBoolean("includeBacklinks", mcp.Description("Include first-degree backlinks for matched notes")),
		mcp.WithBoolean("absolutePaths", mcp.Description("Include absolute paths alongside relative paths")),
		mcp.WithArray("suppressTags", mcp.Description("Additional tags to suppress/exclude from results"), mcp.WithStringItems()),
		mcp.WithBoolean("noSuppress", mcp.Description("Disable all tag suppression (include notes with suppressed tags)")),
	)
	s.AddTool(promptFilesTool, FilesTool(config))

	// Register list_tags tool
	listTagsTool := mcp.NewTool("list_tags",
		mcp.WithDescription(`List all tags with counts as JSON. Returns both exact (individual) and hierarchical (aggregate) counts, sorted by aggregate descending. Supports 'match' filters (find:, tag:, paths) with boolean AND/OR/NOT to scope the scan. Response: {tags:[{name,individualCount,aggregateCount}]}`),
		mcp.WithArray("match", mcp.Description("Optional: restrict scan to files matched by these patterns (find:*, tag:, or paths). Boolean AND/OR/NOT and parentheses are supported."), mcp.WithStringItems()),
	)
	s.AddTool(listTagsTool, ListTagsTool(config))

	// Register list_properties tool
	listPropertiesTool := mcp.NewTool("list_properties",
		mcp.WithDescription(`Inspect properties across the vault (frontmatter + inline 'Key:: Value'). Returns inferred shape/type, note counts, enums, and per-value counts. Supports 'match' filters (find:, tag:, paths) with boolean AND/OR/NOT. Defaults: valueLimit=25 (or maxValues-1 when 'only' is set), valueCounts=true, source=all. Response: {properties:[{name,noteCount,shape,valueType,enumValues?,enumValueCounts?,distinctValueCount,truncatedValueSet?}]}`),
		mcp.WithBoolean("excludeTags", mcp.Description("Exclude the tags field (default false; included by default)")),
		mcp.WithArray("only", mcp.Description("Restrict output to these property names"), mcp.WithStringItems()),
		mcp.WithString("source", mcp.Description("Property source: 'all' (default), 'frontmatter' (YAML only), or 'inline' (Dataview Key:: Value only)")),
		mcp.WithArray("match", mcp.Description("Optional: restrict scan to files matched by these patterns (find:*, tag:, or paths). Boolean AND/OR/NOT and parentheses are supported."), mcp.WithStringItems()),
		mcp.WithNumber("valueLimit", mcp.Description("Emit values when distinct counts are at or below this limit (default 25)"), mcp.Min(1)),
		mcp.WithNumber("maxValues", mcp.Description("Maximum distinct values to track per property (default 500)"), mcp.Min(1)),
		mcp.WithBoolean("verbose", mcp.Description("Emit enums for mixed types and raise value limit to 50")),
		mcp.WithBoolean("valueCounts", mcp.Description("Include per-value note counts (default true)")),
	)
	s.AddTool(listPropertiesTool, ListPropertiesTool(config))

	communityListTool := mcp.NewTool("community_list",
		mcp.WithDescription(`List communities (label propagation) with anchors, top tags, top authority notes (with hub/authority scores), authorityBuckets (coarse quantile-style distribution with example paths), and recency (latest note + count in last 30d). Includes size/fractionOfVault, bridge hints, orphan counts, and weak component sizes. Response: {communities:[{id,size,fractionOfVault,anchor,topTags,topAuthority,density,bridges,bridgesDetailed,authorityBuckets,recency}], stats, orphanCount, orphans?, components:[{id,size}]}.`),
		mcp.WithBoolean("skipAnchors", mcp.Description("Skip wikilinks containing anchors (e.g. [[Note#Section]])")),
		mcp.WithBoolean("skipEmbeds", mcp.Description("Skip embedded wikilinks (e.g. ![[Embedded Note]])")),
		mcp.WithBoolean("includeTags", mcp.Description("Include top tags per community (default true)")),
		mcp.WithBoolean("recencyCascade", mcp.Description("Cascade inferred recency beyond 1 hop (default true)")),
		mcp.WithNumber("maxCommunities", mcp.Description("Maximum communities to return (default 20)"), mcp.Min(1)),
		mcp.WithNumber("maxTopNotes", mcp.Description("Maximum top authority notes per community (default 5)"), mcp.Min(1)),
	)
	s.AddTool(communityListTool, CommunityListTool(config))

	communityDetailTool := mcp.NewTool("community_detail",
		mcp.WithDescription(`Show full detail for a community by ID or file: anchor, density, bridges (with counts), fractionOfVault, top tags/authority, members with hub/authority/in/out (optional tags/neighbors with linksIn/linksOut). Provide either id (from community_list) or file (vault-relative/absolute).`),
		mcp.WithString("id", mcp.Description("Community ID from community_list/graph_stats (e.g., c1234abcd)")),
		mcp.WithString("file", mcp.Description("Vault-relative (or absolute) path to a file contained in the desired community")),
		mcp.WithBoolean("skipAnchors", mcp.Description("Skip wikilinks containing anchors (e.g. [[Note#Section]])")),
		mcp.WithBoolean("skipEmbeds", mcp.Description("Skip embedded wikilinks (e.g. ![[Embedded Note]])")),
		mcp.WithBoolean("includeTags", mcp.Description("Include tags on members (default false)")),
		mcp.WithBoolean("includeNeighbors", mcp.Description("Include neighbor lists on members (default false)")),
		mcp.WithBoolean("recencyCascade", mcp.Description("Cascade inferred recency beyond 1 hop (default true)")),
		mcp.WithNumber("limit", mcp.Description("Limit members returned (default all)"), mcp.Min(1)),
	)
	s.AddTool(communityDetailTool, CommunityDetailTool(config))

	noteContextTool := mcp.NewTool("note_context",
		mcp.WithDescription(`Return graph + community context for one or more notes (contexts returned in input order). Response: {contexts:[...],count}. Each context includes degrees, hub/authority scores (+percentiles), components, orphan/bridge status, neighbors (linksOut/linksIn, same vs cross community), optional backlinks/frontmatter, community summary (size/fractionOfVault/anchor/top tags/authority/bridges), and semantic related notes when embeddings are enabled. Hub measures curation quality (high for MOCs); authority measures reference frequency (high for cornerstone concepts).`),
		mcp.WithArray("files", mcp.Required(), mcp.Description("Array of vault-relative (or absolute) paths to the notes"), mcp.WithStringItems()),
		mcp.WithBoolean("skipAnchors", mcp.Description("Skip wikilinks containing anchors (e.g. [[Note#Section]])")),
		mcp.WithBoolean("skipEmbeds", mcp.Description("Skip embedded wikilinks (e.g. ![[Embedded Note]])")),
		mcp.WithBoolean("includeFrontmatter", mcp.Description("Include parsed frontmatter (default false)")),
		mcp.WithBoolean("includeBacklinks", mcp.Description("Include backlinks targeting these notes (default true)")),
		mcp.WithBoolean("includeNeighbors", mcp.Description("Include neighbors and community classifications (default true)")),
		mcp.WithBoolean("includeTags", mcp.Description("Include top tags in community summary (default true)")),
		mcp.WithNumber("neighborLimit", mcp.Description("Maximum neighbors to return per direction (default 50)"), mcp.Min(1)),
		mcp.WithNumber("backlinksLimit", mcp.Description("Maximum backlinks to return per note (default 50)"), mcp.Min(1)),
		mcp.WithBoolean("recencyCascade", mcp.Description("Cascade inferred recency beyond 1 hop (default true)")),
	)
	s.AddTool(noteContextTool, NoteContextTool(config))

	vaultContextTool := mcp.NewTool("vault_context",
		mcp.WithDescription(`Compact vault summary: graph stats, orphan counts, weak components, top communities (size/fraction/density/anchor/top tags and notes/bridges with hub/authority scores + authorityBuckets shown as coarse quantile-style buckets with examples + recency: latest note and recent count in last 30d), key notes (anchors/bridges/top authority), optional MOC/key-note list sourced from patterns (config or keyPatterns), optional embedded note_context payloads, and optional semanticMatches when semanticQuery is provided and embeddings are enabled. Response: {stats, orphanCount, topOrphans?, components?, communities:[{id,size,fractionOfVault,anchor,density,topTags,topAuthority,bridgesDetailed,authorityBuckets,recency}], keyNotes?, mocs?, keyPatterns?, noteContexts?, semanticQuery?, semanticMatches?}.`),
		mcp.WithBoolean("skipAnchors", mcp.Description("Skip wikilinks containing anchors (e.g. [[Note#Section]])")),
		mcp.WithBoolean("skipEmbeds", mcp.Description("Skip embedded wikilinks (e.g. ![[Embedded Note]])")),
		mcp.WithBoolean("includeTags", mcp.Description("Include top tags per community (default true)")),
		mcp.WithBoolean("recencyCascade", mcp.Description("Cascade inferred recency beyond 1 hop (default true)")),
		mcp.WithNumber("maxCommunities", mcp.Description("Maximum communities to return (default 20)"), mcp.Min(1)),
		mcp.WithNumber("communityTopNotes", mcp.Description("Top authority notes per community (default 5)"), mcp.Min(1)),
		mcp.WithNumber("communityTopTags", mcp.Description("Top tags per community (default 5)"), mcp.Min(1)),
		mcp.WithNumber("bridgeLimit", mcp.Description("Bridges to include per community (default 3)"), mcp.Min(1)),
		mcp.WithNumber("topOrphans", mcp.Description("Top orphan notes to list (default 10)"), mcp.Min(1)),
		mcp.WithNumber("topComponents", mcp.Description("Top weak components to summarize (default 5)"), mcp.Min(1)),
		mcp.WithNumber("topNotes", mcp.Description("Top global authority notes to include (default 10)"), mcp.Min(1)),
		mcp.WithArray("keyPatterns", mcp.Description("Optional patterns (find:/tag:/paths) to surface as key/MOC notes; defaults to vault config keyNotePatterns"), mcp.WithStringItems()),
		mcp.WithArray("contextFiles", mcp.Description("Optional files to include embedded note_context payloads for (vault-relative or absolute, returned in input order)"), mcp.WithStringItems()),
		mcp.WithBoolean("contextIncludeBacklinks", mcp.Description("Include backlinks in noteContexts (default true)")),
		mcp.WithBoolean("contextIncludeFrontmatter", mcp.Description("Include frontmatter in noteContexts (default false)")),
		mcp.WithBoolean("contextIncludeNeighbors", mcp.Description("Include neighbors in noteContexts (default true)")),
		mcp.WithBoolean("contextIncludeTags", mcp.Description("Include tags in noteContexts (default true)")),
		mcp.WithNumber("contextNeighborLimit", mcp.Description("Max neighbors to return per direction in embedded contexts (default 50)"), mcp.Min(1)),
		mcp.WithNumber("contextBacklinksLimit", mcp.Description("Max backlinks to return per note in embedded contexts (default 50)"), mcp.Min(1)),
		mcp.WithString("semanticQuery", mcp.Description("Optional semantic query to return embedding-based matches (requires embeddings index)")),
		mcp.WithNumber("semanticLimit", mcp.Description("Number of semantic matches to return (default 10)"), mcp.Min(1)),
	)
	s.AddTool(vaultContextTool, VaultContextTool(config))

	// Register daily_note tool (returns content)
	dailyNoteTool := mcp.NewTool("daily_note",
		mcp.WithDescription(`Return JSON with path, existence, and content for the daily note. Does not create missing notes; returns empty content when absent. Response: {path,date,exists,content}

**Usage:**
- Defaults to today's date if no date specified
- Date format: YYYY-MM-DD (e.g., "2024-01-15")  
- Follows standard daily notes location: "Daily Notes/YYYY-MM-DD.md"
- Returns empty content for non-existent notes rather than error`),
		mcp.WithString("date", mcp.Description("Date in YYYY-MM-DD format (optional, defaults to today)")),
	)
	s.AddTool(dailyNoteTool, DailyNoteTool(config))

	// Register daily_note_path tool
	dailyNotePathToolDef := mcp.NewTool("daily_note_path",
		mcp.WithDescription(`Return the relative path to the daily note for the specified date.

**Usage:**
- Returns the expected path even if the file doesn't exist yet
- Date format: YYYY-MM-DD (e.g., "2024-01-15")
- Follows standard daily notes location: "Daily Notes/YYYY-MM-DD.md"
- Response structure: { path, date, exists }`),
		mcp.WithString("date", mcp.Description("Date in YYYY-MM-DD format (optional, defaults to today)")),
	)
	s.AddTool(dailyNotePathToolDef, DailyNotePathTool(config))

	// Register rename_note tool
	renameNoteTool := mcp.NewTool("rename_note",
		mcp.WithDescription(`Rename a note or attachment and update backlinks, preserving history in git vaults. Non-markdown files keep their extensions; embeds/links are rewritten when enabled.

Required: source (existing note), target (new note path/title)
Optional: overwrite (default false), updateBacklinks (default true)`),
		mcp.WithString("source", mcp.Required(), mcp.Description("Existing note path/title")),
		mcp.WithString("target", mcp.Required(), mcp.Description("Desired new note path/title")),
		mcp.WithBoolean("overwrite", mcp.Description("Allow replacing an existing target (default false)")),
		mcp.WithBoolean("updateBacklinks", mcp.Description("Rewrite backlinks to the new note (default true)")),
	)
	s.AddTool(renameNoteTool, RenameNoteTool(config))

	// Register move_notes tool
	moveNotesTool := mcp.NewTool("move_notes",
		mcp.WithDescription(`Move or rename one or more notes or attachments within the vault. Backlinks/embeds are rewritten by default (updateBacklinks=true).

Preferred: pass an array of moves [{source,target}]. For simple single-note moves, source/target can be provided directly.

Options:
- overwrite: replace an existing target (default false)
- updateBacklinks: rewrite backlinks to new path (default true)
- open: open the first moved note in Obsidian (default false)`),
		mcp.WithArray("moves",
			mcp.Description("Array of move objects. Each move requires 'source' and 'target' values."),
			mcp.Items(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"source": map[string]any{"type": "string"},
					"target": map[string]any{"type": "string"},
				},
				"required": []string{"source", "target"},
			}),
		),
		mcp.WithString("source", mcp.Description("Single-move shorthand: existing note path/title")),
		mcp.WithString("target", mcp.Description("Single-move shorthand: desired new path/title")),
		mcp.WithBoolean("overwrite", mcp.Description("Allow replacing an existing target (default false)")),
		mcp.WithBoolean("updateBacklinks", mcp.Description("Rewrite backlinks to the new note path (default true)")),
		mcp.WithBoolean("open", mcp.Description("Open the first moved note in Obsidian after moving (default false)")),
	)
	s.AddTool(moveNotesTool, MoveNotesTool(config))

	// --------------------------------------------------------------------
	// Destructive tag/property tools (only when read-write enabled)
	// --------------------------------------------------------------------
	if config.ReadWrite {
		mutateTagsTool := mcp.NewTool("mutate_tags",
			mcp.WithDescription(`Add, delete, or rename tags in one tool.

op=add: requires tags + inputs (find:/tag:/paths).
op=delete: requires tags (optional inputs to scope).
op=rename: requires fromTags + toTag (optional inputs; hierarchical).

Always prefer dryRun=true first.`),
			mcp.WithString("op", mcp.Required(), mcp.Description("add | delete | rename")),
			mcp.WithArray("tags", mcp.Description("For op=add/delete: tags to add or delete"), mcp.WithStringItems()),
			mcp.WithArray("fromTags", mcp.Description("For op=rename: source tags to rename"), mcp.WithStringItems()),
			mcp.WithString("toTag", mcp.Description("For op=rename: destination tag name")),
			mcp.WithArray("inputs", mcp.Description("For op=add: required scope (find:, tag:, paths); for delete/rename: optional scope"), mcp.WithStringItems()),
			mcp.WithBoolean("dryRun", mcp.Description("Preview without writing changes")),
		)
		s.AddTool(mutateTagsTool, MutateTagsTool(config))

		mutatePropertiesTool := mcp.NewTool("mutate_properties",
			mcp.WithDescription(`Set, delete, or rename frontmatter properties.

op=set: property + value (YAML) + inputs required; overwrite optional.
op=delete: properties required; optional inputs to scope.
op=rename: fromProperties + toProperty required; optional inputs; merge defaults to true.`),
			mcp.WithString("op", mcp.Required(), mcp.Description("set | delete | rename")),
			mcp.WithString("property", mcp.Description("For op=set: property name to set")),
			mcp.WithString("value", mcp.Description("For op=set: property value (YAML string accepted)")),
			mcp.WithArray("inputs", mcp.Description("For op=set (required) or op=delete/rename (optional): patterns (find:, tag:, or paths)"), mcp.WithStringItems()),
			mcp.WithBoolean("overwrite", mcp.Description("For op=set: overwrite existing values (default false)")),
			mcp.WithArray("properties", mcp.Description("For op=delete: properties to delete"), mcp.WithStringItems()),
			mcp.WithArray("fromProperties", mcp.Description("For op=rename: properties to rename"), mcp.WithStringItems()),
			mcp.WithString("toProperty", mcp.Description("For op=rename: destination property name")),
			mcp.WithBoolean("merge", mcp.Description("For op=rename: merge values when destination exists (default true)")),
			mcp.WithBoolean("dryRun", mcp.Description("Preview without writing changes")),
		)
		s.AddTool(mutatePropertiesTool, MutatePropertiesTool(config))
	}

	return nil
}
