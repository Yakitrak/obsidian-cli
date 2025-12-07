package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	defaultOpenAIModel    = "text-embedding-3-large"
	defaultOpenAIEndpoint = "https://api.openai.com/v1/embeddings"
)

type openAIProvider struct {
	model      string
	apiKey     string
	endpoint   string
	dims       int
	httpClient *http.Client
}

// NewOpenAIProvider constructs an OpenAI embeddings provider.
func NewOpenAIProvider(cfg ProviderConfig) (Provider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("openai provider requires api key")
	}
	model := cfg.Model
	if model == "" {
		model = defaultOpenAIModel
	}
	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = defaultOpenAIEndpoint
	}
	dims := cfg.Dimensions

	return &openAIProvider{
		model:      model,
		apiKey:     cfg.APIKey,
		endpoint:   endpoint,
		dims:       dims,
		httpClient: http.DefaultClient,
	}, nil
}

func (p *openAIProvider) Dimensions() int {
	return p.dims
}

func (p *openAIProvider) EmbedTexts(ctx context.Context, texts []string) ([]Embedding, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts to embed")
	}

	payload := map[string]any{
		"model": p.model,
		"input": texts,
	}
	if p.dims > 0 {
		payload["dimensions"] = p.dims
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
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("openai embeddings status %d: %s", resp.StatusCode, string(msg))
	}

	var parsed struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	if len(parsed.Data) != len(texts) {
		return nil, fmt.Errorf("embedding count mismatch: want %d got %d", len(texts), len(parsed.Data))
	}

	res := make([]Embedding, len(parsed.Data))
	for i, item := range parsed.Data {
		res[i] = Embedding(item.Embedding)
		if p.dims == 0 {
			p.dims = len(item.Embedding)
		}
	}
	return res, nil
}
