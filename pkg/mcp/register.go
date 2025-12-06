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
		mcp.WithDescription(`List communities (label propagation) with anchors, top tags, and top pagerank notes. Respects include/exclude/minDegree/mutualOnly filters. Response: {communities:[{id,size,nodes,anchor,topTags,topPagerank,density,bridges}], stats}.`),
		mcp.WithBoolean("skipAnchors", mcp.Description("Skip wikilinks containing anchors (e.g. [[Note#Section]])")),
		mcp.WithBoolean("skipEmbeds", mcp.Description("Skip embedded wikilinks (e.g. ![[Embedded Note]])")),
		mcp.WithBoolean("includeTags", mcp.Description("Include top tags per community (may be heavier on large vaults)")),
		mcp.WithArray("exclude", mcp.Description("Exclude notes matching these patterns (same syntax as list/prompt)"), mcp.WithStringItems()),
		mcp.WithArray("include", mcp.Description("Include only notes matching these patterns (same syntax as list/prompt)"), mcp.WithStringItems()),
		mcp.WithNumber("minDegree", mcp.Description("Drop notes whose in+out degree is below this number before analysis (0 = no filter)"), mcp.Min(0)),
		mcp.WithBoolean("mutualOnly", mcp.Description("Only consider mutual (bidirectional) links when building the graph")),
		mcp.WithNumber("maxCommunities", mcp.Description("Maximum communities to return (default 25)"), mcp.Min(1)),
		mcp.WithNumber("maxTopNotes", mcp.Description("Maximum top-pagerank notes per community (default 5)"), mcp.Min(1)),
	)
	s.AddTool(communityListTool, CommunityListTool(config))

	communityDetailTool := mcp.NewTool("community_detail",
		mcp.WithDescription(`Show full detail for a community by ID or file: anchor, density, bridges, top tags/pagerank, members with pagerank/in/out (optional tags/neighbors). Provide either id (from community_list) or file (vault-relative/absolute).`),
		mcp.WithString("id", mcp.Description("Community ID from community_list/graph_stats (e.g., c1234abcd)")),
		mcp.WithString("file", mcp.Description("Vault-relative (or absolute) path to a file contained in the desired community")),
		mcp.WithBoolean("skipAnchors", mcp.Description("Skip wikilinks containing anchors (e.g. [[Note#Section]])")),
		mcp.WithBoolean("skipEmbeds", mcp.Description("Skip embedded wikilinks (e.g. ![[Embedded Note]])")),
		mcp.WithBoolean("includeTags", mcp.Description("Include tags on members (default false)")),
		mcp.WithBoolean("includeNeighbors", mcp.Description("Include neighbor lists on members (default false)")),
		mcp.WithArray("exclude", mcp.Description("Exclude notes matching these patterns (same syntax as list/prompt)"), mcp.WithStringItems()),
		mcp.WithArray("include", mcp.Description("Include only notes matching these patterns (same syntax as list/prompt)"), mcp.WithStringItems()),
		mcp.WithNumber("minDegree", mcp.Description("Drop notes whose in+out degree is below this number before analysis (0 = no filter)"), mcp.Min(0)),
		mcp.WithBoolean("mutualOnly", mcp.Description("Only consider mutual (bidirectional) links when building the graph")),
		mcp.WithNumber("limit", mcp.Description("Limit members returned (default all)"), mcp.Min(1)),
	)
	s.AddTool(communityDetailTool, CommunityDetailTool(config))

	// Register daily_note tool (returns content)
	dailyNoteTool := mcp.NewTool("daily_note",
		mcp.WithDescription(`Return JSON with path, existence, and content for the daily note. Creates an empty note if it doesn't exist. Response: {path,date,exists,content}

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
		mcp.WithDescription(`Rename a note and update backlinks, preserving history in git vaults.

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
		mcp.WithDescription(`Move or rename one or more notes within the vault. Backlinks are NOT rewritten unless updateBacklinks=true.

Preferred: pass an array of moves [{source,target}]. For simple single-note moves, source/target can be provided directly.

Options:
- overwrite: replace an existing target (default false)
- updateBacklinks: rewrite backlinks to new path (default false)
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
		mcp.WithBoolean("updateBacklinks", mcp.Description("Rewrite backlinks to the new note path (default false)")),
		mcp.WithBoolean("open", mcp.Description("Open the first moved note in Obsidian after moving (default false)")),
	)
	s.AddTool(moveNotesTool, MoveNotesTool(config))

	// --------------------------------------------------------------------
	// Destructive tag-management tools (only when read-write enabled)
	// --------------------------------------------------------------------
	if config.ReadWrite {
		// delete_tags – remove one or more tags across all notes
		deleteTagsTool := mcp.NewTool("delete_tags",
			mcp.WithDescription(`Delete specified tags from all notes in the vault. 

**Efficient Usage:**
- **ALWAYS pass multiple tags in a single call** rather than separate calls per tag
- Use ["urgent", "old", "deprecated"] instead of three separate calls
- This processes all files once and applies all deletions together

**Tag Matching:**
- Supports hierarchical tags: deleting "project" removes #project, #project/work, #project/personal
- Exact matching: deleting "work" won't affect "rework" or "working"
- Case-sensitive matching

**Safety:**
- **Always use dryRun=true first** to preview changes before applying
- Shows exactly which files and tags will be affected
- Non-destructive preview shows the same output format as actual operation

**Examples:**
- tags: ["completed", "old"] - Remove both tags from all notes in single operation
- tags: ["project/legacy"], dryRun: true - Preview removal of legacy project tags`),
			mcp.WithArray("tags", mcp.Required(), mcp.Description("List of tags to delete from all notes. Pass multiple tags for efficient batch processing."), mcp.WithStringItems()),
			mcp.WithBoolean("dryRun", mcp.Description("If true, show what would be changed without making actual changes. RECOMMENDED to use first.")),
		)
		s.AddTool(deleteTagsTool, DeleteTagsTool(config))

		// rename_tag – rename a tag (and its hierarchical children) across vault
		renameTagTool := mcp.NewTool("rename_tag",
			mcp.WithDescription(`Rename one or more tags to a single destination tag across the entire vault.

**Efficient Usage:**
- **ALWAYS pass multiple source tags in a single call** when consolidating tags
- Use fromTags: ["old-urgent", "high-priority", "critical"] with toTag: "urgent"
- This processes all files once and applies all renames together

**Tag Renaming Rules:**
- **Hierarchical handling**: Renaming "project" to "work" changes:
  - #project → #work  
  - #project/urgent → #work/urgent
  - #project/meeting/daily → #work/meeting/daily
- **Multiple sources to one destination**: All fromTags are renamed to the same toTag
- **Exact matching**: Only exact tag matches are renamed

**Safety:**
- **Always use dryRun=true first** to preview changes before applying
- Shows exactly which files and tag transformations will occur
- Non-destructive preview shows the same output format as actual operation

**Examples:**
- fromTags: ["todo", "task"], toTag: "action" - Consolidate task-related tags
- fromTags: ["project/old"], toTag: "project/archive", dryRun: true - Preview archiving old projects`),
			mcp.WithArray("fromTags", mcp.Required(), mcp.Description("List of source tags to rename. Pass multiple tags to consolidate them into the destination tag."), mcp.WithStringItems()),
			mcp.WithString("toTag", mcp.Required(), mcp.Description("Single destination tag name that all source tags will be renamed to")),
			mcp.WithBoolean("dryRun", mcp.Description("If true, show what would be changed without making actual changes. RECOMMENDED to use first.")),
		)
		s.AddTool(renameTagTool, RenameTagsTool(config))

		setPropertyTool := mcp.NewTool("set_property",
			mcp.WithDescription(`Set a frontmatter property on matching notes. Value is parsed as YAML (strings, numbers, lists, etc.). Requires inputs to scope the change.`),
			mcp.WithString("property", mcp.Required(), mcp.Description("Property name to set")),
			mcp.WithString("value", mcp.Required(), mcp.Description("Property value (YAML allowed, e.g., \"in-progress\", 3, [a,b])")),
			mcp.WithArray("inputs", mcp.Required(), mcp.Description("Input patterns (find:, tag:, or paths) to select files"), mcp.WithStringItems()),
			mcp.WithBoolean("overwrite", mcp.Description("Overwrite existing values (default false)")),
			mcp.WithBoolean("dryRun", mcp.Description("Preview without writing changes")),
		)
		s.AddTool(setPropertyTool, SetPropertyTool(config))

		deletePropertiesTool := mcp.NewTool("delete_properties",
			mcp.WithDescription(`Delete one or more frontmatter properties across the vault or scoped inputs.`),
			mcp.WithArray("properties", mcp.Required(), mcp.Description("Properties to delete (case-insensitive)"), mcp.WithStringItems()),
			mcp.WithArray("inputs", mcp.Description("Optional: scope deletion to these inputs (find:, tag:, paths)"), mcp.WithStringItems()),
			mcp.WithBoolean("dryRun", mcp.Description("Preview without writing changes")),
		)
		s.AddTool(deletePropertiesTool, DeletePropertiesTool(config))

		renamePropertyTool := mcp.NewTool("rename_property",
			mcp.WithDescription(`Rename one or more properties to a destination property. When merge=true (default), values are merged if the destination already exists.`),
			mcp.WithArray("fromProperties", mcp.Required(), mcp.Description("Properties to rename (case-insensitive)"), mcp.WithStringItems()),
			mcp.WithString("toProperty", mcp.Required(), mcp.Description("Destination property name")),
			mcp.WithBoolean("merge", mcp.Description("Merge values when destination exists (default true)")),
			mcp.WithArray("inputs", mcp.Description("Optional: scope rename to these inputs (find:, tag:, paths)"), mcp.WithStringItems()),
			mcp.WithBoolean("dryRun", mcp.Description("Preview without writing changes")),
		)
		s.AddTool(renamePropertyTool, RenamePropertyTool(config))

		// add_tags – add one or more tags to specific notes matching input criteria
		addTagsTool := mcp.NewTool("add_tags",
			mcp.WithDescription(`Add specified tags to notes matching input criteria.

**Efficient Usage:**
- **ALWAYS pass multiple tags in a single call** rather than separate calls per tag
- Use tags: ["urgent", "review", "priority"] to add all tags to matching notes at once
- **ALWAYS pass multiple input criteria** to target the right notes efficiently
- Use inputs: ["tag:project", "find:meeting*"] to match notes with #project OR "meeting" in filename

**Input Patterns Support:**
- **find:pattern** - Files by filename pattern (supports wildcards * and ?)
- **tag:tagname** - Files containing the specified tag (includes hierarchical children)  
- **literal paths** - Direct file or folder paths relative to vault root

**Tag Addition Rules:**
- Tags are added to notes, not replaced
- Duplicate tags are automatically handled (won't create duplicates)
- Tags can be hierarchical (e.g., "project/urgent")

**Safety:**
- **Always use dryRun=true first** to preview which notes will be affected
- Shows exactly which files will get which tags
- Non-destructive preview shows the same output format as actual operation

**Examples:**
- tags: ["urgent", "review"], inputs: ["tag:project"], dryRun: true - Preview adding tags to project notes
- tags: ["meeting"], inputs: ["find:2024-01-*", "tag:daily"] - Add meeting tag to January 2024 files OR daily notes
- tags: ["archived"], inputs: ["folder/old-projects/"] - Add archived tag to all files in specific folder`),
			mcp.WithArray("tags", mcp.Required(), mcp.Description("List of tags to add to matching notes. Pass multiple tags for efficient batch processing."), mcp.WithStringItems()),
			mcp.WithArray("inputs", mcp.Required(), mcp.Description("List of input criteria (find:pattern, tag:tagname, or literal paths). Multiple criteria are OR'd together to find target notes."), mcp.WithStringItems()),
			mcp.WithBoolean("dryRun", mcp.Description("If true, show what would be changed without making actual changes. RECOMMENDED to use first.")),
		)
		s.AddTool(addTagsTool, AddTagsTool(config))
	}

	return nil
}
