//go:build integration
// +build integration

package cache

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests use the real fsnotify watcher.
// Run with: go test -tags=integration ./pkg/cache/...
//
// These tests are slower and may be flaky on CI due to filesystem timing.
// They're valuable for catching platform-specific watcher behavior.

const (
	// Time to wait for fsnotify events to propagate
	eventDelay = 100 * time.Millisecond
	// Maximum time to wait for cache to reflect changes
	maxWait = 2 * time.Second
)

// waitForCondition polls until condition returns true or timeout expires.
func waitForCondition(t *testing.T, condition func() bool, timeout time.Duration, msg string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for: %s", msg)
}

func TestIntegration_RealWatcher_FileCreate(t *testing.T) {
	tmp := t.TempDir()

	// Create initial file
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "existing.md"), []byte("#existing"), 0o644))

	// Create service with real watcher (no stub)
	svc, err := NewService(tmp, Options{})
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })

	ctx := context.Background()
	require.NoError(t, svc.EnsureReady(ctx))

	// Verify initial state
	paths := svc.Paths()
	assert.Len(t, paths, 1)

	// Create a new file
	newFile := filepath.Join(tmp, "newfile.md")
	require.NoError(t, os.WriteFile(newFile, []byte("#newtag"), 0o644))

	// Wait for watcher to pick it up and refresh
	time.Sleep(eventDelay)

	waitForCondition(t, func() bool {
		_ = svc.Refresh(ctx)
		_, ok := svc.Entry("newfile.md")
		return ok
	}, maxWait, "new file to appear in cache")

	entry, ok := svc.Entry("newfile.md")
	require.True(t, ok)
	assert.Contains(t, entry.Tags, "newtag")
}

func TestIntegration_RealWatcher_FileModify(t *testing.T) {
	tmp := t.TempDir()
	filePath := filepath.Join(tmp, "note.md")
	require.NoError(t, os.WriteFile(filePath, []byte("#original"), 0o644))

	svc, err := NewService(tmp, Options{})
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })

	ctx := context.Background()
	require.NoError(t, svc.EnsureReady(ctx))

	entry, _ := svc.Entry("note.md")
	assert.Contains(t, entry.Tags, "original")

	// Modify the file
	time.Sleep(50 * time.Millisecond) // Ensure mtime changes
	require.NoError(t, os.WriteFile(filePath, []byte("#modified"), 0o644))

	time.Sleep(eventDelay)

	waitForCondition(t, func() bool {
		_ = svc.Refresh(ctx)
		entry, ok := svc.Entry("note.md")
		return ok && len(entry.Tags) > 0 && entry.Tags[0] == "modified"
	}, maxWait, "file modification to be detected")

	entry, _ = svc.Entry("note.md")
	assert.Contains(t, entry.Tags, "modified")
	assert.NotContains(t, entry.Tags, "original")
}

func TestIntegration_RealWatcher_FileDelete(t *testing.T) {
	tmp := t.TempDir()
	filePath := filepath.Join(tmp, "todelete.md")
	require.NoError(t, os.WriteFile(filePath, []byte("#tag"), 0o644))

	svc, err := NewService(tmp, Options{})
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })

	ctx := context.Background()
	require.NoError(t, svc.EnsureReady(ctx))

	_, ok := svc.Entry("todelete.md")
	require.True(t, ok)

	// Delete the file
	require.NoError(t, os.Remove(filePath))

	time.Sleep(eventDelay)

	waitForCondition(t, func() bool {
		_ = svc.Refresh(ctx)
		_, ok := svc.Entry("todelete.md")
		return !ok
	}, maxWait, "deleted file to be removed from cache")

	_, ok = svc.Entry("todelete.md")
	assert.False(t, ok)
}

