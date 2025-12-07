package sqlite

import (
	"context"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/atomicobject/obsidian-cli/pkg/embeddings"

	_ "modernc.org/sqlite"
)

// Store implements embeddings.Index backed by SQLite.
type Store struct {
	db         *sql.DB
	dimensions int
}

// DeleteChunksNotIn removes chunk embeddings whose indices are not in the provided set.
func (s *Store) DeleteChunksNotIn(ctx context.Context, id embeddings.NoteID, indices []int) error {
	row := s.db.QueryRowContext(ctx, `SELECT id FROM notes WHERE note_id = ?`, string(id))
	var rowID int64
	if err := row.Scan(&rowID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	if len(indices) == 0 {
		_, err := s.db.ExecContext(ctx, `DELETE FROM chunk_embeddings WHERE note_row_id = ?`, rowID)
		return err
	}
	holders := make([]string, len(indices))
	args := make([]any, 0, len(indices)+1)
	args = append(args, rowID)
	for i, idx := range indices {
		holders[i] = "?"
		args = append(args, idx)
	}
	stmt := fmt.Sprintf(`DELETE FROM chunk_embeddings WHERE note_row_id = ? AND chunk_index NOT IN (%s)`, strings.Join(holders, ","))
	_, err := s.db.ExecContext(ctx, stmt, args...)
	return err
}

// Open opens (or creates) a SQLite index at path with the expected vector dimensions.
func Open(path string, dimensions int) (*Store, error) {
	if path == "" {
		return nil, errors.New("sqlite path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create index directory: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	store := &Store{db: db, dimensions: dimensions}
	if err := store.EnsureSchema(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

// EnsureSchema creates tables and indices if needed.
func (s *Store) EnsureSchema(ctx context.Context) error {
	stmts := []string{
		`PRAGMA foreign_keys = ON;`,
		`CREATE TABLE IF NOT EXISTS notes (
			id              INTEGER PRIMARY KEY,
			note_id         TEXT NOT NULL UNIQUE,
			title           TEXT NOT NULL,
			path            TEXT NOT NULL,
			last_seen_mtime INTEGER NOT NULL,
			last_seen_size  INTEGER NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS note_embeddings (
			id           INTEGER PRIMARY KEY,
			note_row_id  INTEGER NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
			content_hash TEXT NOT NULL,
			embedding    BLOB NOT NULL,
			dimensions   INTEGER NOT NULL,
			created_at   INTEGER NOT NULL,
			UNIQUE(note_row_id)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_note_embeddings_note_row_id ON note_embeddings(note_row_id);`,
		`CREATE TABLE IF NOT EXISTS chunk_embeddings (
			id           INTEGER PRIMARY KEY,
			note_row_id  INTEGER NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
			chunk_index  INTEGER NOT NULL,
			breadcrumb   TEXT,
			heading      TEXT,
			content_hash TEXT NOT NULL,
			embedding    BLOB NOT NULL,
			dimensions   INTEGER NOT NULL,
			created_at   INTEGER NOT NULL,
			UNIQUE(note_row_id, chunk_index)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_chunk_embeddings_note_row_id ON chunk_embeddings(note_row_id);`,
	}

	for _, stmt := range stmts {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

// Close releases database resources.
func (s *Store) Close() error {
	return s.db.Close()
}

// ChunkHashes returns existing chunk hashes for a note keyed by chunk index.
func (s *Store) ChunkHashes(ctx context.Context, id embeddings.NoteID) (map[int]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT c.chunk_index, c.content_hash
		FROM chunk_embeddings c
		JOIN notes n ON c.note_row_id = n.id
		WHERE n.note_id = ?
	`, string(id))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make(map[int]string)
	for rows.Next() {
		var idx int
		var hash string
		if err := rows.Scan(&idx, &hash); err != nil {
			return nil, err
		}
		res[idx] = hash
	}
	return res, rows.Err()
}

// UpsertNoteMeta inserts or updates metadata for a note.
func (s *Store) UpsertNoteMeta(ctx context.Context, info embeddings.NoteFileInfo) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO notes (note_id, title, path, last_seen_mtime, last_seen_size)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(note_id) DO UPDATE SET
			title = excluded.title,
			path = excluded.path,
			last_seen_mtime = excluded.last_seen_mtime,
			last_seen_size = excluded.last_seen_size
	`, string(info.ID), info.Title, info.Path, info.Mtime.Unix(), info.Size)
	return err
}

// DeleteNotesNotIn removes notes (and embeddings via cascade) not present in existingIDs.
func (s *Store) DeleteNotesNotIn(ctx context.Context, existingIDs []embeddings.NoteID) error {
	if len(existingIDs) == 0 {
		_, err := s.db.ExecContext(ctx, `DELETE FROM notes`)
		return err
	}

	args := make([]any, len(existingIDs))
	holders := make([]string, len(existingIDs))
	for i, id := range existingIDs {
		args[i] = string(id)
		holders[i] = "?"
	}

	query := fmt.Sprintf(`DELETE FROM notes WHERE note_id NOT IN (%s)`, strings.Join(holders, ","))
	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

// ListNotes returns metadata for all tracked notes.
func (s *Store) ListNotes(ctx context.Context) ([]embeddings.NoteFileInfo, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT note_id, title, path, last_seen_mtime, last_seen_size
		FROM notes
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []embeddings.NoteFileInfo
	for rows.Next() {
		var id, title, path string
		var mtimeUnix int64
		var size int64
		if err := rows.Scan(&id, &title, &path, &mtimeUnix, &size); err != nil {
			return nil, err
		}
		res = append(res, embeddings.NoteFileInfo{
			ID:    embeddings.NoteID(id),
			Title: title,
			Path:  path,
			Mtime: time.Unix(mtimeUnix, 0),
			Size:  size,
		})
	}
	return res, rows.Err()
}

// UpsertNoteEmbedding inserts or updates an embedding for a note.
func (s *Store) UpsertNoteEmbedding(ctx context.Context, id embeddings.NoteID, contentHash string, emb embeddings.Embedding) error {
	if s.dimensions > 0 && len(emb) != s.dimensions {
		return fmt.Errorf("embedding dimension mismatch: have %d want %d", len(emb), s.dimensions)
	}
	if len(emb) == 0 {
		return errors.New("empty embedding")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var rowID int64
	if err = tx.QueryRowContext(ctx, `SELECT id FROM notes WHERE note_id = ?`, string(id)).Scan(&rowID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("note metadata missing for %s", id)
		}
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO note_embeddings (note_row_id, content_hash, embedding, dimensions, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(note_row_id) DO UPDATE SET
			content_hash = excluded.content_hash,
			embedding = excluded.embedding,
			dimensions = excluded.dimensions,
			created_at = excluded.created_at
	`, rowID, contentHash, embedToBytes(emb), len(emb), time.Now().Unix())
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err == nil && s.dimensions == 0 {
		s.dimensions = len(emb)
	}
	return err
}

// GetNoteEmbedding returns an embedding and content hash if present.
func (s *Store) GetNoteEmbedding(ctx context.Context, id embeddings.NoteID) (embeddings.Embedding, string, bool, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT e.embedding, e.content_hash
		FROM note_embeddings e
		JOIN notes n ON e.note_row_id = n.id
		WHERE n.note_id = ?
	`, string(id))

	var blob []byte
	var hash string
	err := row.Scan(&blob, &hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", false, nil
		}
		return nil, "", false, err
	}
	return bytesToEmbed(blob), hash, true, nil
}

// UpsertNoteChunks upserts embeddings for provided chunk indices (skips unchanged if caller omits).
func (s *Store) UpsertNoteChunks(ctx context.Context, id embeddings.NoteID, chunks []embeddings.ChunkInput, vecs []embeddings.Embedding) error {
	if len(chunks) != len(vecs) {
		return fmt.Errorf("chunks/embeddings length mismatch: %d vs %d", len(chunks), len(vecs))
	}
	if len(chunks) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var rowID int64
	if err = tx.QueryRowContext(ctx, `SELECT id FROM notes WHERE note_id = ?`, string(id)).Scan(&rowID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("note metadata missing for %s", id)
		}
		return err
	}

	if _, err = tx.ExecContext(ctx, `DELETE FROM chunk_embeddings WHERE note_row_id = ?`, rowID); err != nil {
		return err
	}

	now := time.Now().Unix()
	for i, chunk := range chunks {
		vec := vecs[i]
		if len(vec) == 0 {
			continue
		}
		if s.dimensions > 0 && len(vec) != s.dimensions {
			return fmt.Errorf("chunk dimension mismatch: have %d want %d", len(vec), s.dimensions)
		}
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO chunk_embeddings (note_row_id, chunk_index, breadcrumb, heading, content_hash, embedding, dimensions, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(note_row_id, chunk_index) DO UPDATE SET
				breadcrumb = excluded.breadcrumb,
				heading = excluded.heading,
				content_hash = excluded.content_hash,
				embedding = excluded.embedding,
				dimensions = excluded.dimensions,
				created_at = excluded.created_at
		`, rowID, chunk.Index, chunk.Breadcrumb, chunk.Heading, chunk.Hash, embedToBytes(vec), len(vec), now); err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err == nil && s.dimensions == 0 && len(vecs) > 0 {
		s.dimensions = len(vecs[0])
	}
	return err
}

// SearchChunksByVector performs a brute-force cosine similarity search across chunks.
func (s *Store) SearchChunksByVector(ctx context.Context, query embeddings.Embedding, k int) ([]embeddings.SimilarChunk, error) {
	if len(query) == 0 {
		return nil, errors.New("query embedding is empty")
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT n.note_id, n.title, c.chunk_index, c.breadcrumb, c.heading, c.embedding
		FROM chunk_embeddings c
		JOIN notes n ON c.note_row_id = n.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cands []embeddings.SimilarChunk
	for rows.Next() {
		var id, title, breadcrumb, heading string
		var idx int
		var blob []byte
		if err := rows.Scan(&id, &title, &idx, &breadcrumb, &heading, &blob); err != nil {
			return nil, err
		}
		emb := bytesToEmbed(blob)
		if len(emb) != len(query) {
			continue
		}
		score := cosine(query, emb)
		cands = append(cands, embeddings.SimilarChunk{
			NoteID:     embeddings.NoteID(id),
			Title:      title,
			ChunkIndex: idx,
			Breadcrumb: breadcrumb,
			Heading:    heading,
			Score:      score,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	sort.Slice(cands, func(i, j int) bool { return cands[i].Score > cands[j].Score })
	if k > 0 && len(cands) > k {
		cands = cands[:k]
	}
	return cands, nil
}

// SearchNotesByVector aggregates chunk scores to note-level (max score).
func (s *Store) SearchNotesByVector(ctx context.Context, query embeddings.Embedding, k int) ([]embeddings.SimilarNote, error) {
	chunkLimit := k * 3
	if chunkLimit < k {
		chunkLimit = k
	}
	chunks, err := s.SearchChunksByVector(ctx, query, chunkLimit)
	if err != nil {
		return nil, err
	}
	scoreByNote := make(map[embeddings.NoteID]float64)
	titleByNote := make(map[embeddings.NoteID]string)
	for _, c := range chunks {
		if c.Score > scoreByNote[c.NoteID] {
			scoreByNote[c.NoteID] = c.Score
			titleByNote[c.NoteID] = c.Title
		}
	}
	notes := make([]embeddings.SimilarNote, 0, len(scoreByNote))
	for id, score := range scoreByNote {
		notes = append(notes, embeddings.SimilarNote{
			ID:    id,
			Title: titleByNote[id],
			Score: score,
		})
	}
	sort.Slice(notes, func(i, j int) bool { return notes[i].Score > notes[j].Score })
	if k > 0 && len(notes) > k {
		notes = notes[:k]
	}
	return notes, nil
}

// embedToBytes encodes a float32 slice as little-endian bytes.
func embedToBytes(e embeddings.Embedding) []byte {
	b := make([]byte, 4*len(e))
	for i, f := range e {
		binary.LittleEndian.PutUint32(b[i*4:], math.Float32bits(f))
	}
	return b
}

func bytesToEmbed(b []byte) embeddings.Embedding {
	if len(b)%4 != 0 {
		return nil
	}
	n := len(b) / 4
	e := make(embeddings.Embedding, n)
	for i := 0; i < n; i++ {
		e[i] = math.Float32frombits(binary.LittleEndian.Uint32(b[i*4:]))
	}
	return e
}

func cosine(a, b embeddings.Embedding) float64 {
	var dot, na, nb float64
	for i := range a {
		af := float64(a[i])
		bf := float64(b[i])
		dot += af * bf
		na += af * af
		nb += bf * bf
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

var _ embeddings.Index = (*Store)(nil)
