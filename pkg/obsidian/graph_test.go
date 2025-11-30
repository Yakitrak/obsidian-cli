package obsidian

import (
	"path/filepath"
	"testing"

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
