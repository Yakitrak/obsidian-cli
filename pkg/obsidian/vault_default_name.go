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
	cliConfig, err := readCliConfig()
	if err != nil {
		return err
	}
	cliConfig.DefaultVaultName = name

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
	err = os.MkdirAll(obsConfigDir, 0750)
	if err != nil {
		return errors.New(ObsidianCLIConfigDirWriteEror)
	}

	// create and write file
	err = os.WriteFile(obsConfigFile, jsonContent, 0600)
	if err != nil {
		return errors.New(ObsidianCLIConfigWriteError)
	}

	v.Name = name

	return nil
}

// Settings returns the VaultSettings for the current default vault.
// Returns empty settings if none are configured (not an error).
func (v *Vault) Settings() (VaultSettings, error) {
	name, err := v.DefaultName()
	if err != nil {
		return VaultSettings{}, err
	}

	cfg, err := readCliConfig()
	if err != nil {
		return VaultSettings{}, err
	}

	if cfg.VaultSettings == nil {
		return VaultSettings{}, nil
	}
	settings, ok := cfg.VaultSettings[name]
	if !ok {
		return VaultSettings{}, nil
	}
	return settings, nil
}

// SetSettings saves the VaultSettings for the current default vault.
func (v *Vault) SetSettings(settings VaultSettings) error {
	name, err := v.DefaultName()
	if err != nil {
		return err
	}

	cfg, err := readCliConfig()
	if err != nil {
		return err
	}
	if cfg.VaultSettings == nil {
		cfg.VaultSettings = map[string]VaultSettings{}
	}
	cfg.VaultSettings[name] = settings

	jsonContent, err := JsonMarshal(cfg)
	if err != nil {
		return errors.New(ObsidianCLIConfigGenerateJSONError)
	}

	obsConfigDir, obsConfigFile, err := CliConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(obsConfigDir, 0750); err != nil {
		return errors.New(ObsidianCLIConfigDirWriteEror)
	}
	if err := os.WriteFile(obsConfigFile, jsonContent, 0600); err != nil {
		return errors.New(ObsidianCLIConfigWriteError)
	}
	return nil
}

// readCliConfig reads the CLI configuration from the preferences file.
// Returns an empty config (not an error) if the file doesn't exist yet.
func readCliConfig() (CliConfig, error) {
	_, cliConfigFile, err := CliConfigPath()
	if err != nil {
		return CliConfig{}, err
	}
	content, err := os.ReadFile(cliConfigFile)
	if err != nil {
		return CliConfig{}, nil
	}
	cliConfig := CliConfig{}
	if err := json.Unmarshal(content, &cliConfig); err != nil {
		return CliConfig{}, errors.New(ObsidianCLIConfigParseError)
	}
	return cliConfig, nil
}
