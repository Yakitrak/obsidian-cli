package actions

import (
	"context"
	"errors"

	"github.com/atomicobject/obsidian-cli/pkg/cache"
)

// cachedNoteManager is a lightweight NoteManager that serves cached entries and tracks calls.
type cachedNoteManager struct {
	entries          []cache.Entry
	contentsCalls    int
	notesListCalls   int
	allowNotesReturn bool
}

func (c *cachedNoteManager) EntriesSnapshot(ctx context.Context) ([]cache.Entry, error) {
	return c.entries, nil
}

func (c *cachedNoteManager) Move(string, string) error                { return nil }
func (c *cachedNoteManager) Delete(string) error                      { return nil }
func (c *cachedNoteManager) UpdateLinks(string, string, string) error { return nil }
func (c *cachedNoteManager) GetContents(string, string) (string, error) {
	c.contentsCalls++
	return "", errors.New("GetContents should not be called when cache is used")
}
func (c *cachedNoteManager) GetNotesList(string) ([]string, error) {
	c.notesListCalls++
	if c.allowNotesReturn {
		return []string{}, nil
	}
	return nil, errors.New("GetNotesList should not be called when cache is used")
}
