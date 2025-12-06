package obsidian_test

import (
	"encoding/json"
	"github.com/atomicobject/obsidian-cli/mocks"
	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestSavePathToPreferences(t *testing.T) {
	originalCliConfigPath := obsidian.CliConfigPath
	defer func() { obsidian.CliConfigPath = originalCliConfigPath }()

	t.Run("saves new vault path", func(t *testing.T) {
		mockCliConfigDir, mockCliConfigFile := mocks.CreateMockCliConfigDirectories(t)
		obsidian.CliConfigPath = func() (string, string, error) {
			return mockCliConfigDir, mockCliConfigFile, nil
		}

		vaultDir := t.TempDir()
		vault := obsidian.Vault{Name: "vault-name"}
		err := vault.SavePathToPreferences(vaultDir, false)
		assert.NoError(t, err)

		content, err := os.ReadFile(mockCliConfigFile)
		assert.NoError(t, err)

		cfg := obsidian.CliConfig{}
		assert.NoError(t, json.Unmarshal(content, &cfg))
		assert.Equal(t, vaultDir, cfg.Vaults["vault-name"].Path)
	})

	t.Run("prevents overwrite without force", func(t *testing.T) {
		mockCliConfigDir, mockCliConfigFile := mocks.CreateMockCliConfigDirectories(t)
		obsidian.CliConfigPath = func() (string, string, error) {
			return mockCliConfigDir, mockCliConfigFile, nil
		}

		vaultDir := t.TempDir()
		existingDir := filepath.Join(vaultDir, "existing")
		_ = os.Mkdir(existingDir, 0755)
		content, err := json.Marshal(obsidian.CliConfig{
			Vaults: map[string]obsidian.VaultPathEntry{
				"vault-name": {Path: existingDir},
			},
		})
		assert.NoError(t, err)
		err = os.WriteFile(mockCliConfigFile, content, 0644)
		assert.NoError(t, err)

		newDir := filepath.Join(vaultDir, "new")
		_ = os.Mkdir(newDir, 0755)
		vault := obsidian.Vault{Name: "vault-name"}
		err = vault.SavePathToPreferences(newDir, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), obsidian.ObsidianCLIVaultExistsError)
	})

	t.Run("allows overwrite with force", func(t *testing.T) {
		mockCliConfigDir, mockCliConfigFile := mocks.CreateMockCliConfigDirectories(t)
		obsidian.CliConfigPath = func() (string, string, error) {
			return mockCliConfigDir, mockCliConfigFile, nil
		}

		vaultDir := t.TempDir()
		existingDir := filepath.Join(vaultDir, "existing")
		_ = os.Mkdir(existingDir, 0755)
		content, err := json.Marshal(obsidian.CliConfig{
			Vaults: map[string]obsidian.VaultPathEntry{
				"vault-name": {Path: existingDir},
			},
		})
		assert.NoError(t, err)
		err = os.WriteFile(mockCliConfigFile, content, 0644)
		assert.NoError(t, err)

		newDir := filepath.Join(vaultDir, "new")
		_ = os.Mkdir(newDir, 0755)
		vault := obsidian.Vault{Name: "vault-name"}
		err = vault.SavePathToPreferences(newDir, true)
		assert.NoError(t, err)

		content, err = os.ReadFile(mockCliConfigFile)
		assert.NoError(t, err)
		cfg := obsidian.CliConfig{}
		assert.NoError(t, json.Unmarshal(content, &cfg))
		assert.Equal(t, newDir, cfg.Vaults["vault-name"].Path)
	})

	t.Run("rejects invalid path", func(t *testing.T) {
		mockCliConfigDir, mockCliConfigFile := mocks.CreateMockCliConfigDirectories(t)
		obsidian.CliConfigPath = func() (string, string, error) {
			return mockCliConfigDir, mockCliConfigFile, nil
		}

		vault := obsidian.Vault{Name: "vault-name"}
		err := vault.SavePathToPreferences(filepath.Join(t.TempDir(), "missing"), false)
		assert.EqualError(t, err, obsidian.ObsidianCLIVaultPathInvalidError)
	})
}

func TestRemoveFromPreferences(t *testing.T) {
	originalCliConfigPath := obsidian.CliConfigPath
	defer func() { obsidian.CliConfigPath = originalCliConfigPath }()

	t.Run("removes existing vault", func(t *testing.T) {
		mockCliConfigDir, mockCliConfigFile := mocks.CreateMockCliConfigDirectories(t)
		obsidian.CliConfigPath = func() (string, string, error) {
			return mockCliConfigDir, mockCliConfigFile, nil
		}

		vaultDir := t.TempDir()
		content, err := json.Marshal(obsidian.CliConfig{
			Vaults: map[string]obsidian.VaultPathEntry{
				"vault-name": {Path: vaultDir},
			},
		})
		assert.NoError(t, err)
		err = os.WriteFile(mockCliConfigFile, content, 0644)
		assert.NoError(t, err)

		vault := obsidian.Vault{Name: "vault-name"}
		err = vault.RemoveFromPreferences()
		assert.NoError(t, err)

		content, err = os.ReadFile(mockCliConfigFile)
		assert.NoError(t, err)
		cfg := obsidian.CliConfig{}
		assert.NoError(t, json.Unmarshal(content, &cfg))
		_, exists := cfg.Vaults["vault-name"]
		assert.False(t, exists)
	})

	t.Run("errors when vault missing", func(t *testing.T) {
		mockCliConfigDir, mockCliConfigFile := mocks.CreateMockCliConfigDirectories(t)
		obsidian.CliConfigPath = func() (string, string, error) {
			return mockCliConfigDir, mockCliConfigFile, nil
		}

		vault := obsidian.Vault{Name: "vault-name"}
		err := vault.RemoveFromPreferences()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), obsidian.ObsidianCLIVaultNotFoundError)
	})
}
