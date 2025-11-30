package obsidian

type CliConfig struct {
	DefaultVaultName string                    `json:"default_vault_name"`
	Vaults           map[string]VaultPathEntry `json:"vaults,omitempty"`
}

type ObsidianVaultConfig struct {
	Vaults map[string]VaultPathEntry `json:"vaults"`
}

type VaultManager interface {
	DefaultName() (string, error)
	SetDefaultName(name string) error
	Path() (string, error)
}

type Vault struct {
	Name string
}

type VaultPathEntry struct {
	Path string `json:"path"`
}
