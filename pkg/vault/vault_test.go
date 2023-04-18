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

func createMockObsidianConfigFile(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %s", err)
	}
	return tmpDir + "/obsidian.json"
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

func TestVaultPath(t *testing.T) {
	originalObsidianConfigFile := vault.ObsidianConfigFile
	defer func() { vault.ObsidianConfigFile = originalObsidianConfigFile }()

	obsidianConfig := `{
		"vaults": {
			"random1": {
				"path": "/path/to/vault1"
			},
			"random2": {
				"path": "/path/to/vault2"
			}
		}
	}`
	mockObsidianConfigFile := createMockObsidianConfigFile(t)
	vault.ObsidianConfigFile = func() (string, error) {
		return mockObsidianConfigFile, nil
	}
	err := os.WriteFile(mockObsidianConfigFile, []byte(obsidianConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create obsidian.json file: %v", err)
	}

	t.Run("Get vault path from valid vault name without errors", func(t *testing.T) {
		// Act
		vault := vault.Vault{Name: "vault1"}
		vaultPath, err := vault.Path()
		// Assert
		assert.Equal(t, nil, err)
		assert.Equal(t, "/path/to/vault1", vaultPath)
	})

	t.Run("Error in getting obsidian config file ", func(t *testing.T) {
		// Arrange
		vault.ObsidianConfigFile = func() (string, error) {
			return "", os.ErrNotExist
		}
		// Act
		vault := vault.Vault{Name: "vault1"}
		_, err := vault.Path()
		// Assert
		assert.ErrorContains(t, err, "failed to get obsidian config file")
	})

	t.Run("Error in reading obsidian config file", func(t *testing.T) {
		// Arrange
		mockObsidianConfigFile := createMockObsidianConfigFile(t)
		vault.ObsidianConfigFile = func() (string, error) {
			return mockObsidianConfigFile, nil
		}
		err := os.WriteFile(mockObsidianConfigFile, []byte(``), 0000)
		if err != nil {
			t.Fatalf("Failed to create obsidian.json file: %v", err)
		}
		// Act
		vault := vault.Vault{Name: "vault1"}
		_, err = vault.Path()
		// Assert
		assert.ErrorContains(t, err, "obsidian config file cannot be found")

	})

	t.Run("Error in unmarshalling obsidian config file", func(t *testing.T) {
		// Arrange
		vault.ObsidianConfigFile = func() (string, error) {
			return mockObsidianConfigFile, nil
		}

		err := os.WriteFile(mockObsidianConfigFile, []byte(`abc`), 0644)
		if err != nil {
			t.Fatalf("Failed to create obsidian.json file: %v", err)
		}
		// Act
		vault := vault.Vault{Name: "vault1"}
		_, err = vault.Path()
		// Assert
		assert.ErrorContains(t, err, "obsidian config file cannot be parsed")

	})

	t.Run("No vault found with given name", func(t *testing.T) {
		// Arrange
		vault.ObsidianConfigFile = func() (string, error) {
			return mockObsidianConfigFile, nil
		}

		err := os.WriteFile(mockObsidianConfigFile, []byte(`{"vaults":{}}`), 0644)

		// Act
		vault := vault.Vault{Name: "vault3"}
		_, err = vault.Path()
		// Assert
		assert.ErrorContains(t, err, "obsidian vault cannot be found. Please ensure the vault is set up on Obsidian")

	})
}

func TestVaultUpdateNoteLinks(t *testing.T) {
	// TODO
}
