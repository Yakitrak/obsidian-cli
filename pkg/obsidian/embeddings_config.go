package obsidian

import "github.com/atomicobject/obsidian-cli/pkg/embeddings"

// LoadEmbeddingsConfig returns embeddings config merged with defaults for the vault.
func LoadEmbeddingsConfig(vaultPath string) (embeddings.Config, error) {
	cfg, err := LoadVaultConfig(vaultPath)
	if err != nil {
		return embeddings.Config{}, err
	}
	base := embeddings.DefaultConfig(vaultPath)
	if cfg.Embeddings != nil {
		base = base.Merge(*cfg.Embeddings)
		base.Enabled = cfg.Embeddings.Enabled
	}
	return base, nil
}

// SaveEmbeddingsConfig updates the embeddings section in the vault config.
func SaveEmbeddingsConfig(vaultPath string, embCfg embeddings.Config) error {
	cfg, err := LoadVaultConfig(vaultPath)
	if err != nil {
		return err
	}
	cfg.Embeddings = &embCfg
	return SaveVaultConfig(vaultPath, cfg)
}
