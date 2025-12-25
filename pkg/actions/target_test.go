package actions_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/config"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/stretchr/testify/assert"
)

func TestAppendToTarget_DryRunAndWrite(t *testing.T) {
	originalUserConfigDirectory := config.UserConfigDirectory
	defer func() { config.UserConfigDirectory = originalUserConfigDirectory }()

	cfgRoot := t.TempDir()
	config.UserConfigDirectory = func() (string, error) { return cfgRoot, nil }

	cliDir, targetsFile, err := config.TargetsPath()
	assert.NoError(t, err)
	assert.NoError(t, os.MkdirAll(cliDir, 0750))
	assert.NoError(t, os.WriteFile(targetsFile, []byte("inbox:\n  type: file\n  file: Inbox.md\n"), 0600))

	vaultRoot := t.TempDir()
	now := time.Date(2024, 1, 15, 14, 30, 52, 0, time.UTC)

	originalObsidianConfigFile := obsidian.ObsidianConfigFile
	defer func() { obsidian.ObsidianConfigFile = originalObsidianConfigFile }()
	mockCfg := `{"vaults":{"id":{"path":"` + vaultRoot + `/vault"}}}`
	cfgFile := filepath.Join(t.TempDir(), "obsidian.json")
	assert.NoError(t, os.WriteFile(cfgFile, []byte(mockCfg), 0644))
	obsidian.ObsidianConfigFile = func() (string, error) { return cfgFile, nil }
	v := &obsidian.Vault{Name: "vault"}

	plan, err := actions.AppendToTarget(v, "inbox", "x", now, true)
	assert.NoError(t, err)
	assert.True(t, plan.WillCreateFile)
	_, err = os.Stat(filepath.Join(vaultRoot, "vault", "Inbox.md"))
	assert.True(t, os.IsNotExist(err))

	plan, err = actions.AppendToTarget(v, "inbox", "x", now, false)
	assert.NoError(t, err)
	b, err := os.ReadFile(filepath.Join(vaultRoot, "vault", "Inbox.md"))
	assert.NoError(t, err)
	assert.Equal(t, "x\n", string(b))
}

func TestAppendToTarget_AppliesTemplateOnCreate(t *testing.T) {
	originalUserConfigDirectory := config.UserConfigDirectory
	defer func() { config.UserConfigDirectory = originalUserConfigDirectory }()

	cfgRoot := t.TempDir()
	config.UserConfigDirectory = func() (string, error) { return cfgRoot, nil }

	vaultRoot := t.TempDir()

	originalObsidianConfigFile := obsidian.ObsidianConfigFile
	defer func() { obsidian.ObsidianConfigFile = originalObsidianConfigFile }()
	mockCfg := `{"vaults":{"id":{"path":"` + vaultRoot + `/vault"}}}`
	cfgFile := filepath.Join(t.TempDir(), "obsidian.json")
	assert.NoError(t, os.WriteFile(cfgFile, []byte(mockCfg), 0644))
	obsidian.ObsidianConfigFile = func() (string, error) { return cfgFile, nil }

	cliDir, targetsFile, err := config.TargetsPath()
	assert.NoError(t, err)
	assert.NoError(t, os.MkdirAll(cliDir, 0750))

	// Template lives in the vault.
	assert.NoError(t, os.MkdirAll(filepath.Join(vaultRoot, "vault", "Templates"), 0750))
	assert.NoError(t, os.WriteFile(filepath.Join(vaultRoot, "vault", "Templates", "T.md"), []byte("Title={{title}}\nDate={{date:YYYYMMDDHHmmss}}\n"), 0600))

	assert.NoError(t, os.WriteFile(targetsFile, []byte("t:\n  type: file\n  file: Inbox.md\n  template: Templates/T\n"), 0600))

	now := time.Date(2024, 1, 15, 14, 30, 52, 0, time.UTC)
	v := &obsidian.Vault{Name: "vault"}

	_, err = actions.AppendToTarget(v, "t", "hello", now, false)
	assert.NoError(t, err)

	b, err := os.ReadFile(filepath.Join(vaultRoot, "vault", "Inbox.md"))
	assert.NoError(t, err)
	s := string(b)
	assert.Contains(t, s, "Title=Inbox")
	assert.Contains(t, s, "Date=20240115143052")
	assert.Contains(t, s, "hello\n")
}
