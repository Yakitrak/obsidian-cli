package obsidian_test

import (
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
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
	originalCliConfigPath := obsidian.CliConfigPath
	defer func() { obsidian.CliConfigPath = originalCliConfigPath }()

	t.Run("default obsidian name set without errors", func(t *testing.T) {
		// Arrange
		mockCliConfigDir, mockCliConfigFile := createMockCliConfigDirectories(t)
		obsidian.CliConfigPath = func() (string, string, error) {
			return mockCliConfigDir, mockCliConfigFile, nil
		}
		// Act
		vault := obsidian.Vault{}
		err := vault.SetDefaultName("obsidian-name")

		// Assert
		assert.Equal(t, nil, err)
		content, err := os.ReadFile(mockCliConfigFile)
		assert.Equal(t, nil, err)
		assert.Equal(t, `{"default_vault_name":"obsidian-name"}`, string(content))
	})

	t.Run("default obsidian name not set due to error", func(t *testing.T) {
		t.Run("Error in CliConfigPath", func(t *testing.T) {
			// Arrange
			obsidian.CliConfigPath = func() (string, string, error) {
				return "", "", os.ErrNotExist
			}
			// Act
			vault := obsidian.Vault{}
			err := vault.SetDefaultName("obsidian-name")
			// Assert
			assert.ErrorContains(t, err, "failed to get user config directory")
		})

		t.Run("Error in json marshal", func(t *testing.T) {
			// Arrange
			originalJsonMarshal := obsidian.JsonMarshal
			defer func() { obsidian.JsonMarshal = originalJsonMarshal }()
			obsidian.JsonMarshal = func(v interface{}) ([]byte, error) {
				return nil, os.ErrNotExist
			}
			// Act
			vault := obsidian.Vault{}
			err := vault.SetDefaultName("invalid json")
			// Assert
			assert.ErrorContains(t, err, "failed to save default obsidian to json")
		})

		t.Run("Error in creating default obsidian config directory", func(t *testing.T) {
			// Arrange
			obsidian.CliConfigPath = func() (string, string, error) {
				return "", "" + "/preferences.json", nil
			}
			// Act
			vault := obsidian.Vault{}
			err := vault.SetDefaultName("obsidian-name")
			// Assert
			assert.ErrorContains(t, err, "failed to create default obsidian directory")
		})

		t.Run("Error in creating and writing to default obsidian config file", func(t *testing.T) {
			// Arrange
			mockCliConfigDir, _ := createMockCliConfigDirectories(t)
			obsidian.CliConfigPath = func() (string, string, error) {
				return mockCliConfigDir + "/unwrittable", mockCliConfigDir + "unwrittable/preferences.json", nil
			}
			err := os.Mkdir(mockCliConfigDir+"/unwrittable", 0444)
			// Act
			vault := obsidian.Vault{}
			err = vault.SetDefaultName("obsidian-name")
			// Assert
			assert.ErrorContains(t, err, "failed to create default obsidian configuration file")
		})
	})

}
