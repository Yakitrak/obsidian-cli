package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

var config Config

func GetDefaultVault(vaultName string, obsConfigFilePath string) (string, error) {
	if vaultName != "" {
		return vaultName, nil
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
}

//configDir, err := os.UserConfigDir()
//if err != nil {
//	return "", fmt.Errorf("user config directory not found: %s", err)
//}
//obsConfig := configDir + ObsConfigName
