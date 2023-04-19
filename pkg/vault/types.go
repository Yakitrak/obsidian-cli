package vault

type CliConfig struct {
	DefaultVaultName string `json:"default_vault_name"`
}

type ObsidianVaultConfig struct {
	Vaults map[string]struct {
		Path string `json:"path"`
	} `json:"vaults"`
}

type Vault struct {
	Name string
}
