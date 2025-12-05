package cache

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
	"github.com/fsnotify/fsnotify"
)

// Service maintains an in-memory cache of vault files, metadata, and tags.
//
// Operational story (read before editing):
//  1. EnsureReady() performs a one-time crawl to populate indices and install
//     directory watches. It is concurrency-safe and uses a simple spin gate.
//  2. watchLoop translates fsnotify events into "dirty" markers (or, on watcher
//     failure, flips a stale flag).
//  3. Refresh() is the front door callers hit before reading; it revalidates
//     stale caches and applies dirty markers by re-reading or deleting paths.
//
// The intent is to keep the dataflow legible rather than aggressively abstracted:
// callers see the crawl, the watcher, and the refresh step as distinct phases.
type Service struct {
	vaultPath string

	mu        sync.RWMutex
	ready     bool
	crawling  bool // guards against concurrent initial crawls
	stale     bool // set when watcher fails; forces revalidation
	fileIndex map[string]*Entry
	tagIndex  map[string]map[string]struct{}
	dirIndex  map[string]struct{} // directories currently watched
	dirty     map[string]DirtyKind
	ignored   []string // loaded from .obsidianignore

	watcher        Watcher
	watcherFactory func() (Watcher, error)
	watchOnce      sync.Once
	ctx            context.Context
	cancel         context.CancelFunc
	version        uint64
	staleInterval  time.Duration
}

// Entry represents cached metadata for a single file.
type Entry struct {
	Path        string
	ModTime     time.Time
	Size        int64
	Tags        []string
	Frontmatter map[string]interface{}
	InlineProps map[string][]string
	Content     string
}

// DirtyKind captures why a path was marked dirty.
type DirtyKind string

const (
	DirtyUnknown  DirtyKind = "unknown"
	DirtyCreated  DirtyKind = "created"
	DirtyModified DirtyKind = "modified"
	DirtyRemoved  DirtyKind = "removed"
	DirtyRenamed  DirtyKind = "renamed"
)

// Options controls cache behavior.
type Options struct {
	Watcher        Watcher
	WatcherFactory func() (Watcher, error)
	StaleInterval  time.Duration
}

// Watcher abstracts filesystem notifications for modular backends.
type Watcher interface {
	Add(name string) error
	Close() error
	Events() <-chan fsnotify.Event
	Errors() <-chan error
}

type fsNotifyWatcher struct {
	*fsnotify.Watcher
}

// Events exposes fsnotify events as required by the Watcher interface.
func (f *fsNotifyWatcher) Events() <-chan fsnotify.Event { return f.Watcher.Events }

// Errors exposes fsnotify errors as required by the Watcher interface.
func (f *fsNotifyWatcher) Errors() <-chan error { return f.Watcher.Errors }

