package handler_test

import (
	"github.com/Yakitrak/obsidian-cli/handler"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func createTmpTestDirectories(t *testing.T) (string, string) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %s", err)
	}
	return tmpDir, tmpDir + "/preferences.json"
}

func TestVaultSetDefaultName(t *testing.T) {
	// set the config function
	originalCliConfigPath := handler.CliConfigPath
	defer func() { handler.CliConfigPath = originalCliConfigPath }()

	t.Run("default vault name set without errors", func(t *testing.T) {
		// Arrange
		mockCliConfigDir, mockCliConfigFile := createTmpTestDirectories(t)
		handler.CliConfigPath = func() (string, string, error) {
			return mockCliConfigDir, mockCliConfigFile, nil
		}
		// Act
		vaultHandler := handler.Vault{}
		err := vaultHandler.SetDefaultName("vault-name")

		// Assert
		assert.Equal(t, nil, err)
		content, err := os.ReadFile(mockCliConfigFile)
		assert.Equal(t, nil, err)
		assert.Equal(t, `{"default_vault_name":"vault-name"}`, string(content))
	})

	t.Run("default vault name not set due to error", func(t *testing.T) {
		t.Run("Error in CliConfigPath", func(t *testing.T) {
			// Arrange
			handler.CliConfigPath = func() (string, string, error) {
				return "", "", os.ErrNotExist
			}
			// Act
			vaultHandler := handler.Vault{}
			err := vaultHandler.SetDefaultName("vault-name")
			// Assert
			assert.ErrorContains(t, err, "failed to get user config directory")
		})

		t.Run("Error in json marshal", func(t *testing.T) {
			// Arrange
			originalJsonMarshal := handler.JsonMarshal
			defer func() { handler.JsonMarshal = originalJsonMarshal }()
			handler.JsonMarshal = func(v interface{}) ([]byte, error) {
				return nil, os.ErrNotExist
			}
			// Act
			vaultHandler := handler.Vault{}
			err := vaultHandler.SetDefaultName("invalid json")
			// Assert
			assert.ErrorContains(t, err, "failed to save default vault to json")
		})

		t.Run("Error in creating default vault config directory", func(t *testing.T) {
			// Arrange
			handler.CliConfigPath = func() (string, string, error) {
				return "", "" + "/preferences.json", nil
			}
			// Act
			vaultHandler := handler.Vault{}
			err := vaultHandler.SetDefaultName("vault-name")
			// Assert
			assert.ErrorContains(t, err, "failed to create default vault directory")
		})

		t.Run("Error in creating and writing to default vault config file", func(t *testing.T) {
			// Arrange
			mockCliConfigDir, _ := createTmpTestDirectories(t)
			handler.CliConfigPath = func() (string, string, error) {
				return mockCliConfigDir + "/unwrittable", mockCliConfigDir + "unwrittable/preferences.json", nil
			}
			err := os.Mkdir(mockCliConfigDir+"/unwrittable", 0444)
			// Act
			vaultHandler := handler.Vault{}
			err = vaultHandler.SetDefaultName("vault-name")
			// Assert
			assert.ErrorContains(t, err, "failed to create default vault configuration file")
		})
	})

}

func TestVaultDefaultName(t *testing.T) {
	// set the config function
	originalCliConfigPath := handler.CliConfigPath
	defer func() { handler.CliConfigPath = originalCliConfigPath }()
	t.Run("Get vault name without errors", func(t *testing.T) {
		t.Run("Get vault name from struct", func(t *testing.T) {
			// Act
			vaultHandler := handler.Vault{Name: "my-vault"}
			vaultName, err := vaultHandler.DefaultName()
			// Assert
			assert.Equal(t, nil, err)
			assert.Equal(t, "my-vault", vaultName)
		})

		t.Run("Get vault name from file", func(t *testing.T) {
			// Arrange
			mockCliConfigDir, mockCliConfigFile := createTmpTestDirectories(t)
			handler.CliConfigPath = func() (string, string, error) {
				return mockCliConfigDir, mockCliConfigFile, nil
			}
			err := os.WriteFile(mockCliConfigFile, []byte(`{"default_vault_name":"example-vault"}`), 0644)
			// Act
			vaultHandler := handler.Vault{}
			vaultName, err := vaultHandler.DefaultName()
			// Assert
			assert.Equal(t, nil, err)
			assert.Equal(t, "example-vault", vaultName)
		})
	})

	t.Run("Could not get vault name", func(t *testing.T) {

		t.Run("Error in CliConfigPath", func(t *testing.T) {
			// Arrange
			handler.CliConfigPath = func() (string, string, error) {
				return "", "", os.ErrNotExist
			}
			// Act
			vaultHandler := handler.Vault{}
			_, err := vaultHandler.DefaultName()
			// Assert
			assert.ErrorContains(t, err, "failed to get user config directory")
		})

		t.Run("Error in reading default vault config file", func(t *testing.T) {
			// Arrange
			mockCliConfigDir, mockCliConfigFile := createTmpTestDirectories(t)
			handler.CliConfigPath = func() (string, string, error) {
				return mockCliConfigDir, mockCliConfigFile, nil
			}
			// Act
			vaultHandler := handler.Vault{}
			_, err := vaultHandler.DefaultName()
			// Assert
			assert.ErrorContains(t, err, "cannot find vault config. please use set-default command to set default vault or use --vault")
		})

		t.Run("Error in unmarshalling default vault config file", func(t *testing.T) {
			// Arrange
			mockCliConfigDir, mockCliConfigFile := createTmpTestDirectories(t)
			handler.CliConfigPath = func() (string, string, error) {
				return mockCliConfigDir, mockCliConfigFile, nil
			}
			err := os.WriteFile(mockCliConfigFile, []byte(`{"default_vault_name""example-vault`), 0644)
			// Act
			vaultHandler := handler.Vault{}
			_, err = vaultHandler.DefaultName()
			// Assert
			assert.ErrorContains(t, err, "could not retrieve default vault")
		})

		t.Run("Error DefaultVaultName empty", func(t *testing.T) {
			// Arrange
			mockCliConfigDir, mockCliConfigFile := createTmpTestDirectories(t)
			handler.CliConfigPath = func() (string, string, error) {
				return mockCliConfigDir, mockCliConfigFile, nil
			}
			err := os.WriteFile(mockCliConfigFile, []byte(`{"default_vault_name":""}`), 0644)
			// Act
			vaultHandler := handler.Vault{}
			_, err = vaultHandler.DefaultName()
			// Assert
			assert.ErrorContains(t, err, "could not read value of default vault")
		})
	})

}
