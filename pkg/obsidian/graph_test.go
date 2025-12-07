package obsidian

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeGraphStatsAndOrphans(t *testing.T) {
	vaultPath := filepath.Join("..", "..", "mocks", "vaults", "graph")
	stats, err := ComputeGraphStats(vaultPath, &Note{}, DefaultWikilinkOptions)
	require.NoError(t, err)

	expected := map[string]NodeStats{
		"alpha.md":    {Inbound: 1, Outbound: 1},
		"beta.md":     {Inbound: 2, Outbound: 1},
		"gamma.md":    {Inbound: 0, Outbound: 1},
		"orphan.md":   {Inbound: 0, Outbound: 0},
		"selflink.md": {Inbound: 0, Outbound: 0},
	}

	assert.Equal(t, expected, stats.Nodes)

	expectedComponents := [][]string{
		{"alpha.md", "beta.md"},
		{"gamma.md"},
		{"orphan.md"},
		{"selflink.md"},
	}
	assert.Equal(t, expectedComponents, stats.Components)

	assert.Equal(t, []string{"orphan.md", "selflink.md"}, stats.Orphans())
}

func TestCommunitiesLabelPropagation(t *testing.T) {
	vaultPath := filepath.Join("..", "..", "mocks", "vaults", "graph")
	stats, err := ComputeGraphStats(vaultPath, &Note{}, DefaultWikilinkOptions)
	require.NoError(t, err)

	communities := stats.Communities()

	var cluster []string
	for _, c := range communities {
		if len(c.Nodes) >= 3 {
			cluster = c.Nodes
			break
		}
	}

	assert.ElementsMatch(t, []string{"alpha.md", "beta.md", "gamma.md"}, cluster)
}

func TestGraphAnalysisRespectsExcludes(t *testing.T) {
	vaultPath := filepath.Join("..", "..", "mocks", "vaults", "graph")
	analysis, err := ComputeGraphAnalysis(vaultPath, &Note{}, GraphAnalysisOptions{
		WikilinkOptions: DefaultWikilinkOptions,
		ExcludedPaths: map[string]struct{}{
			"gamma.md": {},
		},
	})
	require.NoError(t, err)

	_, hasGamma := analysis.Nodes["gamma.md"]
	assert.False(t, hasGamma)
}

func TestTopAuthorityNodes(t *testing.T) {
	nodes := map[string]GraphNode{
		"a.md": {Authority: 0.9, Hub: 0.2},
		"b.md": {Authority: 0.7, Hub: 0.5},
		"c.md": {Authority: 0.7, Hub: 0.1},
		"d.md": {Authority: 0.4, Hub: 0.9},
	}
	members := []string{"a.md", "b.md", "c.md", "d.md"}

	top := topAuthorityNodes(members, nodes, 3)

	require.Len(t, top, 3)
	assert.Equal(t, AuthorityScore{Path: "a.md", Authority: 0.9, Hub: 0.2}, top[0])
	// Tie on authority breaks by path
	assert.Equal(t, AuthorityScore{Path: "b.md", Authority: 0.7, Hub: 0.5}, top[1])
	assert.Equal(t, AuthorityScore{Path: "c.md", Authority: 0.7, Hub: 0.1}, top[2])
}

func TestAuthorityBuckets(t *testing.T) {
	nodes := map[string]GraphNode{
		"a.md": {Authority: 1.0},
		"b.md": {Authority: 0.8},
		"c.md": {Authority: 0.6},
		"d.md": {Authority: 0.4},
		"e.md": {Authority: 0.2},
	}
	members := []string{"a.md", "b.md", "c.md", "d.md", "e.md"}

	buckets := authorityBuckets(members, nodes, 2)

	require.Len(t, buckets, 2)

	// First bucket should include top half
	assert.InEpsilon(t, 1.0, buckets[0].High, 1e-9)
	assert.InEpsilon(t, 0.6, buckets[0].Low, 1e-9)
	assert.Equal(t, 3, buckets[0].Count)

	// Second bucket covers the rest
	assert.InEpsilon(t, 0.4, buckets[1].High, 1e-9)
	assert.InEpsilon(t, 0.2, buckets[1].Low, 1e-9)
	assert.Equal(t, 2, buckets[1].Count)

	// Examples come from within each bucket
	assert.Contains(t, members, buckets[0].Example)
	assert.Contains(t, members, buckets[1].Example)
}

func TestCommunityRecency(t *testing.T) {
	now := time.Now()
	modTimes := map[string]time.Time{
		"a.md": now.Add(-48 * time.Hour),      // 2 days ago
		"b.md": now.Add(-10 * 24 * time.Hour), // 10 days ago
		"c.md": now.Add(-40 * 24 * time.Hour), // 40 days ago (outside 30d window)
	}
	members := []string{"a.md", "b.md", "c.md"}

	recency := communityRecency(members, modTimes, 30)

	require.NotNil(t, recency)
	assert.Equal(t, "a.md", recency.LatestPath)
	assert.InDelta(t, 2.0, recency.LatestAgeDays, 0.3)
	assert.Equal(t, 2, recency.RecentCount) // a and b are within 30d
	assert.Equal(t, 30, recency.WindowDays)
}
