package actions

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Yakitrak/obsidian-cli/mocks"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/stretchr/testify/assert"
)

func TestAppendToDailyNoteErrorsWhenNotConfigured(t *testing.T) {
	withTempVaultAndConfig(t, func(vault *obsidian.Vault, _ string) {
		assert.NoError(t, vault.SetSettings(obsidian.VaultSettings{
			DailyNote: obsidian.DailyNoteSettings{
				Folder:          "",
				FilenamePattern: "{YYYY-MM-DD}",
			},
		}))

		err := AppendToDailyNote(vault, "hello")
		assert.Error(t, err)
	})
}

func TestAppendToDailyNoteCreatesAndAppends(t *testing.T) {
	withTempVaultAndConfig(t, func(vault *obsidian.Vault, vaultPath string) {
		assert.NoError(t, vault.SetSettings(obsidian.VaultSettings{
			DailyNote: obsidian.DailyNoteSettings{
				Folder:          "Daily",
				FilenamePattern: "{YYYY-MM-DD}",
			},
		}))

		assert.NoError(t, AppendToDailyNote(vault, "first line"))
		assert.NoError(t, AppendToDailyNote(vault, "second line"))

		filename := time.Now().Format("2006-01-02") + ".md"
		noteAbs := filepath.Join(vaultPath, "Daily", filename)
		b, err := os.ReadFile(noteAbs)
		assert.NoError(t, err)
		assert.Contains(t, string(b), "first line\n\nsecond line\n")
	})
}

func TestAppendToDailyNoteRejectsEscapePaths(t *testing.T) {
	withTempVaultAndConfig(t, func(vault *obsidian.Vault, _ string) {
		assert.NoError(t, vault.SetSettings(obsidian.VaultSettings{
			DailyNote: obsidian.DailyNoteSettings{
				Folder:          "../escape",
				FilenamePattern: "{YYYY-MM-DD}",
			},
		}))

		err := AppendToDailyNote(vault, "hello")
		assert.Error(t, err)
	})
}

func withTempVaultAndConfig(t *testing.T, fn func(vault *obsidian.Vault, vaultPath string)) {
	t.Helper()

	originalObsidianConfigFile := obsidian.ObsidianConfigFile
	t.Cleanup(func() { obsidian.ObsidianConfigFile = originalObsidianConfigFile })
	mockObsidianConfigFile := mocks.CreateMockObsidianConfigFile(t)
	obsidian.ObsidianConfigFile = func() (string, error) {
		return mockObsidianConfigFile, nil
	}

	originalCliConfigPath := obsidian.CliConfigPath
	t.Cleanup(func() { obsidian.CliConfigPath = originalCliConfigPath })
	mockCliConfigDir, mockCliConfigFile := mocks.CreateMockCliConfigDirectories(t)
	obsidian.CliConfigPath = func() (string, string, error) {
		return mockCliConfigDir, mockCliConfigFile, nil
	}

	vaultName := "Example Vault"
	vaultPath := filepath.Join(t.TempDir(), vaultName)
	assert.NoError(t, os.MkdirAll(vaultPath, 0750))

	obsConfig := obsidian.ObsidianVaultConfig{
		Vaults: map[string]struct {
			Path string `json:"path"`
		}{
			vaultName: {Path: vaultPath},
		},
	}
	obsBytes, err := json.Marshal(obsConfig)
	assert.NoError(t, err)
	assert.NoError(t, os.WriteFile(mockObsidianConfigFile, obsBytes, 0600))

	vault := obsidian.Vault{}
	assert.NoError(t, vault.SetDefaultName(vaultName))

	fn(&vault, vaultPath)
}
