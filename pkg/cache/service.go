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

// ═══════════════════════════════════════════════════════════════════════════════
// TYPES AND DATA STRUCTURES
// ═══════════════════════════════════════════════════════════════════════════════
//
// This section defines the core data structures for the cache service:
//   - Service: the main cache coordinator
//   - Entry: cached metadata for a single markdown file
//   - DirtyKind: why a path needs revalidation
//   - Watcher: abstraction over filesystem notifications

// Service maintains an in-memory cache of vault files, metadata, and tags.
//
// Architecture overview:
//
//	┌─────────────────────────────────────────────────────────────────────┐
//	│                           Service                                   │
//	│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐              │
//	│  │  fileIndex   │  │   tagIndex   │  │   dirIndex   │              │
//	│  │ path→Entry   │  │  tag→paths   │  │watched dirs  │              │
//	│  └──────────────┘  └──────────────┘  └──────────────┘              │
//	│         ▲                                    ▲                      │
//	│         │ refreshPath                        │ addWatch             │
//	│         │                                    │                      │
//	│  ┌──────┴───────────────────────────────────┴──────┐               │
//	│  │                  Refresh()                       │               │
//	│  │  Consumes dirty markers, revalidates stale data  │               │
//	│  └──────────────────────▲──────────────────────────┘               │
//	│                         │                                           │
//	│         ┌───────────────┴───────────────┐                          │
//	│         │         dirty map             │                          │
//	│         │   path → DirtyKind            │                          │
//	│         └───────────────▲───────────────┘                          │
//	│                         │ markDirty                                 │
//	│                         │                                           │
//	│  ┌──────────────────────┴──────────────────────┐                   │
//	│  │              watchLoop (goroutine)           │                   │
//	│  │   Translates fsnotify events → dirty markers │                   │
//	│  └──────────────────────────────────────────────┘                   │
//	└─────────────────────────────────────────────────────────────────────┘
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

	// Mutex guards all mutable state below. Use RLock for reads, Lock for writes.
	mu        sync.RWMutex
	watchMu   sync.Mutex                     // guards watcher lifecycle (ctx/cancel/watchers/once)
	ready     bool                           // true after initial crawl completes
	crawling  bool                           // guards against concurrent initial crawls
	stale     bool                           // set when watcher fails; forces full revalidation
	fileIndex map[string]*Entry              // path → cached entry
	tagIndex  map[string]map[string]struct{} // tag → set of paths
	dirIndex  map[string]struct{}            // directories currently watched
	dirty     map[string]DirtyKind           // paths needing revalidation
	ignored   []string                       // patterns from .obsidianignore

	// Watcher subsystem (may be nil if unavailable)
	watcher        Watcher
	watcherFactory func() (Watcher, error)
	watchOnce      sync.Once // ensures watchLoop starts exactly once per lifecycle
	tickerOnce     sync.Once // ensures stale ticker starts exactly once per lifecycle

	// Lifecycle management
	ctx           context.Context    // guarded by watchMu
	cancel        context.CancelFunc // guarded by watchMu
	version       uint64             // monotonic counter, bumped on changes
	staleInterval time.Duration      // if >0, periodically mark stale for polling fallback

	// Channel for queueing watches from watchLoop to avoid deadlock on Windows.
	// On Windows, watcher.Add() communicates with the internal read goroutine.
	// If watchLoop calls Add() and blocks, it can't drain Events, which causes
	// the internal goroutine to block, which causes Add() to never complete.
	pendingWatches chan string
}

// Entry represents cached metadata for a single markdown file.
// All fields are extracted during refreshPath and stored for fast access.
type Entry struct {
	Path        string                 // relative path within vault (normalized)
	ModTime     time.Time              // last modification time from filesystem
	Size        int64                  // file size in bytes
	Tags        []string               // normalized tags (lowercase, no # prefix)
	Frontmatter map[string]interface{} // parsed YAML frontmatter
	InlineProps map[string][]string    // Dataview-style inline properties
	Content     string                 // full file content
	ContentTime time.Time              // derived content timestamp (frontmatter/filename/heading)
}

// DirtyKind captures why a path was marked dirty. The kind affects how
// Refresh() processes the path (e.g., remove vs re-read).
type DirtyKind string

