package obsidian_test

import (
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestVaultDefaultName(t *testing.T) {
	// set the config function
	originalCliConfigPath := obsidian.CliConfigPath
	defer func() { obsidian.CliConfigPath = originalCliConfigPath }()
	t.Run("Get obsidian name without errors", func(t *testing.T) {
		t.Run("Get obsidian name from struct", func(t *testing.T) {
			// Act
			vault := obsidian.Vault{Name: "my-obsidian"}
			vaultName, err := vault.DefaultName()
			// Assert
			assert.Equal(t, nil, err)
			assert.Equal(t, "my-obsidian", vaultName)
		})

		t.Run("Get obsidian name from file", func(t *testing.T) {
			// Arrange
			mockCliConfigDir, mockCliConfigFile := createMockCliConfigDirectories(t)
			obsidian.CliConfigPath = func() (string, string, error) {
				return mockCliConfigDir, mockCliConfigFile, nil
			}
			err := os.WriteFile(mockCliConfigFile, []byte(`{"default_vault_name":"example-obsidian"}`), 0644)
			// Act
			vault := obsidian.Vault{}
			vaultName, err := vault.DefaultName()
			// Assert
			assert.Equal(t, nil, err)
			assert.Equal(t, "example-obsidian", vaultName)
		})
	})

	t.Run("Could not get obsidian name", func(t *testing.T) {

		t.Run("Error in CliConfigPath", func(t *testing.T) {
			// Arrange
			obsidian.CliConfigPath = func() (string, string, error) {
				return "", "", os.ErrNotExist
			}
			// Act
			vault := obsidian.Vault{}
			_, err := vault.DefaultName()
			// Assert
			assert.ErrorContains(t, err, "failed to get user config directory")
		})

		t.Run("Error in reading default obsidian config file", func(t *testing.T) {
			// Arrange
			mockCliConfigDir, mockCliConfigFile := createMockCliConfigDirectories(t)
			obsidian.CliConfigPath = func() (string, string, error) {
				return mockCliConfigDir, mockCliConfigFile, nil
			}
			// Act
			vault := obsidian.Vault{}
			_, err := vault.DefaultName()
			// Assert
			assert.ErrorContains(t, err, "cannot find obsidian config. please use set-default command to set default obsidian or use --obsidian")
		})

		t.Run("Error in unmarshalling default obsidian config file", func(t *testing.T) {
			// Arrange
			mockCliConfigDir, mockCliConfigFile := createMockCliConfigDirectories(t)
			obsidian.CliConfigPath = func() (string, string, error) {
				return mockCliConfigDir, mockCliConfigFile, nil
			}
			err := os.WriteFile(mockCliConfigFile, []byte(`{"default_vault_name""example-obsidian`), 0644)
			// Act
			vault := obsidian.Vault{}
			_, err = vault.DefaultName()
			// Assert
			assert.ErrorContains(t, err, "could not retrieve default obsidian")
		})

		t.Run("Error DefaultVaultName empty", func(t *testing.T) {
			// Arrange
			mockCliConfigDir, mockCliConfigFile := createMockCliConfigDirectories(t)
			obsidian.CliConfigPath = func() (string, string, error) {
				return mockCliConfigDir, mockCliConfigFile, nil
			}
			err := os.WriteFile(mockCliConfigFile, []byte(`{"default_vault_name":""}`), 0644)
			// Act
			vault := obsidian.Vault{}
			_, err = vault.DefaultName()
			// Assert
			assert.ErrorContains(t, err, "could not read value of default obsidian")
		})
	})
}