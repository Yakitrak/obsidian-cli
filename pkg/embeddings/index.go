package embeddings

import (
	"context"
	"time"
)

// SimilarNote captures similarity results prior to graph re-ranking.
type SimilarNote struct {
	ID         NoteID
	Title      string
	Score      float64
	GraphScore float64
	FinalScore float64
}

// Index abstracts persistence and vector search for note embeddings.
type Index interface {
	EnsureSchema(ctx context.Context) error
	UpsertNoteMeta(ctx context.Context, info NoteFileInfo) error
	DeleteNotesNotIn(ctx context.Context, existingIDs []NoteID) error
	ListNotes(ctx context.Context) ([]NoteFileInfo, error)
	GetNoteEmbedding(ctx context.Context, id NoteID) (Embedding, string, bool, error)
	UpsertNoteEmbedding(ctx context.Context, id NoteID, contentHash string, emb Embedding) error
	DeleteNote(ctx context.Context, id NoteID) error
	ChunkHashes(ctx context.Context, id NoteID) (map[int]string, error)
	UpsertNoteChunks(ctx context.Context, id NoteID, chunks []ChunkInput, embeddings []Embedding) error
	DeleteChunksNotIn(ctx context.Context, id NoteID, indices []int) error
	SearchChunksByVector(ctx context.Context, query Embedding, k int) ([]SimilarChunk, int, error)
	SearchNotesByVector(ctx context.Context, query Embedding, k int) ([]SimilarNote, int, error)
	Metadata(ctx context.Context) (IndexMetadata, bool, error)
	ValidateOrInitMetadata(ctx context.Context, meta IndexMetadata) error
	UpdateLastSync(ctx context.Context, ts time.Time) error
	Stats(ctx context.Context) (notes int, chunks int, err error)
	Close() error
}