const (
	DirtyUnknown   DirtyKind = "unknown"
	DirtyCreated   DirtyKind = "created"   // new file appeared
	DirtyModified  DirtyKind = "modified"  // existing file changed
	DirtyRemoved   DirtyKind = "removed"   // file was deleted
	DirtyRenamed   DirtyKind = "renamed"   // file was renamed (old path)
	DirtyRecreated DirtyKind = "recreated" // rapid delete+create sequence
)

// Options controls cache behavior.
type Options struct {
	Watcher        Watcher                 // inject a custom watcher (for testing)
	WatcherFactory func() (Watcher, error) // factory to rebuild watcher after failure
	StaleInterval  time.Duration           // if >0, periodically force revalidation
}

// Watcher abstracts filesystem notifications. The default implementation wraps
// fsnotify, but tests can inject a stub.
type Watcher interface {
	Add(name string) error
	Close() error
	Events() <-chan fsnotify.Event
	Errors() <-chan error
}

// fsNotifyWatcher wraps the real fsnotify.Watcher to implement our interface.
type fsNotifyWatcher struct {
	*fsnotify.Watcher
}

func (f *fsNotifyWatcher) Events() <-chan fsnotify.Event { return f.Watcher.Events }
func (f *fsNotifyWatcher) Errors() <-chan error          { return f.Watcher.Errors }

// ═══════════════════════════════════════════════════════════════════════════════
// CONSTRUCTOR AND LIFECYCLE
// ═══════════════════════════════════════════════════════════════════════════════
//
// These functions manage the Service's lifecycle: creation, initialization, and
// cleanup. The service starts "cold" and lazily initializes on first use.

