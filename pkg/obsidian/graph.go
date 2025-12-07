package obsidian

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

// AuthorityScore captures a note and its authority/hub scores.
type AuthorityScore struct {
	Path      string
	Authority float64
	Hub       float64
}

// AuthorityBucket summarizes distribution of authority scores within a community.
type AuthorityBucket struct {
	Low     float64
	High    float64
	Count   int
	Example string
}

// GraphRecency summarizes modification recency for a group of notes.
type GraphRecency struct {
	LatestPath      string
	LatestAgeDays   float64
	LatestTimestamp time.Time `json:"-"`
	RecentCount     int
	WindowDays      int
}

// GraphTimings captures rough durations for major analysis phases.
type GraphTimings struct {
	LoadEntries time.Duration
	BuildGraph  time.Duration
	HITS        time.Duration
	LabelProp   time.Duration
	Recency     time.Duration
	Total       time.Duration
}

// GraphTimingsMillis exposes timings as milliseconds for JSON-friendly output.
type GraphTimingsMillis struct {
	LoadEntries int64 `json:"loadEntriesMs"`
	BuildGraph  int64 `json:"buildGraphMs"`
	HITS        int64 `json:"hitsMs"`
	LabelProp   int64 `json:"labelPropMs"`
	Recency     int64 `json:"recencyMs"`
	Total       int64 `json:"totalMs"`
}

// ToMillis converts durations to ms, clamping any non-zero value to at least 1ms.
func (t GraphTimings) ToMillis() GraphTimingsMillis {
	conv := func(d time.Duration) int64 {
		if d <= 0 {
			return 0
		}
		ms := d.Milliseconds()
		if ms == 0 {
			return 1
		}
		return ms
	}
	return GraphTimingsMillis{
		LoadEntries: conv(t.LoadEntries),
		BuildGraph:  conv(t.BuildGraph),
		HITS:        conv(t.HITS),
		LabelProp:   conv(t.LabelProp),
		Recency:     conv(t.Recency),
		Total:       conv(t.Total),
	}
}

// AuthorityStats captures coarse percentiles/mean for authority scores.
type AuthorityStats struct {
	Mean float64
	P50  float64
	P75  float64
	P90  float64
	P95  float64
	P99  float64
	Max  float64
}

// NodeStats captures inbound/outbound link counts for a note.
type NodeStats struct {
	Inbound  int
	Outbound int
}

// GraphStats contains degree counts and strongly connected components for a vault's wikilink graph.
type GraphStats struct {
	Nodes      map[string]NodeStats // keyed by normalized note path (with .md)
	Components [][]string           // strongly connected components (each sorted)
	Adjacency  map[string]map[string]struct{}
}

// Community represents a loosely connected cluster of notes.
type Community struct {
	ID    string
	Nodes []string
}

// CommunitySummary captures metadata for a detected community.
type CommunitySummary struct {
	ID               string
	Nodes            []string
	TopTags          []TagCount
	TopAuthority     []AuthorityScore
	AuthorityBuckets []AuthorityBucket
	AuthorityStats   *AuthorityStats
	Recency          *GraphRecency
	Anchor           string
	Density          float64
	Bridges          []string
}

// TagCount represents a tag and its count.
type TagCount struct {
	Tag   string
	Count int
}

