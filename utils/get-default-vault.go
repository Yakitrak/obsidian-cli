package utils

import (
	"encoding/json"
	"log"
	"os"
)

var config Config

func GetDefaultVault(vaultName string) string {
	if vaultName != "" {
		return vaultName
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal("User config directory not found", err)
	}

	obsConfig := configDir + ObsConfigName
	content, err := os.ReadFile(obsConfig)

	if err != nil {
		log.Fatal("Please use set-default command to set default vault or use --vault", err)
	}

	err = json.Unmarshal(content, &config)

	if err != nil {
		log.Fatal("Could not retrieve default vault", err)
	}

	if config.DefaultVaultName == "" {
		log.Fatal("Could not read value of default vault", err)
	}

	return string(config.DefaultVaultName)
}
