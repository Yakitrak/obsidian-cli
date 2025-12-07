package embeddings

import "fmt"

// NewProvider instantiates a provider based on configuration.
func NewProvider(cfg ProviderConfig) (Provider, error) {
	switch cfg.Provider {
	case "openai", "":
		return NewOpenAIProvider(cfg)
	case "ollama":
		return NewOllamaProvider(cfg)
	case "none":
		return nil, fmt.Errorf("no embeddings provider configured")
	default:
		return nil, fmt.Errorf("unknown embeddings provider %q", cfg.Provider)
	}
}
