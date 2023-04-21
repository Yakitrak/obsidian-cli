package obsidian

type CliConfig struct {
	DefaultVaultName string `json:"default_vault_name"`
}

type ObsidianVaultConfig struct {
	Vaults map[string]struct {
		Path string `json:"path"`
	} `json:"vaults"`
}

type VaultOperator interface {
	DefaultName() (string, error)
	SetDefaultName(name string) error
	Path() (string, error)
}

type Vault struct {
	Name string
}
