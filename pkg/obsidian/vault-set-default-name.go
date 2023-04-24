package obsidian

import (
	"encoding/json"
	"fmt"
	"os"
)

var JsonMarshal = json.Marshal

func (v *Vault) SetDefaultName(name string) error {
	// marshal obsidian name to json
	cliConfig := CliConfig{DefaultVaultName: name}
	jsonContent, err := JsonMarshal(cliConfig)
	if err != nil {
		return fmt.Errorf("failed to save default obsidian to json: %s", err)
	}

	// get cliConfig path
	obsConfigDir, obsConfigFile, err := CliConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get user config directory %s", err)
	}
	// create directory
	err = os.MkdirAll(obsConfigDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create default obsidian directory %s", err)
	}

	// create and write file
	err = os.WriteFile(obsConfigFile, jsonContent, 0644)
	if err != nil {
		return fmt.Errorf("failed to create default obsidian configuration file %s", err)
	}

	v.Name = name

	return nil
}