func TestIntegration_RealWatcher_FileRename(t *testing.T) {
	tmp := t.TempDir()
	oldPath := filepath.Join(tmp, "oldname.md")
	newPath := filepath.Join(tmp, "newname.md")
	require.NoError(t, os.WriteFile(oldPath, []byte("#tag"), 0o644))

	svc, err := NewService(tmp, Options{})
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })

	ctx := context.Background()
	require.NoError(t, svc.EnsureReady(ctx))

	_, ok := svc.Entry("oldname.md")
	require.True(t, ok)

	// Rename the file
	require.NoError(t, os.Rename(oldPath, newPath))

	time.Sleep(eventDelay)

	waitForCondition(t, func() bool {
		_ = svc.Refresh(ctx)
		_, oldOk := svc.Entry("oldname.md")
		_, newOk := svc.Entry("newname.md")
		return !oldOk && newOk
	}, maxWait, "rename to be reflected in cache")

	_, oldOk := svc.Entry("oldname.md")
	_, newOk := svc.Entry("newname.md")
	assert.False(t, oldOk, "old name should be gone")
	assert.True(t, newOk, "new name should exist")
}

func TestIntegration_RealWatcher_DirectoryCreate(t *testing.T) {
	tmp := t.TempDir()

	svc, err := NewService(tmp, Options{})
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })

	ctx := context.Background()
	require.NoError(t, svc.EnsureReady(ctx))

	// Create a new directory with a file inside
	newDir := filepath.Join(tmp, "subdir")
	require.NoError(t, os.Mkdir(newDir, 0o755))
	time.Sleep(eventDelay) // Let watcher register the new directory

	// Now create a file in the new directory
	require.NoError(t, os.WriteFile(filepath.Join(newDir, "note.md"), []byte("#nested"), 0o644))

	time.Sleep(eventDelay)

	waitForCondition(t, func() bool {
		_ = svc.Refresh(ctx)
		_, ok := svc.Entry("subdir/note.md")
		return ok
	}, maxWait, "file in new directory to appear")

	entry, ok := svc.Entry("subdir/note.md")
	require.True(t, ok)
	assert.Contains(t, entry.Tags, "nested")
}

func TestIntegration_RealWatcher_DirectoryDelete(t *testing.T) {
	tmp := t.TempDir()

	// Create directory structure
	subdir := filepath.Join(tmp, "subdir")
	require.NoError(t, os.Mkdir(subdir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(subdir, "note1.md"), []byte("#one"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(subdir, "note2.md"), []byte("#two"), 0o644))

	svc, err := NewService(tmp, Options{})
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })

	ctx := context.Background()
	require.NoError(t, svc.EnsureReady(ctx))

	_, ok1 := svc.Entry("subdir/note1.md")
	_, ok2 := svc.Entry("subdir/note2.md")
	require.True(t, ok1 && ok2)

	// Delete entire directory
	require.NoError(t, os.RemoveAll(subdir))

	time.Sleep(eventDelay)

	waitForCondition(t, func() bool {
		_ = svc.Refresh(ctx)
		_, ok1 := svc.Entry("subdir/note1.md")
		_, ok2 := svc.Entry("subdir/note2.md")
		return !ok1 && !ok2
	}, maxWait, "directory contents to be removed from cache")

	_, ok1 = svc.Entry("subdir/note1.md")
	_, ok2 = svc.Entry("subdir/note2.md")
	assert.False(t, ok1)
	assert.False(t, ok2)
}

func TestIntegration_RealWatcher_RapidCreateDelete(t *testing.T) {
	tmp := t.TempDir()

	svc, err := NewService(tmp, Options{})
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })

	ctx := context.Background()
	require.NoError(t, svc.EnsureReady(ctx))

	filePath := filepath.Join(tmp, "rapid.md")

	// Rapid create-delete-create sequence
	require.NoError(t, os.WriteFile(filePath, []byte("#first"), 0o644))
	require.NoError(t, os.Remove(filePath))
	require.NoError(t, os.WriteFile(filePath, []byte("#second"), 0o644))

	time.Sleep(eventDelay)

	waitForCondition(t, func() bool {
		_ = svc.Refresh(ctx)
		entry, ok := svc.Entry("rapid.md")
		return ok && len(entry.Tags) > 0 && entry.Tags[0] == "second"
	}, maxWait, "rapid recreation to settle with final state")

	entry, ok := svc.Entry("rapid.md")
	require.True(t, ok, "file should exist")
	assert.Contains(t, entry.Tags, "second")
	assert.NotContains(t, entry.Tags, "first")
}