// NewService constructs a cache service for a vault.
func NewService(vaultPath string, opts Options) (*Service, error) {
	if vaultPath == "" {
		return nil, errors.New("vaultPath is required")
	}

	var watcher Watcher
	var watcherFactory func() (Watcher, error)
	if opts.Watcher != nil {
		watcher = opts.Watcher
		watcherFactory = opts.WatcherFactory
	} else {
		watcherFactory = func() (Watcher, error) {
			w, err := fsnotify.NewWatcher()
			if err != nil {
				return nil, fmt.Errorf("create watcher: %w", err)
			}
			return &fsNotifyWatcher{Watcher: w}, nil
		}
		w, err := watcherFactory()
		if err != nil {
			// Fall back to polling-only mode when watcher setup fails.
			watcherFactory = nil
			watcher = nil
			if opts.StaleInterval == 0 {
				opts.StaleInterval = 30 * time.Second
			}
			log.Printf("cache: watcher unavailable (%v); falling back to polling with stale interval %s", err, opts.StaleInterval)
		} else {
			watcher = w
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Service{
		vaultPath:      vaultPath,
		fileIndex:      make(map[string]*Entry),
		tagIndex:       make(map[string]map[string]struct{}),
		dirIndex:       make(map[string]struct{}),
		dirty:          make(map[string]DirtyKind),
		watcher:        watcher,
		watcherFactory: watcherFactory,
		ctx:            ctx,
		cancel:         cancel,
		staleInterval:  opts.StaleInterval,
	}, nil
}

// Close stops the watcher and releases resources.
func (s *Service) Close() error {
	s.cancel()
	if s.watcher != nil {
		return s.watcher.Close()
	}
	return nil
}

// EnsureReady performs the initial crawl (once) and starts the watcher.
// It is safe to call concurrently; only one goroutine will perform the initial crawl.
func (s *Service) EnsureReady(ctx context.Context) error {
	s.mu.Lock()
	if s.ready {
		s.mu.Unlock()
		return s.Refresh(ctx)
	}

	// From here exactly one goroutine owns the initial crawl: either we wait
	// for the in-flight crawl to finish, or we claim the work ourselves.
	// If another goroutine is already crawling, wait for it.
	if s.crawling {
		s.mu.Unlock()
		// Spin-wait for crawl to complete (simple approach; could use cond var for efficiency).
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(10 * time.Millisecond):
				s.mu.RLock()
				ready := s.ready
				s.mu.RUnlock()
				if ready {
					return s.Refresh(ctx)
				}
			}
		}
	}
	s.crawling = true
	s.mu.Unlock()

	if err := s.initialCrawl(ctx); err != nil {
		s.mu.Lock()
		s.crawling = false
		s.mu.Unlock()
		return err
	}
	s.startWatcher()
	s.startStaleTicker()
	return s.Refresh(ctx)
}

func (s *Service) loadIgnorePatterns() {
	// Ignore patterns are loaded once at startup; we do not attempt to hot-reload
	// .obsidianignore changes because the watcher already guards the main data paths
	// and this keeps the cache predictable for the lifetime of the process.
	ignorePath := filepath.Join(s.vaultPath, ".obsidianignore")
	patterns := obsidian.DefaultIgnorePatterns()
	content, err := os.ReadFile(ignorePath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			// Log or silently ignore read errors? For now silent as per spec assumptions.
		}
		s.mu.Lock()
		s.ignored = patterns
		s.mu.Unlock()
		return
	}
	lines := strings.Split(string(content), "\n")
	var ignored []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			ignored = append(ignored, line)
		}
	}
	s.mu.Lock()
	if len(ignored) == 0 {
		s.ignored = patterns
	} else {
		s.ignored = ignored
	}
	s.mu.Unlock()
}

// Paths returns all cached paths.
func (s *Service) Paths() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	paths := make([]string, 0, len(s.fileIndex))
	for p := range s.fileIndex {
		paths = append(paths, p)
	}
	return paths
}

// Version returns a monotonic counter that increments whenever the cache
// processes dirty changes or is resynced. Callers can use it to invalidate
// derived caches (e.g., backlinks, graph analysis).
func (s *Service) Version() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.version
}

