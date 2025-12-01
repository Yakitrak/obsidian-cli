package obsidian

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
)

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
	ID          string
	Nodes       []string
	TopTags     []TagCount
	TopPagerank []string
	Anchor      string
	Density     float64
	Bridges     []string
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
	Pagerank   float64  `json:"pagerank"`
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

// ComputeGraphAnalysis builds a rich graph view including pagerank, communities, and tags.
func ComputeGraphAnalysis(vaultPath string, note NoteManager, options GraphAnalysisOptions) (*GraphAnalysis, error) {
	allNotes, err := note.GetNotesList(vaultPath)
	if err != nil {
		return nil, err
	}

	cache := BuildNotePathCache(allNotes)

	excluded := options.ExcludedPaths
	included := options.IncludedPaths

	adjacency := make(map[string]map[string]struct{}, len(allNotes))
	tagMap := make(map[string][]string)

	for _, path := range allNotes {
		normalized := NormalizePath(AddMdSuffix(path))
		if _, skip := excluded[normalized]; skip {
			continue
		}
		if len(included) > 0 {
			if _, ok := included[normalized]; !ok {
				continue
			}
		}
		adjacency[normalized] = make(map[string]struct{})
	}

	for _, path := range allNotes {
		normalized := NormalizePath(AddMdSuffix(path))
		if _, skip := excluded[normalized]; skip {
			continue
		}
		if len(included) > 0 {
			if _, ok := included[normalized]; !ok {
				continue
			}
		}

		content, err := note.GetContents(vaultPath, path)
		if err != nil {
			return nil, err
		}

		if options.IncludeTags {
			tagMap[normalized] = extractTags(content)
		}

		links := scanWikilinks(content, options.WikilinkOptions)
		src := normalized

		for _, link := range links {
			resolved, ok := cache.ResolveNote(link.Target)
			if !ok {
				continue
			}

			dst := NormalizePath(AddMdSuffix(resolved))
			if src == dst {
				continue
			}
			adjacency[src][dst] = struct{}{}
		}
	}

	if options.MutualOnly {
		adjacency = filterMutualEdges(adjacency)
	}
	if options.MinDegree > 0 {
		adjacency = filterByMinDegree(adjacency, options.MinDegree)
	}

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
	comms := labelPropagation(adjacency)
	pagerank := computePagerank(adjacency)

	graphNodes := make(map[string]GraphNode, len(nodes))
	for path, deg := range nodes {
		graphNodes[path] = GraphNode{
			Path:       path,
			Title:      RemoveMdSuffix(filepath.Base(path)),
			Inbound:    deg.Inbound,
			Outbound:   deg.Outbound,
			Pagerank:   pagerank[path],
			Community:  comms[path],
			SCC:        sccID[path],
			Neighbors:  sortedKeys(adjacency[path]),
			Tags:       tagMap[path],
			WeakCompID: weakID[path],
		}
	}

	communities := summarizeCommunities(comms, graphNodes, tagMap)
	bridges := computeBridges(adjacency, graphNodes, communities)
	attachBridges(communities, bridges)

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
	}, nil
}

func summarizeCommunities(labels map[string]string, nodes map[string]GraphNode, tags map[string][]string) []CommunitySummary {
	grouped := make(map[string][]string)
	for node, label := range labels {
		grouped[label] = append(grouped[label], node)
	}

	var summaries []CommunitySummary
	for id, members := range grouped {
		sort.Strings(members)
		topPR := topPagerankNodes(members, nodes, 5)
		topTags := topTagsForCommunity(members, tags, 5)
		anchor := anchorForCommunity(members, nodes)
		summaries = append(summaries, CommunitySummary{
			ID:          communityID(id, anchor, members),
			Nodes:       members,
			TopPagerank: topPR,
			TopTags:     topTags,
			Anchor:      anchor,
			Density:     density(members, nodes),
		})
	}

	sort.Slice(summaries, func(i, j int) bool {
		if len(summaries[i].Nodes) == len(summaries[j].Nodes) {
			return summaries[i].ID < summaries[j].ID
		}
		return len(summaries[i].Nodes) > len(summaries[j].Nodes)
	})
	return summaries
}

func topPagerankNodes(members []string, nodes map[string]GraphNode, limit int) []string {
	type pr struct {
		path string
		val  float64
	}
	var list []pr
	for _, m := range members {
		list = append(list, pr{path: m, val: nodes[m].Pagerank})
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
		list = append(list, pr{path: m, val: nodes[m].Pagerank})
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
				return nodes[candidates[i]].Pagerank > nodes[candidates[j]].Pagerank
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

func computePagerank(adjacency map[string]map[string]struct{}) map[string]float64 {
	const damping = 0.85
	const iterations = 30

	n := len(adjacency)
	if n == 0 {
		return map[string]float64{}
	}
	init := 1.0 / float64(n)

	rank := make(map[string]float64, n)
	for node := range adjacency {
		rank[node] = init
	}

	for i := 0; i < iterations; i++ {
		next := make(map[string]float64, n)
		sinkSum := 0.0
		for node, targets := range adjacency {
			if len(targets) == 0 {
				sinkSum += rank[node]
			}
		}
		for node := range adjacency {
			next[node] = (1 - damping) / float64(n)
			next[node] += damping * sinkSum / float64(n)
		}
		for src, targets := range adjacency {
			if len(targets) == 0 {
				continue
			}
			share := damping * rank[src] / float64(len(targets))
			for dst := range targets {
				next[dst] += share
			}
		}
		rank = next
	}
	return rank
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
