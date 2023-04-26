package obsidian_test

import (
	"github.com/Yakitrak/obsidian-cli/mocks"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestVaultDefaultName(t *testing.T) {
	// Temporarily override the CliConfigPath function
	originalCliConfigPath := obsidian.CliConfigPath
	defer func() { obsidian.CliConfigPath = originalCliConfigPath }()
	t.Run("Get vault name successfully without errors", func(t *testing.T) {
		t.Run("Get vault name from struct", func(t *testing.T) {
			// Arrange
			vault := obsidian.Vault{Name: "my-vault"}
			// Act
			vaultName, err := vault.DefaultName()
			// Assert
			assert.Equal(t, nil, err)
			assert.Equal(t, "my-vault", vaultName)
		})

		t.Run("Get vault name from config file", func(t *testing.T) {
			// Arrange
			mockCliConfigDir, mockCliConfigFile := mocks.CreateMockCliConfigDirectories(t)
			obsidian.CliConfigPath = func() (string, string, error) {
				return mockCliConfigDir, mockCliConfigFile, nil
			}
			err := os.WriteFile(mockCliConfigFile, []byte(`{"default_vault_name":"example-obsidian"}`), 0644)
			vault := obsidian.Vault{}
			// Act
			vaultName, err := vault.DefaultName()
			// Assert
			assert.Equal(t, nil, err)
			assert.Equal(t, "example-obsidian", vaultName)
		})
	})

	t.Run("Could not get vault name", func(t *testing.T) {

		t.Run("Error in config.CliPath", func(t *testing.T) {
			// Arrange
			obsidian.CliConfigPath = func() (string, string, error) {
				return "", "", os.ErrNotExist
			}
			vault := obsidian.Vault{}
			// Act
			_, err := vault.DefaultName()
			// Assert
			assert.Equal(t, os.ErrNotExist, err)
		})

		t.Run("Error in reading default vault config file", func(t *testing.T) {
			// Arrange
			mockCliConfigDir, mockCliConfigFile := mocks.CreateMockCliConfigDirectories(t)
			obsidian.CliConfigPath = func() (string, string, error) {
				return mockCliConfigDir, mockCliConfigFile, nil
			}
			vault := obsidian.Vault{}
			// Act
			_, err := vault.DefaultName()
			// Assert
			assert.Equal(t, err.Error(), obsidian.ObsidianCLIConfigReadError)
		})

		t.Run("Error in unmarshalling default vault config file", func(t *testing.T) {
			// Arrange
			mockCliConfigDir, mockCliConfigFile := mocks.CreateMockCliConfigDirectories(t)
			obsidian.CliConfigPath = func() (string, string, error) {
				return mockCliConfigDir, mockCliConfigFile, nil
			}
			err := os.WriteFile(mockCliConfigFile, []byte(`{"default_vault_name""example-obsidian`), 0644)
			vault := obsidian.Vault{}
			// Act
			_, err = vault.DefaultName()
			// Assert
			assert.Equal(t, err.Error(), obsidian.ObsidianCLIConfigParseError)
		})

		t.Run("Error DefaultVaultName empty", func(t *testing.T) {
			// Arrange
			mockCliConfigDir, mockCliConfigFile := mocks.CreateMockCliConfigDirectories(t)
			obsidian.CliConfigPath = func() (string, string, error) {
				return mockCliConfigDir, mockCliConfigFile, nil
			}
			err := os.WriteFile(mockCliConfigFile, []byte(`{"default_vault_name":""}`), 0644)
			vault := obsidian.Vault{}
			// Act
			_, err = vault.DefaultName()
			// Assert
			assert.Equal(t, err.Error(), obsidian.ObsidianCLIConfigParseError)
		})
	})
}