// GraphNode captures node-centric details for analysis output.
type GraphNode struct {
	Path       string   `json:"path"`
	Title      string   `json:"title"`
	Inbound    int      `json:"inbound"`
	Outbound   int      `json:"outbound"`
	Hub        float64  `json:"hub"`       // HITS hub score: measures how well this note curates/aggregates links
	Authority  float64  `json:"authority"` // HITS authority score: measures how often this note is referenced
	Community  string   `json:"community"`
	SCC        string   `json:"scc"`
	Neighbors  []string `json:"neighbors,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	WeakCompID string   `json:"weakComponent,omitempty"`
}

// GraphAnalysis contains a richer representation for agents.
type GraphAnalysis struct {
	Nodes            map[string]GraphNode
	Communities      []CommunitySummary
	StrongComponents [][]string
	WeakComponents   [][]string
	Orphans          []string
	Stats            GraphStatsSummary
	EffectiveTimes   map[string]time.Time `json:"-"`
	Timings          GraphTimings         `json:"-"`
}

// GraphStatsSummary captures high-level totals.
type GraphStatsSummary struct {
	NodeCount int `json:"nodeCount"`
	EdgeCount int `json:"edgeCount"`
}

// GraphAnalysisOptions controls optional data.
type GraphAnalysisOptions struct {
	WikilinkOptions
	IncludeTags   bool
	ExcludedPaths map[string]struct{}
	IncludedPaths map[string]struct{}
	MinDegree     int
	MutualOnly    bool
	// RecencyCascade controls whether inferred recency is propagated beyond 1 hop.
	// When true (default), a bounded multi-pass propagation will cascade freshness.
	// When false, only direct neighbors can boost an undated note.
	RecencyCascade bool
	// RecencyCascadeSet distinguishes explicit caller choice from defaulting behavior.
	RecencyCascadeSet bool
}

// NoteEntry captures cached note data that callers can reuse to avoid
// re-reading files when computing graph metadata.
type NoteEntry struct {
	Path        string
	Content     string
	Frontmatter map[string]interface{}
	Tags        []string
	ContentTime time.Time
}

// NoteEntriesProvider exposes cached note entries for efficient analysis.
// Implemented by cache-backed adapters.
type NoteEntriesProvider interface {
	NoteEntriesSnapshot(ctx context.Context) ([]NoteEntry, error)
}

// ComputeGraphStats scans the vault and returns degree counts and strongly connected components.
func ComputeGraphStats(vaultPath string, note NoteManager, options WikilinkOptions) (*GraphStats, error) {
	allNotes, err := note.GetNotesList(vaultPath)
	if err != nil {
		return nil, err
	}

	cache := BuildNotePathCache(allNotes)

	adjacency := make(map[string]map[string]struct{}, len(allNotes))
	for _, path := range allNotes {
		normalized := NormalizePath(AddMdSuffix(path))
		adjacency[normalized] = make(map[string]struct{})
	}

	for _, path := range allNotes {
		content, err := note.GetContents(vaultPath, path)
		if err != nil {
			return nil, err
		}

		links := scanWikilinks(content, options)
		src := NormalizePath(AddMdSuffix(path))

		for _, link := range links {
			resolved, ok := cache.ResolveNote(link.Target)
			if !ok {
				continue
			}

			dst := NormalizePath(AddMdSuffix(resolved))
			if src == dst {
				continue // ignore self-links; they don't connect notes for orphan detection
			}
			adjacency[src][dst] = struct{}{}
		}
	}

	nodes := make(map[string]NodeStats, len(adjacency))
	for node := range adjacency {
		nodes[node] = NodeStats{}
	}

	for src, targets := range adjacency {
		stats := nodes[src]
		stats.Outbound = len(targets)
		nodes[src] = stats

		for dst := range targets {
			stats := nodes[dst]
			stats.Inbound++
			nodes[dst] = stats
		}
	}

	components := stronglyConnectedComponents(adjacency)

	return &GraphStats{
		Nodes:      nodes,
		Components: components,
		Adjacency:  adjacency,
	}, nil
}

// Communities returns communities using label propagation over the adjacency map.
func (g *GraphStats) Communities() []Community {
	if g == nil {
		return nil
	}

	labels := labelPropagation(g.Adjacency)

	grouped := make(map[string][]string)
	for node, label := range labels {
		grouped[label] = append(grouped[label], node)
	}

	var result []Community
	for id, nodes := range grouped {
		sort.Strings(nodes)
		result = append(result, Community{ID: id, Nodes: nodes})
	}

	sort.Slice(result, func(i, j int) bool {
		if len(result[i].Nodes) == len(result[j].Nodes) {
			return result[i].Nodes[0] < result[j].Nodes[0]
		}
		return len(result[i].Nodes) > len(result[j].Nodes)
	})

	return result
}

// Orphans returns notes with zero inbound and outbound links.
func (g *GraphStats) Orphans() []string {
	if g == nil {
		return nil
	}
	var orphans []string
	for path, stats := range g.Nodes {
		if stats.Inbound == 0 && stats.Outbound == 0 {
			orphans = append(orphans, path)
		}
	}
	sort.Strings(orphans)
	return orphans
}

func stronglyConnectedComponents(adjacency map[string]map[string]struct{}) [][]string {
	index := 0
	indexMap := make(map[string]int)
	lowlink := make(map[string]int)
	onStack := make(map[string]bool)
	var stack []string
	var components [][]string

	var visit func(v string)
	visit = func(v string) {
		indexMap[v] = index
		lowlink[v] = index
		index++
		stack = append(stack, v)
		onStack[v] = true

		for w := range adjacency[v] {
			if _, seen := indexMap[w]; !seen {
				visit(w)
				if lowlink[w] < lowlink[v] {
					lowlink[v] = lowlink[w]
				}
			} else if onStack[w] && indexMap[w] < lowlink[v] {
				lowlink[v] = indexMap[w]
			}
		}

		if lowlink[v] == indexMap[v] {
			var component []string
			for {
				n := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[n] = false
				component = append(component, n)
				if n == v {
					break
				}
			}
			sort.Strings(component)
			components = append(components, component)
		}
	}

	var nodes []string
	for node := range adjacency {
		nodes = append(nodes, node)
	}
	sort.Strings(nodes)

	for _, node := range nodes {
		if _, visited := indexMap[node]; !visited {
			visit(node)
		}
	}

	sort.Slice(components, func(i, j int) bool {
		if len(components[i]) == len(components[j]) {
			return components[i][0] < components[j][0]
		}
		return len(components[i]) > len(components[j])
	})

	return components
}

// ComputeGraphAnalysis builds a rich graph view including hub/authority scores, communities, and tags.
func ComputeGraphAnalysis(vaultPath string, note NoteManager, options GraphAnalysisOptions) (*GraphAnalysis, error) {
	startTotal := time.Now()
	var timings GraphTimings

	loadStart := time.Now()
	var (
		allNotes []string
		entries  []NoteEntry
	)

	if provider, ok := note.(NoteEntriesProvider); ok {
		if snapshot, err := provider.NoteEntriesSnapshot(context.Background()); err == nil {
			entries = snapshot
		}
	}

	var err error
	if len(entries) == 0 {
		allNotes, err = note.GetNotesList(vaultPath)
		if err != nil {
			return nil, err
		}
		// Build entries lazily from GetContents for the legacy path below.
		entries = make([]NoteEntry, 0, len(allNotes))
		for _, path := range allNotes {
			content, err := note.GetContents(vaultPath, path)
			if err != nil {
				return nil, err
			}
			entries = append(entries, NoteEntry{
				Path:    path,
				Content: content,
			})
		}
	} else {
		allNotes = make([]string, 0, len(entries))
		for _, e := range entries {
			allNotes = append(allNotes, e.Path)
		}
	}
	timings.LoadEntries = time.Since(loadStart)

	cache := BuildNotePathCache(allNotes)

	excluded := options.ExcludedPaths
	included := options.IncludedPaths

	adjacency := make(map[string]map[string]struct{}, len(allNotes))
	tagMap := make(map[string][]string)
	contentTimes := make(map[string]time.Time, len(allNotes))
	normalizedEntries := make([]NoteEntry, 0, len(entries))

	buildStart := time.Now()
	for _, entry := range entries {
		normalized := NormalizePath(AddMdSuffix(entry.Path))
		if _, skip := excluded[normalized]; skip {
			continue
		}
		if len(included) > 0 {
			if _, ok := included[normalized]; !ok {
				continue
			}
		}
		adjacency[normalized] = make(map[string]struct{})
		normalizedEntries = append(normalizedEntries, NoteEntry{
			Path:        normalized,
			Content:     entry.Content,
			Frontmatter: entry.Frontmatter,
			Tags:        entry.Tags,
			ContentTime: entry.ContentTime,
		})
	}

	for _, entry := range normalizedEntries {
		if options.IncludeTags {
			if len(entry.Tags) > 0 {
				tagMap[entry.Path] = entry.Tags
			} else {
				tagMap[entry.Path] = extractTags(entry.Content)
			}
		}

		if ct := entry.ContentTime; !ct.IsZero() {
			contentTimes[entry.Path] = ct
		} else if ct, ok := ResolveContentTime(entry.Path, entry.Content); ok {
			contentTimes[entry.Path] = ct
		}

		links := scanWikilinks(entry.Content, options.WikilinkOptions)
		src := entry.Path

		for _, link := range links {
			resolved, ok := cache.ResolveNote(link.Target)
			if !ok {
				continue
			}

			dst := NormalizePath(AddMdSuffix(resolved))
			if src == dst {
				continue
			}
			// Only retain edges to nodes that survived include/exclude filtering.
			if _, ok := adjacency[src]; ok {
				if _, ok := adjacency[dst]; ok {
					adjacency[src][dst] = struct{}{}
				}
			}
		}
	}

	if options.MutualOnly {
		adjacency = filterMutualEdges(adjacency)
	}
	if options.MinDegree > 0 {
		adjacency = filterByMinDegree(adjacency, options.MinDegree)
	}
	timings.BuildGraph = time.Since(buildStart)

	nodes := make(map[string]NodeStats, len(adjacency))
	for node := range adjacency {
		nodes[node] = NodeStats{}
	}

	edgeCount := 0
	for src, targets := range adjacency {
		stats := nodes[src]
		stats.Outbound = len(targets)
		nodes[src] = stats
		edgeCount += len(targets)

		for dst := range targets {
			stats := nodes[dst]
			stats.Inbound++
			nodes[dst] = stats
		}
	}

	sccs := stronglyConnectedComponents(adjacency)
	sccID := assignIDs(sccs, "scc")
	weak := weakComponents(adjacency)
	weakID := assignIDs(weak, "comp")
	lpStart := time.Now()
	comms := labelPropagation(adjacency)
	timings.LabelProp = time.Since(lpStart)

	hitsStart := time.Now()
	hits := computeHITS(adjacency)
	timings.HITS = time.Since(hitsStart)

	graphNodes := make(map[string]GraphNode, len(nodes))
	for path, deg := range nodes {
		graphNodes[path] = GraphNode{
			Path:       path,
			Title:      RemoveMdSuffix(filepath.Base(path)),
			Inbound:    deg.Inbound,
			Outbound:   deg.Outbound,
			Hub:        hits.Hubs[path],
			Authority:  hits.Authorities[path],
			Community:  comms[path],
			SCC:        sccID[path],
			Neighbors:  sortedKeys(adjacency[path]),
			Tags:       tagMap[path],
			WeakCompID: weakID[path],
		}
	}

	cascade := options.RecencyCascade
	if !options.RecencyCascadeSet {
		cascade = true
	}

	recencyStart := time.Now()
	effectiveTimes := applyNeighborRecency(adjacency, contentTimes, time.Now(), cascade)
	timings.Recency = time.Since(recencyStart)
	communities := summarizeCommunities(comms, graphNodes, tagMap, effectiveTimes)
	bridges := computeBridges(adjacency, graphNodes, communities)
	attachBridges(communities, bridges)

	timings.Total = time.Since(startTotal)

	// Ensure timings are non-zero to make them visible when marshaled.
	if timings.LoadEntries == 0 {
		timings.LoadEntries = 1 * time.Nanosecond
	}
	if timings.BuildGraph == 0 {
		timings.BuildGraph = 1 * time.Nanosecond
	}
	if timings.HITS == 0 {
		timings.HITS = 1 * time.Nanosecond
	}
	if timings.LabelProp == 0 {
		timings.LabelProp = 1 * time.Nanosecond
	}
	if timings.Recency == 0 {
		timings.Recency = 1 * time.Nanosecond
	}
	if timings.Total == 0 {
		timings.Total = 1 * time.Nanosecond
	}

	return &GraphAnalysis{
		Nodes:            graphNodes,
		Communities:      communities,
		StrongComponents: sccs,
		WeakComponents:   weak,
		Orphans:          (&GraphStats{Nodes: nodes}).Orphans(),
		Stats: GraphStatsSummary{
			NodeCount: len(nodes),
			EdgeCount: edgeCount,
		},
		EffectiveTimes: effectiveTimes,
		Timings:        timings,
	}, nil
}

func summarizeCommunities(labels map[string]string, nodes map[string]GraphNode, tags map[string][]string, modTimes map[string]time.Time) []CommunitySummary {
	grouped := make(map[string][]string)
	for node, label := range labels {
		grouped[label] = append(grouped[label], node)
	}

	var summaries []CommunitySummary
	for id, members := range grouped {
		sort.Strings(members)
		topAuth := topAuthorityNodes(members, nodes, 5)
		topTags := topTagsForCommunity(members, tags, 5)
		anchor := anchorForCommunity(members, nodes)
		buckets, stats := authorityBuckets(members, nodes)
		recency := communityRecency(members, modTimes, communityRecencyWindowDays)
		summaries = append(summaries, CommunitySummary{
			ID:               communityID(id, anchor, members),
			Nodes:            members,
			TopAuthority:     topAuth,
			AuthorityBuckets: buckets,
			AuthorityStats:   stats,
			Recency:          recency,
			TopTags:          topTags,
			Anchor:           anchor,
			Density:          density(members, nodes),
		})
	}

	sort.Slice(summaries, func(i, j int) bool {
		ri := summaries[i].Recency
		rj := summaries[j].Recency

		// Prefer more recently updated communities (smaller age).
		if ri != nil && rj != nil && ri.LatestAgeDays != rj.LatestAgeDays {
			return ri.LatestAgeDays < rj.LatestAgeDays
		}
		if ri != nil && rj == nil {
			return true
		}
		if ri == nil && rj != nil {
			return false
		}

		// Fallback: larger communities first.
		if len(summaries[i].Nodes) == len(summaries[j].Nodes) {
			return summaries[i].ID < summaries[j].ID
		}
		return len(summaries[i].Nodes) > len(summaries[j].Nodes)
	})
	return summaries
}

const (
	// neighborFreshWindow defines how far back neighbor activity can boost an undated note.
	// A neighbor older than this won't contribute to inferred recency.
	neighborFreshWindow = 180 * 24 * time.Hour

	// neighborStalenessOffset is subtracted from neighbor timestamps when inferring recency,
	// so a hub's inferred time is at least this many days older than its freshest neighbor.
	neighborStalenessOffset = 7 * 24 * time.Hour

	// neighborSampleLimit caps the number of neighbors considered for recency inference.
	neighborSampleLimit = 5

	// minSaneYear rejects dates before this year as implausible.
	minSaneYear = 1900

	// maxFutureTolerance allows dates up to this far in the future (for scheduled events).
	maxFutureTolerance = 365 * 24 * time.Hour

	// topLinesForDateDetection limits how many lines to scan for ISO dates near the top of a note.
	topLinesForDateDetection = 20

	// communityRecencyWindowDays is the window for counting "recent" notes in a community.
	// This differs from neighborFreshWindow intentionally: neighbor inference looks back 6 months
	// to boost undated hubs, while community RecentCount measures short-term activity (30 days).
	communityRecencyWindowDays = 30

	// recencyPropagationPasses bounds how many times inferred recency can cascade.
	recencyPropagationPasses = 2
)

var (
	isoDateRegex     = regexp.MustCompile(`\d{4}-\d{2}-\d{2}(?:[T _]?\d{2}:?\d{2}(?::?\d{2})?)?`)
	headingDateRegex = regexp.MustCompile(`(?m)^\s{0,3}#*\s*(\d{4}-\d{2}-\d{2}(?:[T ]\d{2}:\d{2}(?::\d{2})?)?)\b`)
)

