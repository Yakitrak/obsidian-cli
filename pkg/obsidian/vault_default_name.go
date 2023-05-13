package obsidian

import (
	"encoding/json"
	"errors"
	"github.com/Yakitrak/obsidian-cli/pkg/config"
	"os"
)

var CliConfigPath = config.CliPath
var JsonMarshal = json.Marshal

func (v *Vault) DefaultName() (string, error) {
	if v.Name != "" {
		return v.Name, nil
	}

	// get cliConfig path
	_, cliConfigFile, err := CliConfigPath()
	if err != nil {
		return "", err
	}

	// read file
	content, err := os.ReadFile(cliConfigFile)
	if err != nil {
		return "", errors.New(ObsidianCLIConfigReadError)
	}

	// unmarshal json
	cliConfig := CliConfig{}
	err = json.Unmarshal(content, &cliConfig)

	if err != nil {
		return "", errors.New(ObsidianCLIConfigParseError)
	}

	if cliConfig.DefaultVaultName == "" {
		return "", errors.New(ObsidianCLIConfigParseError)
	}

	v.Name = cliConfig.DefaultVaultName
	return cliConfig.DefaultVaultName, nil
}

func (v *Vault) SetDefaultName(name string) error {
	// marshal obsidian name to json
	cliConfig := CliConfig{DefaultVaultName: name}
	jsonContent, err := JsonMarshal(cliConfig)
	if err != nil {
		return errors.New(ObsidianCLIConfigGenerateJSONError)
	}

	// get cliConfig path
	obsConfigDir, obsConfigFile, err := CliConfigPath()
	if err != nil {
		return err
	}

	// create directory
	err = os.MkdirAll(obsConfigDir, os.ModePerm)
	if err != nil {
		return errors.New(ObsidianCLIConfigDirWriteEror)
	}

	// create and write file
	err = os.WriteFile(obsConfigFile, jsonContent, 0644)
	if err != nil {
		return errors.New(ObsidianCLIConfigWriteError)
	}

	v.Name = name

	return nil
}
