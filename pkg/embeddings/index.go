package embeddings

import "context"

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
	ChunkHashes(ctx context.Context, id NoteID) (map[int]string, error)
	UpsertNoteChunks(ctx context.Context, id NoteID, chunks []ChunkInput, embeddings []Embedding) error
	DeleteChunksNotIn(ctx context.Context, id NoteID, indices []int) error
	SearchChunksByVector(ctx context.Context, query Embedding, k int) ([]SimilarChunk, error)
	SearchNotesByVector(ctx context.Context, query Embedding, k int) ([]SimilarNote, error)
	Close() error
}