func applyNeighborRecency(adjacency map[string]map[string]struct{}, baseTimes map[string]time.Time, now time.Time, cascade bool) map[string]time.Time {
	inbound := make(map[string][]string)
	for src, targets := range adjacency {
		for dst := range targets {
			inbound[dst] = append(inbound[dst], src)
		}
	}

	nodes := make([]string, 0, len(adjacency))
	for node := range adjacency {
		nodes = append(nodes, node)
	}
	sort.Strings(nodes)

	clampedBase := make(map[string]time.Time, len(baseTimes))
	for path, ts := range baseTimes {
		if ts.IsZero() {
			continue
		}
		if ts.After(now) {
			ts = now
		}
		clampedBase[path] = ts
	}

	if !cascade {
		return recencyPass(nodes, adjacency, inbound, clampedBase, clampedBase, now, nil)
	}

	current := make(map[string]time.Time, len(adjacency))
	for _, node := range nodes {
		if ts := clampedBase[node]; !ts.IsZero() {
			current[node] = ts
		}
	}

	for pass := 0; pass < recencyPropagationPasses; pass++ {
		next := recencyPass(nodes, adjacency, inbound, clampedBase, current, now, current)
		current = next
	}

	return current
}

func recencyPass(nodes []string, adjacency map[string]map[string]struct{}, inbound map[string][]string, base map[string]time.Time, neighborTimes map[string]time.Time, now time.Time, previous map[string]time.Time) map[string]time.Time {
	type neighborTime struct {
		path string
		ts   time.Time
	}

	effective := make(map[string]time.Time, len(adjacency))

	workerCount := runtime.NumCPU()
	if workerCount < 1 {
		workerCount = 1
	}
	if workerCount > len(nodes) {
		workerCount = len(nodes)
	}

	chunkSize := (len(nodes) + workerCount - 1) / workerCount
	partials := make([]map[string]time.Time, workerCount)

	var wg sync.WaitGroup
	for w := 0; w < workerCount; w++ {
		start := w * chunkSize
		end := start + chunkSize
		if end > len(nodes) {
			end = len(nodes)
		}
		if start >= end {
			continue
		}
		wg.Add(1)
		go func(idx int, slice []string) {
			defer wg.Done()
			local := make(map[string]time.Time, len(slice))
			for _, node := range slice {
				best := base[node]
				if prev, ok := previous[node]; ok && prev.After(best) {
					best = prev
				}

				seen := make(map[string]struct{})
				var neighbors []neighborTime

				for dst := range adjacency[node] {
					if ts := neighborTimes[dst]; !ts.IsZero() {
						if ts.After(now) {
							ts = now
						}
						if _, dup := seen[dst]; !dup {
							seen[dst] = struct{}{}
							neighbors = append(neighbors, neighborTime{path: dst, ts: ts})
						}
					}
				}
				for _, src := range inbound[node] {
					if ts := neighborTimes[src]; !ts.IsZero() {
						if ts.After(now) {
							ts = now
						}
						if _, dup := seen[src]; !dup {
							seen[src] = struct{}{}
							neighbors = append(neighbors, neighborTime{path: src, ts: ts})
						}
					}
				}

				sort.Slice(neighbors, func(i, j int) bool {
					return neighbors[i].ts.After(neighbors[j].ts)
				})
				if len(neighbors) > neighborSampleLimit {
					neighbors = neighbors[:neighborSampleLimit]
				}

				for _, n := range neighbors {
					age := now.Sub(n.ts)
					if age < -maxFutureTolerance {
						continue
					}
					if age > neighborFreshWindow {
						continue
					}
					adjusted := n.ts.Add(-neighborStalenessOffset)
					if best.IsZero() || adjusted.After(best) {
						best = adjusted
					}
				}

				if !best.IsZero() {
					local[node] = best
				}
			}
			partials[idx] = local
		}(w, nodes[start:end])
	}

	wg.Wait()

	for _, part := range partials {
		for k, v := range part {
			effective[k] = v
		}
	}

	return effective
}