// EntriesSnapshot returns a shallow copy of all cached entries after ensuring freshness.
func (s *Service) EntriesSnapshot(ctx context.Context) ([]Entry, error) {
	if err := s.Refresh(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	entries := make([]Entry, 0, len(s.fileIndex))
	for _, e := range s.fileIndex {
		copy := *e
		entries = append(entries, copy)
	}
	return entries, nil
}

// Entry returns a copy of the cached entry for the given path.
func (s *Service) Entry(path string) (Entry, bool) {
	norm := obsidian.NormalizePath(obsidian.AddMdSuffix(path))

	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.fileIndex[norm]
	if !ok {
		return Entry{}, false
	}
	// Return a shallow copy to prevent callers from mutating the cache.
	out := *entry
	return out, true
}

// Refresh reconciles in-memory state with the filesystem. Think of it as the
// “checkpoint” callers hit before reading:
//   - If this is the first call, it delegates to EnsureReady().
//   - If the watcher signaled trouble (stale flag), it revalidates every entry.
//   - Otherwise it consumes any dirty markers emitted by watchLoop().
func (s *Service) Refresh(ctx context.Context) error {
	s.mu.Lock()
	if !s.ready {
		s.mu.Unlock()
		return s.EnsureReady(ctx)
	}

	// Snapshot mutable state so we can release the lock while touching disk.
	stale := s.stale
	if stale {
		s.stale = false
	}
	dirty := s.dirty
	s.dirty = make(map[string]DirtyKind)
	s.mu.Unlock()

	// If the watcher failed, rescan everything to avoid missing new files or renames.
	if stale {
		return s.resync(ctx)
	}

	if len(dirty) == 0 {
		return nil
	}

	changed := false

	// Consume dirty markers: remove, rescan parent (for renames), or refresh the file.
	for path, kind := range dirty {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		switch kind {
		case DirtyRemoved, DirtyRenamed:
			s.removeTree(path)
			changed = true
			if kind == DirtyRenamed {
				parent := filepath.Dir(filepath.Join(s.vaultPath, path))
				_ = s.rescanDir(parent, true)
			}
		default:
			if err := s.refreshPath(path); err != nil {
				// Keep the entry dirty so we retry next time.
				s.markDirty(path, DirtyModified)
			} else {
				changed = true
			}
		}
	}

	if changed {
		s.bumpVersion()
	}

	return nil
}

// resync rebuilds watcher state (when possible) and performs a full crawl to ensure freshness.
func (s *Service) resync(ctx context.Context) error {
	// Stop the existing watch loop to avoid duplicate goroutines.
	s.cancel()
	s.ctx, s.cancel = context.WithCancel(context.Background())

	_, err := s.rebuildWatcher()
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.ready = false
	s.crawling = false
	s.fileIndex = make(map[string]*Entry)
	s.tagIndex = make(map[string]map[string]struct{})
	s.dirIndex = make(map[string]struct{})
	s.dirty = make(map[string]DirtyKind)
	s.watchOnce = sync.Once{}
	s.mu.Unlock()

	if err := s.initialCrawl(ctx); err != nil {
		return err
	}
	s.startWatcher()
	return nil
}

func (s *Service) initialCrawl(ctx context.Context) error {
	s.loadIgnorePatterns()

	type job struct {
		path string
		info fs.DirEntry
	}

	// Phase 1: walk and collect work. We collect into jobs so that the second
	// phase (actual reads + watcher registration) can run with minimal locking
	// and without mixing concerns inside WalkDir callbacks.
	var jobs []job
	err := filepath.WalkDir(s.vaultPath, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		rel, err := filepath.Rel(s.vaultPath, path)
		if err != nil {
			return err
		}

		s.mu.RLock()
		ignored := s.ignored
		s.mu.RUnlock()

		if d.IsDir() {
			if obsidian.ShouldIgnorePath(s.vaultPath, path, ignored) {
				return filepath.SkipDir
			}
			// Always watch directories (even empty) to catch new files.
			jobs = append(jobs, job{path: path, info: d})
			return nil
		}

		if shouldSkip(s.vaultPath, rel, d, ignored) {
			return nil
		}
		jobs = append(jobs, job{path: path, info: d})
		return nil
	})
	if err != nil {
		return err
	}

	// Phase 2: execute the work. Directories are watched; files are refreshed.
	for _, j := range jobs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if j.info.IsDir() {
			s.addWatch(j.path)
			continue
		}
		if err := s.refreshPath(j.path); err != nil {
			return err
		}
	}

	s.mu.Lock()
	s.ready = true
	s.version++
	s.mu.Unlock()
	return nil
}

