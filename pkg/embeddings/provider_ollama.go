package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const defaultOllamaEndpoint = "http://localhost:11434/api/embeddings"

type ollamaProvider struct {
	model      string
	endpoint   string
	dims       int
	httpClient *http.Client
}

// NewOllamaProvider constructs an Ollama embeddings provider.
func NewOllamaProvider(cfg ProviderConfig) (Provider, error) {
	if cfg.Model == "" {
		return nil, fmt.Errorf("ollama provider requires model")
	}
	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = defaultOllamaEndpoint
	}
	return &ollamaProvider{
		model:      cfg.Model,
		endpoint:   endpoint,
		dims:       cfg.Dimensions,
		httpClient: http.DefaultClient,
	}, nil
}

func (p *ollamaProvider) Dimensions() int {
	return p.dims
}

func (p *ollamaProvider) EmbedTexts(ctx context.Context, texts []string) ([]Embedding, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts to embed")
	}
	results := make([]Embedding, 0, len(texts))
	for _, t := range texts {
		emb, err := p.embedOne(ctx, t)
		if err != nil {
			return nil, err
		}
		results = append(results, emb)
	}
	return results, nil
}

func (p *ollamaProvider) embedOne(ctx context.Context, text string) (Embedding, error) {
	payload := map[string]any{
		"model":  p.model,
		"prompt": text,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("ollama embeddings status %d: %s", resp.StatusCode, string(msg))
	}

	var parsed struct {
		Embedding []float32 `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	if p.dims == 0 {
		p.dims = len(parsed.Embedding)
	}
	return Embedding(parsed.Embedding), nil
}
