package embeddings

import "context"

// ProviderConfig captures configuration for an embedding provider.
type ProviderConfig struct {
	Provider       string // openai, ollama, none
	Model          string
	APIKey         string
	Endpoint       string
	Dimensions     int
	MaxConcurrency int // optional hint for batching/concurrency
	BatchSize      int // optional hint for batching requests
}

// Provider produces embeddings for one or more text inputs.
type Provider interface {
	EmbedTexts(ctx context.Context, texts []string) ([]Embedding, error)
	Dimensions() int
}