// NewService constructs a cache service for a vault. The service is not yet
// ready; call EnsureReady() to trigger the initial crawl.
func NewService(vaultPath string, opts Options) (*Service, error) {
	if vaultPath == "" {
		return nil, errors.New("vaultPath is required")
	}

	var watcher Watcher
	var watcherFactory func() (Watcher, error)
	if opts.Watcher != nil {
		// Injected watcher (typically for testing)
		watcher = opts.Watcher
		watcherFactory = opts.WatcherFactory
	} else {
		// Production: create a real fsnotify watcher
		watcherFactory = func() (Watcher, error) {
			w, err := fsnotify.NewWatcher()
			if err != nil {
				return nil, fmt.Errorf("create watcher: %w", err)
			}
			return &fsNotifyWatcher{Watcher: w}, nil
		}
		w, err := watcherFactory()
		if err != nil {
			// Graceful degradation: fall back to polling-only mode
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
		pendingWatches: make(chan string, 100),
	}, nil
}

// Close stops the watcher and releases resources. Safe to call multiple times.
func (s *Service) Close() error {
	s.watchMu.Lock()
	if s.cancel != nil {
		s.cancel()
	}
	w := s.watcher
	s.watchMu.Unlock()
	if w != nil {
		return w.Close()
	}
	return nil
}

// EnsureReady performs the initial crawl (once) and starts the watcher.
// It is safe to call concurrently; only one goroutine will perform the initial crawl.
// Subsequent calls delegate to Refresh().
func (s *Service) EnsureReady(ctx context.Context) error {
	s.mu.Lock()
	if s.ready {
		s.mu.Unlock()
		return s.Refresh(ctx)
	}

	// Concurrency gate: if another goroutine is already crawling, wait for it.
	if s.crawling {
		s.mu.Unlock()
		// Spin-wait for crawl to complete. A condition variable would be more
		// efficient but adds complexity for a rare case.
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

	// Start watching BEFORE crawling to avoid missing events during the crawl.
	s.startWatcher()
	if err := s.initialCrawl(ctx); err != nil {
		s.mu.Lock()
		s.crawling = false
		s.mu.Unlock()
		return err
	}
	s.startStaleTicker()
	return s.Refresh(ctx)
}

// ═══════════════════════════════════════════════════════════════════════════════
// PUBLIC API (READING)
// ═══════════════════════════════════════════════════════════════════════════════
//
// These methods provide read access to cached data. Most do NOT call Refresh()
// automatically; callers should call Refresh() first to ensure freshness.
// EntriesSnapshot is the exception—it refreshes before returning.

// Paths returns all cached paths.
// Note: Callers should call Refresh() first to ensure freshness.
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

// EntriesSnapshot returns a deep copy of all cached entries after ensuring freshness.
// This is the safest way to get a consistent view of the entire cache.
func (s *Service) EntriesSnapshot(ctx context.Context) ([]Entry, error) {
	if err := s.Refresh(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	entries := make([]Entry, 0, len(s.fileIndex))
	for _, e := range s.fileIndex {
		entries = append(entries, e.clone())
	}
	return entries, nil
}

// Entry returns a copy of the cached entry for the given path.
// Note: Callers should call Refresh() first to ensure freshness.
func (s *Service) Entry(path string) (Entry, bool) {
	norm := obsidian.NormalizePath(obsidian.AddMdSuffix(path))

	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.fileIndex[norm]
	if !ok {
		return Entry{}, false
	}
	// Return a deep copy to prevent callers from mutating the cache.
	return entry.clone(), true
}

// clone returns a deep copy of the Entry, protecting internal slices and maps.
func (e *Entry) clone() Entry {
	out := Entry{
		Path:        e.Path,
		ModTime:     e.ModTime,
		Size:        e.Size,
		Content:     e.Content,
		ContentTime: e.ContentTime,
	}
	if len(e.Tags) > 0 {
		out.Tags = make([]string, len(e.Tags))
		copy(out.Tags, e.Tags)
	}
	if len(e.Frontmatter) > 0 {
		out.Frontmatter = make(map[string]interface{}, len(e.Frontmatter))
		for k, v := range e.Frontmatter {
			out.Frontmatter[k] = v // Note: nested structures still share references
		}
	}
	if len(e.InlineProps) > 0 {
		out.InlineProps = make(map[string][]string, len(e.InlineProps))
		for k, v := range e.InlineProps {
			cp := make([]string, len(v))
			copy(cp, v)
			out.InlineProps[k] = cp
		}
	}
	return out
}

// ═══════════════════════════════════════════════════════════════════════════════
// REFRESH AND RECONCILIATION
// ═══════════════════════════════════════════════════════════════════════════════
//
// Refresh is the "checkpoint" that reconciles in-memory state with the filesystem.
// It consumes dirty markers accumulated by watchLoop and applies them. If the
// watcher failed (stale flag), it triggers a full resync.

// Refresh reconciles in-memory state with the filesystem.
//   - If this is the first call, it delegates to EnsureReady().
//   - If the watcher signaled trouble (stale flag), it revalidates every entry.
//   - Otherwise it consumes dirty markers emitted by watchLoop().
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

	// Process each dirty marker according to its kind.
	for path, kind := range dirty {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		switch kind {
		case DirtyRemoved, DirtyRenamed:
			// Remove the entry (and children if it's a directory).
			s.removeTree(path)
			changed = true
			if kind == DirtyRenamed {
				// Rescan the parent directory to pick up the new name.
				parent := filepath.Dir(filepath.Join(s.vaultPath, path))
				_ = s.rescanDir(parent, true)
			}

		case DirtyRecreated:
			// A rapid delete+create sequence. Remove old data first.
			s.removeTree(path)
			fallthrough

		default:
			// Created, Modified, or Recreated: refresh from disk.
			absPath := filepath.Join(s.vaultPath, path)
			info, err := os.Stat(absPath)
			if err == nil && info.IsDir() {
				// It's a directory—scan it for files.
				if err := s.rescanDir(absPath, true); err != nil {
					s.markDirty(absPath, DirtyModified)
				} else {
					changed = true
				}
			} else {
				// It's a file—refresh it.
				if err := s.refreshPath(path); err != nil {
					// Keep the entry dirty so we retry next time.
					s.markDirty(absPath, DirtyModified)
				} else {
					changed = true
				}
			}
		}
	}

	if changed {
		s.bumpVersion()
	}

	return nil
}

// resync performs a full cache rebuild. Called when the watcher fails or the
// cache becomes too stale to trust incremental updates.
func (s *Service) resync(ctx context.Context) error {
	// Stop the existing watch loop to avoid duplicate goroutines.
	s.watchMu.Lock()
	if s.cancel != nil {
		s.cancel()
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())

	// Reset lifecycle state for new goroutines.
	s.watchOnce = sync.Once{}
	s.tickerOnce = sync.Once{}

	// Drain any stale pending watches from the previous lifecycle.
drainLoop:
	for {
		select {
		case <-s.pendingWatches:
		default:
			break drainLoop
		}
	}

	// Rebuild the watcher (if factory is available).
	if _, err := s.rebuildWatcherLocked(); err != nil {
		s.watchMu.Unlock()
		return err
	}
	s.watchMu.Unlock()

	// Reset all state and re-crawl.
	s.mu.Lock()
	s.ready = false
	s.crawling = false
	s.fileIndex = make(map[string]*Entry)
	s.tagIndex = make(map[string]map[string]struct{})
	s.dirIndex = make(map[string]struct{})
	s.dirty = make(map[string]DirtyKind)
	s.watchOnce = sync.Once{}
	s.tickerOnce = sync.Once{}
	s.mu.Unlock()

	s.startWatcher()
	if err := s.initialCrawl(ctx); err != nil {
		return err
	}
	s.startStaleTicker()
	return nil
}

// ═══════════════════════════════════════════════════════════════════════════════
// CRAWLING AND SCANNING
// ═══════════════════════════════════════════════════════════════════════════════
//
// These functions walk the filesystem and populate the cache. initialCrawl runs
// once at startup; rescanDir handles incremental updates for new directories.

// initialCrawl walks the vault and populates the cache. It runs in two phases:
//  1. Walk the tree, install directory watches, and collect file paths.
//  2. Read each file and extract metadata.
//
// This two-phase approach keeps the WalkDir callback simple and allows the
// watcher to catch events that occur during the read phase.
func (s *Service) initialCrawl(ctx context.Context) error {
	s.loadIgnorePatterns()

	type job struct {
		path string
		info fs.DirEntry
	}

	// Phase 1: Walk and collect work.
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
			// Watch directories immediately to catch files created during Phase 2.
			s.addWatch(path)
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

	// Phase 2: Read files and build indices.
	for _, j := range jobs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := s.refreshPath(j.path); err != nil {
			// Best effort: mark dirty for retry rather than failing the whole crawl.
			s.markDirty(j.path, DirtyModified)
		}
	}

	s.mu.Lock()
	s.ready = true
	s.version++
	s.mu.Unlock()
	return nil
}

// refreshPath reads a single file from disk and updates the cache.
// It handles both absolute and relative paths.
func (s *Service) refreshPath(absPath string) error {
	if !filepath.IsAbs(absPath) {
		absPath = filepath.Join(s.vaultPath, absPath)
	}

	// Normalize the path for consistent map keys.
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

	// Single disk read; reuse content for all extractors.
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

	if ct, ok := obsidian.ResolveContentTime(rel, entry.Content); ok {
		entry.ContentTime = ct
	} else if !info.ModTime().IsZero() {
		entry.ContentTime = info.ModTime()
	}

	// Extract metadata from content.
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

	// Update indices atomically.
	s.mu.Lock()
	s.removeTagsForPathLocked(rel)
	s.fileIndex[rel] = entry
	s.indexTags(rel, entry.Tags)
	s.mu.Unlock()
	return nil
}

// rescanDir refreshes all files in a directory. Used after a new directory
// is created or after a rename to pick up the new contents.
func (s *Service) rescanDir(absDir string, recursive bool) error {
	entries, err := os.ReadDir(absDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			s.addWatch(filepath.Join(absDir, entry.Name()))
			if recursive {
				_ = s.rescanDir(filepath.Join(absDir, entry.Name()), true)
			}
			continue
		}
		if err := s.refreshPath(filepath.Join(absDir, entry.Name())); err != nil {
			s.markDirty(filepath.Join(absDir, entry.Name()), DirtyModified)
			continue
		}
	}
	return nil
}

