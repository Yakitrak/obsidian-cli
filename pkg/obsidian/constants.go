package obsidian

const (
	ExecuteUriError                    = "Failed to execute Obsidian URI"
	NoteDoesNotExistError              = "Cannot find note in vault"
	VaultAccessError                   = "Failed to access vault directory"
	VaultReadError                     = "Failed to read notes in vault"
	VaultWriteError                    = "Failed to write to update notes in vault"
	ObsidianCLIConfigReadError         = "Cannot find vault config, please use set-default command to set default vault or use --vault flag"
	ObsidianCLIConfigParseError        = "Could not parse vault config file, please use set-default command to set default vault or use --vault flag"
	ObsidianCLIConfigDirWriteEror      = "Failed to create vault config directory. Please ensure you have the correct permissions."
	ObsidianCLIConfigGenerateJSONError = "Failed to generate vault config file. Please ensure vault name does not contain any special characters."
	ObsidianCLIConfigWriteError        = "Failed to write vault config file. Please ensure you have correct permissions."
	ObsidianConfigReadError            = "Failed to read Obsidian config file. Please ensure vault has been set up in Obsidian."
	ObsidianConfigParseError           = "Failed to parse Obsidian config file. Please ensure vault has been set up in Obsidian."
	ObsidianConfigVaultNotFoundError   = "Vault not found in Obsidian config file. Please ensure vault has been set up in Obsidian."
)
