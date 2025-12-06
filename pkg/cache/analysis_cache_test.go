package cache

import (
	"context"
	"errors"
	"testing"

	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
	"github.com/stretchr/testify/assert"
)

type fakeProvider struct {
	version       uint64
	refreshCalls  int
	refreshErr    error
	snapshotCalls int
}

func (p *fakeProvider) EntriesSnapshot(ctx context.Context) ([]Entry, error) {
	p.snapshotCalls++
	return nil, nil
}

func (p *fakeProvider) Version() uint64 {
	return p.version
}

func (p *fakeProvider) Refresh(ctx context.Context) error {
	p.refreshCalls++
	if p.refreshErr != nil {
		return p.refreshErr
	}
	p.version++
	return nil
}

type fakeNoteManager struct {
	notes map[string]string
}

func (n *fakeNoteManager) Move(originalPath string, newPath string) error { return nil }
func (n *fakeNoteManager) Delete(path string) error                       { return nil }
func (n *fakeNoteManager) UpdateLinks(vaultPath string, oldNoteName string, newNoteName string) error {
	return nil
}

func (n *fakeNoteManager) GetContents(vaultPath string, noteName string) (string, error) {
	key := obsidian.NormalizePath(obsidian.AddMdSuffix(noteName))
	content, ok := n.notes[key]
	if !ok {
		return "", errors.New(obsidian.NoteDoesNotExistError)
	}
	return content, nil
}

func (n *fakeNoteManager) GetNotesList(vaultPath string) ([]string, error) {
	out := make([]string, 0, len(n.notes))
	for k := range n.notes {
		out = append(out, k)
	}
	return out, nil
}

func TestAnalysisCacheRefreshesProviderOnCacheHit(t *testing.T) {
	provider := &fakeProvider{}
	cache := NewAnalysisCache(provider)

	noteMgr := &fakeNoteManager{
		notes: map[string]string{
			"a.md": "Link to [[b]]",
			"b.md": "# Target",
		},
	}

	// First call populates cache and bumps provider version via Refresh.
	backlinks1, err := cache.Backlinks("vault", noteMgr, []string{"b"}, obsidian.WikilinkOptions{}, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, provider.refreshCalls)
	if assert.Contains(t, backlinks1, "b.md") {
		assert.Len(t, backlinks1["b.md"], 1)
	}

	// Change note content to remove the backlink.
	noteMgr.notes["a.md"] = "No links here"

	backlinks2, err := cache.Backlinks("vault", noteMgr, []string{"b"}, obsidian.WikilinkOptions{}, nil)
	assert.NoError(t, err)
	assert.Equal(t, 2, provider.refreshCalls, "Refresh should be invoked even on a cache hit to reconcile dirty state")
	if assert.Contains(t, backlinks2, "b.md") {
		assert.Len(t, backlinks2["b.md"], 0, "Backlinks should reflect updated note contents after refresh")
	}
}