// ResolveContentTime extracts a content-driven timestamp from a note using frontmatter,
// filename dates, and in-note headings/lines. It rejects non-ISO or implausible dates.
func ResolveContentTime(path, content string) (time.Time, bool) {
	now := time.Now()

	if t, ok := parseDateFromFrontmatter(content, now); ok {
		return clampFuture(t, now), true
	}
	if t, ok := parseDateFromFilename(path, now); ok {
		return clampFuture(t, now), true
	}
	if t, ok := parseMostRecentHeadingDate(content, now); ok {
		return clampFuture(t, now), true
	}
	if t, ok := parseDateFromTopLines(content, now, topLinesForDateDetection); ok {
		return clampFuture(t, now), true
	}
	return time.Time{}, false
}

func parseDateFromFrontmatter(content string, now time.Time) (time.Time, bool) {
	frontmatter, err := ExtractFrontmatter(content)
	if err != nil || frontmatter == nil {
		return time.Time{}, false
	}
	// Prefer event/meeting, then "last updated" style fields before creation.
	candidates := []string{"event_date", "meeting_date", "updated", "modified", "date", "created"}
	for _, key := range candidates {
		if raw, ok := frontmatter[key]; ok {
			if t, ok := parseDateValue(raw, now); ok {
				return t, true
			}
		}
	}
	return time.Time{}, false
}