// loadIgnorePatterns reads .obsidianignore during crawl/resync. Changes are
// picked up on the next resync or refresh that triggers a rescan.
func (s *Service) loadIgnorePatterns() {
	ignorePath := filepath.Join(s.vaultPath, ".obsidianignore")
	patterns := obsidian.DefaultIgnorePatterns()
	content, err := os.ReadFile(ignorePath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			// Silently ignore read errors; use defaults.
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

// ═══════════════════════════════════════════════════════════════════════════════
// INDEX MANAGEMENT
// ═══════════════════════════════════════════════════════════════════════════════
//
// These functions maintain the fileIndex and tagIndex. They handle adding,
// removing, and updating entries while keeping indices consistent.

// removePath removes a single file from the cache.
func (s *Service) removePath(rel string) {
	rel = obsidian.NormalizePath(rel)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.removeTagsForPathLocked(rel)
	delete(s.fileIndex, rel)
}

// removeTree removes a path and all children (for directory deletions/renames).
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

// removeTagsForPathLocked cleans up the tag index when removing a file.
// Caller must hold s.mu.
func (s *Service) removeTagsForPathLocked(rel string) {
	for tag, paths := range s.tagIndex {
		delete(paths, rel)
		if len(paths) == 0 {
			delete(s.tagIndex, tag)
		}
	}
}

// indexTags adds a file's tags to the tag index. Caller must hold s.mu.
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

// ═══════════════════════════════════════════════════════════════════════════════
// DIRTY TRACKING
// ═══════════════════════════════════════════════════════════════════════════════
//
// The dirty map accumulates filesystem changes between Refresh() calls. The
// markDirty function implements a simple state machine to coalesce events
// (e.g., a rapid delete+create becomes "recreated").

// markDirty records that a path needs revalidation. It implements state
// transitions to handle edge cases like rapid delete+create sequences.
//
// State machine:
//
//	              ┌─────────────┐
//	Created ──────│             │────── Modified
//	              │   (path)    │
//	Removed ◄─────│             │──────► Removed (sticky)
//	    │         └─────────────┘
//	    │              ▲
//	    │    Created   │
//	    └──────────────┴───► Recreated
func (s *Service) markDirty(absPath string, kind DirtyKind) {
	rel, err := filepath.Rel(s.vaultPath, absPath)
	if err != nil {
		return
	}
	rel = obsidian.NormalizePath(rel)

	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, ok := s.dirty[rel]; ok {
		// State transitions for edge cases:

		// Removed → Created/Modified = Recreated (rapid delete+create)
		if existing == DirtyRemoved && (kind == DirtyCreated || kind == DirtyModified) {
			s.dirty[rel] = DirtyRecreated
			return
		}

		// Recreated → Removed = Removed (the recreation was also deleted)
		if existing == DirtyRecreated && kind == DirtyRemoved {
			s.dirty[rel] = DirtyRemoved
			return
		}

		// Removed is sticky: once removed, stay removed (until recreation detected).
		if existing == DirtyRemoved || kind == DirtyRemoved {
			s.dirty[rel] = DirtyRemoved
		}
		return
	}
	s.dirty[rel] = kind
}

// bumpVersion increments the version counter. Called after processing changes.
func (s *Service) bumpVersion() {
	s.mu.Lock()
	s.version++
	s.mu.Unlock()
}

// ═══════════════════════════════════════════════════════════════════════════════
// FILESYSTEM WATCHING
// ═══════════════════════════════════════════════════════════════════════════════
//
// This subsystem translates raw filesystem events into dirty markers. It runs
// in a background goroutine (watchLoop) and uses fsnotify for cross-platform
// file watching. On watcher failure, it sets the stale flag to trigger a full
// resync on the next Refresh().

// startWatcher launches the watch loop goroutine (once per lifecycle).
func (s *Service) startWatcher() {
	s.watchMu.Lock()
	defer s.watchMu.Unlock()
	if s.watcher == nil {
		return
	}
	s.watchOnce.Do(func() {
		go s.watchLoop(s.ctx, s.watcher)
		go s.pendingWatchLoop(s.ctx)
	})
}

// pendingWatchLoop processes directories queued by watchLoop.
// This runs in a separate goroutine so that watchLoop can continue draining
// events while watcher.Add() blocks (which it does on Windows).
func (s *Service) pendingWatchLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case path, ok := <-s.pendingWatches:
			if !ok {
				return
			}
			s.addWatch(path)
		}
	}
}

