package obsidian

import (
	"encoding/json"
	"errors"
)

var JsonMarshal = json.Marshal

func (v *Vault) DefaultName() (string, error) {
	if v.Name != "" {
		return v.Name, nil
	}

	cliConfig, err := readCliConfig(false)
	if err != nil {
		return "", err
	}

	if cliConfig.DefaultVaultName == "" {
		return "", errors.New(ObsidianCLIConfigParseError)
	}

	v.Name = cliConfig.DefaultVaultName
	return cliConfig.DefaultVaultName, nil
}

func (v *Vault) SetDefaultName(name string) error {
	cliConfig, err := readCliConfig(true)
	if err != nil {
		return err
	}

	cliConfig.DefaultVaultName = name
	if err := writeCliConfig(cliConfig); err != nil {
		return err
	}

	v.Name = name

	return nil
}