func parseDateValue(raw interface{}, now time.Time) (time.Time, bool) {
	switch v := raw.(type) {
	case string:
		return parseISOTimestamp(v, now)
	case time.Time:
		if saneTimestamp(v, now) {
			return v, true
		}
	case *time.Time:
		if v != nil && saneTimestamp(*v, now) {
			return *v, true
		}
	case []interface{}:
		for _, item := range v {
			if t, ok := parseDateValue(item, now); ok {
				return t, true
			}
		}
	}
	return time.Time{}, false
}

func parseDateFromFilename(path string, now time.Time) (time.Time, bool) {
	match := isoDateRegex.FindString(path)
	if match == "" {
		return time.Time{}, false
	}
	return parseISOTimestamp(match, now)
}

func parseDateFromTopLines(content string, now time.Time, limit int) (time.Time, bool) {
	lines := strings.Split(content, "\n")
	if limit > len(lines) {
		limit = len(lines)
	}
	for i := 0; i < limit; i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if match := isoDateRegex.FindString(line); match != "" {
			if t, ok := parseISOTimestamp(match, now); ok {
				return t, true
			}
		}
	}
	return time.Time{}, false
}

func parseMostRecentHeadingDate(content string, now time.Time) (time.Time, bool) {
	matches := headingDateRegex.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return time.Time{}, false
	}
	var latest time.Time
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		candidate := m[1]
		if t, ok := parseISOTimestamp(candidate, now); ok {
			if latest.IsZero() || t.After(latest) {
				latest = t
			}
		}
	}
	if latest.IsZero() {
		return time.Time{}, false
	}
	return latest, true
}

func parseISOTimestamp(value string, now time.Time) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02",
		"2006-01-02T15:04",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04",
		"2006-01-02 15:04:05",
		"2006-01-02T1504",
		"2006-01-02T150405",
	}
	candidate := value
	if m := isoDateRegex.FindString(value); m != "" {
		candidate = m
	}
	for _, layout := range layouts {
		if ts, err := time.Parse(layout, candidate); err == nil {
			if saneTimestamp(ts, now) {
				return ts.UTC(), true
			}
		}
	}
	return time.Time{}, false
}

func clampFuture(ts time.Time, now time.Time) time.Time {
	if ts.After(now) {
		return now
	}
	return ts
}

func saneTimestamp(ts time.Time, now time.Time) bool {
	if ts.IsZero() {
		return false
	}
	if ts.Year() < minSaneYear {
		return false
	}
	if ts.After(now.Add(maxFutureTolerance)) {
		return false
	}
	return true
}

