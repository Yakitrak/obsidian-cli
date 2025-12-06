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

func TestServiceStaleResyncFindsNewFiles(t *testing.T) {
	tmp := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "Existing.md"), []byte("#old"), 0o644))

	w := newStubWatcher()
	svc, err := NewService(tmp, Options{Watcher: w})
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })
	require.NoError(t, svc.EnsureReady(context.Background()))

	// Add a new file after the watcher goes stale.
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "New.md"), []byte("#new"), 0o644))
	svc.markStale()

	require.NoError(t, svc.Refresh(context.Background()))

	_, oldOk := svc.Entry("Existing.md")
	assert.True(t, oldOk)
	_, newOk := svc.Entry("New.md")
	assert.True(t, newOk, "resync should discover newly created files")
}

func TestServiceRenameDirectoryRescansChildren(t *testing.T) {
	tmp := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmp, "Folder", "Sub"), 0o755))
	origFile := filepath.Join(tmp, "Folder", "Sub", "Note.md")
	require.NoError(t, os.WriteFile(origFile, []byte("#tag"), 0o644))

	w := newStubWatcher()
	svc, err := NewService(tmp, Options{Watcher: w})
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })
	require.NoError(t, svc.EnsureReady(context.Background()))

	renamedDir := filepath.Join(tmp, "Renamed")
	require.NoError(t, os.Rename(filepath.Join(tmp, "Folder"), renamedDir))

	svc.markDirty(filepath.Join(tmp, "Folder"), DirtyRenamed)
	require.NoError(t, svc.Refresh(context.Background()))

	_, oldOk := svc.Entry("Folder/Sub/Note.md")
	assert.False(t, oldOk, "old path should be removed after directory rename")

	entry, newOk := svc.Entry("Renamed/Sub/Note.md")
	assert.True(t, newOk, "renamed path should be indexed")
	assert.Contains(t, entry.Tags, "tag")
}

func TestServiceStaleTickerMarksStale(t *testing.T) {
	tmp := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "Note.md"), []byte("#tag"), 0o644))

	w := newStubWatcher()
	svc, err := NewService(tmp, Options{Watcher: w, StaleInterval: 10 * time.Millisecond})
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })
	require.NoError(t, svc.EnsureReady(context.Background()))

	initialVersion := svc.Version()
	time.Sleep(25 * time.Millisecond)
	require.NoError(t, svc.Refresh(context.Background()))
	assert.GreaterOrEqual(t, svc.Version(), initialVersion+1, "stale ticker should trigger resync/version bump")
}

func TestServiceStaleTickerRestartsAfterResync(t *testing.T) {
	tmp := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "Note.md"), []byte("#tag"), 0o644))

	w := newStubWatcher()
	svc, err := NewService(tmp, Options{
		Watcher:        w,
		WatcherFactory: func() (Watcher, error) { return newStubWatcher(), nil },
		StaleInterval:  15 * time.Millisecond,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })
	require.NoError(t, svc.EnsureReady(context.Background()))

	// Force a resync by marking stale
	svc.markStale()
	require.NoError(t, svc.Refresh(context.Background()))

	// After resync, the stale ticker should still be running
	versionAfterResync := svc.Version()
	time.Sleep(40 * time.Millisecond)
	require.NoError(t, svc.Refresh(context.Background()))
	assert.Greater(t, svc.Version(), versionAfterResync, "stale ticker should continue working after resync")
}

