package cache

import (
	"context"

	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
)

// NoteAdapter implements obsidian.NoteManager backed by the cache. It falls back
// to the underlying NoteManager when entries are missing.
type NoteAdapter struct {
	cache *Service
	base  obsidian.NoteManager
}

// cachedEntriesProvider exposes cached entries to callers that can take advantage of them.
type cachedEntriesProvider interface {
	EntriesSnapshot(context.Context) ([]Entry, error)
}

// NewNoteAdapter constructs a cached NoteManager.
func NewNoteAdapter(cache *Service, base obsidian.NoteManager) *NoteAdapter {
	return &NoteAdapter{
		cache: cache,
		base:  base,
	}
}

func (n *NoteAdapter) Move(originalPath string, newPath string) error {
	return n.base.Move(originalPath, newPath)
}

func (n *NoteAdapter) Delete(path string) error {
	return n.base.Delete(path)
}

func (n *NoteAdapter) UpdateLinks(vaultPath string, oldNoteName string, newNoteName string) error {
	return n.base.UpdateLinks(vaultPath, oldNoteName, newNoteName)
}

func (n *NoteAdapter) GetContents(vaultPath string, noteName string) (string, error) {
	if err := n.cache.Refresh(context.Background()); err != nil {
		return "", err
	}
	if entry, ok := n.cache.Entry(noteName); ok && entry.Content != "" {
		return entry.Content, nil
	}
	return n.base.GetContents(vaultPath, noteName)
}

func (n *NoteAdapter) GetNotesList(vaultPath string) ([]string, error) {
	if err := n.cache.Refresh(context.Background()); err != nil {
		return nil, err
	}
	return n.cache.Paths(), nil
}

// Refresh forces the underlying cache to reconcile watcher events.
func (n *NoteAdapter) Refresh(ctx context.Context) error {
	return n.cache.Refresh(ctx)
}

// EntriesSnapshot exposes cached entries when the consumer can take advantage of them.
func (n *NoteAdapter) EntriesSnapshot(ctx context.Context) ([]Entry, error) {
	return n.cache.EntriesSnapshot(ctx)
}

// Version exposes the cache version for downstream caches.
func (n *NoteAdapter) Version() uint64 {
	return n.cache.Version()
}
