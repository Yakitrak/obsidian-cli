package obsidian

import (
	"encoding/json"
	"errors"
	"github.com/atomicobject/obsidian-cli/pkg/config"
	"os"
	"strings"
)

var ObsidianConfigFile = config.ObsidianFile

func (v *Vault) Path() (string, error) {
	if v.Name == "" {
		return "", errors.New(ObsidianCLIVaultNameRequiredError)
	}

	if path, err := v.pathFromCliConfig(); err != nil {
		return "", err
	} else if path != "" {
		return path, nil
	}

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

func (v *Vault) pathFromCliConfig() (string, error) {
	cfg, err := readCliConfig(true)
	if err != nil {
		return "", err
	}

	entry, ok := cfg.Vaults[v.Name]
	if !ok {
		return "", nil
	}

	if entry.Path == "" {
		return "", errors.New(ObsidianCLIConfigParseError)
	}

	return entry.Path, nil
}

func (v *Vault) SavePathToPreferences(path string, force bool) error {
	return setCliVaultPath(v.Name, path, force)
}

func (v *Vault) RemoveFromPreferences() error {
	return removeCliVaultPath(v.Name)
}

func ListPreferenceVaults() (map[string]VaultPathEntry, string, error) {
	return listCliVaults()
}

func ListObsidianVaults() (map[string]VaultPathEntry, error) {
	obsidianConfigFile, err := ObsidianConfigFile()
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(obsidianConfigFile)
	if err != nil {
		return nil, errors.New(ObsidianConfigReadError)
	}

	vaultsContent := ObsidianVaultConfig{}
	if err := json.Unmarshal(content, &vaultsContent); err != nil {
		return nil, errors.New(ObsidianConfigParseError)
	}

	if vaultsContent.Vaults == nil {
		vaultsContent.Vaults = make(map[string]VaultPathEntry)
	}
	return vaultsContent.Vaults, nil
}
