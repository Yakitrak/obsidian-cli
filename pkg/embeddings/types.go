package embeddings

import "time"

// NoteID uniquely identifies a note, typically by vault-relative path.
type NoteID string

// NoteFileInfo captures filesystem metadata used to detect changes.
type NoteFileInfo struct {
	ID    NoteID
	Path  string
	Title string
	Size  int64
	Mtime time.Time
}

// Embedding is a dense vector representation of text.
type Embedding []float32

// ChunkInput is the text plus metadata to embed.
type ChunkInput struct {
	Index      int
	Text       string
	Breadcrumb string
	Heading    string
	Hash       string
}

// SimilarChunk captures chunk-level similarity results.
type SimilarChunk struct {
	NoteID     NoteID
	Title      string
	ChunkIndex int
	Breadcrumb string
	Heading    string
	Score      float64
	GraphScore float64
	FinalScore float64
}
