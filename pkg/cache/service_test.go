package cache

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubWatcher implements Watcher for tests without relying on actual fsnotify events.
type stubWatcher struct {
	events chan fsnotify.Event
	errors chan error
	adds   []string
	mu     sync.Mutex
	closed bool
}

func newStubWatcher() *stubWatcher {
	return &stubWatcher{
		events: make(chan fsnotify.Event, 16),
		errors: make(chan error, 1),
	}
}

func (w *stubWatcher) Add(name string) error {
	w.mu.Lock()
	w.adds = append(w.adds, name)
	w.mu.Unlock()
	return nil
}

func (w *stubWatcher) Close() error {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return nil
	}
	w.closed = true
	close(w.events)
	close(w.errors)
	w.mu.Unlock()
	return nil
}

func (w *stubWatcher) Events() <-chan fsnotify.Event { return w.events }
func (w *stubWatcher) Errors() <-chan error          { return w.errors }

func TestServiceInitialCrawlCachesTags(t *testing.T) {
	tmp := t.TempDir()
	content := `---
tags: ["Project"]
---

# Heading
#todo something
`
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "Note.md"), []byte(content), 0o644))

	w := newStubWatcher()
	svc, err := NewService(tmp, Options{Watcher: w})
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })

	require.NoError(t, svc.EnsureReady(context.Background()))

	paths := svc.Paths()
	assert.Len(t, paths, 1)
	assert.Equal(t, "Note.md", paths[0])

	entry, ok := svc.Entry("Note.md")
	require.True(t, ok)
	assert.Contains(t, entry.Tags, "project")
	assert.Contains(t, entry.Tags, "todo")
}

func TestServiceRefreshUpdatesModifiedFile(t *testing.T) {
	tmp := t.TempDir()
	filePath := filepath.Join(tmp, "Note.md")
	require.NoError(t, os.WriteFile(filePath, []byte("#old"), 0o644))

	w := newStubWatcher()
	svc, err := NewService(tmp, Options{Watcher: w})
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })
	require.NoError(t, svc.EnsureReady(context.Background()))

	// Modify file content and send a write event.
	require.NoError(t, os.WriteFile(filePath, []byte("#newtag"), 0o644))
	svc.markDirty(filePath, DirtyModified)

	ctx := context.Background()
	var entry Entry
	var ok bool
	svc.mu.RLock()
	dirtyLen := len(svc.dirty)
	svc.mu.RUnlock()
	require.Equal(t, 1, dirtyLen)

	for i := 0; i < 5; i++ {
		require.NoError(t, svc.Refresh(ctx))
		entry, ok = svc.Entry("Note.md")
		if ok && strings.Contains(entry.Content, "#newtag") {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	require.True(t, ok)
	assert.Contains(t, entry.Tags, "newtag")
	assert.NotContains(t, entry.Tags, "old")
	assert.Contains(t, entry.Content, "#newtag")
}

func TestServiceRefreshRemovesDeletedFile(t *testing.T) {
	tmp := t.TempDir()
	filePath := filepath.Join(tmp, "Note.md")
	require.NoError(t, os.WriteFile(filePath, []byte("#tag"), 0o644))

	w := newStubWatcher()
	svc, err := NewService(tmp, Options{Watcher: w})
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })
	require.NoError(t, svc.EnsureReady(context.Background()))

	require.NoError(t, os.Remove(filePath))
	svc.markDirty(filePath, DirtyRemoved)

	require.NoError(t, svc.Refresh(context.Background()))

	_, ok := svc.Entry("Note.md")
	assert.False(t, ok)
	assert.Empty(t, svc.Paths())
}

func TestServiceHandlesRenameEvent(t *testing.T) {
	tmp := t.TempDir()
	orig := filepath.Join(tmp, "Old.md")
	require.NoError(t, os.WriteFile(orig, []byte("#tag"), 0o644))

	w := newStubWatcher()
	svc, err := NewService(tmp, Options{Watcher: w})
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })
	require.NoError(t, svc.EnsureReady(context.Background()))

	newPath := filepath.Join(tmp, "New.md")
	require.NoError(t, os.Rename(orig, newPath))

	svc.markDirty(orig, DirtyRenamed)
	require.NoError(t, svc.Refresh(context.Background()))

	_, oldOk := svc.Entry("Old.md")
	assert.False(t, oldOk, "old name should be removed")

	entry, newOk := svc.Entry("New.md")
	assert.True(t, newOk, "new name should be indexed")
	assert.Contains(t, entry.Tags, "tag")
}

func TestServiceConcurrentEnsureReady(t *testing.T) {
	tmp := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "Note.md"), []byte("#tag"), 0o644))

	w := newStubWatcher()
	svc, err := NewService(tmp, Options{Watcher: w})
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })

	// Launch multiple concurrent EnsureReady calls
	const goroutines = 5
	var wg sync.WaitGroup
	errs := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := svc.EnsureReady(context.Background()); err != nil {
				errs <- err
			}
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		require.NoError(t, err)
	}

	// Verify cache is populated correctly
	paths := svc.Paths()
	assert.Len(t, paths, 1)
	assert.Equal(t, "Note.md", paths[0])
}

func TestServiceStaleRevalidation(t *testing.T) {
	tmp := t.TempDir()
	filePath := filepath.Join(tmp, "Note.md")
	require.NoError(t, os.WriteFile(filePath, []byte("#original"), 0o644))

	w := newStubWatcher()
	svc, err := NewService(tmp, Options{Watcher: w})
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })
	require.NoError(t, svc.EnsureReady(context.Background()))

	// Modify the file behind the watcher's back
	time.Sleep(10 * time.Millisecond) // Ensure mtime changes
	require.NoError(t, os.WriteFile(filePath, []byte("#updated"), 0o644))

	// Manually mark cache as stale (simulating watcher failure)
	svc.markStale()

	// Refresh should revalidate and detect the change
	require.NoError(t, svc.Refresh(context.Background()))

	entry, ok := svc.Entry("Note.md")
	require.True(t, ok)
	assert.Contains(t, entry.Tags, "updated")
	assert.NotContains(t, entry.Tags, "original")
}

func TestServiceRespectsObsidianIgnore(t *testing.T) {
	tmp := t.TempDir()

	// Create ignored file
	ignoredPath := filepath.Join(tmp, "Ignored.md")
	require.NoError(t, os.WriteFile(ignoredPath, []byte("#secret"), 0o644))

	// Create included file
	includedPath := filepath.Join(tmp, "Included.md")
	require.NoError(t, os.WriteFile(includedPath, []byte("#public"), 0o644))

	// Create .obsidianignore
	ignoreFile := filepath.Join(tmp, ".obsidianignore")
	require.NoError(t, os.WriteFile(ignoreFile, []byte("Ignored.md\n"), 0o644))

	w := newStubWatcher()
	svc, err := NewService(tmp, Options{Watcher: w})
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })
	require.NoError(t, svc.EnsureReady(context.Background()))

	paths := svc.Paths()
	assert.Len(t, paths, 1)
	assert.Equal(t, "Included.md", paths[0])

	_, ok := svc.Entry("Ignored.md")
	assert.False(t, ok, "ignored file should not be indexed")
}
