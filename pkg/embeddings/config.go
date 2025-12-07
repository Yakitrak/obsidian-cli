package embeddings

import (
	"os"
	"path/filepath"
)

const (
	DefaultBatchSize     = 8
	DefaultMaxConcurrent = 4
)

// Config captures user-configurable settings for embeddings.
type Config struct {
	Enabled        bool   `json:"enabled,omitempty"`
	IndexPath      string `json:"indexPath,omitempty"`
	Provider       string `json:"provider,omitempty"`
	Model          string `json:"model,omitempty"`
	Endpoint       string `json:"endpoint,omitempty"`
	Dimensions     int    `json:"dimensions,omitempty"`
	BatchSize      int    `json:"batchSize,omitempty"`
	MaxConcurrency int    `json:"maxConcurrency,omitempty"`
}

// DefaultConfig returns a config populated with sensible defaults for a vault.
func DefaultConfig(vaultPath string) Config {
	return Config{
		Provider:       "openai",
		Model:          defaultOpenAIModel,
		Endpoint:       defaultOpenAIEndpoint,
		Dimensions:     0, // auto-detect from provider when unset
		IndexPath:      DefaultIndexPath(vaultPath),
		BatchSize:      DefaultBatchSize,
		MaxConcurrency: DefaultMaxConcurrent,
	}
}

// Merge applies non-zero/empty fields from override onto base.
func (c Config) Merge(override Config) Config {
	result := c
	if override.IndexPath != "" {
		result.IndexPath = override.IndexPath
	}
	if override.Provider != "" {
		result.Provider = override.Provider
	}
	if override.Model != "" {
		result.Model = override.Model
	}
	if override.Endpoint != "" {
		result.Endpoint = override.Endpoint
	}
	if override.Dimensions > 0 {
		result.Dimensions = override.Dimensions
	}
	if override.BatchSize > 0 {
		result.BatchSize = override.BatchSize
	}
	if override.MaxConcurrency > 0 {
		result.MaxConcurrency = override.MaxConcurrency
	}
	if override.Enabled {
		result.Enabled = true
	}
	return result
}

// ProviderCfg converts to a ProviderConfig, injecting the API key.
func (c Config) ProviderCfg(apiKey string) ProviderConfig {
	return ProviderConfig{
		Provider:       c.Provider,
		Model:          c.Model,
		APIKey:         apiKey,
		Endpoint:       c.Endpoint,
		Dimensions:     c.Dimensions,
		MaxConcurrency: c.MaxConcurrency,
		BatchSize:      c.BatchSize,
	}
}

// DefaultIndexPath returns the default path for the local SQLite index.
func DefaultIndexPath(vaultPath string) string {
	return filepath.Join(vaultPath, ".obsidian-cli", "semantic-index.sqlite")
}

// ResolveAPIKey chooses the first non-empty API key from explicit flag or env.
func ResolveAPIKey(explicit string) string {
	if explicit != "" {
		return explicit
	}
	if val := os.Getenv("OBSIDIAN_CLI_OPENAI_API_KEY"); val != "" {
		return val
	}
	if val := os.Getenv("OPENAI_API_KEY"); val != "" {
		return val
	}
	return ""
}
