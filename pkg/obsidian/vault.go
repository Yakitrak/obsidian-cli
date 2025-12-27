package obsidian

// CliConfig represents the obsidian-cli configuration stored in preferences.json.
// It contains the default vault name and optional per-vault settings.
type CliConfig struct {
	DefaultVaultName string                   `json:"default_vault_name"`
	VaultSettings    map[string]VaultSettings `json:"vault_settings,omitempty"`
}

// VaultSettings contains per-vault configuration options.
type VaultSettings struct {
	DailyNote DailyNoteSettings `json:"daily_note,omitempty"`
}

// DailyNoteSettings configures how daily notes are created and managed.
type DailyNoteSettings struct {
	Folder          string `json:"folder,omitempty"`            // Folder path relative to vault root
	FilenamePattern string `json:"filename_pattern,omitempty"`  // Pattern with {YYYY-MM-DD} placeholder
	TemplatePath    string `json:"template_path,omitempty"`     // Optional template file path
	CreateIfMissing bool   `json:"create_if_missing,omitempty"` // Auto-create notes if they don't exist
}

type ObsidianVaultConfig struct {
	Vaults map[string]struct {
		Path string `json:"path"`
	} `json:"vaults"`
}

type VaultManager interface {
	DefaultName() (string, error)
	SetDefaultName(name string) error
	Path() (string, error)
}

type Vault struct {
	Name string
}
