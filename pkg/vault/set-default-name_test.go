package vault_test

import (
	"github.com/Yakitrak/obsidian-cli/pkg/vault"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func createMockCliConfigDirectories(t *testing.T) (string, string) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %s", err)
	}
	return tmpDir, tmpDir + "/preferences.json"
}

func TestVaultSetDefaultName(t *testing.T) {
	// set the config function
	originalCliConfigPath := vault.CliConfigPath
	defer func() { vault.CliConfigPath = originalCliConfigPath }()

	t.Run("default vault name set without errors", func(t *testing.T) {
		// Arrange
		mockCliConfigDir, mockCliConfigFile := createMockCliConfigDirectories(t)
		vault.CliConfigPath = func() (string, string, error) {
			return mockCliConfigDir, mockCliConfigFile, nil
		}
		// Act
		vault := vault.Vault{}
		err := vault.SetDefaultName("vault-name")

		// Assert
		assert.Equal(t, nil, err)
		content, err := os.ReadFile(mockCliConfigFile)
		assert.Equal(t, nil, err)
		assert.Equal(t, `{"default_vault_name":"vault-name"}`, string(content))
	})

	t.Run("default vault name not set due to error", func(t *testing.T) {
		t.Run("Error in CliConfigPath", func(t *testing.T) {
			// Arrange
			vault.CliConfigPath = func() (string, string, error) {
				return "", "", os.ErrNotExist
			}
			// Act
			vault := vault.Vault{}
			err := vault.SetDefaultName("vault-name")
			// Assert
			assert.ErrorContains(t, err, "failed to get user config directory")
		})

		t.Run("Error in json marshal", func(t *testing.T) {
			// Arrange
			originalJsonMarshal := vault.JsonMarshal
			defer func() { vault.JsonMarshal = originalJsonMarshal }()
			vault.JsonMarshal = func(v interface{}) ([]byte, error) {
				return nil, os.ErrNotExist
			}
			// Act
			vault := vault.Vault{}
			err := vault.SetDefaultName("invalid json")
			// Assert
			assert.ErrorContains(t, err, "failed to save default vault to json")
		})

		t.Run("Error in creating default vault config directory", func(t *testing.T) {
			// Arrange
			vault.CliConfigPath = func() (string, string, error) {
				return "", "" + "/preferences.json", nil
			}
			// Act
			vault := vault.Vault{}
			err := vault.SetDefaultName("vault-name")
			// Assert
			assert.ErrorContains(t, err, "failed to create default vault directory")
		})

		t.Run("Error in creating and writing to default vault config file", func(t *testing.T) {
			// Arrange
			mockCliConfigDir, _ := createMockCliConfigDirectories(t)
			vault.CliConfigPath = func() (string, string, error) {
				return mockCliConfigDir + "/unwrittable", mockCliConfigDir + "unwrittable/preferences.json", nil
			}
			err := os.Mkdir(mockCliConfigDir+"/unwrittable", 0444)
			// Act
			vault := vault.Vault{}
			err = vault.SetDefaultName("vault-name")
			// Assert
			assert.ErrorContains(t, err, "failed to create default vault configuration file")
		})
	})

}
