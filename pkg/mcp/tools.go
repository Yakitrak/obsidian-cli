package mcp

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/mark3labs/mcp-go/mcp"
)

// ListFilesParams defines parameters for the list_files tool
type ListFilesParams struct {
	Inputs        []string `json:"inputs"`
	FollowLinks   bool     `json:"followLinks"`
	MaxDepth      int      `json:"maxDepth"`
	SkipAnchors   bool     `json:"skipAnchors"`
	SkipEmbeds    bool     `json:"skipEmbeds"`
	AbsolutePaths bool     `json:"absolutePaths"`
}

// FileInfoParams defines parameters for the file_info tool
type FileInfoParams struct {
	Path string `json:"path"`
}

// PrintNoteParams defines parameters for the print_note tool
type PrintNoteParams struct {
	Path string `json:"path"`
}

// SearchTextParams defines parameters for the search_text tool
type SearchTextParams struct {
	Query         string `json:"query"`
	CaseSensitive bool   `json:"caseSensitive"`
}

// ListTagsParams defines parameters for the list_tags tool
type ListTagsParams struct {
	// No parameters needed for basic tag listing
}

// OpenInOSParams defines parameters for the open_in_os tool
type OpenInOSParams struct {
	Path string `json:"path"`
}

// DailyNotePathParams defines parameters for the daily_note_path tool
type DailyNotePathParams struct {
	Date string `json:"date"`
}

// ListFilesTool implements the list_files MCP tool
func ListFilesTool(config Config) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if config.Debug {
			log.Printf("MCP list_files called with arguments: %v", request.GetArguments())
		}

		args := request.GetArguments()

		// Extract inputs
		inputsRaw, ok := args["inputs"].([]interface{})
		if !ok {
			return mcp.NewToolResultError("inputs parameter is required and must be an array"), nil
		}

		inputs := make([]string, len(inputsRaw))
		for i, v := range inputsRaw {
			if s, ok := v.(string); ok {
				inputs[i] = s
			} else {
				return mcp.NewToolResultError("all inputs must be strings"), nil
			}
		}

		// Extract other parameters with defaults
		followLinks, _ := args["followLinks"].(bool)
		maxDepth, _ := args["maxDepth"].(float64) // JSON numbers are float64
		skipAnchors, _ := args["skipAnchors"].(bool)
		skipEmbeds, _ := args["skipEmbeds"].(bool)
		absolutePaths, _ := args["absolutePaths"].(bool)

		// Parse inputs using the existing helper
		parsedInputs, err := actions.ParseInputs(inputs)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error parsing inputs: %s", err)), nil
		}

		note := obsidian.Note{}
		var files []string

		// Call ListFiles with the parsed parameters
		files, err = actions.ListFiles(config.Vault, &note, actions.ListParams{
			Inputs:         parsedInputs,
			FollowLinks:    followLinks,
			MaxDepth:       int(maxDepth),
			SkipAnchors:    skipAnchors,
			SkipEmbeds:     skipEmbeds,
			AbsolutePaths:  absolutePaths,
			SuppressedTags: config.SuppressedTags,
		})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error listing files: %s", err)), nil
		}

		// Format output as newline-separated paths
		result := strings.Join(files, "\n")
		if result == "" {
			result = "No files found matching the specified criteria."
		}

		return mcp.NewToolResultText(result), nil
	}
}

// FileInfoTool implements the file_info MCP tool
func FileInfoTool(config Config) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		path, ok := args["path"].(string)
		if !ok {
			return mcp.NewToolResultError("path parameter is required"), nil
		}

		if config.Debug {
			log.Printf("MCP file_info called with path: %s", path)
		}

		note := obsidian.Note{}
		info, err := actions.GetFileInfo(config.Vault, &note, path)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error getting file info: %s", err)), nil
		}

		// Format the info as a readable string
		result := fmt.Sprintf("File: %s\n\nFrontmatter:\n", path)
		if info.Frontmatter != nil {
			for k, v := range info.Frontmatter {
				result += fmt.Sprintf("  %s: %v\n", k, v)
			}
		} else {
			result += "  No frontmatter found\n"
		}

		result += "\nTags:\n"
		if len(info.Tags) > 0 {
			for _, tag := range info.Tags {
				result += fmt.Sprintf("  %s\n", tag)
			}
		} else {
			result += "  No tags found\n"
		}

		return mcp.NewToolResultText(result), nil
	}
}

