package utils

import (
	"encoding/json"
	"log"
	"os"
)

var config Config

func GetDefaultVault() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		handleGetVaultError("User config directory not found", err)
	}

	obsConfig := configDir + ObsConfigName
	content, err := os.ReadFile(obsConfig)

	if err != nil {
		handleGetVaultError("Please use set-default command to set default vault or use --vault", err)
	}

	err = json.Unmarshal(content, &config)

	if err != nil {
		handleGetVaultError("Could not retrieve default vault", err)
	}

	if config.DefaultVaultName == "" {
		handleGetVaultError("Could not read value of default vault", err)
	}

	return string(config.DefaultVaultName)
}

func handleGetVaultError(msg string, err error) {
	log.Fatal(msg, err)
}
