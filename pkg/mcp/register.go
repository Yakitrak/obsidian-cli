package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterAll registers all MCP tools with the given server
func RegisterAll(s *server.MCPServer, config Config) error {
	// Register prompt_files tool (LLM-friendly file contents)
	promptFilesTool := mcp.NewTool("prompt_files",
		mcp.WithDescription(`Return the contents of matching notes formatted for LLM consumption (similar to obsidian-cli prompt command).

**Input Patterns Support:**
- **find:pattern** - Find files by filename pattern (supports wildcards like * and ?)
- **tag:tagname** - Find files containing the specified tag (hierarchical tags supported)
- **literal paths** - Direct file or folder paths relative to vault root

**Examples:**
- ["find:*.md"] - All markdown files
- ["tag:project"] - Notes tagged with #project (includes #project/work, #project/personal)
- ["folder/subfolder"] - All files in a specific folder
- ["Notes/Important.md"] - Specific file
- ["tag:work", "find:meeting*"] - Multiple criteria (files tagged #work OR with "meeting" in filename)

**Best Practices:**
- Use specific patterns to limit results and improve performance
- Combine multiple input patterns in a single call rather than making separate requests
- Use followLinks and maxDepth to automatically include referenced notes
- Set suppressTags to exclude notes with certain tags from results`),
		mcp.WithArray("inputs",
			mcp.Required(),
			mcp.Description("List of input patterns (find:pattern, tag:tagname, or literal folder/file paths). Multiple patterns are OR'd together."),
			mcp.WithStringItems(mcp.Description("Input pattern - use find:*, tag:name, or literal paths")),
		),
		mcp.WithBoolean("followLinks", mcp.Description("Follow wikilinks recursively to include referenced notes")),
		mcp.WithNumber("maxDepth", mcp.Description("Maximum depth for following wikilinks (0 = don't follow, 1 = direct links only)"), mcp.Min(0)),
		mcp.WithBoolean("skipAnchors", mcp.Description("Skip wikilinks containing anchors (e.g. [[Note#Section]])")),
		mcp.WithBoolean("skipEmbeds", mcp.Description("Skip embedded wikilinks (e.g. ![[Embedded Note]])")),
		mcp.WithArray("suppressTags", mcp.Description("Additional tags to suppress/exclude from results"), mcp.WithStringItems()),
		mcp.WithBoolean("noSuppress", mcp.Description("Disable all tag suppression (include notes with suppressed tags)")),
	)
	s.AddTool(promptFilesTool, PromptTool(config))

	// Register file_info tool
	fileInfoTool := mcp.NewTool("file_info",
		mcp.WithDescription(`Get detailed information about a specific file, including frontmatter metadata, tags, word count, character count, creation date, and modification date.

**Usage:**
- Provide relative path from vault root (e.g., "Notes/Project.md")
- Returns structured information about the file's metadata and content statistics
- Useful for understanding file properties before reading full content`),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Relative path to the file from vault root (e.g., 'Notes/Project.md')"),
		),
	)
	s.AddTool(fileInfoTool, FileInfoTool(config))

	// Register print_note tool
	printNoteTool := mcp.NewTool("print_note",
		mcp.WithDescription(`Print the full contents of a specific note file. 

**Usage:**
- Provide relative path from vault root
- Returns the complete raw content of the note including frontmatter
- For multiple notes, use prompt_files instead for better formatting
- Best for reading individual notes when you know the exact path`),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Relative path to the note from vault root (e.g., 'Daily Notes/2024-01-15.md')"),
		),
	)
	s.AddTool(printNoteTool, PrintNoteTool(config))

	// Register search_text tool
	searchTextTool := mcp.NewTool("search_text",
		mcp.WithDescription(`Search for text within all notes in the vault. Returns a list of file paths containing the search term.

**Usage:**
- Searches within note content (not just filenames)
- Returns file paths, not content excerpts
- Use prompt_files afterward to get content of matching files
- Supports both case-sensitive and case-insensitive searches`),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Text to search for within note contents"),
		),
		mcp.WithBoolean("caseSensitive",
			mcp.Description("Whether the search should be case sensitive (default: false)"),
		),
	)
	s.AddTool(searchTextTool, SearchTextTool(config))

	// Register list_tags tool
	listTagsTool := mcp.NewTool("list_tags",
		mcp.WithDescription(`List all unique tags found across all notes in the vault.

**Usage:**
- Returns a simple list of all tags used in the vault
- Includes hierarchical tags (e.g., #project/work, #project/personal)
- Useful for discovering available tags before filtering or tagging operations
- For detailed tag statistics, use the tags command through other interfaces`),
	)
	s.AddTool(listTagsTool, ListTagsTool(config))

	// Register open_in_os tool
	openInOSTool := mcp.NewTool("open_in_os",
		mcp.WithDescription(`Open a file in the default operating system application (e.g., Obsidian, text editor).

**Usage:**
- Triggers external application to open the file
- Useful for allowing users to edit files outside the current session
- Returns confirmation of the action taken
- File must exist in the vault`),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Relative path to the file from vault root"),
		),
	)
	s.AddTool(openInOSTool, OpenInOSTool(config))

	// Register daily_note tool (returns content)
	dailyNoteTool := mcp.NewTool("daily_note",
		mcp.WithDescription(`Return the content of the daily note for the specified date. Creates an empty note if it doesn't exist.

**Usage:**
- Defaults to today's date if no date specified
- Date format: YYYY-MM-DD (e.g., "2024-01-15")  
- Follows standard daily notes location: "Daily Notes/YYYY-MM-DD.md"
- Returns empty content for non-existent notes rather than error`),
		mcp.WithString("date", mcp.Description("Date in YYYY-MM-DD format (optional, defaults to today)")),
	)
	s.AddTool(dailyNoteTool, DailyNoteTool(config))

	// --------------------------------------------------------------------
	// Additional tools not previously registered
	// --------------------------------------------------------------------

	// Register list_files tool (simple path listing)
	listFilesTool := mcp.NewTool("list_files",
		mcp.WithDescription(`List file paths matching input criteria. Returns just the file paths, not content.

**Input Patterns Support:**
- **find:pattern** - Find files by filename pattern (supports wildcards * and ?)
- **tag:tagname** - Find files containing the specified tag (includes hierarchical children)
- **literal paths** - Direct file or folder paths relative to vault root

**Examples:**
- ["find:2024-*.md"] - All markdown files starting with "2024-"
- ["tag:project"] - Files tagged #project (includes #project/work, #project/personal)
- ["Meeting Notes/"] - All files in Meeting Notes folder
- ["tag:urgent", "find:todo*"] - Files tagged #urgent OR with "todo" in filename

**Best Practices:**
- Use this for getting file lists before other operations
- Combine multiple patterns in single call for efficiency  
- Use followLinks to automatically include linked notes
- Set absolutePaths=true if you need full file paths`),
		mcp.WithArray("inputs",
			mcp.Required(),
			mcp.Description("List of input patterns (find:pattern, tag:tagname, or literal folder/file paths). Multiple patterns are OR'd together."),
			mcp.WithStringItems(mcp.Description("Input pattern - use find:*, tag:name, or literal paths")),
		),
		mcp.WithBoolean("followLinks", mcp.Description("Follow wikilinks recursively to include referenced notes")),
		mcp.WithNumber("maxDepth", mcp.Description("Maximum depth for following wikilinks (0 = don't follow, 1 = direct links only)"), mcp.Min(0)),
		mcp.WithBoolean("skipAnchors", mcp.Description("Skip wikilinks containing anchors (e.g. [[Note#Section]])")),
		mcp.WithBoolean("skipEmbeds", mcp.Description("Skip embedded wikilinks (e.g. ![[Embedded Note]])")),
		mcp.WithBoolean("absolutePaths", mcp.Description("Return absolute file paths instead of relative paths")),
	)
	s.AddTool(listFilesTool, ListFilesTool(config))

	// Register daily_note_path tool
	dailyNotePathToolDef := mcp.NewTool("daily_note_path",
		mcp.WithDescription(`Return the relative path to the daily note for the specified date.

**Usage:**
- Returns the expected path even if the file doesn't exist yet
- Date format: YYYY-MM-DD (e.g., "2024-01-15")
- Follows standard daily notes location: "Daily Notes/YYYY-MM-DD.md"
- Useful for determining where to create or find daily notes`),
		mcp.WithString("date", mcp.Description("Date in YYYY-MM-DD format (optional, defaults to today)")),
	)
	s.AddTool(dailyNotePathToolDef, DailyNotePathTool(config))

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
