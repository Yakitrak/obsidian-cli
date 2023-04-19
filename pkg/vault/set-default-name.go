package vault

import (
	"encoding/json"
	"fmt"
	"os"
)

var JsonMarshal = json.Marshal

func (v *Vault) SetDefaultName(name string) error {
	// marshal vault name to json
	cliConfig := CliConfig{DefaultVaultName: name}
	jsonContent, err := JsonMarshal(cliConfig)
	if err != nil {
		return fmt.Errorf("failed to save default vault to json: %s", err)
	}

	// get cliConfig path
	obsConfigDir, obsConfigFile, err := CliConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get user config directory %s", err)
	}
	// create directory
	err = os.MkdirAll(obsConfigDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create default vault directory %s", err)
	}

	// create and write file
	err = os.WriteFile(obsConfigFile, jsonContent, 0644)
	if err != nil {
		return fmt.Errorf("failed to create default vault configuration file %s", err)
	}

	v.Name = name

	return nil
}
