package obsidian_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Yakitrak/obsidian-cli/mocks"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/stretchr/testify/assert"
)

func TestSetDefaultNamePreservesVaultSettings(t *testing.T) {
	mockCliConfigDir, mockCliConfigFile := mocks.CreateMockCliConfigDirectories(t)
	originalCliConfigPath := obsidian.CliConfigPath
	t.Cleanup(func() { obsidian.CliConfigPath = originalCliConfigPath })
	obsidian.CliConfigPath = func() (string, string, error) {
		return mockCliConfigDir, mockCliConfigFile, nil
	}

	initial := obsidian.CliConfig{
		DefaultVaultName: "Old Vault",
		VaultSettings: map[string]obsidian.VaultSettings{
			"Old Vault": {
				DailyNote: obsidian.DailyNoteSettings{
					Folder:          "Daily",
					FilenamePattern: "{YYYY-MM-DD}",
					TemplatePath:    "Templates/Daily.md",
					CreateIfMissing: true,
				},
			},
		},
	}
	b, err := json.Marshal(initial)
	assert.NoError(t, err)
	assert.NoError(t, os.MkdirAll(filepath.Dir(mockCliConfigFile), os.ModePerm))
	assert.NoError(t, os.WriteFile(mockCliConfigFile, b, 0644))

	vault := obsidian.Vault{}
	assert.NoError(t, vault.SetDefaultName("New Vault"))

	updatedBytes, err := os.ReadFile(mockCliConfigFile)
	assert.NoError(t, err)
	var updated obsidian.CliConfig
	assert.NoError(t, json.Unmarshal(updatedBytes, &updated))
	assert.Equal(t, "New Vault", updated.DefaultVaultName)
	assert.NotNil(t, updated.VaultSettings)
	assert.Contains(t, updated.VaultSettings, "Old Vault")
	assert.Equal(t, initial.VaultSettings["Old Vault"].DailyNote.Folder, updated.VaultSettings["Old Vault"].DailyNote.Folder)
}

func TestSetSettingsCreatesVaultSettingsEntry(t *testing.T) {
	mockCliConfigDir, mockCliConfigFile := mocks.CreateMockCliConfigDirectories(t)
	originalCliConfigPath := obsidian.CliConfigPath
	t.Cleanup(func() { obsidian.CliConfigPath = originalCliConfigPath })
	obsidian.CliConfigPath = func() (string, string, error) {
		return mockCliConfigDir, mockCliConfigFile, nil
	}

	vault := obsidian.Vault{}
	assert.NoError(t, vault.SetDefaultName("Example Vault"))

	settings := obsidian.VaultSettings{
		DailyNote: obsidian.DailyNoteSettings{
			Folder:          "Daily",
			FilenamePattern: "{YYYY-MM-DD}",
			TemplatePath:    "Templates/Daily.md",
			CreateIfMissing: true,
		},
	}
	assert.NoError(t, vault.SetSettings(settings))

	updatedBytes, err := os.ReadFile(mockCliConfigFile)
	assert.NoError(t, err)
	var updated obsidian.CliConfig
	assert.NoError(t, json.Unmarshal(updatedBytes, &updated))
	assert.Equal(t, "Example Vault", updated.DefaultVaultName)
	assert.NotNil(t, updated.VaultSettings)
	assert.Contains(t, updated.VaultSettings, "Example Vault")
	assert.Equal(t, "Daily", updated.VaultSettings["Example Vault"].DailyNote.Folder)
}