func TestIntegration_RealWatcher_BulkOperations(t *testing.T) {
	tmp := t.TempDir()

	svc, err := NewService(tmp, Options{})
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })

	ctx := context.Background()
	require.NoError(t, svc.EnsureReady(ctx))

	// Create many files rapidly
	for i := 0; i < 20; i++ {
		path := filepath.Join(tmp, filepath.Base(t.Name())+string(rune('a'+i))+".md")
		require.NoError(t, os.WriteFile(path, []byte("#bulk"), 0o644))
	}

	time.Sleep(eventDelay * 3) // Give more time for bulk events

	waitForCondition(t, func() bool {
		_ = svc.Refresh(ctx)
		return len(svc.Paths()) >= 20
	}, maxWait*2, "all bulk files to appear")

	assert.GreaterOrEqual(t, len(svc.Paths()), 20)
}

func TestIntegration_RealWatcher_NestedDirectoryCreate(t *testing.T) {
	tmp := t.TempDir()

	svc, err := NewService(tmp, Options{})
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })

	ctx := context.Background()
	require.NoError(t, svc.EnsureReady(ctx))

	// Create deeply nested directory structure
	nested := filepath.Join(tmp, "a", "b", "c")
	require.NoError(t, os.MkdirAll(nested, 0o755))
	time.Sleep(eventDelay)

	// Create file in deeply nested directory
	require.NoError(t, os.WriteFile(filepath.Join(nested, "deep.md"), []byte("#deep"), 0o644))

	time.Sleep(eventDelay)

	waitForCondition(t, func() bool {
		_ = svc.Refresh(ctx)
		_, ok := svc.Entry("a/b/c/deep.md")
		return ok
	}, maxWait, "deeply nested file to appear")

	entry, ok := svc.Entry("a/b/c/deep.md")
	require.True(t, ok)
	assert.Contains(t, entry.Tags, "deep")
}

func TestIntegration_RealWatcher_AtomicWrite(t *testing.T) {
	tmp := t.TempDir()
	targetPath := filepath.Join(tmp, "atomic.md")

	// Create initial file
	require.NoError(t, os.WriteFile(targetPath, []byte("#original"), 0o644))

	svc, err := NewService(tmp, Options{})
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })

	ctx := context.Background()
	require.NoError(t, svc.EnsureReady(ctx))

	entry, _ := svc.Entry("atomic.md")
	assert.Contains(t, entry.Tags, "original")

	// Simulate atomic write: write to temp, then rename
	tempPath := filepath.Join(tmp, ".atomic.md.tmp")
	require.NoError(t, os.WriteFile(tempPath, []byte("#atomicnew"), 0o644))
	require.NoError(t, os.Rename(tempPath, targetPath))

	time.Sleep(eventDelay)

	waitForCondition(t, func() bool {
		_ = svc.Refresh(ctx)
		entry, ok := svc.Entry("atomic.md")
		return ok && len(entry.Tags) > 0 && entry.Tags[0] == "atomicnew"
	}, maxWait, "atomic write to be detected")

	entry, ok := svc.Entry("atomic.md")
	require.True(t, ok)
	assert.Contains(t, entry.Tags, "atomicnew")
}

func TestIntegration_RealWatcher_VersionBumps(t *testing.T) {
	tmp := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "note.md"), []byte("#tag"), 0o644))

	svc, err := NewService(tmp, Options{})
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })

	ctx := context.Background()
	require.NoError(t, svc.EnsureReady(ctx))

	initialVersion := svc.Version()

	// Modify file
	time.Sleep(50 * time.Millisecond)
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "note.md"), []byte("#modified"), 0o644))

	time.Sleep(eventDelay)

	waitForCondition(t, func() bool {
		_ = svc.Refresh(ctx)
		return svc.Version() > initialVersion
	}, maxWait, "version to bump after modification")

	assert.Greater(t, svc.Version(), initialVersion)
}