// PrintNoteTool implements the print_note MCP tool
func PrintNoteTool(config Config) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		path, ok := args["path"].(string)
		if !ok {
			return mcp.NewToolResultError("path parameter is required"), nil
		}

		if config.Debug {
			log.Printf("MCP print_note called with path: %s", path)
		}

		note := obsidian.Note{}
		content, err := actions.PrintNote(config.Vault, &note, actions.PrintParams{
			NoteName: path,
		})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error printing note: %s", err)), nil
		}

		return mcp.NewToolResultText(content), nil
	}
}

// SearchTextTool implements the search_text MCP tool
func SearchTextTool(config Config) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		query, ok := args["query"].(string)
		if !ok {
			return mcp.NewToolResultError("query parameter is required"), nil
		}

		caseSensitive, _ := args["caseSensitive"].(bool)

		if config.Debug {
			log.Printf("MCP search_text called with query: %s", query)
		}

		// For search functionality, we'll need to implement a basic search
		// since there's no SearchNotes function that returns results
		// Let's implement a basic grep-like search
		note := obsidian.Note{}
		vaultPath, err := config.Vault.Path()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error getting vault path: %s", err)), nil
		}

		allNotes, err := note.GetNotesList(vaultPath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error getting notes list: %s", err)), nil
		}

		var results []string
		for _, notePath := range allNotes {
			content, err := note.GetContents(vaultPath, notePath)
			if err != nil {
				continue // Skip notes we can't read
			}

			// Simple text search
			searchIn := content
			searchFor := query
			if !caseSensitive {
				searchIn = strings.ToLower(content)
				searchFor = strings.ToLower(query)
			}

			if strings.Contains(searchIn, searchFor) {
				results = append(results, notePath)
			}
		}

		if len(results) == 0 {
			return mcp.NewToolResultText("No matches found."), nil
		}

		// Format results as newline-separated entries
		result := strings.Join(results, "\n")
		return mcp.NewToolResultText(result), nil
	}
}

// ListTagsTool implements the list_tags MCP tool
func ListTagsTool(config Config) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if config.Debug {
			log.Printf("MCP list_tags called")
		}

		note := obsidian.Note{}
		tagSummaries, err := actions.Tags(config.Vault, &note)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error listing tags: %s", err)), nil
		}

		if len(tagSummaries) == 0 {
			return mcp.NewToolResultText("No tags found in vault."), nil
		}

		// Convert TagSummary slice to string slice
		var tags []string
		for _, tagSummary := range tagSummaries {
			tags = append(tags, tagSummary.Name)
		}

		// Format tags as newline-separated entries
		result := strings.Join(tags, "\n")
		return mcp.NewToolResultText(result), nil
	}
}

