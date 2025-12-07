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
)

// Indexer coordinates scanning a vault and keeping the embedding index fresh.
type Indexer struct {
	Index         Index
	Provider      Provider
	Root          string
	BatchSize     int
	MaxConcurrent int
}

// NewIndexer constructs an indexer with sensible defaults for batching and concurrency.
func NewIndexer(idx Index, provider Provider, root string) *Indexer {
	return &Indexer{
		Index:         idx,
		Provider:      provider,
		Root:          root,
		BatchSize:     8,
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

		hash, content, err := contentHashForFile(info.Path)
		if err != nil {
			return fmt.Errorf("hash file %s: %w", info.Path, err)
		}
		if err := ix.indexNote(ctx, info, content, hash, hasEmb && hash == prevHash); err != nil {
			return err
		}
	}

	if err := ix.Index.DeleteNotesNotIn(ctx, ids); err != nil {
		return fmt.Errorf("delete removed notes: %w", err)
	}

	return nil
}

type embedTask struct {
	info    NoteFileInfo
	content string
	hash    string
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
		vecs, err := ix.Provider.EmbedTexts(ctx, changedTexts)
		if err != nil {
			return fmt.Errorf("embed chunks %s: %w", info.Path, err)
		}
		if len(vecs) != len(changedChunks) {
			return fmt.Errorf("expected %d chunk embeddings, got %d", len(changedChunks), len(vecs))
		}
		if err := ix.Index.UpsertNoteChunks(ctx, info.ID, changedChunks, vecs); err != nil {
			return fmt.Errorf("upsert chunks %s: %w", info.Path, err)
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

func contentHashForFile(path string) (string, string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), string(b), nil
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
