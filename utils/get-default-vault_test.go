package utils_test

import (
	"github.com/Yakitrak/obsidian-cli/utils"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestGetDefaultVault(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}

	defer os.RemoveAll(tmpDir)
	// Create a sample obsidian.json file with test data
	obsConfigContent := `{
		"default_vault_name": "test-vault"
	}`

	obsConfigFile := filepath.Join(tmpDir, "/preferences.json")
	err = os.WriteFile(obsConfigFile, []byte(obsConfigContent), 0644)

	// Tests
	t.Run("With vault name", func(t *testing.T) {
		defaultVault, err := utils.GetDefaultVault("named-vault", obsConfigFile)
		assert.Equal(t, "named-vault", defaultVault)
		assert.Nil(t, err)
	})

	t.Run("Without vault name arg", func(t *testing.T) {
		t.Run("valid config file", func(t *testing.T) {
			defaultVault, err := utils.GetDefaultVault("", obsConfigFile)
			assert.Equal(t, "test-vault", defaultVault)
			assert.Nil(t, err)
		})

		t.Run("invalid config file path", func(t *testing.T) {
			defaultVault, err := utils.GetDefaultVault("", "invalid-file")
			assert.Equal(t, "", defaultVault)
			assert.ErrorContains(t, err, "cannot find vault config. please use set-default command")
		})

		t.Run("invalid config file content", func(t *testing.T) {
			obsConfigFile := filepath.Join(tmpDir, "/preferences.json")
			err = os.WriteFile(obsConfigFile, []byte(""), 0644)
			defaultVault, err := utils.GetDefaultVault("", obsConfigFile)
			assert.Equal(t, "", defaultVault)
			assert.ErrorContains(t, err, "could not retrieve default vault")
		})

		t.Run("invalid .default_vault_path property", func(t *testing.T) {
			obsConfigContent := `{
				"default_vault_name": ""
			}`
			obsConfigFile := filepath.Join(tmpDir, "/preferences.json")
			err = os.WriteFile(obsConfigFile, []byte(obsConfigContent), 0644)
			defaultVault, err := utils.GetDefaultVault("", obsConfigFile)
			assert.Equal(t, "", defaultVault)
			assert.ErrorContains(t, err, "could not read value of default vault")
		})
	})
}
