package obsidian

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Yakitrak/notesmd-cli/pkg/config"
)

// AddVault registers a vault path in the Obsidian config file.
// It creates the config file and directory if they don't exist.
// Returns the resolved absolute path on success.
func AddVault(vaultPath string) (string, error) {
	absPath, err := filepath.Abs(vaultPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("path does not exist: %s", absPath)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("path is not a directory: %s", absPath)
	}

	obsidianConfigFile, vaultsConfig, err := readOrCreateObsidianConfig()
	if err != nil {
		return "", err
	}

	// Check if vault already registered
	for _, v := range vaultsConfig.Vaults {
		if filepath.Clean(v.Path) == filepath.Clean(absPath) {
			return "", fmt.Errorf("vault already registered: %s", absPath)
		}
	}

	id, err := generateVaultID()
	if err != nil {
		return "", fmt.Errorf("failed to generate vault ID: %w", err)
	}

	vaultsConfig.Vaults[id] = struct {
		Path string `json:"path"`
	}{Path: absPath}

	return absPath, writeObsidianConfig(obsidianConfigFile, vaultsConfig)
}

// RemoveVault removes a vault from the Obsidian config file by name or path.
// Returns the resolved vault name (directory basename) so callers can use it
// for follow-up operations like clearing the default.
func RemoveVault(input string) (string, error) {
	obsidianConfigFile, err := ObsidianConfigFile()
	if err != nil {
		return "", errors.New(ObsidianConfigReadError)
	}

	content, err := os.ReadFile(obsidianConfigFile)
	if err != nil {
		return "", errors.New(ObsidianConfigReadError)
	}

	vaultsConfig := vaultRegistryConfig{}
	if json.Unmarshal(content, &vaultsConfig) != nil {
		return "", errors.New(ObsidianConfigParseError)
	}

	// Exact path match (only when input looks like a path, not a bare name)
	if filepath.IsAbs(input) || strings.Contains(input, string(filepath.Separator)) || strings.HasPrefix(input, ".") {
		absInput, _ := filepath.Abs(input)
		for id, v := range vaultsConfig.Vaults {
			if filepath.Clean(v.Path) == filepath.Clean(absInput) {
				name := filepath.Base(v.Path)
				delete(vaultsConfig.Vaults, id)
				return name, writeObsidianConfig(obsidianConfigFile, vaultsConfig)
			}
		}
	}

	// Name match -- collect all matches to detect ambiguity
	type match struct {
		id   string
		path string
	}
	var matches []match
	for id, v := range vaultsConfig.Vaults {
		if filepath.Base(v.Path) == input {
			matches = append(matches, match{id: id, path: v.Path})
		}
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("vault %q not found", input)
	}
	if len(matches) > 1 {
		var paths []string
		for _, m := range matches {
			paths = append(paths, fmt.Sprintf("  %s", m.path))
		}
		return "", fmt.Errorf(
			"multiple vaults named %q found. Use the full path to disambiguate:\n%s",
			input, strings.Join(paths, "\n"),
		)
	}

	delete(vaultsConfig.Vaults, matches[0].id)
	return input, writeObsidianConfig(obsidianConfigFile, vaultsConfig)
}

// ClearDefaultIfMatch clears the default vault in CLI config if it matches the given name.
func ClearDefaultIfMatch(name string) error {
	_, cliConfigFile, err := CliConfigPath()
	if err != nil {
		return nil //nolint:nilerr // no config dir means nothing to clear
	}

	content, err := os.ReadFile(cliConfigFile)
	if err != nil {
		return nil //nolint:nilerr // no config file means nothing to clear
	}

	cliConfig := CliConfig{}
	if err := json.Unmarshal(content, &cliConfig); err != nil {
		return nil //nolint:nilerr // unparseable config means nothing to clear
	}

	if cliConfig.DefaultVaultName != name {
		return nil
	}

	v := &Vault{}
	return v.SetDefaultName("")
}

func readOrCreateObsidianConfig() (string, vaultRegistryConfig, error) {
	empty := vaultRegistryConfig{ObsidianVaultConfig: ObsidianVaultConfig{
		Vaults: make(map[string]struct {
			Path string `json:"path"`
		}),
	}}

	// Try to find existing config
	obsidianConfigFile, err := ObsidianConfigFile()
	if err == nil {
		content, readErr := os.ReadFile(obsidianConfigFile)
		if readErr == nil {
			vaultsConfig := vaultRegistryConfig{}
			if err := json.Unmarshal(content, &vaultsConfig); err != nil {
				return "", empty, fmt.Errorf("corrupt obsidian config at %s: %w", obsidianConfigFile, err)
			}
			if vaultsConfig.Vaults == nil {
				vaultsConfig.Vaults = make(map[string]struct {
					Path string `json:"path"`
				})
			}
			return obsidianConfigFile, vaultsConfig, nil
		}
	}

	// Config doesn't exist, create it
	userConfigDir, err := config.UserConfigDirectory()
	if err != nil {
		return "", empty, fmt.Errorf("failed to determine config directory: %w", err)
	}

	configDir := filepath.Join(userConfigDir, config.ObsidianConfigDirectory)
	if err := os.MkdirAll(configDir, os.ModePerm); err != nil {
		return "", empty, fmt.Errorf("failed to create config directory: %w", err)
	}

	configFile := filepath.Join(configDir, config.ObsidianConfigFile)
	return configFile, empty, nil
}

func writeObsidianConfig(path string, cfg vaultRegistryConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func generateVaultID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
