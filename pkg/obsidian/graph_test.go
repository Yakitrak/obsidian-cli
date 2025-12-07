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

	buckets, _ := authorityBuckets(members, nodes)

	require.True(t, len(buckets) >= 5 && len(buckets) <= 10)
	total := 0
	for _, b := range buckets {
		total += b.Count
		assert.Contains(t, members, b.Example)
	}
	assert.Equal(t, len(members), total)
	assert.InEpsilon(t, 1.0, buckets[0].High, 1e-9)
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

func TestResolveContentTimePrefersFrontmatter(t *testing.T) {
	content := `---
date: 2025-05-01
event_date: 2025-06-01
---
Body`
	ts, ok := ResolveContentTime("Notes/sample.md", content)
	require.True(t, ok)
	assert.Equal(t, time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC), ts.UTC())
}

func TestResolveContentTimeFromFilename(t *testing.T) {
	ts, ok := ResolveContentTime("Log/2024-12-31 Planning.md", "Body")
	require.True(t, ok)
	assert.Equal(t, time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC), ts.UTC())
}

func TestResolveContentTimeFromHeading(t *testing.T) {
	content := "# 2024-11-25\nSome text"
	ts, ok := ResolveContentTime("Notes/heading.md", content)
	require.True(t, ok)
	assert.Equal(t, time.Date(2024, 11, 25, 0, 0, 0, 0, time.UTC), ts.UTC())
}

func TestApplyNeighborRecencyBoostsUndated(t *testing.T) {
	now := time.Date(2025, time.January, 15, 12, 0, 0, 0, time.UTC)
	adjacency := map[string]map[string]struct{}{
		"moc.md":    {"recent.md": {}},
		"recent.md": {},
	}
	base := map[string]time.Time{
		"recent.md": now.Add(-24 * time.Hour),
	}

	effective := applyNeighborRecency(adjacency, base, now)

	require.Contains(t, effective, "recent.md")
	assert.WithinDuration(t, now.Add(-24*time.Hour), effective["recent.md"], time.Second)

	moc, ok := effective["moc.md"]
	require.True(t, ok)
	ageDays := now.Sub(moc).Hours() / 24.0
	assert.InDelta(t, 8.0, ageDays, 0.5)
}

func TestAuthorityStatsAndBucketsAdaptive(t *testing.T) {
	nodes := map[string]GraphNode{
		"a.md": {Authority: 1.0},
		"b.md": {Authority: 0.9},
		"c.md": {Authority: 0.8},
		"d.md": {Authority: 0.7},
		"e.md": {Authority: 0.6},
		"f.md": {Authority: 0.5},
		"g.md": {Authority: 0.4},
		"h.md": {Authority: 0.3},
		"i.md": {Authority: 0.2},
		"j.md": {Authority: 0.1},
	}
	members := []string{"a.md", "b.md", "c.md", "d.md", "e.md", "f.md", "g.md", "h.md", "i.md", "j.md"}

	buckets, stats := authorityBuckets(members, nodes)

	require.NotNil(t, stats)
	require.True(t, len(buckets) >= 5 && len(buckets) <= 10)
	assert.InEpsilon(t, 0.55, stats.Mean, 0.02)
	assert.InEpsilon(t, 0.50, stats.P50, 0.05)
	assert.InEpsilon(t, 1.0, stats.Max, 1e-9)
	assert.InEpsilon(t, 0.9, stats.P95, 0.2)
}
