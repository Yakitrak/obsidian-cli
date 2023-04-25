package obsidian_test

import (
	"github.com/Yakitrak/obsidian-cli/mocks"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestVaultSetDefaultName(t *testing.T) {
	// Temporarily override the CliConfigPath function
	originalCliConfigPath := obsidian.CliConfigPath
	defer func() { obsidian.CliConfigPath = originalCliConfigPath }()

	t.Run("default obsidian name set without errors", func(t *testing.T) {
		// Arrange
		mockCliConfigDir, mockCliConfigFile := mocks.CreateMockCliConfigDirectories(t)
		obsidian.CliConfigPath = func() (string, string, error) {
			return mockCliConfigDir, mockCliConfigFile, nil
		}
		vault := obsidian.Vault{}
		// Act
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
			vault := obsidian.Vault{}
			// Act
			err := vault.SetDefaultName("obsidian-name")
			// Assert
			assert.ErrorContains(t, err, "failed to get user config directory")
		})

		t.Run("Error in json marshal", func(t *testing.T) {
			// Temporarily override the JsonMarshal function
			originalJsonMarshal := obsidian.JsonMarshal
			defer func() { obsidian.JsonMarshal = originalJsonMarshal }()
			obsidian.JsonMarshal = func(v interface{}) ([]byte, error) {
				return nil, os.ErrNotExist
			}
			// Arrange
			vault := obsidian.Vault{}
			// Act
			err := vault.SetDefaultName("invalid json")
			// Assert
			assert.ErrorContains(t, err, "failed to save default obsidian to json")
		})

		t.Run("Error in creating default obsidian config directory", func(t *testing.T) {
			// Arrange
			obsidian.CliConfigPath = func() (string, string, error) {
				return "", "" + "/preferences.json", nil
			}
			vault := obsidian.Vault{}
			// Act
			err := vault.SetDefaultName("obsidian-name")
			// Assert
			assert.ErrorContains(t, err, "failed to create default obsidian directory")
		})

		t.Run("Error in creating and writing to default obsidian config file", func(t *testing.T) {
			// Arrange
			mockCliConfigDir, _ := mocks.CreateMockCliConfigDirectories(t)
			obsidian.CliConfigPath = func() (string, string, error) {
				return mockCliConfigDir + "/unwrittable", mockCliConfigDir + "unwrittable/preferences.json", nil
			}
			err := os.Mkdir(mockCliConfigDir+"/unwrittable", 0444)
			vault := obsidian.Vault{}
			// Act
			err = vault.SetDefaultName("obsidian-name")
			// Assert
			assert.ErrorContains(t, err, "failed to create default obsidian configuration file")
		})
	})

}
