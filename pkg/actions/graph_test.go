package actions

import (
	"testing"

	"github.com/atomicobject/obsidian-cli/mocks"
	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGraphStatsCountsLinks(t *testing.T) {
	vault := &mocks.VaultManager{}
	note := &mocks.NoteManager{}

	vault.On("Path").Return("/vault", nil)
	note.On("GetNotesList", "/vault").Return([]string{"a.md", "b.md"}, nil)
	note.On("GetContents", "/vault", "a.md").Return("See [[b]]", nil)
	note.On("GetContents", "/vault", "b.md").Return("", nil)

	stats, err := GraphStats(vault, note, obsidian.DefaultWikilinkOptions)
	require.NoError(t, err)

	assert.Equal(t, 1, stats.Nodes["a.md"].Outbound)
	assert.Equal(t, 0, stats.Nodes["a.md"].Inbound)
	assert.Equal(t, 1, stats.Nodes["b.md"].Inbound)
	assert.Equal(t, 0, stats.Nodes["b.md"].Outbound)

	expectedComponents := [][]string{{"a.md"}, {"b.md"}}
	assert.Equal(t, expectedComponents, stats.Components)

	vault.AssertExpectations(t)
	note.AssertExpectations(t)
}

func TestOrphansFiltersEmptyDegreeNotes(t *testing.T) {
	vault := &mocks.VaultManager{}
	note := &mocks.NoteManager{}

	vault.On("Path").Return("/vault", nil)
	note.On("GetNotesList", "/vault").Return([]string{"lonely.md"}, nil)
	note.On("GetContents", "/vault", "lonely.md").Return("Self [[lonely]] reference", nil)

	orphans, err := Orphans(vault, note, obsidian.DefaultWikilinkOptions)
	require.NoError(t, err)
	assert.Equal(t, []string{"lonely.md"}, orphans)

	vault.AssertExpectations(t)
	note.AssertExpectations(t)
}
