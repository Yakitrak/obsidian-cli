package obsidian

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// VaultGraphConfig holds per-vault graph preferences.
type VaultGraphConfig struct {
	GraphIgnore []string `json:"graphIgnore,omitempty"`
}

// LoadVaultGraphConfig loads .obsidian-cli/config.json inside the vault if present.
// Missing files return an empty config with nil error.
func LoadVaultGraphConfig(vaultPath string) (VaultGraphConfig, error) {
	path := vaultGraphConfigPath(vaultPath)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return VaultGraphConfig{}, nil
		}
		return VaultGraphConfig{}, err
	}
	var cfg VaultGraphConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return VaultGraphConfig{}, err
	}
	return cfg, nil
}

// SaveVaultGraphConfig writes .obsidian-cli/config.json inside the vault.
func SaveVaultGraphConfig(vaultPath string, cfg VaultGraphConfig) error {
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

func vaultGraphConfigPath(vaultPath string) string {
	return filepath.Join(vaultPath, ".obsidian-cli", "config.json")
}
