package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/stretchr/testify/assert"
)

func TestRenameCommand(t *testing.T) {
	// Override obsidian config path for test vault resolution
	originalConfig := obsidian.ObsidianConfigFile
	defer func() { obsidian.ObsidianConfigFile = originalConfig }()

	rootDir := t.TempDir()
	vaultDir := filepath.Join(rootDir, "testvault")
	if err := os.MkdirAll(vaultDir, 0o755); err != nil {
		t.Fatalf("mk vault dir: %v", err)
	}
	configFile := filepath.Join(rootDir, "obsidian.json")
	if err := os.WriteFile(configFile, []byte(`{"vaults":{"random":{"path":"`+vaultDir+`"}}}`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	obsidian.ObsidianConfigFile = func() (string, error) {
		return configFile, nil
	}

	if err := os.WriteFile(filepath.Join(vaultDir, "Old.md"), []byte("Link [[Old]]"), 0o644); err != nil {
		t.Fatalf("seed old: %v", err)
	}
	if err := os.WriteFile(filepath.Join(vaultDir, "Ref.md"), []byte("Ref [[Old|Alias]]"), 0o644); err != nil {
		t.Fatalf("seed ref: %v", err)
	}

	rootCmd.SetArgs([]string{"rename", "Old", "New", "--vault", "testvault"})
	err := rootCmd.Execute()
	rootCmd.SetArgs([]string{})

	assert.NoError(t, err)

	_, oldErr := os.Stat(filepath.Join(vaultDir, "Old.md"))
	assert.Error(t, oldErr)
	newContent, newErr := os.ReadFile(filepath.Join(vaultDir, "Ref.md"))
	assert.NoError(t, newErr)
	assert.Contains(t, string(newContent), "[[New|Alias]]")
}