// PromptTool implements the prompt_files MCP tool (similar to the CLI `prompt` command).
// It returns the full contents of matching notes formatted for LLM consumption.
func PromptTool(config Config) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()

		// Extract and validate required "inputs" parameter.
		inputsRaw, ok := args["inputs"].([]interface{})
		if !ok {
			return mcp.NewToolResultError("inputs parameter is required and must be an array"), nil
		}

		inputs := make([]string, len(inputsRaw))
		for i, v := range inputsRaw {
			s, ok := v.(string)
			if !ok {
				return mcp.NewToolResultError("all inputs must be strings"), nil
			}
			inputs[i] = s
		}

		// Optional parameters with defaults mirroring the CLI.
		followLinks, _ := args["followLinks"].(bool)
		maxDepthFloat, _ := args["maxDepth"].(float64) // JSON numbers => float64
		maxDepth := int(maxDepthFloat)
		skipAnchors, _ := args["skipAnchors"].(bool)
		skipEmbeds, _ := args["skipEmbeds"].(bool)
		suppressTagsRaw, _ := args["suppressTags"].([]interface{})
		noSuppress, _ := args["noSuppress"].(bool)

		// Convert suppressTagsRaw to []string
		var suppressTags []string
		for _, v := range suppressTagsRaw {
			if s, ok := v.(string); ok {
				suppressTags = append(suppressTags, s)
			}
		}

		// Build final suppression list.
		suppressedTags := config.SuppressedTags
		if noSuppress {
			suppressedTags = []string{}
		} else if len(suppressTags) > 0 {
			suppressedTags = append(suppressedTags, suppressTags...)
		}

		if config.Debug {
			log.Printf("MCP prompt_files called with inputs=%v followLinks=%v maxDepth=%d", inputs, followLinks, maxDepth)
		}

		// Parse inputs using existing helper.
		parsedInputs, err := actions.ParseInputs(inputs)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error parsing inputs: %s", err)), nil
		}

		vaultName := config.Vault.Name
		note := obsidian.Note{}

		// Buffer for output.
		var output strings.Builder
		fmt.Fprintf(&output, "<obsidian-vault name=\"%s\">\n\n", vaultName)

		// Use a map to track uniqueness and preserve insertion order using slice.
		unique := make(map[string]bool)
		var order []string

		// Gather files via ListFiles with OnMatch callback.
		_, err = actions.ListFiles(config.Vault, &note, actions.ListParams{
			Inputs:         parsedInputs,
			FollowLinks:    followLinks || maxDepth > 0,
			MaxDepth:       maxDepth,
			SkipAnchors:    skipAnchors,
			SkipEmbeds:     skipEmbeds,
			AbsolutePaths:  false,
			SuppressedTags: suppressedTags,
			OnMatch: func(file string) {
				if !unique[file] {
					unique[file] = true
					order = append(order, file)
				}
			},
		})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error listing files: %s", err)), nil
		}

		// Get vault path for reading contents.
		vaultPath := config.VaultPath

		// Read contents and build output.
		for _, file := range order {
			content, err := note.GetContents(vaultPath, file)
			if err != nil {
				// Skip unreadable files but mention in debug.
				if config.Debug {
					log.Printf("Unable to read file %s: %v", file, err)
				}
				continue
			}
			fmt.Fprintf(&output, "<file path=\"%s\">\n%s\n</file>\n\n", file, content)
		}

		fmt.Fprintf(&output, "</obsidian-vault>")

		return mcp.NewToolResultText(output.String()), nil
	}
}

// DailyNoteTool implements the daily_note MCP tool which returns the content of the (optionally dated) daily note.
func DailyNoteTool(config Config) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		dateStr, _ := args["date"].(string)

		// If no date provided, use today.
		if dateStr == "" {
			dateStr = time.Now().Format("2006-01-02")
		}

		if config.Debug {
			log.Printf("MCP daily_note called with date: %s", dateStr)
		}

		dailyRelPath := fmt.Sprintf("Daily Notes/%s.md", dateStr)
		note := obsidian.Note{}

		content, err := note.GetContents(config.VaultPath, dailyRelPath)
		if err != nil {
			// If file doesn't exist, return empty content rather than error.
			if err.Error() == obsidian.NoteDoesNotExistError {
				content = "" // treat as empty note
			} else {
				return mcp.NewToolResultError(fmt.Sprintf("Error reading daily note: %s", err)), nil
			}
		}

		return mcp.NewToolResultText(content), nil
	}
}

