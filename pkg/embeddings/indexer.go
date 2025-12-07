package embeddings

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Indexer coordinates scanning a vault and keeping the embedding index fresh.
type Indexer struct {
	Index         Index
	Provider      Provider
	ProviderInfo  ProviderConfig
	Root          string
	BatchSize     int
	MaxConcurrent int
}

// NewIndexer constructs an indexer with sensible defaults for batching and concurrency.
func NewIndexer(idx Index, provider Provider, info ProviderConfig, root string) *Indexer {
	return &Indexer{
		Index:         idx,
		Provider:      provider,
		ProviderInfo:  info,
		Root:          root,
		BatchSize:     DefaultBatchSize,
		MaxConcurrent: DefaultMaxConcurrent,
	}
}

// ScanVault walks the vault root and returns markdown files with metadata.
func (ix *Indexer) ScanVault() ([]NoteFileInfo, error) {
	if ix.Root == "" {
		return nil, errors.New("indexer root is required")
	}

	var notes []NoteFileInfo
	err := filepath.WalkDir(ix.Root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if shouldSkipDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(d.Name()) != ".md" || strings.HasPrefix(d.Name(), ".") {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(ix.Root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		notes = append(notes, NoteFileInfo{
			ID:    NoteID(rel),
			Path:  path,
			Title: titleFromPath(path),
			Size:  info.Size(),
			Mtime: info.ModTime(),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return notes, nil
}

// SyncVault refreshes the index for all notes under Root.
func (ix *Indexer) SyncVault(ctx context.Context) error {
	if ix.Index == nil || ix.Provider == nil {
		return errors.New("indexer is missing index or provider")
	}
	if err := ix.Index.EnsureSchema(ctx); err != nil {
		return err
	}

	if err := ix.Index.ValidateOrInitMetadata(ctx, IndexMetadata{
		Provider:   ix.ProviderInfo.Provider,
		Model:      ix.ProviderInfo.Model,
		Dimensions: ix.Provider.Dimensions(),
	}); err != nil {
		return err
	}

	existing, err := ix.Index.ListNotes(ctx)
	if err != nil {
		return err
	}
	existingMap := make(map[NoteID]NoteFileInfo, len(existing))
	for _, n := range existing {
		existingMap[n.ID] = n
	}

	files, err := ix.ScanVault()
	if err != nil {
		return err
	}

	ids := make([]NoteID, 0, len(files))
	var tasks []embedTask
	for _, info := range files {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		ids = append(ids, info.ID)
		if err := ix.Index.UpsertNoteMeta(ctx, info); err != nil {
			return fmt.Errorf("upsert meta %s: %w", info.Path, err)
		}

		prev, ok := existingMap[info.ID]
		metaChanged := !ok || !prev.Mtime.Equal(info.Mtime) || prev.Size != info.Size
		if !metaChanged {
			continue
		}

		_, prevHash, hasEmb, err := ix.Index.GetNoteEmbedding(ctx, info.ID)
		if err != nil {
			return fmt.Errorf("get embedding %s: %w", info.Path, err)
		}

		tasks = append(tasks, embedTask{
			info:     info,
			prevHash: prevHash,
			hasEmb:   hasEmb,
		})
	}

	if err := ix.processTasks(ctx, tasks); err != nil {
		return err
	}

	if err := ix.Index.DeleteNotesNotIn(ctx, ids); err != nil {
		return fmt.Errorf("delete removed notes: %w", err)
	}

	_ = ix.Index.UpdateLastSync(ctx, time.Now())
	return nil
}

type embedTask struct {
	info     NoteFileInfo
	prevHash string
	hasEmb   bool
}

func (ix *Indexer) processTasks(ctx context.Context, tasks []embedTask) error {
	if len(tasks) == 0 {
		return nil
	}
	workerCount := ix.MaxConcurrent
	if workerCount <= 0 {
		workerCount = DefaultMaxConcurrent
	}
	if workerCount > len(tasks) {
		workerCount = len(tasks)
	}

	taskCh := make(chan embedTask)
	errCh := make(chan error, 1)
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskCh {
				select {
				case <-ctx.Done():
					return
				default:
				}

				hash, content, err := contentHashForFile(task.info.Path)
				if err != nil {
					select {
					case errCh <- fmt.Errorf("hash file %s: %w", task.info.Path, err):
					default:
					}
					return
				}
				unchanged := task.hasEmb && hash == task.prevHash
				if err := ix.indexNote(ctx, task.info, content, hash, unchanged); err != nil {
					select {
					case errCh <- err:
					default:
					}
					return
				}
			}
		}()
	}

	for _, task := range tasks {
		select {
		case <-ctx.Done():
			close(taskCh)
			wg.Wait()
			return ctx.Err()
		case err := <-errCh:
			close(taskCh)
			wg.Wait()
			return err
		case taskCh <- task:
		}
	}
	close(taskCh)
	wg.Wait()

	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

func (ix *Indexer) indexNote(ctx context.Context, info NoteFileInfo, content string, fileHash string, unchanged bool) error {
	chunks, err := ChunkNote(info.Path, info.Title, content)
	if err != nil {
		return fmt.Errorf("chunk note %s: %w", info.Path, err)
	}

	existingHashes, err := ix.Index.ChunkHashes(ctx, info.ID)
	if err != nil {
		return fmt.Errorf("load chunk hashes %s: %w", info.Path, err)
	}

	allIndices := make([]int, 0, len(chunks))
	var changedChunks []ChunkInput
	var changedTexts []string
	for _, ch := range chunks {
		allIndices = append(allIndices, ch.Index)
		if existingHashes[ch.Index] == ch.Hash {
			continue
		}
		changedChunks = append(changedChunks, ch)
		changedTexts = append(changedTexts, ch.Text)
	}

	if err := ix.Index.DeleteChunksNotIn(ctx, info.ID, allIndices); err != nil {
		return fmt.Errorf("delete stale chunks %s: %w", info.Path, err)
	}

	if len(changedChunks) > 0 {
		batchSize := ix.batchSize()
		for start := 0; start < len(changedChunks); start += batchSize {
			end := start + batchSize
			if end > len(changedChunks) {
				end = len(changedChunks)
			}
			batchChunks := changedChunks[start:end]
			batchTexts := changedTexts[start:end]
			vecs, err := ix.Provider.EmbedTexts(ctx, batchTexts)
			if err != nil {
				return fmt.Errorf("embed chunks %s: %w", info.Path, err)
			}
			if len(vecs) != len(batchChunks) {
				return fmt.Errorf("expected %d chunk embeddings, got %d", len(batchChunks), len(vecs))
			}
			if err := ix.Index.UpsertNoteChunks(ctx, info.ID, batchChunks, vecs); err != nil {
				return fmt.Errorf("upsert chunks %s: %w", info.Path, err)
			}
		}
	}

	if unchanged {
		return nil
	}

	vecs, err := ix.Provider.EmbedTexts(ctx, []string{content})
	if err != nil {
		return fmt.Errorf("embed note %s: %w", info.Path, err)
	}
	if len(vecs) != 1 {
		return fmt.Errorf("expected 1 embedding for note %s, got %d", info.Path, len(vecs))
	}
	return ix.Index.UpsertNoteEmbedding(ctx, info.ID, fileHash, vecs[0])
}

// UpdateNote indexes a single note given its metadata and content (skips unchanged chunks).
func (ix *Indexer) UpdateNote(ctx context.Context, info NoteFileInfo, content string) error {
	if ix.Index == nil || ix.Provider == nil {
		return errors.New("indexer is missing index or provider")
	}
	if err := ix.Index.UpsertNoteMeta(ctx, info); err != nil {
		return fmt.Errorf("upsert meta %s: %w", info.Path, err)
	}
	_, prevHash, hasEmb, err := ix.Index.GetNoteEmbedding(ctx, info.ID)
	if err != nil {
		return fmt.Errorf("get embedding %s: %w", info.Path, err)
	}
	hash := hashContent(content)
	return ix.indexNote(ctx, info, content, hash, hasEmb && hash == prevHash)
}

func contentHashForFile(path string) (string, string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), string(b), nil
}

func hashContent(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

func (ix *Indexer) batchSize() int {
	if ix.BatchSize > 0 {
		return ix.BatchSize
	}
	return DefaultBatchSize
}

func shouldSkipDir(name string) bool {
	name = strings.ToLower(name)
	if strings.HasPrefix(name, ".") {
		return true
	}
	switch name {
	case "node_modules", "vendor", "bin", ".git", ".obsidian", ".obsidian-cli":
		return true
	default:
		return false
	}
}

func titleFromPath(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
