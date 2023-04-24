package obsidian

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
		return "", fmt.Errorf("cannot find obsidian config. please use set-default command to set default obsidian or use --obsidian: %s", err)
	}

	// retrieve value
	cliConfig := CliConfig{}
	err = json.Unmarshal(content, &cliConfig)

	if err != nil {
		return "", fmt.Errorf("could not retrieve default obsidian %s", err)
	}

	if cliConfig.DefaultVaultName == "" {
		return "", errors.New("could not read value of default obsidian %s")
	}

	v.Name = cliConfig.DefaultVaultName
	return cliConfig.DefaultVaultName, nil
}
