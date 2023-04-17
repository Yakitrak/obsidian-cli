package handler

import (
	"encoding/json"
	"fmt"
	"os"
)

type Vault struct {
	name string
}

type Config struct {
	DefaultVaultName string `json:"default_vault_name"` // TODO make lower case?
}

var config Config

func (v *Vault) SetDefaultName(name string) error {
	return nil
}

func (v *Vault) DefaultName(obsConfigFilePath string) (string, error) {
	if v.name != "" {
		return v.name, nil
	}
	content, err := os.ReadFile(obsConfigFilePath)

	if err != nil {
		return "", fmt.Errorf("cannot find vault config. please use set-default command to set default vault or use --vault: %s", err)
	}

	err = json.Unmarshal(content, &config)

	if err != nil {
		return "", fmt.Errorf("could not retrieve default vault %s", err)
	}

	if config.DefaultVaultName == "" {
		return "", fmt.Errorf("could not read value of default vault %s", err)
	}

	return string(config.DefaultVaultName), nil
	return "", nil
}

func (v *Vault) Path() (string, error) {
	return "", nil
}
