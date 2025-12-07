package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
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

// BridgePayload captures cross-community bridge strength for a node.
type BridgePayload struct {
	Path                string `json:"path"`
	CrossCommunityEdges int    `json:"crossCommunityEdges,omitempty"`
}

// GraphNodePayload captures node-level metrics for MCP clients.
type GraphNodePayload struct {
	Path        string   `json:"path"`
	Title       string   `json:"title"`
	Inbound     int      `json:"inbound"`
	Outbound    int      `json:"outbound"`
	Hub         float64  `json:"hub"`       // HITS hub score: measures how well this note curates/aggregates links
	Authority   float64  `json:"authority"` // HITS authority score: measures how often this note is referenced
	Community   string   `json:"community,omitempty"`
	SCC         string   `json:"scc"`
	Neighbors   []string `json:"neighbors,omitempty"`
	LinksOut    []string `json:"linksOut,omitempty"`
	LinksIn     []string `json:"linksIn,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	WeakComp    string   `json:"weakComponent,omitempty"`
	IsBridge    bool     `json:"isBridge,omitempty"`
	BridgeEdges int      `json:"bridgeEdges,omitempty"`
}

// AuthorityScorePayload carries a note and its hub/authority scores.
type AuthorityScorePayload struct {
	Path      string  `json:"path"`
	Authority float64 `json:"authority"`
	Hub       float64 `json:"hub,omitempty"`
}

// AuthorityBucketPayload summarizes authority distribution for a community.
type AuthorityBucketPayload struct {
	Low     float64 `json:"low,omitempty"`
	High    float64 `json:"high,omitempty"`
	Count   int     `json:"count,omitempty"`
	Example string  `json:"example,omitempty"`
}

// GraphRecencyPayload summarizes modification recency for a community.
type GraphRecencyPayload struct {
	LatestPath    string  `json:"latestPath,omitempty"`
	LatestAgeDays float64 `json:"latestAgeDays,omitempty"`
	RecentCount   int     `json:"recentCount,omitempty"`
	WindowDays    int     `json:"windowDays,omitempty"`
}

// AuthorityStatsPayload captures coarse percentiles/mean for authority scores.
type AuthorityStatsPayload struct {
	Mean float64 `json:"mean,omitempty"`
	P50  float64 `json:"p50,omitempty"`
	P75  float64 `json:"p75,omitempty"`
	P90  float64 `json:"p90,omitempty"`
	P95  float64 `json:"p95,omitempty"`
	P99  float64 `json:"p99,omitempty"`
	Max  float64 `json:"max,omitempty"`
}

// GraphCommunityPayload summarizes a community.
type GraphCommunityPayload struct {
	ID               string                   `json:"id"`
	Size             int                      `json:"size"`
	FractionOfVault  float64                  `json:"fractionOfVault,omitempty"`
	Nodes            []string                 `json:"nodes,omitempty"`
	TopTags          []obsidian.TagCount      `json:"topTags,omitempty"`
	TopAuthority     []AuthorityScorePayload  `json:"topAuthority,omitempty"`
	AuthorityBuckets []AuthorityBucketPayload `json:"authorityBuckets,omitempty"`
	AuthorityStats   *AuthorityStatsPayload   `json:"authorityStats,omitempty"`
	Recency          *GraphRecencyPayload     `json:"recency,omitempty"`
	Anchor           string                   `json:"anchor,omitempty"`
	Density          float64                  `json:"density,omitempty"`
	Bridges          []string                 `json:"bridges,omitempty"`
	BridgesDetailed  []BridgePayload          `json:"bridgesDetailed,omitempty"`
}

// CommunityListResponse summarizes communities.
type CommunityListResponse struct {
	Communities []GraphCommunityPayload    `json:"communities"`
	Stats       obsidian.GraphStatsSummary `json:"stats"`
	OrphanCount int                        `json:"orphanCount,omitempty"`
	Orphans     []string                   `json:"orphans,omitempty"`
	Components  []ComponentSummary         `json:"components,omitempty"`
}

// ComponentSummary captures weak component sizes for global structure awareness.
type ComponentSummary struct {
	ID              string  `json:"id"`
	Size            int     `json:"size"`
	FractionOfVault float64 `json:"fractionOfVault,omitempty"`
}

// CommunityDetailResponse provides full detail for a single community.
type CommunityDetailResponse struct {
	ID               string                   `json:"id"`
	Anchor           string                   `json:"anchor,omitempty"`
	Size             int                      `json:"size"`
	FractionOfVault  float64                  `json:"fractionOfVault,omitempty"`
	Density          float64                  `json:"density,omitempty"`
	Bridges          []string                 `json:"bridges,omitempty"`
	BridgesDetailed  []BridgePayload          `json:"bridgesDetailed,omitempty"`
	TopTags          []obsidian.TagCount      `json:"topTags,omitempty"`
	TopAuthority     []AuthorityScorePayload  `json:"topAuthority,omitempty"`
	AuthorityBuckets []AuthorityBucketPayload `json:"authorityBuckets,omitempty"`
	AuthorityStats   *AuthorityStatsPayload   `json:"authorityStats,omitempty"`
	Recency          *GraphRecencyPayload     `json:"recency,omitempty"`
	Members          []GraphNodePayload       `json:"members"`
	InternalEdges    int                      `json:"internalEdges,omitempty"`
}

// OrphansResponse describes orphaned note paths.
type OrphansResponse struct {
	Orphans []string `json:"orphans"`
}

// NeighborRef captures a neighbor path with its community for richer context.
type NeighborRef struct {
	Path      string `json:"path"`
	Community string `json:"community,omitempty"`
}

// NoteGraphContext summarizes graph metrics for a single note.
type NoteGraphContext struct {
	Inbound             int     `json:"inbound"`
	Outbound            int     `json:"outbound"`
	Hub                 float64 `json:"hub"`                           // HITS hub score
	HubPercentile       float64 `json:"hubPercentile,omitempty"`       // Percentile rank for hub score
	Authority           float64 `json:"authority"`                     // HITS authority score
	AuthorityPercentile float64 `json:"authorityPercentile,omitempty"` // Percentile rank for authority score
	IsOrphan            bool    `json:"isOrphan"`
	WeakComponent       string  `json:"weakComponent,omitempty"`
	StrongComponent     string  `json:"strongComponent,omitempty"`
}

// NoteCommunityContext captures the community around a note.
type NoteCommunityContext struct {
	ID               string                   `json:"id"`
	Size             int                      `json:"size"`
	FractionOfVault  float64                  `json:"fractionOfVault,omitempty"`
	Density          float64                  `json:"density,omitempty"`
	Anchor           string                   `json:"anchor,omitempty"`
	TopTags          []obsidian.TagCount      `json:"topTags,omitempty"`
	TopAuthority     []AuthorityScorePayload  `json:"topAuthority,omitempty"`
	AuthorityBuckets []AuthorityBucketPayload `json:"authorityBuckets,omitempty"`
	AuthorityStats   *AuthorityStatsPayload   `json:"authorityStats,omitempty"`
	Recency          *GraphRecencyPayload     `json:"recency,omitempty"`
	Bridges          []BridgePayload          `json:"bridges,omitempty"`
	IsBridge         bool                     `json:"isBridge,omitempty"`
}

// NoteNeighbors distinguishes inbound/outbound and community boundaries.
type NoteNeighbors struct {
	LinksOut       []NeighborRef `json:"linksOut,omitempty"`
	LinksIn        []NeighborRef `json:"linksIn,omitempty"`
	SameCommunity  []string      `json:"sameCommunity,omitempty"`
	CrossCommunity []NeighborRef `json:"crossCommunity,omitempty"`
}

// NoteContextResponse is returned by the note_context tool.
type NoteContextResponse struct {
	Path               string                 `json:"path"`
	Title              string                 `json:"title,omitempty"`
	Error              string                 `json:"error,omitempty"`
	Tags               []string               `json:"tags,omitempty"`
	Frontmatter        map[string]interface{} `json:"frontmatter,omitempty"`
	Graph              NoteGraphContext       `json:"graph,omitempty"`
	Community          NoteCommunityContext   `json:"community,omitempty"`
	Neighbors          NoteNeighbors          `json:"neighbors,omitempty"`
	Backlinks          []obsidian.Backlink    `json:"backlinks,omitempty"`
	NeighborsTruncated bool                   `json:"neighborsTruncated,omitempty"`
	NeighborsLimit     int                    `json:"neighborsLimit,omitempty"`
	BacklinksTruncated bool                   `json:"backlinksTruncated,omitempty"`
	BacklinksLimit     int                    `json:"backlinksLimit,omitempty"`
}

// VaultContextResponse summarizes the vault for agents.
type VaultContextResponse struct {
	Stats        obsidian.GraphStatsSummary `json:"stats"`
	OrphanCount  int                        `json:"orphanCount"`
	TopOrphans   []string                   `json:"topOrphans,omitempty"`
	Components   []ComponentSummary         `json:"components,omitempty"`
	Communities  []CommunityOverview        `json:"communities"`
	KeyNotes     []string                   `json:"keyNotes,omitempty"`
	MOCs         []KeyNoteMatch             `json:"mocs,omitempty"`
	KeyPatterns  []string                   `json:"keyPatterns,omitempty"`
	NoteContexts []NoteContextResponse      `json:"noteContexts,omitempty"`
}

// CommunityOverview is a lightweight community summary for vault_context.
type CommunityOverview struct {
	ID               string                   `json:"id"`
	Size             int                      `json:"size"`
	FractionOfVault  float64                  `json:"fractionOfVault,omitempty"`
	Anchor           string                   `json:"anchor,omitempty"`
	Density          float64                  `json:"density,omitempty"`
	TopTags          []obsidian.TagCount      `json:"topTags,omitempty"`
	TopAuthority     []AuthorityScorePayload  `json:"topAuthority,omitempty"`
	AuthorityBuckets []AuthorityBucketPayload `json:"authorityBuckets,omitempty"`
	AuthorityStats   *AuthorityStatsPayload   `json:"authorityStats,omitempty"`
	Recency          *GraphRecencyPayload     `json:"recency,omitempty"`
	BridgesDetailed  []BridgePayload          `json:"bridgesDetailed,omitempty"`
}

// KeyNoteMatch captures a key/MOC note and which pattern matched.
type KeyNoteMatch struct {
	Path    string `json:"path"`
	Pattern string `json:"pattern,omitempty"`
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

// parseStringArray ensures a JSON array of strings, returning the slice or an error message.
func parseStringArray(raw interface{}, field string) ([]string, string) {
	items, ok := raw.([]interface{})
	if !ok {
		return nil, fmt.Sprintf("%s parameter is required and must be an array", field)
	}

	out := make([]string, len(items))
	for i, v := range items {
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Sprintf("all %s items must be strings", field)
		}
		out[i] = s
	}
	return out, ""
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

		baseSuppressed := config.SuppressedTags
		suppressedTags := make([]string, len(baseSuppressed))
		copy(suppressedTags, baseSuppressed)
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

		reverseNeighbors := buildReverseNeighbors(analysis.Nodes)
		communityLookup := obsidian.CommunityMembershipLookup(analysis.Communities)
		bridgeCounts := crossCommunityEdgeCounts(analysis.Nodes, reverseNeighbors, communityLookup)

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
			size := len(comm.Nodes)
			topAuthorityPayload := authorityScoresToPayload(comm.TopAuthority, maxTopNotes)
			bucketPayload := authorityBucketsToPayload(comm.AuthorityBuckets)
			statsPayload := authorityStatsToPayload(comm.AuthorityStats)
			recencyPayload := recencyToPayload(comm.Recency)
			comms = append(comms, GraphCommunityPayload{
				ID:               comm.ID,
				Size:             size,
				FractionOfVault:  fractionOfVault(size, analysis.Stats.NodeCount),
				Nodes:            nil, // omit members from list response
				TopTags:          comm.TopTags,
				TopAuthority:     topAuthorityPayload,
				AuthorityBuckets: bucketPayload,
				AuthorityStats:   statsPayload,
				Recency:          recencyPayload,
				Anchor:           comm.Anchor,
				Density:          comm.Density,
				Bridges:          comm.Bridges,
				BridgesDetailed:  bridgePayloads(comm.Bridges, bridgeCounts),
			})
		}

		resp := CommunityListResponse{
			Communities: comms,
			Stats:       analysis.Stats,
			OrphanCount: len(analysis.Orphans),
			Orphans:     analysis.Orphans,
			Components:  componentSummariesFromWeak(analysis.WeakComponents, analysis.Nodes, analysis.Stats.NodeCount),
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

		reverseNeighbors := buildReverseNeighbors(analysis.Nodes)
		communityLookup := obsidian.CommunityMembershipLookup(analysis.Communities)
		bridgeCounts := crossCommunityEdgeCounts(analysis.Nodes, reverseNeighbors, communityLookup)

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
			target = communityLookup[normalized]
			if target == nil {
				return mcp.NewToolResultError(fmt.Sprintf("file %s is not assigned to a community under current filters", file)), nil
			}
		}

		resp := communityDetailPayload(target, analysis, includeTags, includeNeighbors, limit, reverseNeighbors, bridgeCounts)

		encoded, err := json.Marshal(resp)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("error marshaling community detail: %s", err)), nil
		}
		return mcp.NewToolResultText(string(encoded)), nil
	}
}

type rankedMember struct {
	path      string
	title     string
	hub       float64
	authority float64
	in        int
	out       int
	tags      []string
}

func rankMembers(members []string, nodes map[string]obsidian.GraphNode) []rankedMember {
	var list []rankedMember
	for _, p := range members {
		n := nodes[p]
		list = append(list, rankedMember{
			path:      p,
			title:     n.Title,
			hub:       n.Hub,
			authority: n.Authority,
			in:        n.Inbound,
			out:       n.Outbound,
			tags:      n.Tags,
		})
	}
	// Sort by authority (cornerstone concepts first)
	sort.Slice(list, func(i, j int) bool {
		if list[i].authority == list[j].authority {
			return list[i].path < list[j].path
		}
		return list[i].authority > list[j].authority
	})
	return list
}

func buildReverseNeighbors(nodes map[string]obsidian.GraphNode) map[string][]string {
	reverse := make(map[string][]string, len(nodes))
	for path := range nodes {
		reverse[path] = nil
	}
	for src, node := range nodes {
		for _, dst := range node.Neighbors {
			reverse[dst] = append(reverse[dst], src)
		}
	}
	for path := range reverse {
		sort.Strings(reverse[path])
	}
	return reverse
}

func crossCommunityEdgeCounts(nodes map[string]obsidian.GraphNode, reverse map[string][]string, membership map[string]*obsidian.CommunitySummary) map[string]int {
	counts := make(map[string]int, len(nodes))

	for src, node := range nodes {
		var srcComm string
		if comm := membership[src]; comm != nil {
			srcComm = comm.ID
		}
		for _, dst := range node.Neighbors {
			var dstComm string
			if comm := membership[dst]; comm != nil {
				dstComm = comm.ID
			}
			if srcComm != "" && dstComm != "" && srcComm != dstComm {
				counts[src]++
			}
		}
	}

	for dst, sources := range reverse {
		var dstComm string
		if comm := membership[dst]; comm != nil {
			dstComm = comm.ID
		}
		for _, src := range sources {
			var srcComm string
			if comm := membership[src]; comm != nil {
				srcComm = comm.ID
			}
			if srcComm != "" && dstComm != "" && srcComm != dstComm {
				counts[dst]++
			}
		}
	}

	return counts
}

func bridgePayloads(paths []string, counts map[string]int) []BridgePayload {
	if len(paths) == 0 {
		return nil
	}
	out := make([]BridgePayload, 0, len(paths))
	for _, p := range paths {
		out = append(out, BridgePayload{
			Path:                p,
			CrossCommunityEdges: counts[p],
		})
	}
	return out
}

func componentSummariesFromWeak(components [][]string, nodes map[string]obsidian.GraphNode, total int) []ComponentSummary {
	if len(components) == 0 {
		return nil
	}
	summaries := make([]ComponentSummary, 0, len(components))
	for idx, comp := range components {
		if len(comp) == 0 {
			summaries = append(summaries, ComponentSummary{
				ID:   fmt.Sprintf("comp%d", idx),
				Size: 0,
			})
			continue
		}
		id := nodes[comp[0]].WeakCompID
		if id == "" {
			id = fmt.Sprintf("comp%d", idx)
		}
		size := len(comp)
		summaries = append(summaries, ComponentSummary{
			ID:              id,
			Size:            size,
			FractionOfVault: fractionOfVault(size, total),
		})
	}
	return summaries
}

func fractionOfVault(size int, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(size) / float64(total)
}

func isBridgeNode(path string, bridgeSet map[string]struct{}, counts map[string]int) bool {
	if _, ok := bridgeSet[path]; ok {
		return true
	}
	return counts[path] > 0
}

func neighborRefs(paths []string, membership map[string]*obsidian.CommunitySummary) []NeighborRef {
	if len(paths) == 0 {
		return nil
	}
	out := make([]NeighborRef, 0, len(paths))
	for _, p := range paths {
		ref := NeighborRef{
			Path:      p,
			Community: "",
		}
		if comm := membership[p]; comm != nil {
			ref.Community = comm.ID
		}
		out = append(out, ref)
	}
	return out
}

func sortedStringsFromSet(set map[string]struct{}) []string {
	if len(set) == 0 {
		return nil
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func sortedNeighborRefsFromMap(m map[string]NeighborRef) []NeighborRef {
	if len(m) == 0 {
		return nil
	}
	out := make([]NeighborRef, 0, len(m))
	for _, ref := range m {
		out = append(out, ref)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Path < out[j].Path
	})
	return out
}

func hubPercentile(nodes map[string]obsidian.GraphNode, target string, pct *percentilesCache) float64 {
	if len(nodes) == 0 || pct == nil {
		return 0
	}
	return percentileLookup(nodes[target].Hub, pct.Hubs)
}

func authorityPercentile(nodes map[string]obsidian.GraphNode, target string, pct *percentilesCache) float64 {
	if len(nodes) == 0 || pct == nil {
		return 0
	}
	return percentileLookup(nodes[target].Authority, pct.Authorities)
}

type percentilesCache struct {
	Hubs        []float64
	Authorities []float64
}

func buildPercentilesCache(nodes map[string]obsidian.GraphNode) *percentilesCache {
	if len(nodes) == 0 {
		return &percentilesCache{}
	}
	hubs := make([]float64, 0, len(nodes))
	auths := make([]float64, 0, len(nodes))
	for _, n := range nodes {
		hubs = append(hubs, n.Hub)
		auths = append(auths, n.Authority)
	}
	sort.Float64s(hubs)
	sort.Float64s(auths)
	return &percentilesCache{Hubs: hubs, Authorities: auths}
}

func percentileLookup(target float64, sorted []float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	// Position of the last value <= target.
	idx := sort.Search(len(sorted), func(i int) bool {
		return sorted[i] > target
	})
	return float64(idx) / float64(len(sorted))
}

func authorityScoresToPayload(scores []obsidian.AuthorityScore, limit int) []AuthorityScorePayload {
	if limit > 0 && len(scores) > limit {
		scores = scores[:limit]
	}
	out := make([]AuthorityScorePayload, 0, len(scores))
	for _, s := range scores {
		out = append(out, AuthorityScorePayload{
			Path:      s.Path,
			Authority: s.Authority,
			Hub:       s.Hub,
		})
	}
	return out
}

func authorityBucketsToPayload(buckets []obsidian.AuthorityBucket) []AuthorityBucketPayload {
	if len(buckets) == 0 {
		return nil
	}
	out := make([]AuthorityBucketPayload, 0, len(buckets))
	for _, b := range buckets {
		out = append(out, AuthorityBucketPayload{
			Low:     b.Low,
			High:    b.High,
			Count:   b.Count,
			Example: b.Example,
		})
	}
	return out
}

func authorityStatsToPayload(stats *obsidian.AuthorityStats) *AuthorityStatsPayload {
	if stats == nil {
		return nil
	}
	return &AuthorityStatsPayload{
		Mean: stats.Mean,
		P50:  stats.P50,
		P75:  stats.P75,
		P90:  stats.P90,
		P95:  stats.P95,
		P99:  stats.P99,
		Max:  stats.Max,
	}
}

func recencyToPayload(r *obsidian.GraphRecency) *GraphRecencyPayload {
	if r == nil {
		return nil
	}
	return &GraphRecencyPayload{
		LatestPath:    r.LatestPath,
		LatestAgeDays: r.LatestAgeDays,
		RecentCount:   r.RecentCount,
		WindowDays:    r.WindowDays,
	}
}

// normalizeInputFile takes a vault-relative or absolute path. Absolute paths
// are relativized to the vault; paths outside the vault will not resolve in
// the graph and will return a filtered error.
func normalizeInputFile(file string, config Config) string {
	normalized := obsidian.NormalizePath(obsidian.AddMdSuffix(file))
	if filepath.IsAbs(file) && config.VaultPath != "" {
		if rel, err := filepath.Rel(config.VaultPath, file); err == nil {
			normalized = obsidian.NormalizePath(obsidian.AddMdSuffix(rel))
		}
	}
	return normalized
}

func extractStringArray(raw interface{}) []string {
	switch v := raw.(type) {
	case []string:
		return v
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case string:
		return []string{v}
	default:
		return nil
	}
}

// topAuthorityAcrossGraph returns top notes by Authority score.
func topAuthorityAcrossGraph(nodes map[string]obsidian.GraphNode, limit int) []string {
	if limit <= 0 {
		return nil
	}
	type pr struct {
		path string
		val  float64
	}
	list := make([]pr, 0, len(nodes))
	for path, n := range nodes {
		list = append(list, pr{path: path, val: n.Authority})
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].val == list[j].val {
			return list[i].path < list[j].path
		}
		return list[i].val > list[j].val
	})
	if len(list) > limit {
		list = list[:limit]
	}
	out := make([]string, len(list))
	for i, item := range list {
		out[i] = item.path
	}
	return out
}

// truncateNeighborRefs sorts neighbors by Authority (descending) before truncating,
// so the most important neighbors are kept when the limit is applied.
func truncateNeighborRefs(refs []NeighborRef, limit int, nodes map[string]obsidian.GraphNode) ([]NeighborRef, bool) {
	if len(refs) == 0 {
		return refs, false
	}
	// Sort by Authority descending (highest first), then by path for stability
	sorted := make([]NeighborRef, len(refs))
	copy(sorted, refs)
	sort.Slice(sorted, func(i, j int) bool {
		authI, authJ := 0.0, 0.0
		if n, ok := nodes[sorted[i].Path]; ok {
			authI = n.Authority
		}
		if n, ok := nodes[sorted[j].Path]; ok {
			authJ = n.Authority
		}
		if authI != authJ {
			return authI > authJ
		}
		return sorted[i].Path < sorted[j].Path
	})
	if limit <= 0 || len(sorted) <= limit {
		return sorted, false
	}
	return sorted[:limit], true
}

// truncateBacklinks sorts backlinks by Authority (descending) before truncating,
// so the most important referrers are kept when the limit is applied.
func truncateBacklinks(backs []obsidian.Backlink, limit int, nodes map[string]obsidian.GraphNode) ([]obsidian.Backlink, bool) {
	if len(backs) == 0 {
		return backs, false
	}
	// Sort by Authority descending (highest first), then by path for stability
	sorted := make([]obsidian.Backlink, len(backs))
	copy(sorted, backs)
	sort.Slice(sorted, func(i, j int) bool {
		authI, authJ := 0.0, 0.0
		if n, ok := nodes[sorted[i].Referrer]; ok {
			authI = n.Authority
		}
		if n, ok := nodes[sorted[j].Referrer]; ok {
			authJ = n.Authority
		}
		if authI != authJ {
			return authI > authJ
		}
		return sorted[i].Referrer < sorted[j].Referrer
	})
	if limit <= 0 || len(sorted) <= limit {
		return sorted, false
	}
	return sorted[:limit], true
}

func findKeyNotes(config Config, patterns []string) ([]KeyNoteMatch, error) {
	if len(patterns) == 0 {
		return nil, nil
	}
	note := resolveNoteManager(config)
	parsed, expr, err := actions.ParseInputsWithExpression(patterns)
	if err != nil {
		return nil, err
	}

	unique := make(map[string]struct{})
	order := make([]string, 0)
	_, err = actions.ListFiles(config.Vault, note, actions.ListParams{
		Inputs:         parsed,
		Expression:     expr,
		MaxDepth:       0,
		SkipAnchors:    false,
		SkipEmbeds:     false,
		AbsolutePaths:  false,
		SuppressedTags: config.SuppressedTags,
		OnMatch: func(file string) {
			if _, ok := unique[file]; !ok {
				unique[file] = struct{}{}
				order = append(order, file)
			}
		},
	})
	if err != nil {
		return nil, err
	}
	// Normalize to vault-relative with .md suffix to stay consistent with graph payloads.
	matches := make([]KeyNoteMatch, 0, len(order))
	for _, p := range order {
		norm := obsidian.NormalizePath(obsidian.AddMdSuffix(p))
		matches = append(matches, KeyNoteMatch{Path: norm})
	}
	// Attach the first pattern as a hint when single-pattern search; when multiple, leave pattern empty (multiple possible).
	if len(patterns) == 1 {
		for i := range matches {
			matches[i].Pattern = patterns[0]
		}
	}
	return matches, nil
}

func buildNoteContext(path string, analysis *obsidian.GraphAnalysis, pct *percentilesCache, reverseNeighbors map[string][]string, membership map[string]*obsidian.CommunitySummary, bridgeCounts map[string]int, note obsidian.NoteManager, config Config, includeTags bool, includeNeighbors bool, includeFrontmatter bool, includeBacklinks bool, backlinks map[string][]obsidian.Backlink, neighborLimit int, backlinksLimit int) (NoteContextResponse, error) {
	node, ok := analysis.Nodes[path]
	if !ok {
		return NoteContextResponse{}, fmt.Errorf("graph filter: file %s not found in graph (use vault-relative path and check include/exclude filters)", path)
	}

	comm := membership[path]
	if comm == nil {
		return NoteContextResponse{}, fmt.Errorf("graph filter: file %s is not assigned to a community under current filters", path)
	}

	bridgeSet := make(map[string]struct{}, len(comm.Bridges))
	for _, b := range comm.Bridges {
		bridgeSet[b] = struct{}{}
	}

	info, err := actions.GetFileInfo(config.Vault, note, path)
	if err != nil {
		return NoteContextResponse{}, fmt.Errorf("error reading file info: %s", err)
	}

	var backs []obsidian.Backlink
	backlinksTruncated := false
	if includeBacklinks {
		if backlinks == nil {
			return NoteContextResponse{}, fmt.Errorf("backlinks are required when includeBacklinks is true")
		}
		if list, ok := backlinks[path]; ok {
			backs, backlinksTruncated = truncateBacklinks(list, backlinksLimit, analysis.Nodes)
		} else {
			backs = []obsidian.Backlink{}
		}
	}

	neighborContext := NoteNeighbors{}
	neighborsTruncated := false
	if includeNeighbors {
		linksOutRefs := neighborRefs(node.Neighbors, membership)
		linksInRefs := neighborRefs(reverseNeighbors[path], membership)
		linksOutRefs, outTrunc := truncateNeighborRefs(linksOutRefs, neighborLimit, analysis.Nodes)
		linksInRefs, inTrunc := truncateNeighborRefs(linksInRefs, neighborLimit, analysis.Nodes)
		sameSet := make(map[string]struct{})
		crossMap := make(map[string]NeighborRef)
		for _, ref := range append(linksOutRefs, linksInRefs...) {
			if ref.Community == "" {
				continue
			}
			if ref.Community == comm.ID {
				sameSet[ref.Path] = struct{}{}
			} else {
				if _, ok := crossMap[ref.Path]; !ok {
					crossMap[ref.Path] = ref
				}
			}
		}
		neighborContext.LinksOut = linksOutRefs
		neighborContext.LinksIn = linksInRefs
		neighborContext.SameCommunity = sortedStringsFromSet(sameSet)
		neighborContext.CrossCommunity = sortedNeighborRefsFromMap(crossMap)
		neighborsTruncated = outTrunc || inTrunc
	}

	graphContext := NoteGraphContext{
		Inbound:             node.Inbound,
		Outbound:            node.Outbound,
		Hub:                 node.Hub,
		HubPercentile:       hubPercentile(analysis.Nodes, path, pct),
		Authority:           node.Authority,
		AuthorityPercentile: authorityPercentile(analysis.Nodes, path, pct),
		IsOrphan:            node.Inbound == 0 && node.Outbound == 0,
		WeakComponent:       node.WeakCompID,
		StrongComponent:     node.SCC,
	}

	communityContext := NoteCommunityContext{
		ID:               comm.ID,
		Size:             len(comm.Nodes),
		FractionOfVault:  fractionOfVault(len(comm.Nodes), analysis.Stats.NodeCount),
		Density:          comm.Density,
		Anchor:           comm.Anchor,
		TopTags:          comm.TopTags,
		TopAuthority:     authorityScoresToPayload(comm.TopAuthority, 0),
		AuthorityBuckets: authorityBucketsToPayload(comm.AuthorityBuckets),
		AuthorityStats:   authorityStatsToPayload(comm.AuthorityStats),
		Recency:          recencyToPayload(comm.Recency),
		Bridges:          bridgePayloads(comm.Bridges, bridgeCounts),
		IsBridge:         isBridgeNode(path, bridgeSet, bridgeCounts),
	}
	if !includeTags {
		communityContext.TopTags = nil
	}

	var frontmatter map[string]interface{}
	if includeFrontmatter && info.Frontmatter != nil {
		frontmatter = info.Frontmatter
	}

	neighborsLimitField := 0
	if includeNeighbors {
		neighborsLimitField = neighborLimit
	}
	backlinksLimitField := 0
	if includeBacklinks {
		backlinksLimitField = backlinksLimit
	}

	return NoteContextResponse{
		Path:               path,
		Title:              node.Title,
		Tags:               info.Tags,
		Frontmatter:        frontmatter,
		Graph:              graphContext,
		Community:          communityContext,
		Neighbors:          neighborContext,
		Backlinks:          backs,
		NeighborsTruncated: neighborsTruncated,
		NeighborsLimit:     neighborsLimitField,
		BacklinksTruncated: backlinksTruncated,
		BacklinksLimit:     backlinksLimitField,
	}, nil
}

func communityDetailPayload(target *obsidian.CommunitySummary, analysis *obsidian.GraphAnalysis, includeTags bool, includeNeighbors bool, limit int, reverse map[string][]string, bridgeCounts map[string]int) CommunityDetailResponse {
	edgeCount := obsidian.CommunityInternalEdges(target, analysis.Nodes)

	members := rankMembers(target.Nodes, analysis.Nodes)
	if limit > 0 && limit < len(members) {
		members = members[:limit]
	}

	payloadMembers := make([]GraphNodePayload, 0, len(members))
	bridgeSet := make(map[string]struct{}, len(target.Bridges))
	for _, b := range target.Bridges {
		bridgeSet[b] = struct{}{}
	}

	for _, m := range members {
		linksOut := analysis.Nodes[m.path].Neighbors
		linksIn := reverse[m.path]
		payload := GraphNodePayload{
			Path:        m.path,
			Title:       m.title,
			Inbound:     m.in,
			Outbound:    m.out,
			Hub:         m.hub,
			Authority:   m.authority,
			Tags:        m.tags,
			WeakComp:    analysis.Nodes[m.path].WeakCompID,
			SCC:         analysis.Nodes[m.path].SCC,
			IsBridge:    isBridgeNode(m.path, bridgeSet, bridgeCounts),
			BridgeEdges: bridgeCounts[m.path],
			Community:   target.ID,
		}
		if includeNeighbors {
			payload.Neighbors = linksOut
			payload.LinksOut = linksOut
			payload.LinksIn = linksIn
		}
		if !includeTags {
			payload.Tags = nil
		}
		payloadMembers = append(payloadMembers, payload)
	}

	return CommunityDetailResponse{
		ID:               target.ID,
		Anchor:           target.Anchor,
		Size:             len(target.Nodes),
		FractionOfVault:  fractionOfVault(len(target.Nodes), analysis.Stats.NodeCount),
		Density:          target.Density,
		Bridges:          target.Bridges,
		BridgesDetailed:  bridgePayloads(target.Bridges, bridgeCounts),
		TopTags:          target.TopTags,
		TopAuthority:     authorityScoresToPayload(target.TopAuthority, 0),
		AuthorityBuckets: authorityBucketsToPayload(target.AuthorityBuckets),
		AuthorityStats:   authorityStatsToPayload(target.AuthorityStats),
		Recency:          recencyToPayload(target.Recency),
		Members:          payloadMembers,
		InternalEdges:    edgeCount,
	}
}

// NoteContextTool returns a single-note context (graph + community + backlinks).
// NoteContextTool returns graph + community context for one or more notes.
func NoteContextTool(config Config) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		files := extractStringArray(args["files"])
		if len(files) == 0 {
			return mcp.NewToolResultError("files is required (array of paths)"), nil
		}

		skipAnchors, _ := args["skipAnchors"].(bool)
		skipEmbeds, _ := args["skipEmbeds"].(bool)
		includeFrontmatter, _ := args["includeFrontmatter"].(bool)
		includeBacklinks := true
		if v, ok := args["includeBacklinks"].(bool); ok {
			includeBacklinks = v
		}
		includeNeighbors := true
		if v, ok := args["includeNeighbors"].(bool); ok {
			includeNeighbors = v
		}
		includeTags := true
		if v, ok := args["includeTags"].(bool); ok {
			includeTags = v
		}
		neighborLimit := 50
		if v, ok := args["neighborLimit"].(float64); ok && int(v) > 0 {
			neighborLimit = int(v)
		}
		backlinksLimit := 50
		if v, ok := args["backlinksLimit"].(float64); ok && int(v) > 0 {
			backlinksLimit = int(v)
		}
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

		note := resolveNoteManager(config)
		analysis, err := actions.GraphAnalysis(config.Vault, note, actions.GraphAnalysisParams{
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

		pct := buildPercentilesCache(analysis.Nodes)
		communityLookup := obsidian.CommunityMembershipLookup(analysis.Communities)
		reverseNeighbors := buildReverseNeighbors(analysis.Nodes)
		bridgeCounts := crossCommunityEdgeCounts(analysis.Nodes, reverseNeighbors, communityLookup)

		normalizedTargets := make([]string, 0, len(files))
		for _, f := range files {
			normalizedTargets = append(normalizedTargets, normalizeInputFile(f, config))
		}

		var backlinks map[string][]obsidian.Backlink
		if includeBacklinks {
			blOptions := obsidian.WikilinkOptions{SkipAnchors: skipAnchors, SkipEmbeds: skipEmbeds}
			if config.AnalysisCache != nil {
				backlinks, err = config.AnalysisCache.Backlinks(config.VaultPath, note, normalizedTargets, blOptions, config.SuppressedTags)
			} else {
				backlinks, err = obsidian.CollectBacklinks(config.VaultPath, note, normalizedTargets, blOptions, config.SuppressedTags)
			}
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("error collecting backlinks: %s", err)), nil
			}
		}

		contexts := make([]NoteContextResponse, 0, len(normalizedTargets))
		for _, norm := range normalizedTargets {
			ctxResp, err := buildNoteContext(norm, analysis, pct, reverseNeighbors, communityLookup, bridgeCounts, note, config, includeTags, includeNeighbors, includeFrontmatter, includeBacklinks, backlinks, neighborLimit, backlinksLimit)
			if err != nil {
				// Return partial result with error instead of failing the whole batch
				contexts = append(contexts, NoteContextResponse{
					Path:  norm,
					Error: err.Error(),
				})
			} else {
				contexts = append(contexts, ctxResp)
			}
		}

		payload := map[string]interface{}{
			"contexts": contexts,
			"count":    len(contexts),
		}
		encoded, err := json.Marshal(payload)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("error marshaling note contexts: %s", err)), nil
		}
		return mcp.NewToolResultText(string(encoded)), nil
	}
}

// VaultContextTool provides a compact, high-signal snapshot of the vault.
func VaultContextTool(config Config) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		skipAnchors, _ := args["skipAnchors"].(bool)
		skipEmbeds, _ := args["skipEmbeds"].(bool)
		includeTags := true
		if v, ok := args["includeTags"].(bool); ok {
			includeTags = v
		}
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

		maxCommunities := 10
		if v, ok := args["maxCommunities"].(float64); ok && int(v) > 0 {
			maxCommunities = int(v)
		}
		communityMemberLimit := 5
		if v, ok := args["communityTopNotes"].(float64); ok && int(v) > 0 {
			communityMemberLimit = int(v)
		}
		communityTagsLimit := 5
		if v, ok := args["communityTopTags"].(float64); ok && int(v) > 0 {
			communityTagsLimit = int(v)
		}
		bridgeLimit := 3
		if v, ok := args["bridgeLimit"].(float64); ok && int(v) > 0 {
			bridgeLimit = int(v)
		}
		topOrphansLimit := 10
		if v, ok := args["topOrphans"].(float64); ok && int(v) > 0 {
			topOrphansLimit = int(v)
		}
		topComponentsLimit := 5
		if v, ok := args["topComponents"].(float64); ok && int(v) > 0 {
			topComponentsLimit = int(v)
		}
		topGlobalPRLimit := 10
		if v, ok := args["topNotes"].(float64); ok && int(v) > 0 {
			topGlobalPRLimit = int(v)
		}
		contextFiles := extractStringArray(args["contextFiles"])
		contextIncludeBacklinks := true
		if v, ok := args["contextIncludeBacklinks"].(bool); ok {
			contextIncludeBacklinks = v
		}
		contextIncludeFrontmatter, _ := args["contextIncludeFrontmatter"].(bool)
		contextIncludeNeighbors := true
		if v, ok := args["contextIncludeNeighbors"].(bool); ok {
			contextIncludeNeighbors = v
		}
		contextIncludeTags := true
		if v, ok := args["contextIncludeTags"].(bool); ok {
			contextIncludeTags = v
		}
		contextNeighborLimit := 50
		if v, ok := args["contextNeighborLimit"].(float64); ok && int(v) > 0 {
			contextNeighborLimit = int(v)
		}
		contextBacklinksLimit := 50
		if v, ok := args["contextBacklinksLimit"].(float64); ok && int(v) > 0 {
			contextBacklinksLimit = int(v)
		}

		keyPatterns := extractStringArray(args["keyPatterns"])
		if len(keyPatterns) == 0 && config.VaultPath != "" {
			if cfg, err := obsidian.LoadVaultGraphConfig(config.VaultPath); err == nil {
				keyPatterns = append(keyPatterns, cfg.KeyNotePatterns...)
			}
		}

		note := resolveNoteManager(config)
		analysis, err := actions.GraphAnalysis(config.Vault, note, actions.GraphAnalysisParams{
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

		pct := buildPercentilesCache(analysis.Nodes)
		reverseNeighbors := buildReverseNeighbors(analysis.Nodes)
		communityLookup := obsidian.CommunityMembershipLookup(analysis.Communities)
		bridgeCounts := crossCommunityEdgeCounts(analysis.Nodes, reverseNeighbors, communityLookup)

		communities := make([]CommunityOverview, 0, len(analysis.Communities))
		for idx, comm := range analysis.Communities {
			if idx >= maxCommunities {
				break
			}
			topTags := comm.TopTags
			if len(topTags) > communityTagsLimit {
				topTags = topTags[:communityTagsLimit]
			}
			topPR := authorityScoresToPayload(comm.TopAuthority, communityMemberLimit)
			buckets := authorityBucketsToPayload(comm.AuthorityBuckets)
			stats := authorityStatsToPayload(comm.AuthorityStats)
			recency := recencyToPayload(comm.Recency)
			bridges := bridgePayloads(comm.Bridges, bridgeCounts)
			if len(bridges) > bridgeLimit {
				bridges = bridges[:bridgeLimit]
			}
			communities = append(communities, CommunityOverview{
				ID:               comm.ID,
				Size:             len(comm.Nodes),
				FractionOfVault:  fractionOfVault(len(comm.Nodes), analysis.Stats.NodeCount),
				Anchor:           comm.Anchor,
				Density:          comm.Density,
				TopTags:          topTags,
				TopAuthority:     topPR,
				AuthorityBuckets: buckets,
				AuthorityStats:   stats,
				Recency:          recency,
				BridgesDetailed:  bridges,
			})
		}

		orphans := analysis.Orphans
		if len(orphans) > topOrphansLimit {
			orphans = orphans[:topOrphansLimit]
		}

		components := componentSummariesFromWeak(analysis.WeakComponents, analysis.Nodes, analysis.Stats.NodeCount)
		if len(components) > topComponentsLimit {
			components = components[:topComponentsLimit]
		}

		keySet := make(map[string]struct{})
		for _, c := range communities {
			if c.Anchor != "" {
				keySet[c.Anchor] = struct{}{}
			}
			for _, b := range c.BridgesDetailed {
				if b.Path != "" {
					keySet[b.Path] = struct{}{}
				}
			}
			for _, p := range c.TopAuthority {
				if p.Path != "" {
					keySet[p.Path] = struct{}{}
				}
			}
		}
		for _, p := range topAuthorityAcrossGraph(analysis.Nodes, topGlobalPRLimit) {
			keySet[p] = struct{}{}
		}
		// Rank key notes by authority (desc) for a more useful orientation list.
		keyNotes := make([]string, 0, len(keySet))
		for k := range keySet {
			keyNotes = append(keyNotes, k)
		}
		sort.Slice(keyNotes, func(i, j int) bool {
			ai := analysis.Nodes[keyNotes[i]].Authority
			aj := analysis.Nodes[keyNotes[j]].Authority
			if ai == aj {
				return keyNotes[i] < keyNotes[j]
			}
			return ai > aj
		})

		mocs, err := findKeyNotes(config, keyPatterns)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("error finding key notes: %s", err)), nil
		}

		var noteContexts []NoteContextResponse
		if len(contextFiles) > 0 {
			normalizedTargets := make([]string, 0, len(contextFiles))
			for _, f := range contextFiles {
				normalizedTargets = append(normalizedTargets, normalizeInputFile(f, config))
			}
			var backlinks map[string][]obsidian.Backlink
			if contextIncludeBacklinks {
				blOptions := obsidian.WikilinkOptions{SkipAnchors: skipAnchors, SkipEmbeds: skipEmbeds}
				if config.AnalysisCache != nil {
					backlinks, err = config.AnalysisCache.Backlinks(config.VaultPath, note, normalizedTargets, blOptions, config.SuppressedTags)
				} else {
					backlinks, err = obsidian.CollectBacklinks(config.VaultPath, note, normalizedTargets, blOptions, config.SuppressedTags)
				}
				if err != nil {
					return mcp.NewToolResultError(fmt.Sprintf("error collecting backlinks for contexts: %s", err)), nil
				}
			}
			for _, norm := range normalizedTargets {
				ctxResp, err := buildNoteContext(norm, analysis, pct, reverseNeighbors, communityLookup, bridgeCounts, note, config, contextIncludeTags, contextIncludeNeighbors, contextIncludeFrontmatter, contextIncludeBacklinks, backlinks, contextNeighborLimit, contextBacklinksLimit)
				if err != nil {
					// Return partial result with error instead of failing the whole request
					noteContexts = append(noteContexts, NoteContextResponse{
						Path:  norm,
						Error: err.Error(),
					})
				} else {
					noteContexts = append(noteContexts, ctxResp)
				}
			}
		}

		resp := VaultContextResponse{
			Stats:        analysis.Stats,
			OrphanCount:  len(analysis.Orphans),
			TopOrphans:   orphans,
			Components:   components,
			Communities:  communities,
			KeyNotes:     keyNotes,
			MOCs:         mocs,
			KeyPatterns:  keyPatterns,
			NoteContexts: noteContexts,
		}

		encoded, err := json.Marshal(resp)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("error marshaling vault context: %s", err)), nil
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

// MoveNotesTool implements the move_notes MCP tool for single or bulk moves (backlinks rewritten by default).
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

// MutateTagsTool consolidates add/delete/rename tag operations.
func MutateTagsTool(config Config) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()

		op, _ := args["op"].(string)
		if op != "add" && op != "delete" && op != "rename" {
			return mcp.NewToolResultError("op must be one of add, delete, rename"), nil
		}

		dryRun, _ := args["dryRun"].(bool)

		if !config.ReadWrite && !dryRun {
			return mcp.NewToolResultError("Server is in read-only mode; either enable --read-write or set dryRun=true"), nil
		}

		note := resolveNoteManager(config)

		var summary actions.TagMutationSummary
		var err error

		switch op {
		case "add":
			tags, errMsg := parseStringArray(args["tags"], "tags")
			if errMsg != "" {
				return mcp.NewToolResultError(errMsg), nil
			}
			inputs, errMsg := parseStringArray(args["inputs"], "inputs")
			if errMsg != "" {
				return mcp.NewToolResultError(errMsg), nil
			}
			if len(tags) == 0 {
				return mcp.NewToolResultError("tags parameter cannot be empty"), nil
			}

			parsedInputs, expr, parseErr := actions.ParseInputsWithExpression(inputs)
			if parseErr != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Error parsing input criteria: %s", parseErr)), nil
			}
			matchingFiles, listErr := actions.ListFiles(config.Vault, note, actions.ListParams{
				Inputs:         parsedInputs,
				MaxDepth:       0,
				SkipAnchors:    false,
				SkipEmbeds:     false,
				AbsolutePaths:  false,
				Expression:     expr,
				SuppressedTags: []string{},
			})
			if listErr != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Error getting matching files: %s", listErr)), nil
			}
			if len(matchingFiles) == 0 {
				return mcp.NewToolResultError("No files match the specified criteria"), nil
			}
			summary, err = actions.AddTagsToFilesWithWorkers(config.Vault, note, tags, matchingFiles, dryRun, runtime.NumCPU())
		case "delete":
			tags, errMsg := parseStringArray(args["tags"], "tags")
			if errMsg != "" {
				return mcp.NewToolResultError(errMsg), nil
			}
			var files []string
			if raw, ok := args["inputs"]; ok {
				inputs, parseErrMsg := parseStringArray(raw, "inputs")
				if parseErrMsg != "" {
					return mcp.NewToolResultError(parseErrMsg), nil
				}
				parsedInputs, expr, parseErr := actions.ParseInputsWithExpression(inputs)
				if parseErr != nil {
					return mcp.NewToolResultError(fmt.Sprintf("Error parsing input criteria: %s", parseErr)), nil
				}
				files, err = actions.ListFiles(config.Vault, note, actions.ListParams{
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
				if len(files) == 0 {
					return mcp.NewToolResultError("No files match the specified criteria"), nil
				}
			}
			if len(files) > 0 {
				summary, err = actions.DeleteTagsFromFiles(config.Vault, note, tags, files, dryRun)
			} else {
				summary, err = actions.DeleteTagsWithWorkers(config.Vault, note, tags, dryRun, runtime.NumCPU())
			}
		case "rename":
			from, errMsg := parseStringArray(args["fromTags"], "fromTags")
			if errMsg != "" {
				return mcp.NewToolResultError(errMsg), nil
			}
			to, _ := args["toTag"].(string)
			if strings.TrimSpace(to) == "" {
				return mcp.NewToolResultError("toTag parameter is required for rename"), nil
			}
			var files []string
			if raw, ok := args["inputs"]; ok {
				inputs, parseErrMsg := parseStringArray(raw, "inputs")
				if parseErrMsg != "" {
					return mcp.NewToolResultError(parseErrMsg), nil
				}
				parsedInputs, expr, parseErr := actions.ParseInputsWithExpression(inputs)
				if parseErr != nil {
					return mcp.NewToolResultError(fmt.Sprintf("Error parsing input criteria: %s", parseErr)), nil
				}
				files, err = actions.ListFiles(config.Vault, note, actions.ListParams{
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
				if len(files) == 0 {
					return mcp.NewToolResultError("No files match the specified criteria"), nil
				}
			}
			if len(files) > 0 {
				summary, err = actions.RenameTagsInFiles(config.Vault, note, from, to, files, dryRun)
			} else {
				summary, err = actions.RenameTagsWithWorkers(config.Vault, note, from, to, dryRun, runtime.NumCPU())
			}
		}

		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error mutating tags: %s", err)), nil
		}

		result := TagMutationResult{
			DryRun:       dryRun,
			NotesTouched: summary.NotesTouched,
			TagChanges:   summary.TagChanges,
			FilesChanged: summary.FilesChanged,
		}

		encoded, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error marshaling mutate_tags result: %s", err)), nil
		}

		return mcp.NewToolResultText(string(encoded)), nil
	}
}

// MutatePropertiesTool consolidates set/delete/rename property operations.
func MutatePropertiesTool(config Config) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()

		op, _ := args["op"].(string)
		if op != "set" && op != "delete" && op != "rename" {
			return mcp.NewToolResultError("op must be one of set, delete, rename"), nil
		}

		dryRun, _ := args["dryRun"].(bool)
		overwrite, _ := args["overwrite"].(bool)
		merge := true
		if v, ok := args["merge"].(bool); ok {
			merge = v
		}

		if !config.ReadWrite && !dryRun {
			return mcp.NewToolResultError("Server is in read-only mode; either enable --read-write or set dryRun=true"), nil
		}

		note := resolveNoteManager(config)
		var summary actions.PropertyMutationSummary
		var err error

		switch op {
		case "set":
			property, _ := args["property"].(string)
			if strings.TrimSpace(property) == "" {
				return mcp.NewToolResultError("property parameter is required for set"), nil
			}
			valueRaw, ok := args["value"].(string)
			if !ok {
				return mcp.NewToolResultError("value parameter is required and must be a string (YAML accepted)"), nil
			}
			var value interface{}
			if err := yaml.Unmarshal([]byte(valueRaw), &value); err != nil {
				value = valueRaw
			}
			inputs, errMsg := parseStringArray(args["inputs"], "inputs")
			if errMsg != "" {
				return mcp.NewToolResultError(errMsg), nil
			}
			parsedInputs, expr, parseErr := actions.ParseInputsWithExpression(inputs)
			if parseErr != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Error parsing input criteria: %s", parseErr)), nil
			}
			matchingFiles, listErr := actions.ListFiles(config.Vault, note, actions.ListParams{
				Inputs:        parsedInputs,
				MaxDepth:      0,
				SkipAnchors:   false,
				SkipEmbeds:    false,
				AbsolutePaths: false,
				Expression:    expr,
			})
			if listErr != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Error getting matching files: %s", listErr)), nil
			}
			if len(matchingFiles) == 0 {
				return mcp.NewToolResultError("No files match the specified criteria"), nil
			}
			summary, err = actions.SetPropertyOnFiles(config.Vault, note, property, value, matchingFiles, overwrite, dryRun)
		case "delete":
			properties, errMsg := parseStringArray(args["properties"], "properties")
			if errMsg != "" {
				return mcp.NewToolResultError(errMsg), nil
			}
			var files []string
			if inputsRaw, ok := args["inputs"]; ok {
				inputs, parseErrMsg := parseStringArray(inputsRaw, "inputs")
				if parseErrMsg != "" {
					return mcp.NewToolResultError(parseErrMsg), nil
				}
				parsedInputs, expr, parseErr := actions.ParseInputsWithExpression(inputs)
				if parseErr != nil {
					return mcp.NewToolResultError(fmt.Sprintf("Error parsing input criteria: %s", parseErr)), nil
				}
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
			summary, err = actions.DeleteProperties(config.Vault, note, properties, files, dryRun)
		case "rename":
			from, errMsg := parseStringArray(args["fromProperties"], "fromProperties")
			if errMsg != "" {
				return mcp.NewToolResultError(errMsg), nil
			}
			to, _ := args["toProperty"].(string)
			if strings.TrimSpace(to) == "" {
				return mcp.NewToolResultError("toProperty parameter is required for rename"), nil
			}
			var files []string
			if inputsRaw, ok := args["inputs"]; ok {
				inputs, parseErrMsg := parseStringArray(inputsRaw, "inputs")
				if parseErrMsg != "" {
					return mcp.NewToolResultError(parseErrMsg), nil
				}
				parsedInputs, expr, parseErr := actions.ParseInputsWithExpression(inputs)
				if parseErr != nil {
					return mcp.NewToolResultError(fmt.Sprintf("Error parsing input criteria: %s", parseErr)), nil
				}
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
			summary, err = actions.RenameProperties(config.Vault, note, from, to, merge, files, dryRun)
		}

		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error mutating properties: %s", err)), nil
		}

		result := PropertyMutationResult{
			DryRun:          dryRun,
			NotesTouched:    summary.NotesTouched,
			PropertyChanges: summary.PropertyChanges,
			FilesChanged:    summary.FilesChanged,
		}

		encoded, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error marshaling mutate_properties result: %s", err)), nil
		}

		return mcp.NewToolResultText(string(encoded)), nil
	}
}
