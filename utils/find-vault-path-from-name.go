package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

var vaultsContent ObsidianVaultConfig

type Vault struct {
	Path string `json:"path"`
	Ts   int64  `json:"ts"`
	Open bool   `json:"open"`
}

type Vaults map[string]Vault

type ObsidianVaultConfig struct {
	Vaults `json:"vaults"`
}

func FindVaultPathFromName(vaultName string) (string, error) {

	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("User config directory not found %g", err)
	}

	obsidianVaultsFile := configDir + "/obsidian/obsidian.json"

	content, err := os.ReadFile(obsidianVaultsFile)

	if err != nil {
		return "", fmt.Errorf("Obsidian config cannot be read %g", err)
	}

	err = json.Unmarshal(content, &vaultsContent)

	for _, element := range vaultsContent.Vaults {
		if strings.Contains(element.Path, vaultName) {
			return element.Path, nil
		}
	}

	return "", fmt.Errorf("Obsidian vault cannot be found. Please ensure the vault is set up on Obsidian %g", err)
}
