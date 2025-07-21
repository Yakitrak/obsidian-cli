package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterAll registers all MCP tools with the given server
func RegisterAll(s *server.MCPServer, config Config) error {
	// Register prompt_files tool (LLM-friendly file contents)
	promptFilesTool := mcp.NewTool("prompt_files",
		mcp.WithDescription("Return the contents of matching notes formatted for LLM consumption (similar to obsidian-cli prompt command)."),
		mcp.WithArray("inputs",
			mcp.Required(),
			mcp.Description("List of inputs (find:pattern, tag:tagname, or literal folder/file names)"),
			mcp.WithStringItems(mcp.Description("Input pattern")),
		),
		mcp.WithBoolean("followLinks", mcp.Description("Follow wikilinks recursively")),
		mcp.WithNumber("maxDepth", mcp.Description("Maximum depth for following wikilinks (0 = don't follow)"), mcp.Min(0)),
		mcp.WithBoolean("skipAnchors", mcp.Description("Skip wikilinks containing anchors (e.g. [[Note#Section]])")),
		mcp.WithBoolean("skipEmbeds", mcp.Description("Skip embedded wikilinks (e.g. ![[Embedded Note]])")),
		mcp.WithArray("suppressTags", mcp.Description("Additional tags to suppress/exclude"), mcp.WithStringItems()),
		mcp.WithBoolean("noSuppress", mcp.Description("Disable all tag suppression")),
	)
	s.AddTool(promptFilesTool, PromptTool(config))

	// Register file_info tool
	fileInfoTool := mcp.NewTool("file_info",
		mcp.WithDescription("Get information about a specific file, including word count, character count, creation date, and modification date."),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Path to the file (relative to vault root)"),
		),
	)
	s.AddTool(fileInfoTool, FileInfoTool(config))

	// Register print_note tool
	printNoteTool := mcp.NewTool("print_note",
		mcp.WithDescription("Print the full contents of a note file. Useful for reading the content of specific notes."),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Path to the note to print (relative to vault root)"),
		),
	)
	s.AddTool(printNoteTool, PrintNoteTool(config))

	// Register search_text tool
	searchTextTool := mcp.NewTool("search_text",
		mcp.WithDescription("Search for text within all notes in the vault. Supports case-sensitive and case-insensitive searches."),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Text to search for within notes"),
		),
		mcp.WithBoolean("caseSensitive",
			mcp.Description("Whether the search should be case sensitive"),
		),
	)
	s.AddTool(searchTextTool, SearchTextTool(config))

	// Register list_tags tool
	listTagsTool := mcp.NewTool("list_tags",
		mcp.WithDescription("List all tags found in the vault. Returns a list of unique tags used across all notes."),
	)
	s.AddTool(listTagsTool, ListTagsTool(config))

	// Register open_in_os tool
	openInOSTool := mcp.NewTool("open_in_os",
		mcp.WithDescription("Open a file in the default operating system application. This is a side-effect operation that launches the file externally."),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Path to the file to open (relative to vault root)"),
		),
	)
	s.AddTool(openInOSTool, OpenInOSTool(config))

	// Register daily_note tool (returns content)
	dailyNoteTool := mcp.NewTool("daily_note",
		mcp.WithDescription("Return the content of the daily note for the specified date (defaults to today). Creates an empty note if it does not exist."),
		mcp.WithString("date", mcp.Description("Date in YYYY-MM-DD format (defaults to today)")),
	)
	s.AddTool(dailyNoteTool, DailyNoteTool(config))

	return nil
}
