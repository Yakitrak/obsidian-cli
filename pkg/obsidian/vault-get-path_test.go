package obsidian_test

import (
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
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
	mockObsidianConfigFile := createMockObsidianConfigFile(t)
	obsidian.ObsidianConfigFile = func() (string, error) {
		return mockObsidianConfigFile, nil
	}
	err := os.WriteFile(mockObsidianConfigFile, []byte(obsidianConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create obsidian.json file: %v", err)
	}

	t.Run("Get obsidian path from valid obsidian name without errors", func(t *testing.T) {
		// Act
		vault := obsidian.Vault{Name: "vault1"}
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
		// Act
		vault := obsidian.Vault{Name: "vault1"}
		_, err := vault.Path()
		// Assert
		assert.ErrorContains(t, err, "failed to get obsidian config file")
	})

	t.Run("Error in reading obsidian config file", func(t *testing.T) {
		// Arrange
		mockObsidianConfigFile := createMockObsidianConfigFile(t)
		obsidian.ObsidianConfigFile = func() (string, error) {
			return mockObsidianConfigFile, nil
		}
		err := os.WriteFile(mockObsidianConfigFile, []byte(``), 0000)
		if err != nil {
			t.Fatalf("Failed to create obsidian.json file: %v", err)
		}
		// Act
		vault := obsidian.Vault{Name: "vault1"}
		_, err = vault.Path()
		// Assert
		assert.ErrorContains(t, err, "obsidian config file cannot be found")

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
		// Act
		vault := obsidian.Vault{Name: "vault1"}
		_, err = vault.Path()
		// Assert
		assert.ErrorContains(t, err, "obsidian config file cannot be parsed")

	})

	t.Run("No obsidian found with given name", func(t *testing.T) {
		// Arrange
		obsidian.ObsidianConfigFile = func() (string, error) {
			return mockObsidianConfigFile, nil
		}

		err := os.WriteFile(mockObsidianConfigFile, []byte(`{"vaults":{}}`), 0644)

		// Act
		vault := obsidian.Vault{Name: "vault3"}
		_, err = vault.Path()
		// Assert
		assert.ErrorContains(t, err, "obsidian obsidian cannot be found. Please ensure the obsidian is set up on Obsidian")

	})
}
