package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/mark3labs/mcp-go/mcp"
)

// FileEntry is the structured payload returned by the files tool
type FileEntry struct {
	Path         string                 `json:"path"`
	AbsolutePath string                 `json:"absolutePath,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	Frontmatter  map[string]interface{} `json:"frontmatter,omitempty"`
	Content      string                 `json:"content,omitempty"`
	Backlinks    []obsidian.Backlink    `json:"backlinks,omitempty"`
}

// FilesResponse wraps the full files response
type FilesResponse struct {
	Vault string      `json:"vault"`
	Count int         `json:"count"`
	Files []FileEntry `json:"files"`
}

// DailyNoteResponse describes the JSON shape for the daily_note tool
type DailyNoteResponse struct {
	Path    string `json:"path"`
	Date    string `json:"date"`
	Exists  bool   `json:"exists"`
	Content string `json:"content,omitempty"`
}

// DailyNotePathResponse describes the JSON shape for the daily_note_path tool
type DailyNotePathResponse struct {
	Path   string `json:"path"`
	Date   string `json:"date"`
	Exists bool   `json:"exists"`
}

// TagListResponse describes the JSON shape for listing tags
type TagListResponse struct {
	Tags []actions.TagSummary `json:"tags"`
}

// TagMutationResult describes the JSON shape returned by tag mutators
type TagMutationResult struct {
	DryRun       bool           `json:"dryRun,omitempty"`
	NotesTouched int            `json:"notesTouched"`
	TagChanges   map[string]int `json:"tagChanges"`
	FilesChanged []string       `json:"filesChanged,omitempty"`
}

// RenameNoteResponse describes the JSON shape returned by the rename_note tool.
type RenameNoteResponse struct {
	Path                string   `json:"path"`
	LinkUpdates         int      `json:"linkUpdates"`
	Skipped             []string `json:"skipped,omitempty"`
	GitHistoryPreserved bool     `json:"gitHistoryPreserved"`
}

// MoveNotesResponse describes the JSON shape returned by the move_notes MCP tool.
type MoveNotesResponse struct {
	Moves            []MoveNoteEntry `json:"moves"`
	TotalLinkUpdates int             `json:"totalLinkUpdates"`
	Skipped          []string        `json:"skipped,omitempty"`
}

// MoveNoteEntry captures per-note move results.
type MoveNoteEntry struct {
	Source              string `json:"source"`
	Target              string `json:"target"`
	LinkUpdates         int    `json:"linkUpdates"`
	GitHistoryPreserved bool   `json:"gitHistoryPreserved"`
}

// FilesTool implements the files MCP tool (paths + optional content/frontmatter as JSON).
func FilesTool(config Config) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()

		rawInputs, ok := args["inputs"].([]interface{})
		if !ok {
			return mcp.NewToolResultError("inputs parameter is required and must be an array"), nil
		}

		inputs := make([]string, len(rawInputs))
		for i, v := range rawInputs {
			s, ok := v.(string)
			if !ok {
				return mcp.NewToolResultError("all inputs must be strings"), nil
			}
			inputs[i] = s
		}

		followLinks, _ := args["followLinks"].(bool)
		maxDepthFloat, _ := args["maxDepth"].(float64)
		maxDepth := int(maxDepthFloat)
		skipAnchors, _ := args["skipAnchors"].(bool)
		skipEmbeds, _ := args["skipEmbeds"].(bool)

		includeContent := true
		if v, ok := args["includeContent"].(bool); ok {
			includeContent = v
		}
		includeFrontmatter, _ := args["includeFrontmatter"].(bool)
		absolutePaths, _ := args["absolutePaths"].(bool)
		includeBacklinks, _ := args["includeBacklinks"].(bool)

		suppressTagsRaw, _ := args["suppressTags"].([]interface{})
		noSuppress, _ := args["noSuppress"].(bool)

		var suppressTags []string
		for _, v := range suppressTagsRaw {
			if s, ok := v.(string); ok {
				suppressTags = append(suppressTags, s)
			}
		}

		suppressedTags := config.SuppressedTags
		if noSuppress {
			suppressedTags = []string{}
		} else if len(suppressTags) > 0 {
			suppressedTags = append(suppressedTags, suppressTags...)
		}

		if config.Debug {
			log.Printf("MCP files args: inputs=%v followLinks=%v maxDepth=%d includeContent=%v includeFrontmatter=%v", inputs, followLinks, maxDepth, includeContent, includeFrontmatter)
		}

		parsedInputs, err := actions.ParseInputs(inputs)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error parsing inputs: %s", err)), nil
		}

		note := obsidian.Note{}

		unique := make(map[string]bool)
		order := make([]string, 0)

		params := actions.ListParams{
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
		}

		var backlinks map[string][]obsidian.Backlink
		if includeBacklinks {
			params.IncludeBacklinks = true
			params.Backlinks = &backlinks
		}

		var primaryMatches []string
		if includeBacklinks {
			params.PrimaryMatches = &primaryMatches
		}

		_, err = actions.ListFiles(config.Vault, &note, params)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error listing files: %s", err)), nil
		}

		response := FilesResponse{
			Vault: config.Vault.Name,
			Files: make([]FileEntry, 0, len(order)),
		}

		vaultPath := config.VaultPath

		primarySet := make(map[string]struct{})
		for _, p := range primaryMatches {
			primarySet[obsidian.NormalizePath(obsidian.AddMdSuffix(p))] = struct{}{}
		}

		for _, file := range order {
			info, err := actions.GetFileInfo(config.Vault, &note, file)
			if err != nil {
				if config.Debug {
					log.Printf("Unable to get info for %s: %v", file, err)
				}
				continue
			}

			entry := FileEntry{
				Path: file,
				Tags: info.Tags,
			}

			if includeFrontmatter && info.Frontmatter != nil {
				entry.Frontmatter = info.Frontmatter
			}

			if includeContent {
				content, err := note.GetContents(vaultPath, file)
				if err != nil {
					if config.Debug {
						log.Printf("Unable to read file %s: %v", file, err)
					}
					continue
				}
				entry.Content = content
			}

			if absolutePaths {
				entry.AbsolutePath = filepath.Join(vaultPath, file)
			}

			if includeBacklinks {
				key := obsidian.NormalizePath(obsidian.AddMdSuffix(file))
				if _, ok := primarySet[key]; ok {
					if backs, ok := backlinks[key]; ok && len(backs) > 0 {
						entry.Backlinks = backs
					} else if ok {
						entry.Backlinks = []obsidian.Backlink{}
					}
				}
			}

			response.Files = append(response.Files, entry)
		}

		response.Count = len(response.Files)

		encoded, err := json.Marshal(response)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error marshaling response: %s", err)), nil
		}

		return mcp.NewToolResultText(string(encoded)), nil
	}
}

// ListTagsTool implements the list_tags MCP tool.
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

		payload := TagListResponse{Tags: tagSummaries}
		encoded, err := json.Marshal(payload)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error marshaling tag list: %s", err)), nil
		}

		return mcp.NewToolResultText(string(encoded)), nil
	}
}

// RenameNoteTool implements the rename_note MCP tool mirroring CLI behavior.
func RenameNoteTool(config Config) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		source, _ := args["source"].(string)
		target, _ := args["target"].(string)
		overwrite, _ := args["overwrite"].(bool)
		updateBacklinks := true
		if v, ok := args["updateBacklinks"].(bool); ok {
			updateBacklinks = v
		}
		ignored := make([]string, 0)
		if arr, ok := args["ignoredPaths"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					ignored = append(ignored, s)
				}
			}
		}

		if strings.TrimSpace(source) == "" || strings.TrimSpace(target) == "" {
			return mcp.NewToolResultError("source and target are required"), nil
		}

		params := actions.RenameParams{
			Source:          source,
			Target:          target,
			Overwrite:       overwrite,
			UpdateBacklinks: updateBacklinks,
			IgnoredPaths:    ignored,
		}

		result, err := actions.RenameNote(config.Vault, params)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("rename failed: %s", err)), nil
		}

		response := RenameNoteResponse{
			Path:                result.RenamedPath,
			LinkUpdates:         result.LinkUpdates,
			Skipped:             result.Skipped,
			GitHistoryPreserved: result.GitHistoryPreserved,
		}

		encoded, err := json.Marshal(response)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("error marshaling response: %s", err)), nil
		}

		return mcp.NewToolResultText(string(encoded)), nil
	}
}

// MoveNotesTool implements the move_notes MCP tool for single or bulk moves (no backlinks rewritten by default).
func MoveNotesTool(config Config) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()

		moves := make([]actions.MoveRequest, 0)
		if rawMoves, ok := args["moves"]; ok {
			switch mv := rawMoves.(type) {
			case []interface{}:
				for _, raw := range mv {
					obj, ok := raw.(map[string]interface{})
					if !ok {
						return mcp.NewToolResultError("each move must be an object with source and target"), nil
					}
					src, _ := obj["source"].(string)
					dst, _ := obj["target"].(string)
					moves = append(moves, actions.MoveRequest{Source: src, Target: dst})
				}
			case []map[string]interface{}:
				for _, obj := range mv {
					src, _ := obj["source"].(string)
					dst, _ := obj["target"].(string)
					moves = append(moves, actions.MoveRequest{Source: src, Target: dst})
				}
			case []map[string]string:
				for _, obj := range mv {
					moves = append(moves, actions.MoveRequest{Source: obj["source"], Target: obj["target"]})
				}
			default:
				return mcp.NewToolResultError("moves must be an array of objects with source and target"), nil
			}
		} else {
			// Back-compat: accept single source/target pair
			src, _ := args["source"].(string)
			dst, _ := args["target"].(string)
			if strings.TrimSpace(src) != "" && strings.TrimSpace(dst) != "" {
				moves = append(moves, actions.MoveRequest{Source: src, Target: dst})
			}
		}

		if len(moves) == 0 {
			return mcp.NewToolResultError("moves array or source/target pair is required"), nil
		}

		overwrite, _ := args["overwrite"].(bool)
		updateBacklinks, _ := args["updateBacklinks"].(bool)
		shouldOpen, _ := args["open"].(bool)

		uri := obsidian.Uri{}
		summary, err := actions.MoveNotes(config.Vault, &uri, actions.MoveParams{
			Moves:           moves,
			Overwrite:       overwrite,
			UpdateBacklinks: updateBacklinks,
			ShouldOpen:      shouldOpen,
		})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("move failed: %s", err)), nil
		}

		resp := MoveNotesResponse{
			TotalLinkUpdates: summary.TotalLinkUpdates,
			Skipped:          summary.Skipped,
		}
		for _, res := range summary.Results {
			resp.Moves = append(resp.Moves, MoveNoteEntry{
				Source:              res.Source,
				Target:              res.Target,
				LinkUpdates:         res.LinkUpdates,
				GitHistoryPreserved: res.GitHistoryPreserved,
			})
		}

		encoded, err := json.Marshal(resp)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("error marshaling response: %s", err)), nil
		}

		return mcp.NewToolResultText(string(encoded)), nil
	}
}

// DailyNoteTool implements the daily_note MCP tool which returns JSON describing the daily note.
func DailyNoteTool(config Config) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		dateStr, _ := args["date"].(string)

		if dateStr == "" {
			dateStr = time.Now().Format("2006-01-02")
		}

		if config.Debug {
			log.Printf("MCP daily_note called with date: %s", dateStr)
		}

		dailyRelPath := fmt.Sprintf("Daily Notes/%s.md", dateStr)
		note := obsidian.Note{}

		content, err := note.GetContents(config.VaultPath, dailyRelPath)
		exists := true
		if err != nil {
			if err.Error() == obsidian.NoteDoesNotExistError {
				exists = false
				content = ""
			} else {
				return mcp.NewToolResultError(fmt.Sprintf("Error reading daily note: %s", err)), nil
			}
		}

		payload := DailyNoteResponse{
			Path:    dailyRelPath,
			Date:    dateStr,
			Exists:  exists,
			Content: content,
		}

		encoded, err := json.Marshal(payload)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error marshaling daily note: %s", err)), nil
		}

		return mcp.NewToolResultText(string(encoded)), nil
	}
}

// DailyNotePathTool implements the daily_note_path MCP tool.
func DailyNotePathTool(config Config) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		date := strings.TrimSpace(fmt.Sprint(args["date"]))

		if date == "" || date == "<nil>" {
			date = time.Now().Format("2006-01-02")
		}

		if config.Debug {
			log.Printf("MCP daily_note_path called with date: %s", date)
		}

		dailyNotePath := fmt.Sprintf("Daily Notes/%s.md", date)

		_, err := os.Stat(filepath.Join(config.VaultPath, dailyNotePath))
		exists := err == nil

		payload := DailyNotePathResponse{
			Path:   dailyNotePath,
			Date:   date,
			Exists: exists,
		}

		encoded, err := json.Marshal(payload)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error marshaling daily note path: %s", err)), nil
		}

		return mcp.NewToolResultText(string(encoded)), nil
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

		result := TagMutationResult{
			DryRun:       dryRun,
			NotesTouched: summary.NotesTouched,
			TagChanges:   summary.TagChanges,
			FilesChanged: summary.FilesChanged,
		}

		encoded, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error marshaling delete_tags result: %s", err)), nil
		}

		return mcp.NewToolResultText(string(encoded)), nil
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

		result := TagMutationResult{
			DryRun:       dryRun,
			NotesTouched: summary.NotesTouched,
			TagChanges:   summary.TagChanges,
			FilesChanged: summary.FilesChanged,
		}

		encoded, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error marshaling rename_tag result: %s", err)), nil
		}

		return mcp.NewToolResultText(string(encoded)), nil
	}
}

// AddTagsTool implements the add_tags MCP tool (destructive; optional dryRun).
func AddTagsTool(config Config) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

		parsedInputs, err := actions.ParseInputs(inputs)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error parsing input criteria: %s", err)), nil
		}

		note := obsidian.Note{}

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

		summary, err := actions.AddTagsToFiles(config.Vault, &note, tags, matchingFiles, dryRun)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error adding tags: %s", err)), nil
		}

		result := TagMutationResult{
			DryRun:       dryRun,
			NotesTouched: summary.NotesTouched,
			TagChanges:   summary.TagChanges,
			FilesChanged: summary.FilesChanged,
		}

		encoded, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error marshaling add_tags result: %s", err)), nil
		}

		return mcp.NewToolResultText(string(encoded)), nil
	}
}