// startStaleTicker launches a periodic stale trigger for polling fallback.
// Used when the watcher is unavailable or as a safety net.
func (s *Service) startStaleTicker() {
	if s.staleInterval <= 0 {
		return
	}
	s.watchMu.Lock()
	ctx := s.ctx
	s.tickerOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(s.staleInterval)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					s.markStale()
				}
			}
		}()
	})
	s.watchMu.Unlock()
}

// addWatch registers a directory for filesystem notifications.
func (s *Service) addWatch(path string) {
	if path == "" {
		return
	}
	s.watchMu.Lock()
	watcher := s.watcher
	s.watchMu.Unlock()
	if watcher == nil {
		return
	}
	s.mu.Lock()
	if _, ok := s.dirIndex[path]; ok {
		s.mu.Unlock()
		return
	}
	s.dirIndex[path] = struct{}{}
	s.mu.Unlock()
	_ = watcher.Add(path)
}

// watchLoop is the main event loop for filesystem notifications. It translates
// fsnotify events into dirty markers and handles watcher errors by setting the
// stale flag.
func (s *Service) watchLoop(ctx context.Context, watcher Watcher) {
	for {
		select {
		case <-ctx.Done():
			return

		case evt, ok := <-watcher.Events():
			if !ok {
				// Channel closed unexpectedly; mark stale for safety.
				s.markStale()
				return
			}

			switch {
			case evt.Op&fsnotify.Create == fsnotify.Create:
				s.markDirty(evt.Name, DirtyCreated)
				// Queue new directories for watching via pendingWatchLoop.
				// We must not call addWatch directly from watchLoop because on Windows,
				// watcher.Add() blocks waiting for the internal read goroutine, but that
				// goroutine may be blocked trying to send events that we're not reading
				// (because we're blocked on Add). Using a channel lets watchLoop continue
				// draining events while pendingWatchLoop handles the Add calls.
				info, err := os.Stat(evt.Name)
				if err == nil && info.IsDir() {
					select {
					case s.pendingWatches <- evt.Name:
					default:
						// Buffer full; directory will be picked up on next Refresh.
					}
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

		case err, ok := <-watcher.Errors():
			if !ok {
				s.markStale()
				return
			}
			// Watcher error; mark stale to trigger full revalidation.
			s.markStale()
			_ = err // TODO: log in debug mode
		}
	}
}

// markStale flags the cache as potentially out of sync. The next Refresh()
// will trigger a full revalidation.
func (s *Service) markStale() {
	s.mu.Lock()
	s.stale = true
	s.mu.Unlock()
}

// dropDirIndex removes a directory from the watch list (after deletion/rename).
func (s *Service) dropDirIndex(path string) {
	s.mu.Lock()
	delete(s.dirIndex, path)
	s.mu.Unlock()
}

// rebuildWatcher closes the old watcher and creates a new one.
// Used during resync to recover from watcher failures.
func (s *Service) rebuildWatcherLocked() (Watcher, error) {
	// watchMu must be held by callers.
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

// ═══════════════════════════════════════════════════════════════════════════════
// UTILITIES
// ═══════════════════════════════════════════════════════════════════════════════
//
// Helper functions for path filtering and tag normalization.

// fileInfo is a minimal interface for path filtering, satisfied by both
// fs.DirEntry and os.FileInfo.
type fileInfo interface {
	Name() string
	IsDir() bool
}

// shouldSkip returns true if a path should be excluded from the cache.
// Skips: hidden files, non-markdown files, and paths matching ignore patterns.
func shouldSkip(vaultPath, relPath string, info fileInfo, ignored []string) bool {
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
	absPath := filepath.Join(vaultPath, relPath)
	return obsidian.ShouldIgnorePath(vaultPath, absPath, ignored)
}

// normalizeTags converts tags to lowercase and removes # prefixes.
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

// stripHashtagPrefix removes # from tags (used before normalization).
func stripHashtagPrefix(tags []string) []string {
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		out = append(out, strings.TrimPrefix(t, "#"))
	}
	return out
}
