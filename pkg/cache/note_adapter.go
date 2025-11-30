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
	_ = n.cache.Refresh(context.Background())
	if entry, ok := n.cache.Entry(noteName); ok && entry.Content != "" {
		return entry.Content, nil
	}
	return n.base.GetContents(vaultPath, noteName)
}

func (n *NoteAdapter) GetNotesList(vaultPath string) ([]string, error) {
	if err := n.cache.EnsureReady(context.Background()); err != nil {
		return nil, err
	}
	return n.cache.Paths(), nil
}