func topAuthorityNodes(members []string, nodes map[string]GraphNode, limit int) []AuthorityScore {
	type pr struct {
		path      string
		authority float64
		hub       float64
	}
	var list []pr
	for _, m := range members {
		list = append(list, pr{path: m, authority: nodes[m].Authority, hub: nodes[m].Hub})
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].authority == list[j].authority {
			return list[i].path < list[j].path
		}
		return list[i].authority > list[j].authority
	})
	if len(list) > limit {
		list = list[:limit]
	}
	out := make([]AuthorityScore, len(list))
	for i, item := range list {
		out[i] = AuthorityScore{
			Path:      item.path,
			Authority: item.authority,
			Hub:       item.hub,
		}
	}
	return out
}

func authorityBuckets(members []string, nodes map[string]GraphNode) ([]AuthorityBucket, *AuthorityStats) {
	if len(members) == 0 {
		return nil, nil
	}
	type pr struct {
		path string
		val  float64
	}
	values := make([]pr, 0, len(members))
	for _, m := range members {
		values = append(values, pr{path: m, val: nodes[m].Authority})
	}
	sort.Slice(values, func(i, j int) bool { return values[i].val > values[j].val })

	bucketCount := bucketCountFor(len(values))
	buckets := make([]AuthorityBucket, 0, bucketCount)
	size := int(math.Ceil(float64(len(values)) / float64(bucketCount)))
	for i := 0; i < len(values); i += size {
		end := i + size
		if end > len(values) {
			end = len(values)
		}
		segment := values[i:end]
		high := segment[0].val
		low := segment[len(segment)-1].val
		example := segment[0].path
		buckets = append(buckets, AuthorityBucket{
			Low:     low,
			High:    high,
			Count:   len(segment),
			Example: example,
		})
	}

	// Convert to simple slice for stats
	vals := make([]struct {
		path string
		val  float64
	}, len(values))
	for i, v := range values {
		vals[i] = struct {
			path string
			val  float64
		}{path: v.path, val: v.val}
	}
	stats := authorityStatsFromValues(vals)
	return buckets, stats
}

func bucketCountFor(size int) int {
	if size <= 0 {
		return 0
	}
	c := int(math.Ceil(math.Sqrt(float64(size))))
	if c < 5 {
		c = 5
	}
	if c > 10 {
		c = 10
	}
	return c
}

func authorityStatsFromValues(values []struct {
	path string
	val  float64
}) *AuthorityStats {
	if len(values) == 0 {
		return nil
	}
	vals := make([]float64, len(values))
	sum := 0.0
	for i, v := range values {
		vals[i] = v.val
		sum += v.val
	}
	mean := sum / float64(len(vals))
	sort.Float64s(vals) // ascending
	p := func(q float64) float64 {
		if len(vals) == 1 {
			return vals[0]
		}
		idx := int(math.Ceil(q*float64(len(vals)))) - 1
		if idx < 0 {
			idx = 0
		}
		if idx >= len(vals) {
			idx = len(vals) - 1
		}
		return vals[idx]
	}
	return &AuthorityStats{
		Mean: mean,
		P50:  p(0.50),
		P75:  p(0.75),
		P90:  p(0.90),
		P95:  p(0.95),
		P99:  p(0.99),
		Max:  vals[len(vals)-1],
	}
}

func communityRecency(members []string, modTimes map[string]time.Time, windowDays int) *GraphRecency {
	if len(members) == 0 || windowDays <= 0 {
		return nil
	}
	latestPath := ""
	var latestTime time.Time
	recentCount := 0
	window := time.Duration(windowDays) * 24 * time.Hour
	now := time.Now()

	for _, m := range members {
		mt, ok := modTimes[m]
		if !ok || mt.IsZero() {
			continue
		}
		if latestTime.IsZero() || mt.After(latestTime) {
			latestTime = mt
			latestPath = m
		}
		if now.Sub(mt) <= window {
			recentCount++
		}
	}
	if latestTime.IsZero() && recentCount == 0 {
		return nil
	}
	age := 0.0
	if !latestTime.IsZero() {
		age = now.Sub(latestTime).Hours() / 24.0
		if age < 0 {
			age = 0
		}
	}
	return &GraphRecency{
		LatestPath:      latestPath,
		LatestAgeDays:   age,
		LatestTimestamp: latestTime,
		RecentCount:     recentCount,
		WindowDays:      windowDays,
	}
}

func anchorForCommunity(members []string, nodes map[string]GraphNode) string {
	if len(members) == 0 {
		return ""
	}
	type pr struct {
		path string
		val  float64
	}
	var list []pr
	for _, m := range members {
		list = append(list, pr{path: m, val: nodes[m].Authority})
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].val == list[j].val {
			return list[i].path < list[j].path
		}
		return list[i].val > list[j].val
	})
	return list[0].path
}

func density(members []string, nodes map[string]GraphNode) float64 {
	if len(members) < 2 {
		return 0
	}
	// Approximate density using average degree within the community / (n-1)
	// (We don't keep intra-community edge counts here; this is a simple heuristic.)
	var totalDegree int
	for _, m := range members {
		totalDegree += nodes[m].Inbound + nodes[m].Outbound
	}
	n := float64(len(members))
	return (float64(totalDegree) / n) / (n - 1)
}

func communityID(label, anchor string, members []string) string {
	h := sha1.New()
	io.WriteString(h, label)
	io.WriteString(h, "|")
	io.WriteString(h, anchor)
	io.WriteString(h, "|")
	for _, m := range members {
		io.WriteString(h, m)
		io.WriteString(h, ";")
	}
	sum := hex.EncodeToString(h.Sum(nil))
	if len(sum) > 8 {
		sum = sum[:8]
	}
	return "c" + sum
}

