package vault_test

import (
	"github.com/Yakitrak/obsidian-cli/pkg/vault"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestVaultDefaultName(t *testing.T) {
	// set the config function
	originalCliConfigPath := vault.CliConfigPath
	defer func() { vault.CliConfigPath = originalCliConfigPath }()
	t.Run("Get vault name without errors", func(t *testing.T) {
		t.Run("Get vault name from struct", func(t *testing.T) {
			// Act
			vault := vault.Vault{Name: "my-vault"}
			vaultName, err := vault.DefaultName()
			// Assert
			assert.Equal(t, nil, err)
			assert.Equal(t, "my-vault", vaultName)
		})

		t.Run("Get vault name from file", func(t *testing.T) {
			// Arrange
			mockCliConfigDir, mockCliConfigFile := createMockCliConfigDirectories(t)
			vault.CliConfigPath = func() (string, string, error) {
				return mockCliConfigDir, mockCliConfigFile, nil
			}
			err := os.WriteFile(mockCliConfigFile, []byte(`{"default_vault_name":"example-vault"}`), 0644)
			// Act
			vault := vault.Vault{}
			vaultName, err := vault.DefaultName()
			// Assert
			assert.Equal(t, nil, err)
			assert.Equal(t, "example-vault", vaultName)
		})
	})

	t.Run("Could not get vault name", func(t *testing.T) {

		t.Run("Error in CliConfigPath", func(t *testing.T) {
			// Arrange
			vault.CliConfigPath = func() (string, string, error) {
				return "", "", os.ErrNotExist
			}
			// Act
			vault := vault.Vault{}
			_, err := vault.DefaultName()
			// Assert
			assert.ErrorContains(t, err, "failed to get user config directory")
		})

		t.Run("Error in reading default vault config file", func(t *testing.T) {
			// Arrange
			mockCliConfigDir, mockCliConfigFile := createMockCliConfigDirectories(t)
			vault.CliConfigPath = func() (string, string, error) {
				return mockCliConfigDir, mockCliConfigFile, nil
			}
			// Act
			vault := vault.Vault{}
			_, err := vault.DefaultName()
			// Assert
			assert.ErrorContains(t, err, "cannot find vault config. please use set-default command to set default vault or use --vault")
		})

		t.Run("Error in unmarshalling default vault config file", func(t *testing.T) {
			// Arrange
			mockCliConfigDir, mockCliConfigFile := createMockCliConfigDirectories(t)
			vault.CliConfigPath = func() (string, string, error) {
				return mockCliConfigDir, mockCliConfigFile, nil
			}
			err := os.WriteFile(mockCliConfigFile, []byte(`{"default_vault_name""example-vault`), 0644)
			// Act
			vault := vault.Vault{}
			_, err = vault.DefaultName()
			// Assert
			assert.ErrorContains(t, err, "could not retrieve default vault")
		})

		t.Run("Error DefaultVaultName empty", func(t *testing.T) {
			// Arrange
			mockCliConfigDir, mockCliConfigFile := createMockCliConfigDirectories(t)
			vault.CliConfigPath = func() (string, string, error) {
				return mockCliConfigDir, mockCliConfigFile, nil
			}
			err := os.WriteFile(mockCliConfigFile, []byte(`{"default_vault_name":""}`), 0644)
			// Act
			vault := vault.Vault{}
			_, err = vault.DefaultName()
			// Assert
			assert.ErrorContains(t, err, "could not read value of default vault")
		})
	})
}
