package cache

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
	"github.com/fsnotify/fsnotify"
)

// Service maintains an in-memory cache of vault files, metadata, and tags.
// It performs a full crawl once, watches for filesystem changes, and refreshes
// dirty entries before serving requests.
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
	ignored   []string            // loaded from .obsidianignore

	watcher   Watcher
	watchOnce sync.Once
	ctx       context.Context
	cancel    context.CancelFunc
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
	Watcher Watcher
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
	if opts.Watcher != nil {
		watcher = opts.Watcher
	} else {
		w, err := fsnotify.NewWatcher()
		if err != nil {
			return nil, fmt.Errorf("create watcher: %w", err)
		}
		watcher = &fsNotifyWatcher{Watcher: w}
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Service{
		vaultPath: vaultPath,
		fileIndex: make(map[string]*Entry),
		tagIndex:  make(map[string]map[string]struct{}),
		dirIndex:  make(map[string]struct{}),
		dirty:     make(map[string]DirtyKind),
		watcher:   watcher,
		ctx:       ctx,
		cancel:    cancel,
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
	return s.Refresh(ctx)
}

func (s *Service) loadIgnorePatterns() {
	ignorePath := filepath.Join(s.vaultPath, ".obsidianignore")
	content, err := os.ReadFile(ignorePath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			// Log or silently ignore read errors? For now silent as per spec assumptions.
		}
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
	s.ignored = ignored
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

// Refresh processes any dirty paths by re-reading or removing them.
// If the cache was marked stale (watcher failure), a full revalidation is performed.
func (s *Service) Refresh(ctx context.Context) error {
	s.mu.Lock()
	if !s.ready {
		s.mu.Unlock()
		return s.EnsureReady(ctx)
	}

	stale := s.stale
	if stale {
		s.stale = false
	}
	dirty := s.dirty
	s.dirty = make(map[string]DirtyKind)
	s.mu.Unlock()

	// If the watcher failed, revalidate all cached entries against disk.
	if stale {
		if err := s.revalidateAll(ctx); err != nil {
			return err
		}
	}

	if len(dirty) == 0 {
		return nil
	}

	for path, kind := range dirty {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		switch kind {
		case DirtyRemoved, DirtyRenamed:
			s.removePath(path)
			if kind == DirtyRenamed {
				parent := filepath.Dir(filepath.Join(s.vaultPath, path))
				_ = s.rescanDir(parent)
			}
		default:
			if err := s.refreshPath(path); err != nil {
				// Keep the entry dirty so we retry next time.
				s.markDirty(path, DirtyModified)
			}
		}
	}

	return nil
}

// revalidateAll checks all cached paths against the filesystem and marks
// any out-of-sync entries dirty. This is called when the watcher failed.
func (s *Service) revalidateAll(ctx context.Context) error {
	s.mu.RLock()
	paths := make([]string, 0, len(s.fileIndex))
	for p := range s.fileIndex {
		paths = append(paths, p)
	}
	s.mu.RUnlock()

	for _, rel := range paths {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		absPath := filepath.Join(s.vaultPath, rel)
		info, err := os.Stat(absPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				s.removePath(rel)
			}
			continue
		}

		s.mu.RLock()
		entry, ok := s.fileIndex[rel]
		s.mu.RUnlock()
		if !ok {
			continue
		}

		// Check if file changed (mtime or size mismatch).
		if info.ModTime() != entry.ModTime || info.Size() != entry.Size {
			if err := s.refreshPath(absPath); err != nil {
				s.markDirty(absPath, DirtyModified)
			}
		}
	}

	return nil
}

func (s *Service) initialCrawl(ctx context.Context) error {
	s.loadIgnorePatterns()

	type job struct {
		path string
		info fs.DirEntry
	}

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
	s.mu.Unlock()
	return nil
}

func (s *Service) refreshPath(absPath string) error {
	if !filepath.IsAbs(absPath) {
		absPath = filepath.Join(s.vaultPath, absPath)
	}

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

func (s *Service) removeTagsForPathLocked(rel string) {
	for tag, paths := range s.tagIndex {
		delete(paths, rel)
		if len(paths) == 0 {
			delete(s.tagIndex, tag)
		}
	}
}

// rescanDir refreshes all files in a directory (non-recursive).
func (s *Service) rescanDir(absDir string) error {
	entries, err := os.ReadDir(absDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			s.addWatch(filepath.Join(absDir, entry.Name()))
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
		// Prefer removal markers to avoid stale entries.
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
	s.watchOnce.Do(func() {
		go s.watchLoop()
	})
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
					_ = s.rescanDir(evt.Name)
				}
			case evt.Op&fsnotify.Write == fsnotify.Write:
				s.markDirty(evt.Name, DirtyModified)
			case evt.Op&fsnotify.Remove == fsnotify.Remove:
				s.markDirty(evt.Name, DirtyRemoved)
			case evt.Op&fsnotify.Rename == fsnotify.Rename:
				s.markDirty(evt.Name, DirtyRenamed)
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
	
	// Check ignore patterns
	absPath := filepath.Join(vaultPath, relPath)
	return obsidian.ShouldIgnorePath(vaultPath, absPath, ignored)
}