func topTagsForCommunity(members []string, tags map[string][]string, limit int) []TagCount {
	counts := make(map[string]int)
	for _, m := range members {
		for _, t := range tags[m] {
			counts[strings.ToLower(t)]++
		}
	}
	type kv struct {
		tag   string
		count int
	}
	var list []kv
	for tag, count := range counts {
		list = append(list, kv{tag: tag, count: count})
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].count == list[j].count {
			return list[i].tag < list[j].tag
		}
		return list[i].count > list[j].count
	})
	if len(list) > limit {
		list = list[:limit]
	}
	out := make([]TagCount, len(list))
	for i, item := range list {
		out[i] = TagCount{Tag: item.tag, Count: item.count}
	}
	return out
}

// CommunityMembershipLookup returns a map from note path to its community summary for quick lookups.
func CommunityMembershipLookup(communities []CommunitySummary) map[string]*CommunitySummary {
	lookup := make(map[string]*CommunitySummary)
	for i := range communities {
		comm := &communities[i]
		for _, node := range comm.Nodes {
			lookup[node] = comm
		}
	}
	return lookup
}

// CommunityInternalEdges counts directed edges whose endpoints both live in the community.
func CommunityInternalEdges(comm *CommunitySummary, nodes map[string]GraphNode) int {
	if comm == nil {
		return 0
	}
	memberSet := make(map[string]struct{}, len(comm.Nodes))
	for _, n := range comm.Nodes {
		memberSet[n] = struct{}{}
	}

	edgeCount := 0
	for _, n := range comm.Nodes {
		node := nodes[n]
		for _, neigh := range node.Neighbors {
			if _, ok := memberSet[neigh]; ok {
				edgeCount++
			}
		}
	}
	return edgeCount
}

func assignIDs(components [][]string, prefix string) map[string]string {
	ids := make(map[string]string)
	for idx, comp := range components {
		id := fmt.Sprintf("%s%d", prefix, idx)
		for _, node := range comp {
			ids[node] = id
		}
	}
	return ids
}

func weakComponents(adjacency map[string]map[string]struct{}) [][]string {
	visited := make(map[string]bool)
	var comps [][]string

	for node := range adjacency {
		if visited[node] {
			continue
		}
		var queue []string
		queue = append(queue, node)
		visited[node] = true
		var comp []string
		for len(queue) > 0 {
			cur := queue[0]
			queue = queue[1:]
			comp = append(comp, cur)

			for neigh := range adjacency[cur] {
				if !visited[neigh] {
					visited[neigh] = true
					queue = append(queue, neigh)
				}
			}
			for neigh, targets := range adjacency {
				if _, ok := targets[cur]; ok && !visited[neigh] {
					visited[neigh] = true
					queue = append(queue, neigh)
				}
			}
		}
		sort.Strings(comp)
		comps = append(comps, comp)
	}

	sort.Slice(comps, func(i, j int) bool {
		if len(comps[i]) == len(comps[j]) {
			return comps[i][0] < comps[j][0]
		}
		return len(comps[i]) > len(comps[j])
	})
	return comps
}

func computeBridges(adjacency map[string]map[string]struct{}, nodes map[string]GraphNode, comms []CommunitySummary) map[string][]string {
	communityByNode := make(map[string]string)
	for _, c := range comms {
		for _, n := range c.Nodes {
			communityByNode[n] = c.ID
		}
	}

	// Betweenness approx: count cross-community edges.
	bridgeCount := make(map[string]int)
	for src, targets := range adjacency {
		for dst := range targets {
			if communityByNode[src] != "" && communityByNode[src] != communityByNode[dst] {
				bridgeCount[src]++
				bridgeCount[dst]++
			}
		}
	}

	result := make(map[string][]string)
	for _, c := range comms {
		var candidates []string
		for _, n := range c.Nodes {
			if bridgeCount[n] > 0 {
				candidates = append(candidates, n)
			}
		}
		sort.Slice(candidates, func(i, j int) bool {
			if bridgeCount[candidates[i]] == bridgeCount[candidates[j]] {
				return nodes[candidates[i]].Authority > nodes[candidates[j]].Authority
			}
			return bridgeCount[candidates[i]] > bridgeCount[candidates[j]]
		})
		if len(candidates) > 5 {
			candidates = candidates[:5]
		}
		result[c.ID] = candidates
	}
	return result
}

func attachBridges(comms []CommunitySummary, bridges map[string][]string) {
	for i := range comms {
		if bs, ok := bridges[comms[i].ID]; ok {
			comms[i].Bridges = bs
		}
	}
}

// HITSResult contains the hub and authority scores from HITS algorithm.
type HITSResult struct {
	Hubs        map[string]float64
	Authorities map[string]float64
}

