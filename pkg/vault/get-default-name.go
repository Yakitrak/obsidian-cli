package vault

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Yakitrak/obsidian-cli/pkg/config"
	"os"
)

var CliConfigPath = config.CliPath

func (v *Vault) DefaultName() (string, error) {
	if v.Name != "" {
		return v.Name, nil
	}

	// get cliConfig paths
	_, cliConfigFile, err := CliConfigPath()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory %s", err)
	}

	// read cliConfig
	content, err := os.ReadFile(cliConfigFile)
	if err != nil {
		return "", fmt.Errorf("cannot find vault config. please use set-default command to set default vault or use --vault: %s", err)
	}

	// retrieve value
	cliConfig := CliConfig{}
	err = json.Unmarshal(content, &cliConfig)

	if err != nil {
		return "", fmt.Errorf("could not retrieve default vault %s", err)
	}

	if cliConfig.DefaultVaultName == "" {
		return "", errors.New("could not read value of default vault %s")
	}

	v.Name = cliConfig.DefaultVaultName
	return cliConfig.DefaultVaultName, nil
}
