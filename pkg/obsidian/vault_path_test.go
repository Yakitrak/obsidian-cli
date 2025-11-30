package obsidian_test

import (
	"github.com/atomicobject/obsidian-cli/mocks"
	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestVaultPath(t *testing.T) {
	// Temporarily override the ObsidianConfigFile function
	originalObsidianConfigFile := obsidian.ObsidianConfigFile
	defer func() { obsidian.ObsidianConfigFile = originalObsidianConfigFile }()

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
	mockObsidianConfigFile := mocks.CreateMockObsidianConfigFile(t)
	obsidian.ObsidianConfigFile = func() (string, error) {
		return mockObsidianConfigFile, nil
	}
	err := os.WriteFile(mockObsidianConfigFile, []byte(obsidianConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create obsidian.json file: %v", err)
	}

	t.Run("Gets vault path successfully from vault name without errors", func(t *testing.T) {
		// Arrange
		vault := obsidian.Vault{Name: "vault1"}
		// Act
		vaultPath, err := vault.Path()
		// Assert
		assert.Equal(t, nil, err)
		assert.Equal(t, "/path/to/vault1", vaultPath)
	})

	t.Run("Error in getting obsidian config file ", func(t *testing.T) {
		// Arrange
		obsidian.ObsidianConfigFile = func() (string, error) {
			return "", os.ErrNotExist
		}
		vault := obsidian.Vault{Name: "vault1"}
		// Act
		_, err := vault.Path()
		// Assert
		assert.Equal(t, os.ErrNotExist, err)
	})

	t.Run("Error in reading obsidian config file", func(t *testing.T) {
		// Arrange
		mockObsidianConfigFile := mocks.CreateMockObsidianConfigFile(t)
		obsidian.ObsidianConfigFile = func() (string, error) {
			return mockObsidianConfigFile, nil
		}
		err := os.WriteFile(mockObsidianConfigFile, []byte(``), 0000)
		if err != nil {
			t.Fatalf("Failed to create obsidian.json file: %v", err)
		}
		vault := obsidian.Vault{Name: "vault1"}
		// Act
		_, err = vault.Path()
		// Assert
		assert.Equal(t, err.Error(), obsidian.ObsidianConfigReadError)

	})

	t.Run("Error in unmarshalling obsidian config file", func(t *testing.T) {
		// Arrange
		obsidian.ObsidianConfigFile = func() (string, error) {
			return mockObsidianConfigFile, nil
		}

		err := os.WriteFile(mockObsidianConfigFile, []byte(`abc`), 0644)
		if err != nil {
			t.Fatalf("Failed to create obsidian.json file: %v", err)
		}
		vault := obsidian.Vault{Name: "vault1"}
		// Act
		_, err = vault.Path()
		// Assert
		assert.Equal(t, err.Error(), obsidian.ObsidianConfigParseError)

	})

	t.Run("No vault found with given name", func(t *testing.T) {
		// Arrange
		obsidian.ObsidianConfigFile = func() (string, error) {
			return mockObsidianConfigFile, nil
		}
		err := os.WriteFile(mockObsidianConfigFile, []byte(`{"vaults":{}}`), 0644)
		vault := obsidian.Vault{Name: "vault3"}
		// Act
		_, err = vault.Path()
		// Assert
		assert.Equal(t, err.Error(), obsidian.ObsidianConfigVaultNotFoundError)
	})
}
