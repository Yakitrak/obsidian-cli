package obsidian

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Yakitrak/obsidian-cli/pkg/config"
	"os"
	"strings"
)

var ObsidianConfigFile = config.ObsidianFile

func (v *Vault) Path() (string, error) {
	obsidianConfigFile, err := ObsidianConfigFile()
	if err != nil {
		return "", fmt.Errorf("failed to get obsidian config file %s", err)
	}

	content, err := os.ReadFile(obsidianConfigFile)

	if err != nil {
		return "", fmt.Errorf("obsidian config file cannot be found: %s", err)
	}

	vaultsContent := ObsidianVaultConfig{}
	err = json.Unmarshal(content, &vaultsContent)

	if err != nil {
		return "", fmt.Errorf("obsidian config file cannot be parsed: %s", err)
	}

	for _, element := range vaultsContent.Vaults {
		if strings.HasSuffix(element.Path, v.Name) {
			return element.Path, nil
		}
	}

	return "", errors.New("obsidian obsidian cannot be found. Please ensure the obsidian is set up on Obsidian.")
}