// OpenInOSTool implements the open_in_os MCP tool
func OpenInOSTool(config Config) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		path, ok := args["path"].(string)
		if !ok {
			return mcp.NewToolResultError("path parameter is required"), nil
		}

		if config.Debug {
			log.Printf("MCP open_in_os called with path: %s", path)
		}

		uri := obsidian.Uri{}
		err := actions.OpenNote(config.Vault, &uri, actions.OpenParams{
			NoteName: path,
		})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error opening file: %s", err)), nil
		}

		// Get absolute path for response
		fullPath := filepath.Join(config.VaultPath, path)
		return mcp.NewToolResultText(fmt.Sprintf("Opened file: %s", fullPath)), nil
	}
}

// DeleteTagsTool implements the delete_tags MCP tool (destructive; optional dryRun).
func DeleteTagsTool(config Config) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()

		tagsRaw, ok := args["tags"].([]interface{})
		if !ok {
			return mcp.NewToolResultError("tags parameter is required and must be an array"), nil
		}

		var tags []string
		for _, v := range tagsRaw {
			if s, ok := v.(string); ok {
				tags = append(tags, s)
			} else {
				return mcp.NewToolResultError("all tags must be strings"), nil
			}
		}

		dryRun, _ := args["dryRun"].(bool)

		if !config.ReadWrite && !dryRun {
			return mcp.NewToolResultError("Server is in read-only mode; either enable --read-write or set dryRun=true"), nil
		}

		note := obsidian.Note{}

		summary, err := actions.DeleteTags(config.Vault, &note, tags, dryRun)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error deleting tags: %s", err)), nil
		}

		// Build human-readable summary
		var sb strings.Builder
		if dryRun {
			sb.WriteString("[DRY RUN] ")
		}
		fmt.Fprintf(&sb, "Notes touched: %d\n", summary.NotesTouched)
		if len(summary.TagChanges) > 0 {
			sb.WriteString("Tag deletions:\n")
			for tag, cnt := range summary.TagChanges {
				fmt.Fprintf(&sb, "  – %s: %d note(s)\n", tag, cnt)
			}
		}
		if len(summary.FilesChanged) > 0 {
			sb.WriteString("Files changed:\n")
			for _, f := range summary.FilesChanged {
				fmt.Fprintf(&sb, "  %s\n", f)
			}
		}

		return mcp.NewToolResultText(sb.String()), nil
	}
}

// RenameTagsTool implements the rename_tag MCP tool.
func RenameTagsTool(config Config) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()

		fromRaw, ok := args["fromTags"].([]interface{})
		if !ok {
			return mcp.NewToolResultError("fromTags parameter is required and must be an array"), nil
		}

		var fromTags []string
		for _, v := range fromRaw {
			if s, ok := v.(string); ok {
				fromTags = append(fromTags, s)
			} else {
				return mcp.NewToolResultError("all fromTags values must be strings"), nil
			}
		}

		toTag, ok := args["toTag"].(string)
		if !ok || strings.TrimSpace(toTag) == "" {
			return mcp.NewToolResultError("toTag parameter is required and must be a non-empty string"), nil
		}

		dryRun, _ := args["dryRun"].(bool)

		if !config.ReadWrite && !dryRun {
			return mcp.NewToolResultError("Server is in read-only mode; either enable --read-write or set dryRun=true"), nil
		}

		note := obsidian.Note{}

		summary, err := actions.RenameTags(config.Vault, &note, fromTags, toTag, dryRun)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error renaming tags: %s", err)), nil
		}

		// Build human-readable summary
		var sb strings.Builder
		if dryRun {
			sb.WriteString("[DRY RUN] ")
		}
		fmt.Fprintf(&sb, "Notes touched: %d\n", summary.NotesTouched)
		if len(summary.TagChanges) > 0 {
			sb.WriteString("Tag renames:\n")
			for tag, cnt := range summary.TagChanges {
				fmt.Fprintf(&sb, "  – %s → %s : %d note(s)\n", tag, toTag, cnt)
			}
		}
		if len(summary.FilesChanged) > 0 {
			sb.WriteString("Files changed:\n")
			for _, f := range summary.FilesChanged {
				fmt.Fprintf(&sb, "  %s\n", f)
			}
		}

		return mcp.NewToolResultText(sb.String()), nil
	}
}

