package utils_test

import (
	"fmt"
	"github.com/Yakitrak/obsidian-cli/utils"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestFindVaultPathFromName(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}

	defer os.RemoveAll(tmpDir)
	// Create a sample obsidian.json file with test data
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

	obsidianConfigFile := filepath.Join(tmpDir, "/obsidian.json")
	err = os.WriteFile(obsidianConfigFile, []byte(obsidianConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create obsidian.json file: %v", err)
	}

	t.Run("Valid vault name", func(t *testing.T) {
		vaultPath, err := utils.FindVaultPathFromConfig("vault1", obsidianConfigFile)
		assert.Nil(t, err)
		assert.Equal(t, "/path/to/vault1", vaultPath)
	})

	t.Run("Invalid vault name", func(t *testing.T) {
		vaultPath, err := utils.FindVaultPathFromConfig("vault3", obsidianConfigFile)
		assert.Error(t, err, fmt.Sprintf("obsidian vault cannot be found. Please ensure the vault is set up on Obsidian %s", err))
		assert.Equal(t, "", vaultPath)
	})

	t.Run("Obsidian config file not found", func(t *testing.T) {
		vaultPath, err := utils.FindVaultPathFromConfig("vault1", "invalid-path")
		assert.Error(t, err, "obsidian config file cannot be found")
		assert.Equal(t, "", vaultPath)
	})

}
