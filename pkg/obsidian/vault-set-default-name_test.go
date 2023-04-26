package obsidian_test

import (
	"errors"
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

	t.Run("Default vault name successfully set without errors", func(t *testing.T) {
		// Arrange
		mockCliConfigDir, mockCliConfigFile := mocks.CreateMockCliConfigDirectories(t)
		obsidian.CliConfigPath = func() (string, string, error) {
			return mockCliConfigDir, mockCliConfigFile, nil
		}
		vault := obsidian.Vault{}
		// Act
		err := vault.SetDefaultName("vault-name")

		// Assert
		assert.Equal(t, nil, err)
		content, err := os.ReadFile(mockCliConfigFile)
		assert.Equal(t, nil, err)
		assert.Equal(t, `{"default_vault_name":"vault-name"}`, string(content))
	})

	t.Run("Error in config.CliPath", func(t *testing.T) {
		// Arrange
		obsidian.CliConfigPath = func() (string, string, error) {
			return "", "", os.ErrNotExist
		}
		vault := obsidian.Vault{}
		// Act
		err := vault.SetDefaultName("vault-name")
		// Assert
		assert.Equal(t, os.ErrNotExist, err)
	})

	t.Run("Error in json marshal", func(t *testing.T) {
		// Temporarily override the JsonMarshal function
		originalJsonMarshal := obsidian.JsonMarshal
		defer func() { obsidian.JsonMarshal = originalJsonMarshal }()
		obsidian.JsonMarshal = func(v interface{}) ([]byte, error) {
			return nil, errors.New("json marshal error")
		}
		// Arrange
		vault := obsidian.Vault{}
		// Act
		err := vault.SetDefaultName("invalid json")
		// Assert
		assert.Equal(t, err.Error(), obsidian.ObsidianCLIConfigGenerateJSONError)
	})

	t.Run("Error in creating default vault config directory", func(t *testing.T) {
		// Arrange
		obsidian.CliConfigPath = func() (string, string, error) {
			return "", "" + "/preferences.json", nil
		}
		vault := obsidian.Vault{}
		// Act
		err := vault.SetDefaultName("vault-name")
		// Assert
		assert.Equal(t, err.Error(), obsidian.ObsidianCLIConfigDirWriteEror)
	})

	t.Run("Error in writing to default vault config file", func(t *testing.T) {
		// Arrange
		mockCliConfigDir, _ := mocks.CreateMockCliConfigDirectories(t)
		obsidian.CliConfigPath = func() (string, string, error) {
			return mockCliConfigDir + "/unwrittable", mockCliConfigDir + "unwrittable/preferences.json", nil
		}
		err := os.Mkdir(mockCliConfigDir+"/unwrittable", 0444)
		vault := obsidian.Vault{}
		// Act
		err = vault.SetDefaultName("vault-name")
		// Assert
		assert.Equal(t, err.Error(), obsidian.ObsidianCLIConfigWriteError)
	})

}
