package obsidian

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/atomicobject/obsidian-cli/pkg/embeddings"
)

// VaultConfig holds per-vault agent preferences (graph, semantic search, etc.).
type VaultConfig struct {
	GraphIgnore     []string           `json:"graphIgnore,omitempty"`
	KeyNotePatterns []string           `json:"keyNotePatterns,omitempty"`
	Embeddings      *embeddings.Config `json:"embeddings,omitempty"`
}

// VaultGraphConfig remains for backwards compatibility.
type VaultGraphConfig = VaultConfig

// LoadVaultConfig loads .obsidian-cli/config.json inside the vault if present.
// Missing files return an empty config with nil error.
func LoadVaultConfig(vaultPath string) (VaultConfig, error) {
	path := vaultGraphConfigPath(vaultPath)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return VaultConfig{}, nil
		}
		return VaultConfig{}, err
	}
	var cfg VaultConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return VaultConfig{}, err
	}
	return cfg, nil
}

// LoadVaultGraphConfig is retained for callers expecting the old name.
func LoadVaultGraphConfig(vaultPath string) (VaultGraphConfig, error) {
	return LoadVaultConfig(vaultPath)
}

// SaveVaultConfig writes .obsidian-cli/config.json inside the vault.
func SaveVaultConfig(vaultPath string, cfg VaultConfig) error {
	dir := filepath.Join(vaultPath, ".obsidian-cli")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := vaultGraphConfigPath(vaultPath)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// SaveVaultGraphConfig is retained for callers expecting the old name.
func SaveVaultGraphConfig(vaultPath string, cfg VaultGraphConfig) error {
	return SaveVaultConfig(vaultPath, cfg)
}

func vaultGraphConfigPath(vaultPath string) string {
	return filepath.Join(vaultPath, ".obsidian-cli", "config.json")
}