func (s *Service) refreshPath(absPath string) error {
	if !filepath.IsAbs(absPath) {
		absPath = filepath.Join(s.vaultPath, absPath)
	}

	// Normalize early so all downstream indexing keys are consistent.
	rel, err := filepath.Rel(s.vaultPath, absPath)
	if err != nil {
		return err
	}
	rel = obsidian.NormalizePath(rel)

	s.mu.RLock()
	ignored := s.ignored
	s.mu.RUnlock()

	info, err := os.Stat(absPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			s.removePath(rel)
			return nil
		}
		return err
	}
	if info.IsDir() || shouldSkip(s.vaultPath, rel, info, ignored) {
		return nil
	}

	// Read the file once; reuse the content for all extractors to avoid multiple disk hits.
	content, err := os.ReadFile(absPath)
	if err != nil {
		return err
	}

	entry := &Entry{
		Path:    rel,
		ModTime: info.ModTime(),
		Size:    info.Size(),
		Content: string(content),
	}

	// Extract frontmatter + inline props + tags in one pass per refresh.
	fm, _ := obsidian.ExtractFrontmatter(entry.Content)
	if fm != nil {
		entry.Frontmatter = fm
		if tags, ok := fm["tags"].([]string); ok {
			entry.Tags = append(entry.Tags, normalizeTags(tags)...)
		}
	}

	hashtags := obsidian.ExtractHashtags(entry.Content)
	entry.Tags = append(entry.Tags, normalizeTags(stripHashtagPrefix(hashtags))...)

	entry.InlineProps = obsidian.ExtractInlineProperties(entry.Content)

	s.mu.Lock()
	s.removeTagsForPathLocked(rel)
	s.fileIndex[rel] = entry
	s.indexTags(rel, entry.Tags)
	s.mu.Unlock()
	return nil
}

func (s *Service) removePath(rel string) {
	rel = obsidian.NormalizePath(rel)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.removeTagsForPathLocked(rel)
	delete(s.fileIndex, rel)
}

func (s *Service) removeTree(rel string) {
	rel = obsidian.NormalizePath(rel)
	s.mu.Lock()
	defer s.mu.Unlock()
	for path := range s.fileIndex {
		if path == rel || strings.HasPrefix(path, rel+"/") {
			s.removeTagsForPathLocked(path)
			delete(s.fileIndex, path)
		}
	}
}

func (s *Service) removeTagsForPathLocked(rel string) {
	for tag, paths := range s.tagIndex {
		delete(paths, rel)
		if len(paths) == 0 {
			delete(s.tagIndex, tag)
		}
	}
}

// rescanDir refreshes all files in a directory (optionally recursive).
func (s *Service) rescanDir(absDir string, recursive bool) error {
	entries, err := os.ReadDir(absDir)
	if err != nil {
		return err
	}
	// Non-recursive: watcher events will cascade for deeper levels.
	for _, entry := range entries {
		if entry.IsDir() {
			s.addWatch(filepath.Join(absDir, entry.Name()))
			if recursive {
				_ = s.rescanDir(filepath.Join(absDir, entry.Name()), true)
			}
			continue
		}
		if err := s.refreshPath(filepath.Join(absDir, entry.Name())); err != nil {
			// continue on individual file errors to avoid missing other files
			continue
		}
	}
	return nil
}

func (s *Service) indexTags(path string, tags []string) {
	for _, t := range tags {
		if t == "" {
			continue
		}
		if _, ok := s.tagIndex[t]; !ok {
			s.tagIndex[t] = make(map[string]struct{})
		}
		s.tagIndex[t][path] = struct{}{}
	}
}

func normalizeTags(tags []string) []string {
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		nt := strings.TrimSpace(strings.TrimPrefix(t, "#"))
		nt = strings.ToLower(nt)
		if nt != "" {
			out = append(out, nt)
		}
	}
	return out
}

func stripHashtagPrefix(tags []string) []string {
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		out = append(out, strings.TrimPrefix(t, "#"))
	}
	return out
}

func (s *Service) markDirty(absPath string, kind DirtyKind) {
	rel, err := filepath.Rel(s.vaultPath, absPath)
	if err != nil {
		return
	}
	rel = obsidian.NormalizePath(rel)

	s.mu.Lock()
	if existing, ok := s.dirty[rel]; ok {
		// Prefer removal markers to avoid stale entries; once a path is gone we
		// do not want later modify events to resurrect it.
		if existing == DirtyRemoved || kind == DirtyRemoved {
			s.dirty[rel] = DirtyRemoved
		}
		s.mu.Unlock()
		return
	}
	s.dirty[rel] = kind
	s.mu.Unlock()
}

