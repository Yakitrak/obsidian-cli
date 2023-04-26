package obsidian

import (
	"encoding/json"
	"errors"
	"github.com/Yakitrak/obsidian-cli/pkg/config"
	"os"
	"strings"
)

var ObsidianConfigFile = config.ObsidianFile

func (v *Vault) Path() (string, error) {
	obsidianConfigFile, err := ObsidianConfigFile()
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(obsidianConfigFile)

	if err != nil {
		return "", errors.New(ObsidianConfigReadError)
	}

	vaultsContent := ObsidianVaultConfig{}
	err = json.Unmarshal(content, &vaultsContent)

	if err != nil {
		return "", errors.New(ObsidianConfigParseError)
	}

	for _, element := range vaultsContent.Vaults {
		if strings.HasSuffix(element.Path, v.Name) {
			return element.Path, nil
		}
	}

	return "", errors.New(ObsidianConfigVaultNotFoundError)
}
