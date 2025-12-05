package cache

import (
	"context"
	"sort"
	"strings"
	"sync"

	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
)

// SnapshotProvider is implemented by cache.Service (and adapters) to expose entries and a version counter.
type SnapshotProvider interface {
	EntriesSnapshot(context.Context) ([]Entry, error)
	Version() uint64
}

// AnalysisCache memoizes backlink and graph computations keyed by cache version and options.
type AnalysisCache struct {
	provider SnapshotProvider

	mu           sync.Mutex
	version      uint64
	backlinks    map[backlinkKey]map[string][]obsidian.Backlink
	graphs       map[graphKey]*obsidian.GraphAnalysis
	maxBacklinks int
	maxGraphs    int
	backlinkKeys []backlinkKey
	graphKeys    []graphKey
}

// NewAnalysisCache constructs a cache bound to a snapshot provider (typically Service or NoteAdapter).
func NewAnalysisCache(provider SnapshotProvider) *AnalysisCache {
	return &AnalysisCache{
		provider:     provider,
		backlinks:    make(map[backlinkKey]map[string][]obsidian.Backlink),
		graphs:       make(map[graphKey]*obsidian.GraphAnalysis),
		maxBacklinks: 64,
		maxGraphs:    32,
	}
}

// Backlinks returns cached backlinks when the provider version matches; otherwise it recomputes.
func (c *AnalysisCache) Backlinks(vaultPath string, note obsidian.NoteManager, targets []string, options obsidian.WikilinkOptions, suppressedTags []string) (map[string][]obsidian.Backlink, error) {
	version := c.provider.Version()

	key := backlinkKey{
		targets:       hashStrings(normalizeTargets(targets)),
		skipAnchors:   options.SkipAnchors,
		skipEmbeds:    options.SkipEmbeds,
		suppressedKey: hashStrings(normalizeTagsLower(suppressedTags)),
	}

	c.mu.Lock()
	if version != c.version {
		c.backlinks = make(map[backlinkKey]map[string][]obsidian.Backlink)
		c.graphs = make(map[graphKey]*obsidian.GraphAnalysis)
		c.version = version
	}
	if cached, ok := c.backlinks[key]; ok {
		c.mu.Unlock()
		return cloneBacklinks(cached), nil
	}
	c.mu.Unlock()

	result, err := obsidian.CollectBacklinks(vaultPath, note, targets, options, suppressedTags)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.backlinks[key] = result
	c.backlinkKeys = append(c.backlinkKeys, key)
	if c.maxBacklinks > 0 && len(c.backlinkKeys) > c.maxBacklinks {
		oldest := c.backlinkKeys[0]
		c.backlinkKeys = c.backlinkKeys[1:]
		delete(c.backlinks, oldest)
	}
	c.mu.Unlock()

	return cloneBacklinks(result), nil
}

// GraphAnalysis returns cached graph analysis keyed by options and provider version.
func (c *AnalysisCache) GraphAnalysis(vaultPath string, note obsidian.NoteManager, options obsidian.GraphAnalysisOptions) (*obsidian.GraphAnalysis, error) {
	version := c.provider.Version()

	key := graphKey{
		skipAnchors: options.SkipAnchors,
		skipEmbeds:  options.SkipEmbeds,
		includeTags: options.IncludeTags,
		minDegree:   options.MinDegree,
		mutualOnly:  options.MutualOnly,
		excludedKey: hashStrings(mapKeys(options.ExcludedPaths)),
		includedKey: hashStrings(mapKeys(options.IncludedPaths)),
	}

	c.mu.Lock()
	if version != c.version {
		c.backlinks = make(map[backlinkKey]map[string][]obsidian.Backlink)
		c.graphs = make(map[graphKey]*obsidian.GraphAnalysis)
		c.version = version
	}
	if cached, ok := c.graphs[key]; ok {
		c.mu.Unlock()
		return cloneGraphAnalysis(cached), nil
	}
	c.mu.Unlock()

	result, err := obsidian.ComputeGraphAnalysis(vaultPath, note, options)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.graphs[key] = result
	c.graphKeys = append(c.graphKeys, key)
	if c.maxGraphs > 0 && len(c.graphKeys) > c.maxGraphs {
		oldest := c.graphKeys[0]
		c.graphKeys = c.graphKeys[1:]
		delete(c.graphs, oldest)
	}
	c.mu.Unlock()

	return cloneGraphAnalysis(result), nil
}

type backlinkKey struct {
	targets       string
	skipAnchors   bool
	skipEmbeds    bool
	suppressedKey string
}

type graphKey struct {
	skipAnchors bool
	skipEmbeds  bool
	includeTags bool
	minDegree   int
	mutualOnly  bool
	excludedKey string
	includedKey string
}

func normalizeTargets(targets []string) []string {
	out := make([]string, 0, len(targets))
	for _, t := range targets {
		out = append(out, obsidian.NormalizePath(obsidian.AddMdSuffix(t)))
	}
	sort.Strings(out)
	return out
}

func normalizeTagsLower(tags []string) []string {
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		nt := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(t, "#")))
		if nt != "" {
			out = append(out, nt)
		}
	}
	sort.Strings(out)
	return out
}

func mapKeys(m map[string]struct{}) []string {
	if len(m) == 0 {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func hashStrings(items []string) string {
	return strings.Join(items, "|")
}

func cloneBacklinks(src map[string][]obsidian.Backlink) map[string][]obsidian.Backlink {
	out := make(map[string][]obsidian.Backlink, len(src))
	for k, v := range src {
		copied := make([]obsidian.Backlink, len(v))
		copy(copied, v)
		out[k] = copied
	}
	return out
}

func cloneGraphAnalysis(src *obsidian.GraphAnalysis) *obsidian.GraphAnalysis {
	if src == nil {
		return nil
	}
	copyAnalysis := *src

	copyAnalysis.Nodes = make(map[string]obsidian.GraphNode, len(src.Nodes))
	for k, v := range src.Nodes {
		node := v
		node.Neighbors = append([]string(nil), v.Neighbors...)
		node.Tags = append([]string(nil), v.Tags...)
		copyAnalysis.Nodes[k] = node
	}

	copyAnalysis.Communities = append([]obsidian.CommunitySummary(nil), src.Communities...)
	copyAnalysis.StrongComponents = clone2D(src.StrongComponents)
	copyAnalysis.WeakComponents = clone2D(src.WeakComponents)
	copyAnalysis.Orphans = append([]string(nil), src.Orphans...)
	return &copyAnalysis
}

func clone2D(src [][]string) [][]string {
	out := make([][]string, len(src))
	for i, arr := range src {
		out[i] = append([]string(nil), arr...)
	}
	return out
}
