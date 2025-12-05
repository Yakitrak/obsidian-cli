package cache

import (
	"context"
	"testing"

	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
	"github.com/stretchr/testify/require"
)

type stubSnapshotProvider struct {
	version uint64
}

func (s *stubSnapshotProvider) EntriesSnapshot(context.Context) ([]Entry, error) { return nil, nil }
func (s *stubSnapshotProvider) Version() uint64                                  { return s.version }

type countingNote struct {
	notes            map[string]string
	contentsCalls    int
	notesListCalls   int
	forceNotesListOK bool
}

func (s *countingNote) Move(string, string) error                { return nil }
func (s *countingNote) Delete(string) error                      { return nil }
func (s *countingNote) UpdateLinks(string, string, string) error { return nil }
func (s *countingNote) GetContents(_ string, name string) (string, error) {
	s.contentsCalls++
	return s.notes[name], nil
}
func (s *countingNote) GetNotesList(string) ([]string, error) {
	s.notesListCalls++
	keys := make([]string, 0, len(s.notes))
	for k := range s.notes {
		keys = append(keys, k)
	}
	return keys, nil
}

func TestAnalysisCacheBacklinksUsesVersion(t *testing.T) {
	provider := &stubSnapshotProvider{version: 1}
	cache := NewAnalysisCache(provider)

	note := &countingNote{
		notes: map[string]string{
			"A.md": "[[Target]]",
		},
	}

	_, err := cache.Backlinks("", note, []string{"Target"}, obsidian.WikilinkOptions{}, nil)
	require.NoError(t, err)
	require.Equal(t, 1, note.contentsCalls)

	// Same version should hit cache.
	_, err = cache.Backlinks("", note, []string{"Target"}, obsidian.WikilinkOptions{}, nil)
	require.NoError(t, err)
	require.Equal(t, 1, note.contentsCalls, "second call should use cache")

	// Version change should invalidate cache and recompute.
	provider.version = 2
	_, err = cache.Backlinks("", note, []string{"Target"}, obsidian.WikilinkOptions{}, nil)
	require.NoError(t, err)
	require.Equal(t, 2, note.contentsCalls, "version change should force recompute")
}

func TestAnalysisCacheGraphVersionInvalidates(t *testing.T) {
	provider := &stubSnapshotProvider{version: 1}
	cache := NewAnalysisCache(provider)
	note := &countingNote{
		notes: map[string]string{
			"A.md": "[[B]]",
			"B.md": "#tag",
		},
	}

	opts := obsidian.GraphAnalysisOptions{}

	analysis1, err := cache.GraphAnalysis("", note, opts)
	require.NoError(t, err)
	require.NotNil(t, analysis1)
	require.Equal(t, 2, note.contentsCalls)

	// Repeat should not re-read.
	_, err = cache.GraphAnalysis("", note, opts)
	require.NoError(t, err)
	require.Equal(t, 2, note.contentsCalls)

	provider.version = 2
	_, err = cache.GraphAnalysis("", note, opts)
	require.NoError(t, err)
	require.Greater(t, note.contentsCalls, 2, "version change should cause re-read")
}
