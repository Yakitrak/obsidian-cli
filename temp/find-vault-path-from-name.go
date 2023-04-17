package temp

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type ObsidianVaultConfig struct {
	Vaults map[string]struct {
		Path string `json:"path"`
	} `json:"vaults"`
}

var vaultsContent ObsidianVaultConfig

func FindVaultPathFromConfig(vaultName string, configFilePath string) (string, error) {

	content, err := os.ReadFile(configFilePath)

	if err != nil {
		return "", fmt.Errorf("obsidian config file cannot be found: %s", err)
	}

	err = json.Unmarshal(content, &vaultsContent)

	for _, element := range vaultsContent.Vaults {
		if strings.HasSuffix(element.Path, "/"+vaultName) {
			return element.Path, nil
		}
	}

	return "", fmt.Errorf("obsidian vault cannot be found. Please ensure the vault is set up on Obsidian %s", err)
}
