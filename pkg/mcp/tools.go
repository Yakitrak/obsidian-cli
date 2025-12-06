package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/atomicobject/obsidian-cli/pkg/actions"
	"github.com/atomicobject/obsidian-cli/pkg/cache"
	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
	"github.com/mark3labs/mcp-go/mcp"
	"gopkg.in/yaml.v3"
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

// PropertyListResponse describes the JSON shape for listing properties
type PropertyListResponse struct {
	Properties []actions.PropertySummary `json:"properties"`
}

// GraphNodePayload captures node-level metrics for MCP clients.
type GraphNodePayload struct {
	Path      string   `json:"path"`
	Title     string   `json:"title"`
	Inbound   int      `json:"inbound"`
	Outbound  int      `json:"outbound"`
	Pagerank  float64  `json:"pagerank"`
	Community string   `json:"community"`
	SCC       string   `json:"scc"`
	Neighbors []string `json:"neighbors,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	WeakComp  string   `json:"weakComponent,omitempty"`
}

// GraphCommunityPayload summarizes a community.
type GraphCommunityPayload struct {
	ID          string              `json:"id"`
	Nodes       []string            `json:"nodes"`
	TopTags     []obsidian.TagCount `json:"topTags,omitempty"`
	TopPagerank []string            `json:"topPagerank,omitempty"`
	Anchor      string              `json:"anchor,omitempty"`
	Density     float64             `json:"density,omitempty"`
	Bridges     []string            `json:"bridges,omitempty"`
}

// CommunityListResponse summarizes communities.
type CommunityListResponse struct {
	Communities []GraphCommunityPayload    `json:"communities"`
	Stats       obsidian.GraphStatsSummary `json:"stats"`
}

// CommunityDetailResponse provides full detail for a single community.
type CommunityDetailResponse struct {
	ID            string              `json:"id"`
	Anchor        string              `json:"anchor,omitempty"`
	Size          int                 `json:"size"`
	Density       float64             `json:"density,omitempty"`
	Bridges       []string            `json:"bridges,omitempty"`
	TopTags       []obsidian.TagCount `json:"topTags,omitempty"`
	TopPagerank   []string            `json:"topPagerank,omitempty"`
	Members       []GraphNodePayload  `json:"members"`
	InternalEdges int                 `json:"internalEdges,omitempty"`
}

// OrphansResponse describes orphaned note paths.
type OrphansResponse struct {
	Orphans []string `json:"orphans"`
}

// TagMutationResult describes the JSON shape returned by tag mutators
type TagMutationResult struct {
	DryRun       bool           `json:"dryRun,omitempty"`
	NotesTouched int            `json:"notesTouched"`
	TagChanges   map[string]int `json:"tagChanges"`
	FilesChanged []string       `json:"filesChanged,omitempty"`
}

// PropertyMutationResult describes the JSON shape returned by property mutators.
type PropertyMutationResult struct {
	DryRun          bool           `json:"dryRun,omitempty"`
	NotesTouched    int            `json:"notesTouched"`
	PropertyChanges map[string]int `json:"propertyChanges"`
	FilesChanged    []string       `json:"filesChanged,omitempty"`
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

func resolveNoteManager(config Config) obsidian.NoteManager {
	if config.Cache != nil {
		return cache.NewNoteAdapter(config.Cache, &obsidian.Note{})
	}
	return &obsidian.Note{}
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
			log.Printf("MCP files args: inputs=%v maxDepth=%d includeContent=%v includeFrontmatter=%v", inputs, maxDepth, includeContent, includeFrontmatter)
		}

		parsedInputs, expr, err := actions.ParseInputsWithExpression(inputs)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error parsing inputs: %s", err)), nil
		}

		note := resolveNoteManager(config)

		unique := make(map[string]bool)
		order := make([]string, 0)

		params := actions.ListParams{
			Inputs:         parsedInputs,
			MaxDepth:       maxDepth,
			SkipAnchors:    skipAnchors,
			SkipEmbeds:     skipEmbeds,
			AbsolutePaths:  false,
			Expression:     expr,
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
			params.AnalysisCache = config.AnalysisCache
		}

		var primaryMatches []string
		if includeBacklinks {
			params.PrimaryMatches = &primaryMatches
		}

		_, err = actions.ListFiles(config.Vault, note, params)
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
			info, err := actions.GetFileInfo(config.Vault, note, file)
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

func parseMatchPatterns(raw interface{}) ([]string, error) {
	if raw == nil {
		return nil, nil
	}
	items, ok := raw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("match must be an array of strings")
	}
	out := make([]string, 0, len(items))
	for _, v := range items {
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("match must be an array of strings")
		}
		out = append(out, s)
	}
	return out, nil
}

// ListTagsTool implements the list_tags MCP tool.
func ListTagsTool(config Config) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if config.Debug {
			log.Printf("MCP list_tags called")
		}

		args := request.GetArguments()
		inputs, err := parseMatchPatterns(args["match"])
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		var scanNotes []string
		if len(inputs) > 0 {
			parsed, expr, err := actions.ParseInputsWithExpression(inputs)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Error parsing inputs: %s", err)), nil
			}
			note := resolveNoteManager(config)
			matchingFiles, err := actions.ListFiles(config.Vault, note, actions.ListParams{
				Inputs:         parsed,
				Expression:     expr,
				MaxDepth:       0,
				SkipAnchors:    false,
				SkipEmbeds:     false,
				AbsolutePaths:  false,
				SuppressedTags: []string{},
			})
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Error filtering files: %s", err)), nil
			}
			scanNotes = matchingFiles
		}

		note := resolveNoteManager(config)
		tagSummaries, err := actions.Tags(config.Vault, note, actions.TagsOptions{Notes: scanNotes})
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

// ListPropertiesTool implements the list_properties MCP tool.
func ListPropertiesTool(config Config) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if config.Debug {
			log.Printf("MCP list_properties called")
		}

		args := request.GetArguments()
		excludeTags, _ := args["excludeTags"].(bool)
		sourceArg, _ := args["source"].(string)
		var source actions.PropertySource
		switch sourceArg {
		case "", "all":
			source = actions.PropertySourceAll
		case "frontmatter":
			source = actions.PropertySourceFrontmatter
		case "inline":
			source = actions.PropertySourceInline
		default:
			return mcp.NewToolResultError(fmt.Sprintf("invalid source value %q: must be all, frontmatter, or inline", sourceArg)), nil
		}
		inputs, err := parseMatchPatterns(args["match"])
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		onlyProps := make([]string, 0)
		if raw, ok := args["only"]; ok {
			switch v := raw.(type) {
			case []interface{}:
				for _, item := range v {
					if s, ok := item.(string); ok {
						onlyProps = append(onlyProps, s)
					}
				}
			case []string:
				onlyProps = v
			case string:
				onlyProps = append(onlyProps, v)
			}
		}

		valueLimit := 25
		valueLimitSet := false
		if v, ok := args["valueLimit"].(float64); ok {
			valueLimit = int(v)
			valueLimitSet = true
		} else if v, ok := args["enumThreshold"].(float64); ok { // backward compatibility
			valueLimit = int(v)
			valueLimitSet = true
		}

		maxValues := 500
		if v, ok := args["maxValues"].(float64); ok {
			maxValues = int(v)
		}
		if maxValues <= 0 {
			maxValues = 500
		}

		includeValueCounts := true
		if v, ok := args["valueCounts"].(bool); ok {
			includeValueCounts = v
		} else if v, ok := args["includeEnumCounts"].(bool); ok { // backward compatibility
			includeValueCounts = v
		}

		forceEnumMixed := false
		if v, ok := args["verbose"].(bool); ok && v {
			forceEnumMixed = true
			if valueLimit < 50 {
				valueLimit = 50
			}
		}
		if len(onlyProps) > 0 && !valueLimitSet {
			if maxValues > 1 {
				valueLimit = maxValues - 1
			} else {
				valueLimit = maxValues
			}
		}
		if maxValues < valueLimit+1 {
			maxValues = valueLimit + 1
		}

		var scanNotes []string
		if len(inputs) > 0 {
			parsed, expr, err := actions.ParseInputsWithExpression(inputs)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Error parsing inputs: %s", err)), nil
			}
			note := resolveNoteManager(config)
			matchingFiles, err := actions.ListFiles(config.Vault, note, actions.ListParams{
				Inputs:         parsed,
				Expression:     expr,
				MaxDepth:       0,
				SkipAnchors:    false,
				SkipEmbeds:     false,
				AbsolutePaths:  false,
				SuppressedTags: []string{},
			})
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Error filtering files: %s", err)), nil
			}
			scanNotes = matchingFiles
		}

		note := resolveNoteManager(config)
		summaries, err := actions.Properties(config.Vault, note, actions.PropertiesOptions{
			ExcludeTags:        excludeTags,
			Source:             source,
			ValueLimit:         valueLimit,
			MaxValues:          maxValues,
			Notes:              scanNotes,
			Only:               onlyProps,
			ForceEnumMixed:     forceEnumMixed,
			IncludeValueCounts: includeValueCounts,
		})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error listing properties: %s", err)), nil
		}

		payload := PropertyListResponse{Properties: summaries}
		encoded, err := json.Marshal(payload)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error marshaling property list: %s", err)), nil
		}

		return mcp.NewToolResultText(string(encoded)), nil
	}
}

// GraphStatsTool has been removed; kept for backward compatibility to return an error.
func GraphStatsTool(config Config) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultError("graph_stats removed; use community_list/community_detail"), nil
	}
}

// CommunityListTool returns community summaries only.
func CommunityListTool(config Config) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		skipAnchors, _ := args["skipAnchors"].(bool)
		skipEmbeds, _ := args["skipEmbeds"].(bool)
		includeTags, _ := args["includeTags"].(bool)
		minDegree := 0
		if v, ok := args["minDegree"].(float64); ok {
			minDegree = int(v)
		}
		mutualOnly, _ := args["mutualOnly"].(bool)
		var exclude []string
		var include []string
		if raw, ok := args["exclude"].([]interface{}); ok {
			for _, v := range raw {
				if s, ok := v.(string); ok {
					exclude = append(exclude, s)
				}
			}
		} else if raw, ok := args["exclude"].([]string); ok {
			exclude = raw
		}
		if raw, ok := args["include"].([]interface{}); ok {
			for _, v := range raw {
				if s, ok := v.(string); ok {
					include = append(include, s)
				}
			}
		} else if raw, ok := args["include"].([]string); ok {
			include = raw
		}
		exclude = actions.ExpandPatterns(exclude)
		include = actions.ExpandPatterns(include)

		analysis, err := actions.GraphAnalysis(config.Vault, resolveNoteManager(config), actions.GraphAnalysisParams{
			UseConfig: true,
			Options: obsidian.GraphAnalysisOptions{
				WikilinkOptions: obsidian.WikilinkOptions{
					SkipAnchors: skipAnchors,
					SkipEmbeds:  skipEmbeds,
				},
				IncludeTags: includeTags,
				MinDegree:   minDegree,
				MutualOnly:  mutualOnly,
			},
			ExcludePatterns: exclude,
			IncludePatterns: include,
			AnalysisCache:   config.AnalysisCache,
		})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("error computing communities: %s", err)), nil
		}

		maxCommunities := 25
		if v, ok := args["maxCommunities"].(float64); ok && int(v) > 0 {
			maxCommunities = int(v)
		}
		maxTopNotes := 5
		if v, ok := args["maxTopNotes"].(float64); ok && int(v) > 0 {
			maxTopNotes = int(v)
		}

		var comms []GraphCommunityPayload
		for idx, comm := range analysis.Communities {
			if idx >= maxCommunities {
				break
			}
			topPagerank := comm.TopPagerank
			if len(topPagerank) > maxTopNotes {
				topPagerank = topPagerank[:maxTopNotes]
			}
			comms = append(comms, GraphCommunityPayload{
				ID:          comm.ID,
				Nodes:       nil, // omit members from list response
				TopTags:     comm.TopTags,
				TopPagerank: topPagerank,
				Anchor:      comm.Anchor,
				Density:     comm.Density,
				Bridges:     comm.Bridges,
			})
		}

		resp := CommunityListResponse{
			Communities: comms,
			Stats:       analysis.Stats,
		}
		encoded, err := json.Marshal(resp)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("error marshaling community list: %s", err)), nil
		}
		return mcp.NewToolResultText(string(encoded)), nil
	}
}

// CommunityDetailTool returns full detail for a specific community.
func CommunityDetailTool(config Config) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		id, _ := args["id"].(string)
		file, _ := args["file"].(string)
		if strings.TrimSpace(id) == "" && strings.TrimSpace(file) == "" {
			return mcp.NewToolResultError("id or file is required"), nil
		}
		if strings.TrimSpace(id) != "" && strings.TrimSpace(file) != "" {
			return mcp.NewToolResultError("provide only one of id or file"), nil
		}
		skipAnchors, _ := args["skipAnchors"].(bool)
		skipEmbeds, _ := args["skipEmbeds"].(bool)
		includeTags, _ := args["includeTags"].(bool)
		includeNeighbors, _ := args["includeNeighbors"].(bool)
		minDegree := 0
		if v, ok := args["minDegree"].(float64); ok {
			minDegree = int(v)
		}
		mutualOnly, _ := args["mutualOnly"].(bool)
		var exclude []string
		var include []string
		if raw, ok := args["exclude"].([]interface{}); ok {
			for _, v := range raw {
				if s, ok := v.(string); ok {
					exclude = append(exclude, s)
				}
			}
		} else if raw, ok := args["exclude"].([]string); ok {
			exclude = raw
		}
		if raw, ok := args["include"].([]interface{}); ok {
			for _, v := range raw {
				if s, ok := v.(string); ok {
					include = append(include, s)
				}
			}
		} else if raw, ok := args["include"].([]string); ok {
			include = raw
		}
		exclude = actions.ExpandPatterns(exclude)
		include = actions.ExpandPatterns(include)

		limit := 0
		if v, ok := args["limit"].(float64); ok && int(v) > 0 {
			limit = int(v)
		}

		analysis, err := actions.GraphAnalysis(config.Vault, resolveNoteManager(config), actions.GraphAnalysisParams{
			UseConfig: true,
			Options: obsidian.GraphAnalysisOptions{
				WikilinkOptions: obsidian.WikilinkOptions{
					SkipAnchors: skipAnchors,
					SkipEmbeds:  skipEmbeds,
				},
				IncludeTags: includeTags,
				MinDegree:   minDegree,
				MutualOnly:  mutualOnly,
			},
			ExcludePatterns: exclude,
			IncludePatterns: include,
			AnalysisCache:   config.AnalysisCache,
		})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("error computing graph: %s", err)), nil
		}

		var target *obsidian.CommunitySummary
		if strings.TrimSpace(id) != "" {
			for i := range analysis.Communities {
				if analysis.Communities[i].ID == id {
					target = &analysis.Communities[i]
					break
				}
			}
			if target == nil {
				return mcp.NewToolResultError(fmt.Sprintf("community %s not found under current filters", id)), nil
			}
		} else {
			normalized := obsidian.NormalizePath(obsidian.AddMdSuffix(file))
			if filepath.IsAbs(file) && config.VaultPath != "" {
				if rel, err := filepath.Rel(config.VaultPath, file); err == nil {
					normalized = obsidian.NormalizePath(obsidian.AddMdSuffix(rel))
				}
			}
			if _, ok := analysis.Nodes[normalized]; !ok {
				return mcp.NewToolResultError(fmt.Sprintf("file %s not found in graph (use vault-relative path and check include/exclude filters)", file)), nil
			}
			lookup := obsidian.CommunityMembershipLookup(analysis.Communities)
			target = lookup[normalized]
			if target == nil {
				return mcp.NewToolResultError(fmt.Sprintf("file %s is not assigned to a community under current filters", file)), nil
			}
		}

		resp := communityDetailPayload(target, analysis, includeTags, includeNeighbors, limit)

		encoded, err := json.Marshal(resp)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("error marshaling community detail: %s", err)), nil
		}
		return mcp.NewToolResultText(string(encoded)), nil
	}
}

type rankedMember struct {
	path  string
	title string
	pr    float64
	in    int
	out   int
	tags  []string
}

func rankMembers(members []string, nodes map[string]obsidian.GraphNode) []rankedMember {
	var list []rankedMember
	for _, p := range members {
		n := nodes[p]
		list = append(list, rankedMember{
			path:  p,
			title: n.Title,
			pr:    n.Pagerank,
			in:    n.Inbound,
			out:   n.Outbound,
			tags:  n.Tags,
		})
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].pr == list[j].pr {
			return list[i].path < list[j].path
		}
		return list[i].pr > list[j].pr
	})
	return list
}

func communityDetailPayload(target *obsidian.CommunitySummary, analysis *obsidian.GraphAnalysis, includeTags bool, includeNeighbors bool, limit int) CommunityDetailResponse {
	edgeCount := obsidian.CommunityInternalEdges(target, analysis.Nodes)

	members := rankMembers(target.Nodes, analysis.Nodes)
	if limit > 0 && limit < len(members) {
		members = members[:limit]
	}

	payloadMembers := make([]GraphNodePayload, 0, len(members))
	for _, m := range members {
		payload := GraphNodePayload{
			Path:      m.path,
			Title:     m.title,
			Inbound:   m.in,
			Outbound:  m.out,
			Pagerank:  m.pr,
			Community: target.ID,
			Tags:      m.tags,
			WeakComp:  analysis.Nodes[m.path].WeakCompID,
			SCC:       analysis.Nodes[m.path].SCC,
		}
		if includeNeighbors {
			payload.Neighbors = analysis.Nodes[m.path].Neighbors
		}
		if !includeTags {
			payload.Tags = nil
		}
		payloadMembers = append(payloadMembers, payload)
	}

	return CommunityDetailResponse{
		ID:            target.ID,
		Anchor:        target.Anchor,
		Size:          len(target.Nodes),
		Density:       target.Density,
		Bridges:       target.Bridges,
		TopTags:       target.TopTags,
		TopPagerank:   target.TopPagerank,
		Members:       payloadMembers,
		InternalEdges: edgeCount,
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
		updateBacklinks := true
		if v, ok := args["updateBacklinks"].(bool); ok {
			updateBacklinks = v
		}
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
		note := resolveNoteManager(config)

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

		note := resolveNoteManager(config)

		summary, err := actions.DeleteTags(config.Vault, note, tags, dryRun)
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

		note := resolveNoteManager(config)

		summary, err := actions.RenameTags(config.Vault, note, fromTags, toTag, dryRun)
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

func SetPropertyTool(config Config) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()

		property, ok := args["property"].(string)
		if !ok || strings.TrimSpace(property) == "" {
			return mcp.NewToolResultError("property parameter is required and must be a non-empty string"), nil
		}

		valueRaw, ok := args["value"].(string)
		if !ok {
			return mcp.NewToolResultError("value parameter is required and must be a string (YAML accepted)"), nil
		}

		inputsRaw, ok := args["inputs"].([]interface{})
		if !ok {
			return mcp.NewToolResultError("inputs parameter is required and must be an array"), nil
		}

		dryRun, _ := args["dryRun"].(bool)
		overwrite, _ := args["overwrite"].(bool)

		if !config.ReadWrite && !dryRun {
			return mcp.NewToolResultError("Server is in read-only mode; either enable --read-write or set dryRun=true"), nil
		}

		var value interface{}
		if err := yaml.Unmarshal([]byte(valueRaw), &value); err != nil {
			value = valueRaw
		}

		inputs := make([]string, len(inputsRaw))
		for i, v := range inputsRaw {
			s, ok := v.(string)
			if !ok {
				return mcp.NewToolResultError("all inputs must be strings"), nil
			}
			inputs[i] = s
		}

		parsedInputs, expr, err := actions.ParseInputsWithExpression(inputs)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error parsing input criteria: %s", err)), nil
		}

		note := resolveNoteManager(config)

		matchingFiles, err := actions.ListFiles(config.Vault, note, actions.ListParams{
			Inputs:        parsedInputs,
			MaxDepth:      0,
			SkipAnchors:   false,
			SkipEmbeds:    false,
			AbsolutePaths: false,
			Expression:    expr,
		})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error getting matching files: %s", err)), nil
		}
		if len(matchingFiles) == 0 {
			return mcp.NewToolResultError("No files match the specified criteria"), nil
		}

		summary, err := actions.SetPropertyOnFiles(config.Vault, note, property, value, matchingFiles, overwrite, dryRun)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error setting property: %s", err)), nil
		}

		result := PropertyMutationResult{
			DryRun:          dryRun,
			NotesTouched:    summary.NotesTouched,
			PropertyChanges: summary.PropertyChanges,
			FilesChanged:    summary.FilesChanged,
		}

		encoded, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error marshaling set_property result: %s", err)), nil
		}

		return mcp.NewToolResultText(string(encoded)), nil
	}
}

func DeletePropertiesTool(config Config) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()

		propsRaw, ok := args["properties"].([]interface{})
		if !ok {
			return mcp.NewToolResultError("properties parameter is required and must be an array"), nil
		}
		var properties []string
		for _, v := range propsRaw {
			s, ok := v.(string)
			if !ok {
				return mcp.NewToolResultError("all properties must be strings"), nil
			}
			properties = append(properties, s)
		}

		var files []string
		if inputsRaw, ok := args["inputs"].([]interface{}); ok {
			inputs := make([]string, len(inputsRaw))
			for i, v := range inputsRaw {
				s, ok := v.(string)
				if !ok {
					return mcp.NewToolResultError("all inputs must be strings"), nil
				}
				inputs[i] = s
			}
			parsedInputs, expr, err := actions.ParseInputsWithExpression(inputs)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Error parsing input criteria: %s", err)), nil
			}

			note := resolveNoteManager(config)
			files, err = actions.ListFiles(config.Vault, note, actions.ListParams{
				Inputs:        parsedInputs,
				MaxDepth:      0,
				SkipAnchors:   false,
				SkipEmbeds:    false,
				AbsolutePaths: false,
				Expression:    expr,
			})
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Error getting matching files: %s", err)), nil
			}
			if len(files) == 0 {
				return mcp.NewToolResultError("No files match the specified criteria"), nil
			}
		}

		dryRun, _ := args["dryRun"].(bool)
		if !config.ReadWrite && !dryRun {
			return mcp.NewToolResultError("Server is in read-only mode; either enable --read-write or set dryRun=true"), nil
		}

		note := resolveNoteManager(config)

		summary, err := actions.DeleteProperties(config.Vault, note, properties, files, dryRun)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error deleting properties: %s", err)), nil
		}

		result := PropertyMutationResult{
			DryRun:          dryRun,
			NotesTouched:    summary.NotesTouched,
			PropertyChanges: summary.PropertyChanges,
			FilesChanged:    summary.FilesChanged,
		}

		encoded, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error marshaling delete_properties result: %s", err)), nil
		}

		return mcp.NewToolResultText(string(encoded)), nil
	}
}

func RenamePropertyTool(config Config) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()

		fromRaw, ok := args["fromProperties"].([]interface{})
		if !ok {
			return mcp.NewToolResultError("fromProperties parameter is required and must be an array"), nil
		}
		var from []string
		for _, v := range fromRaw {
			s, ok := v.(string)
			if !ok {
				return mcp.NewToolResultError("all fromProperties values must be strings"), nil
			}
			from = append(from, s)
		}

		to, ok := args["toProperty"].(string)
		if !ok || strings.TrimSpace(to) == "" {
			return mcp.NewToolResultError("toProperty parameter is required and must be a non-empty string"), nil
		}

		var files []string
		if inputsRaw, ok := args["inputs"].([]interface{}); ok {
			inputs := make([]string, len(inputsRaw))
			for i, v := range inputsRaw {
				s, ok := v.(string)
				if !ok {
					return mcp.NewToolResultError("all inputs must be strings"), nil
				}
				inputs[i] = s
			}

			parsedInputs, expr, err := actions.ParseInputsWithExpression(inputs)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Error parsing input criteria: %s", err)), nil
			}

			note := resolveNoteManager(config)
			files, err = actions.ListFiles(config.Vault, note, actions.ListParams{
				Inputs:        parsedInputs,
				MaxDepth:      0,
				SkipAnchors:   false,
				SkipEmbeds:    false,
				AbsolutePaths: false,
				Expression:    expr,
			})
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Error getting matching files: %s", err)), nil
			}
			if len(files) == 0 {
				return mcp.NewToolResultError("No files match the specified criteria"), nil
			}
		}

		merge := true
		if v, ok := args["merge"].(bool); ok {
			merge = v
		}
		dryRun, _ := args["dryRun"].(bool)

		if !config.ReadWrite && !dryRun {
			return mcp.NewToolResultError("Server is in read-only mode; either enable --read-write or set dryRun=true"), nil
		}

		note := resolveNoteManager(config)

		summary, err := actions.RenameProperties(config.Vault, note, from, to, merge, files, dryRun)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error renaming properties: %s", err)), nil
		}

		result := PropertyMutationResult{
			DryRun:          dryRun,
			NotesTouched:    summary.NotesTouched,
			PropertyChanges: summary.PropertyChanges,
			FilesChanged:    summary.FilesChanged,
		}

		encoded, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error marshaling rename_property result: %s", err)), nil
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

		parsedInputs, expr, err := actions.ParseInputsWithExpression(inputs)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error parsing input criteria: %s", err)), nil
		}

		note := resolveNoteManager(config)

		matchingFiles, err := actions.ListFiles(config.Vault, note, actions.ListParams{
			Inputs:         parsedInputs,
			MaxDepth:       0,
			SkipAnchors:    false,
			SkipEmbeds:     false,
			AbsolutePaths:  false,
			Expression:     expr,
			SuppressedTags: []string{},
		})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error getting matching files: %s", err)), nil
		}

		summary, err := actions.AddTagsToFiles(config.Vault, note, tags, matchingFiles, dryRun)
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