// AddTagsTool implements the add_tags MCP tool (destructive; optional dryRun).
func AddTagsTool(config Config) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()

		// Extract tags to add
		tagsRaw, ok := args["tags"].([]interface{})
		if !ok {
			return mcp.NewToolResultError("tags parameter is required and must be an array"), nil
		}

		var tags []string
		for _, v := range tagsRaw {
			if s, ok := v.(string); ok {
				tags = append(tags, s)
			} else {
				return mcp.NewToolResultError("all tags must be strings"), nil
			}
		}

		// Extract input criteria
		inputsRaw, ok := args["inputs"].([]interface{})
		if !ok {
			return mcp.NewToolResultError("inputs parameter is required and must be an array"), nil
		}

		inputs := make([]string, len(inputsRaw))
		for i, v := range inputsRaw {
			s, ok := v.(string)
			if !ok {
				return mcp.NewToolResultError("all inputs must be strings"), nil
			}
			inputs[i] = s
		}

		dryRun, _ := args["dryRun"].(bool)

		if !config.ReadWrite && !dryRun {
			return mcp.NewToolResultError("Server is in read-only mode; either enable --read-write or set dryRun=true"), nil
		}

		// Parse input criteria to get matching files
		parsedInputs, err := actions.ParseInputs(inputs)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error parsing input criteria: %s", err)), nil
		}

		note := obsidian.Note{}

		// Get list of files matching the input criteria
		matchingFiles, err := actions.ListFiles(config.Vault, &note, actions.ListParams{
			Inputs:         parsedInputs,
			FollowLinks:    false,
			MaxDepth:       0,
			SkipAnchors:    false,
			SkipEmbeds:     false,
			AbsolutePaths:  false,
			SuppressedTags: []string{},
		})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error getting matching files: %s", err)), nil
		}

		if len(matchingFiles) == 0 {
			return mcp.NewToolResultText("No files match the specified criteria."), nil
		}

		// Add tags to the specific matching files
		summary, err := actions.AddTagsToFiles(config.Vault, &note, tags, matchingFiles, dryRun)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error adding tags: %s", err)), nil
		}

		// Build human-readable summary
		var sb strings.Builder
		if dryRun {
			sb.WriteString("[DRY RUN] ")
		}
		fmt.Fprintf(&sb, "Notes touched: %d\n", summary.NotesTouched)
		if len(summary.TagChanges) > 0 {
			sb.WriteString("Tag additions:\n")
			for tag, cnt := range summary.TagChanges {
				fmt.Fprintf(&sb, "  + %s: %d note(s)\n", tag, cnt)
			}
		}
		if len(summary.FilesChanged) > 0 {
			sb.WriteString("Files changed:\n")
			for _, f := range summary.FilesChanged {
				fmt.Fprintf(&sb, "  %s\n", f)
			}
		}

		return mcp.NewToolResultText(sb.String()), nil
	}
}

// DailyNotePathTool implements the daily_note_path MCP tool
func DailyNotePathTool(config Config) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		date, _ := args["date"].(string) // Optional parameter

		if config.Debug {
			log.Printf("MCP daily_note_path called with date: %s", date)
		}

		// For daily note path, we need to check if there's a Daily function that returns a path
		// Looking at the daily.go, it seems to open the daily note rather than return the path
		// Let's implement a simple daily note path generator

		// This is a simplified implementation - in a real scenario, you'd want to
		// check Obsidian's daily note settings for the actual format and location
		if date == "" {
			// Use today's date in YYYY-MM-DD format
			date = "today" // Placeholder - would need proper date formatting
		}

		dailyNotePath := fmt.Sprintf("Daily Notes/%s.md", date)
		return mcp.NewToolResultText(dailyNotePath), nil
	}
}