func TestServiceEntryDeepCopy(t *testing.T) {
	tmp := t.TempDir()
	content := `---
tags: ["original"]
custom: value
---
#inline
`
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "Note.md"), []byte(content), 0o644))

	w := newStubWatcher()
	svc, err := NewService(tmp, Options{Watcher: w})
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })
	require.NoError(t, svc.EnsureReady(context.Background()))

	// Get entry and mutate it
	entry, ok := svc.Entry("Note.md")
	require.True(t, ok)

	// Mutate the returned entry
	entry.Tags[0] = "MUTATED"
	entry.Content = "MUTATED CONTENT"

	// Get entry again and verify cache wasn't affected
	entry2, ok := svc.Entry("Note.md")
	require.True(t, ok)
	assert.NotEqual(t, "MUTATED", entry2.Tags[0], "mutating returned entry should not affect cache")
	assert.NotContains(t, entry2.Content, "MUTATED", "mutating returned entry should not affect cache")
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

func TestServiceUsesDefaultIgnoreWhenMissing(t *testing.T) {
	tmp := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(tmp, "node_modules", "mod"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "node_modules", "mod", "ignored.md"), []byte("#ignored"), 0o644))

	included := filepath.Join(tmp, "Included.md")
	require.NoError(t, os.WriteFile(included, []byte("#public"), 0o644))

	w := newStubWatcher()
	svc, err := NewService(tmp, Options{Watcher: w})
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })
	require.NoError(t, svc.EnsureReady(context.Background()))

	paths := svc.Paths()
	assert.Len(t, paths, 1)
	assert.Equal(t, "Included.md", paths[0])
}

func TestServiceHandlesRecreation(t *testing.T) {
	tmp := t.TempDir()
	filePath := filepath.Join(tmp, "Note.md")
	require.NoError(t, os.WriteFile(filePath, []byte("#old"), 0o644))

	w := newStubWatcher()
	svc, err := NewService(tmp, Options{Watcher: w})
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })
	require.NoError(t, svc.EnsureReady(context.Background()))

	// Simulate rapid remove then create (recreation)
	// 1. Remove
	require.NoError(t, os.Remove(filePath))
	svc.markDirty(filePath, DirtyRemoved)

	// 2. Create (new content)
	require.NoError(t, os.WriteFile(filePath, []byte("#new"), 0o644))
	svc.markDirty(filePath, DirtyCreated)

	// Refresh should handle the transition: remove old entry, read new entry
	require.NoError(t, svc.Refresh(context.Background()))

	entry, ok := svc.Entry("Note.md")
	require.True(t, ok, "file should exist after recreation")
	assert.Contains(t, entry.Content, "#new")
	assert.Contains(t, entry.Tags, "new")
	assert.NotContains(t, entry.Tags, "old")
}

func TestServiceHandlesDirectoryRecreation(t *testing.T) {
	tmp := t.TempDir()
	dirPath := filepath.Join(tmp, "Folder")
	require.NoError(t, os.Mkdir(dirPath, 0o755))
	filePath := filepath.Join(dirPath, "Note.md")
	require.NoError(t, os.WriteFile(filePath, []byte("#old"), 0o644))

	w := newStubWatcher()
	svc, err := NewService(tmp, Options{Watcher: w})
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })
	require.NoError(t, svc.EnsureReady(context.Background()))

	// Verify initial state
	_, ok := svc.Entry("Folder/Note.md")
	require.True(t, ok)

	// Simulate rapid remove dir then create dir
	require.NoError(t, os.RemoveAll(dirPath))
	svc.markDirty(filepath.Join(tmp, "Folder"), DirtyRemoved)

	require.NoError(t, os.Mkdir(dirPath, 0o755))
	// New file in new dir
	newFilePath := filepath.Join(dirPath, "NewNote.md")
	require.NoError(t, os.WriteFile(newFilePath, []byte("#new"), 0o644))

	svc.markDirty(filepath.Join(tmp, "Folder"), DirtyCreated)

	// Refresh should remove old tree and scan new dir
	require.NoError(t, svc.Refresh(context.Background()))

	_, oldOk := svc.Entry("Folder/Note.md")
	assert.False(t, oldOk, "old file should be gone")

	entry, newOk := svc.Entry("Folder/NewNote.md")
	require.True(t, newOk, "new file should be found")
	assert.Contains(t, entry.Content, "#new")
}
