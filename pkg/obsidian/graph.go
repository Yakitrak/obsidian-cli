package obsidian

import (
	"sort"
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
	}, nil
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
