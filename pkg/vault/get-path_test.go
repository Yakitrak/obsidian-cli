package vault_test

import (
	"github.com/Yakitrak/obsidian-cli/pkg/vault"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func createMockObsidianConfigFile(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %s", err)
	}
	return tmpDir + "/obsidian.json"
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
