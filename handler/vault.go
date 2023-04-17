package handler

import (
	"encoding/json"
	"fmt"
	"github.com/Yakitrak/obsidian-cli/utils"
	"os"
)

type Vault struct {
	Name string
}

type Config struct {
	DefaultVaultName string `json:"default_vault_name"`
}

var CliConfigPath = utils.CliConfigPath
var JsonMarshal = json.Marshal

func (v *Vault) SetDefaultName(name string) error {
	// marshal vault name to json
	config := Config{DefaultVaultName: name}
	jsonContent, err := JsonMarshal(config)
	if err != nil {
		return fmt.Errorf("failed to save default vault to json: %s", err)
	}

	// get config path
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

func (v *Vault) DefaultName() (string, error) {
	config := Config{}

	if v.Name != "" {
		return v.Name, nil
	}

	// get config paths
	_, cliConfigFile, err := CliConfigPath()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory %s", err)
	}

	// read config
	content, err := os.ReadFile(cliConfigFile)
	if err != nil {
		return "", fmt.Errorf("cannot find vault config. please use set-default command to set default vault or use --vault: %s", err)
	}

	// retrieve value
	err = json.Unmarshal(content, &config)

	if err != nil {
		return "", fmt.Errorf("could not retrieve default vault %s", err)
	}

	if config.DefaultVaultName == "" {
		return "", fmt.Errorf("could not read value of default vault %s", err)
	}

	return config.DefaultVaultName, nil
}
