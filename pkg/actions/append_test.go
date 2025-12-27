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

func TestAppendToDailyNoteAppliesTemplateOnCreate(t *testing.T) {
	withTempVaultAndConfig(t, func(vault *obsidian.Vault, vaultPath string) {
		now := time.Now()

		assert.NoError(t, os.MkdirAll(filepath.Join(vaultPath, "Templates"), 0750))
		assert.NoError(t, os.WriteFile(filepath.Join(vaultPath, "Templates", "T.md"), []byte("Title={{title}}\nDate={{date:YYYYMMDDHHmmss}}\n"), 0600))

		assert.NoError(t, vault.SetSettings(obsidian.VaultSettings{
			DailyNote: obsidian.DailyNoteSettings{
				Folder:          "Daily",
				FilenamePattern: "{YYYY-MM-DD}",
				TemplatePath:    "Templates/T",
			},
		}))

		assert.NoError(t, AppendToDailyNote(vault, "hello"))

		filename := now.Format("2006-01-02") + ".md"
		noteAbs := filepath.Join(vaultPath, "Daily", filename)
		b, err := os.ReadFile(noteAbs)
		assert.NoError(t, err)
		s := string(b)
		assert.Contains(t, s, "Title="+now.Format("2006-01-02")+"\n") // title is the note filename
		assert.Contains(t, s, "Date=")
		assert.Contains(t, s, "hello\n")
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

func TestPlanDailyAppendSupportsTimeBasedPatterns(t *testing.T) {
	withTempVaultAndConfig(t, func(vault *obsidian.Vault, _ string) {
		now := time.Date(2025, 12, 25, 17, 31, 45, 0, time.UTC)
		assert.NoError(t, vault.SetSettings(obsidian.VaultSettings{
			DailyNote: obsidian.DailyNoteSettings{
				Folder:          "Daily",
				FilenamePattern: "YYYY-MM-DD_HHmmss",
			},
		}))

		plan, err := PlanDailyAppend(vault, now)
		assert.NoError(t, err)
		assert.Equal(t, "2025-12-25_173145", plan.Filename)
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
