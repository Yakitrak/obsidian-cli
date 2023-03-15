package utils

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	DefaultVaultName string `json:"default_vault_name"`
}

func SetDefaultVault(name string) {
	jsonContent, err := json.Marshal(Config{DefaultVaultName: name})

	if err != nil {
		log.Fatal(err)
		handlSetVaultError("Vault name has unexpected character", err)
	}

	// Get default user config dir
	dirname, err := os.UserConfigDir()
	if err != nil {
		handlSetVaultError("User config directory not found", err)
		log.Fatal(err)
	}

	// create config dir
	err = os.MkdirAll(dirname+ObsDir, os.ModePerm)
	if err != nil {
		handlSetVaultError("Failed to save default vault to configuration", err)
	}

	// create file
	obsConfig := dirname + ObsConfigName
	f, err := os.Create(obsConfig)

	if err != nil {
		handlSetVaultError("Failed to save default vault to configuration", err)
	}

	// write file
	_, err = f.WriteString(string(jsonContent))
	if err != nil {
		handlSetVaultError("Failed to write default vault to configuration", err)

	}
}

func handlSetVaultError(msg string, err error) {
	log.Fatal(msg, err)
}