func (s *Service) startWatcher() {
	if s.watcher == nil {
		return
	}
	// watchLoop is intentionally fire-and-forget; errors bubble through markStale().
	s.watchOnce.Do(func() {
		go s.watchLoop()
	})
}

func (s *Service) startStaleTicker() {
	if s.staleInterval <= 0 {
		return
	}
	go func() {
		ticker := time.NewTicker(s.staleInterval)
		defer ticker.Stop()
		for {
			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				s.markStale()
			}
		}
	}()
}

func (s *Service) addWatch(path string) {
	if s.watcher == nil {
		return
	}
	if path == "" {
		return
	}
	s.mu.Lock()
	if _, ok := s.dirIndex[path]; ok {
		s.mu.Unlock()
		return
	}
	s.dirIndex[path] = struct{}{}
	s.mu.Unlock()
	_ = s.watcher.Add(path)
}

func (s *Service) watchLoop() {
	// This loop translates filesystem noise into coarse-grained signals:
	// mark paths dirty so Refresh() can reconcile, or mark the whole cache
	// stale when the watcher itself fails.
	for {
		select {
		case <-s.ctx.Done():
			return
		case evt, ok := <-s.watcher.Events():
			if !ok {
				// Channel closed; mark cache stale for safety.
				s.markStale()
				return
			}
			switch {
			case evt.Op&fsnotify.Create == fsnotify.Create:
				s.markDirty(evt.Name, DirtyCreated)
				// If a new directory is created, start watching it and scan for files.
				info, err := os.Stat(evt.Name)
				if err == nil && info.IsDir() {
					s.addWatch(evt.Name)
					// Scan the new directory for any files it may already contain
					// (e.g., created via git checkout or archive extraction).
					_ = s.rescanDir(evt.Name, true)
				}
			case evt.Op&fsnotify.Write == fsnotify.Write:
				s.markDirty(evt.Name, DirtyModified)
			case evt.Op&fsnotify.Remove == fsnotify.Remove:
				s.markDirty(evt.Name, DirtyRemoved)
				s.dropDirIndex(evt.Name)
			case evt.Op&fsnotify.Rename == fsnotify.Rename:
				s.markDirty(evt.Name, DirtyRenamed)
				s.dropDirIndex(evt.Name)
			}
		case err, ok := <-s.watcher.Errors():
			if !ok {
				// Channel closed; mark cache stale.
				s.markStale()
				return
			}
			// Watcher error; mark cache stale so next read triggers revalidation.
			s.markStale()
			_ = err // Log in debug mode if needed
		}
	}
}

// markStale flags the cache as potentially out of sync. The next Refresh
// will trigger a full revalidation to recover from watcher failures.
func (s *Service) markStale() {
	s.mu.Lock()
	s.stale = true
	s.mu.Unlock()
}

func (s *Service) bumpVersion() {
	s.mu.Lock()
	s.version++
	s.mu.Unlock()
}

func (s *Service) dropDirIndex(path string) {
	s.mu.Lock()
	delete(s.dirIndex, path)
	s.mu.Unlock()
}

func (s *Service) rebuildWatcher() (Watcher, error) {
	if s.watcherFactory == nil {
		return s.watcher, nil
	}
	if s.watcher != nil {
		_ = s.watcher.Close()
	}
	w, err := s.watcherFactory()
	if err != nil {
		return nil, err
	}
	s.watcher = w
	return w, nil
}

type fileInfo interface {
	Name() string
	IsDir() bool
}

func shouldSkip(vaultPath, relPath string, info fileInfo, ignored []string) bool {
	// Skip hidden files or non-markdown files. Treat hidden directories as skipped.
	name := info.Name()
	if strings.HasPrefix(name, ".") {
		return true
	}
	if info.IsDir() {
		return false
	}
	if filepath.Ext(relPath) != ".md" {
		return true
	}

	// Check ignore patterns (.obsidianignore)
	absPath := filepath.Join(vaultPath, relPath)
	return obsidian.ShouldIgnorePath(vaultPath, absPath, ignored)
}