// computeHITS computes HITS (Hyperlink-Induced Topic Search) scores.
// Hub score: measures how well a note curates/aggregates links to good authorities.
// Authority score: measures how often a note is referenced by good hubs.
func computeHITS(adjacency map[string]map[string]struct{}) HITSResult {
	const iterations = 30

	n := len(adjacency)
	if n == 0 {
		return HITSResult{
			Hubs:        map[string]float64{},
			Authorities: map[string]float64{},
		}
	}

	// Build reverse adjacency for efficient authority computation
	// reverse[dst] = set of nodes that link TO dst
	reverse := make(map[string]map[string]struct{}, n)
	for node := range adjacency {
		reverse[node] = make(map[string]struct{})
	}
	for src, targets := range adjacency {
		for dst := range targets {
			if _, ok := reverse[dst]; !ok {
				reverse[dst] = make(map[string]struct{})
			}
			reverse[dst][src] = struct{}{}
		}
	}

	// Initialize scores
	hub := make(map[string]float64, n)
	auth := make(map[string]float64, n)
	for node := range adjacency {
		hub[node] = 1.0
		auth[node] = 1.0
	}

	// Iterative refinement
	for i := 0; i < iterations; i++ {
		// Update authority scores: auth(p) = sum of hub(q) for all q that link to p
		newAuth := make(map[string]float64, n)
		for node := range adjacency {
			sum := 0.0
			for src := range reverse[node] {
				sum += hub[src]
			}
			newAuth[node] = sum
		}

		// Update hub scores: hub(p) = sum of auth(q) for all q that p links to
		newHub := make(map[string]float64, n)
		for node, targets := range adjacency {
			sum := 0.0
			for dst := range targets {
				sum += newAuth[dst] // Use updated authority scores
			}
			newHub[node] = sum
		}

		// Normalize to prevent score explosion
		authNorm := 0.0
		hubNorm := 0.0
		for node := range adjacency {
			authNorm += newAuth[node] * newAuth[node]
			hubNorm += newHub[node] * newHub[node]
		}
		authNorm = math.Sqrt(authNorm)
		hubNorm = math.Sqrt(hubNorm)

		if authNorm > 0 {
			for node := range adjacency {
				newAuth[node] /= authNorm
			}
		}
		if hubNorm > 0 {
			for node := range adjacency {
				newHub[node] /= hubNorm
			}
		}

		auth = newAuth
		hub = newHub
	}

	return HITSResult{
		Hubs:        hub,
		Authorities: auth,
	}
}

func sortedKeys(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func extractTags(content string) []string {
	tags := make(map[string]struct{})

	frontmatter, err := ExtractFrontmatter(content)
	if err == nil && frontmatter != nil {
		if raw, ok := frontmatter["tags"]; ok {
			if slice, ok := raw.([]string); ok {
				for _, t := range slice {
					normalized := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(t), "#"))
					if normalized != "" {
						tags[normalized] = struct{}{}
					}
				}
			}
		}
	}

	for _, ht := range ExtractHashtags(content) {
		normalized := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(ht), "#"))
		if normalized != "" {
			tags[normalized] = struct{}{}
		}
	}

	if len(tags) == 0 {
		return nil
	}
	return sortedKeys(tags)
}

func filterMutualEdges(adjacency map[string]map[string]struct{}) map[string]map[string]struct{} {
	mutual := make(map[string]map[string]struct{}, len(adjacency))
	for src := range adjacency {
		mutual[src] = make(map[string]struct{})
	}
	for src, targets := range adjacency {
		for dst := range targets {
			if _, ok := adjacency[dst][src]; ok {
				mutual[src][dst] = struct{}{}
			}
		}
	}
	return mutual
}

func filterByMinDegree(adjacency map[string]map[string]struct{}, min int) map[string]map[string]struct{} {
	if min <= 0 {
		return adjacency
	}

	adj := adjacency
	for {
		inDeg := make(map[string]int)
		for src, targets := range adj {
			for dst := range targets {
				inDeg[dst]++
			}
			if _, ok := inDeg[src]; !ok {
				inDeg[src] = inDeg[src] // ensure key exists
			}
		}

		var toRemove []string
		for node, targets := range adj {
			deg := len(targets) + inDeg[node]
			if deg < min {
				toRemove = append(toRemove, node)
			}
		}

		if len(toRemove) == 0 {
			break
		}

		for _, n := range toRemove {
			delete(adj, n)
		}
		for _, targets := range adj {
			for dst := range targets {
				if _, ok := adj[dst]; !ok {
					delete(targets, dst)
				}
			}
		}
	}

	return adj
}

// labelPropagation performs synchronous label propagation on an undirected view of the graph.
func labelPropagation(adjacency map[string]map[string]struct{}) map[string]string {
	// Build undirected neighbor sets
	neighbors := make(map[string]map[string]struct{}, len(adjacency))
	for src, targets := range adjacency {
		if _, ok := neighbors[src]; !ok {
			neighbors[src] = make(map[string]struct{})
		}
		for dst := range targets {
			if src == dst {
				continue
			}
			neighbors[src][dst] = struct{}{}
			if _, ok := neighbors[dst]; !ok {
				neighbors[dst] = make(map[string]struct{})
			}
			neighbors[dst][src] = struct{}{}
		}
	}
	for node := range adjacency {
		if _, ok := neighbors[node]; !ok {
			neighbors[node] = make(map[string]struct{})
		}
	}

	labels := make(map[string]string, len(neighbors))
	for node := range neighbors {
		labels[node] = node
	}

	nodes := make([]string, 0, len(neighbors))
	for node := range neighbors {
		nodes = append(nodes, node)
	}
	sort.Strings(nodes)

	const maxIter = 20
	for iter := 0; iter < maxIter; iter++ {
		changed := false
		for _, node := range nodes {
			counts := make(map[string]int)
			for neigh := range neighbors[node] {
				counts[labels[neigh]]++
			}
			if len(counts) == 0 {
				continue
			}
			var bestLabel string
			bestCount := -1
			for label, count := range counts {
				if count > bestCount || (count == bestCount && label < bestLabel) {
					bestLabel = label
					bestCount = count
				}
			}
			if bestLabel != "" && bestLabel != labels[node] {
				labels[node] = bestLabel
				changed = true
			}
		}
		if !changed {
			break
		}
	}

	return labels
}
