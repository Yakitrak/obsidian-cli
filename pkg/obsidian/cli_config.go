package obsidian

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/atomicobject/obsidian-cli/pkg/config"
	"os"
	"path/filepath"
)

var CliConfigPath = config.CliPath

// readCliConfig loads obsidian-cli preferences. If allowMissing is true, a missing file
// returns an empty config and nil error.
func readCliConfig(allowMissing bool) (CliConfig, error) {
	_, cliConfigFile, err := CliConfigPath()
	if err != nil {
		return CliConfig{}, err
	}

	content, err := os.ReadFile(cliConfigFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && allowMissing {
			return CliConfig{}, nil
		}
		return CliConfig{}, errors.New(ObsidianCLIConfigReadError)
	}

	cfg := CliConfig{}
	if err := json.Unmarshal(content, &cfg); err != nil {
		return CliConfig{}, errors.New(ObsidianCLIConfigParseError)
	}

	return cfg, nil
}

func writeCliConfig(cfg CliConfig) error {
	obsConfigDir, obsConfigFile, err := CliConfigPath()
	if err != nil {
		return err
	}

	if obsConfigDir == "" {
		return errors.New(ObsidianCLIConfigDirWriteEror)
	}

	if err := os.MkdirAll(obsConfigDir, os.ModePerm); err != nil {
		return errors.New(ObsidianCLIConfigDirWriteEror)
	}

	jsonContent, err := JsonMarshal(cfg)
	if err != nil {
		return errors.New(ObsidianCLIConfigGenerateJSONError)
	}

	if err := os.WriteFile(obsConfigFile, jsonContent, 0644); err != nil {
		return errors.New(ObsidianCLIConfigWriteError)
	}

	return nil
}

func validateVaultPath(path string) error {
	if path == "" {
		return errors.New(ObsidianCLIVaultPathInvalidError)
	}
	cleaned := filepath.Clean(path)
	info, err := os.Stat(cleaned)
	if err != nil || !info.IsDir() {
		return errors.New(ObsidianCLIVaultPathInvalidError)
	}
	return nil
}

// setCliVaultPath stores/overwrites a vault path in the CLI preferences.
func setCliVaultPath(name string, path string, force bool) error {
	if name == "" {
		return errors.New(ObsidianCLIVaultNameRequiredError)
	}
	if err := validateVaultPath(path); err != nil {
		return err
	}

	cfg, err := readCliConfig(true)
	if err != nil {
		return err
	}

	if cfg.Vaults == nil {
		cfg.Vaults = make(map[string]VaultPathEntry)
	}

	if existing, ok := cfg.Vaults[name]; ok && existing.Path != path && !force {
		return fmt.Errorf("%s: %s", ObsidianCLIVaultExistsError, name)
	}

	cfg.Vaults[name] = VaultPathEntry{Path: filepath.Clean(path)}
	return writeCliConfig(cfg)
}

// removeCliVaultPath removes a vault entry from CLI preferences.
func removeCliVaultPath(name string) error {
	if name == "" {
		return errors.New(ObsidianCLIVaultNameRequiredError)
	}

	cfg, err := readCliConfig(true)
	if err != nil {
		return err
	}

	if len(cfg.Vaults) == 0 {
		return fmt.Errorf("%s: %s", ObsidianCLIVaultNotFoundError, name)
	}

	if _, ok := cfg.Vaults[name]; !ok {
		return fmt.Errorf("%s: %s", ObsidianCLIVaultNotFoundError, name)
	}

	delete(cfg.Vaults, name)
	return writeCliConfig(cfg)
}

// listCliVaults returns manual vault mappings stored in preferences.
func listCliVaults() (map[string]VaultPathEntry, string, error) {
	cfg, err := readCliConfig(true)
	if err != nil {
		return nil, "", err
	}

	vaults := cfg.Vaults
	if vaults == nil {
		vaults = make(map[string]VaultPathEntry)
	}

	return vaults, cfg.DefaultVaultName, nil
}
