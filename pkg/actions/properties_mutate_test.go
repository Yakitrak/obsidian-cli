package actions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
	"github.com/stretchr/testify/require"
)

func TestSetPropertyOnFiles(t *testing.T) {
	dir := t.TempDir()
	notePath := filepath.Join(dir, "note.md")
	require.NoError(t, os.WriteFile(notePath, []byte("# Note"), 0644))

	vault := obsidian.Vault{Name: dir}
	note := obsidian.Note{}

	summary, err := SetPropertyOnFiles(&vault, &note, "status", "open", []string{"note.md"}, false, false)
	require.NoError(t, err)
	require.Equal(t, 1, summary.NotesTouched)
	require.Equal(t, 1, summary.PropertyChanges["status"])

	data, err := os.ReadFile(notePath)
	require.NoError(t, err)
	require.Contains(t, string(data), "status: open")
}

func TestDeletePropertiesAcrossVault(t *testing.T) {
	dir := t.TempDir()
	notePath := filepath.Join(dir, "note.md")
	require.NoError(t, os.WriteFile(notePath, []byte("---\nstatus: open\n---\nbody"), 0644))

	vault := obsidian.Vault{Name: dir}
	note := obsidian.Note{}

	summary, err := DeleteProperties(&vault, &note, []string{"status"}, nil, false)
	require.NoError(t, err)
	require.Equal(t, 1, summary.NotesTouched)
	require.Equal(t, 1, summary.PropertyChanges["status"])

	data, err := os.ReadFile(notePath)
	require.NoError(t, err)
	require.NotContains(t, string(data), "status")
}

func TestRenamePropertiesMerge(t *testing.T) {
	dir := t.TempDir()
	notePath := filepath.Join(dir, "note.md")
	require.NoError(t, os.WriteFile(notePath, []byte("---\nold: a\nnew:\n  - b\n---\nbody"), 0644))

	vault := obsidian.Vault{Name: dir}
	note := obsidian.Note{}

	summary, err := RenameProperties(&vault, &note, []string{"old"}, "new", true, nil, false)
	require.NoError(t, err)
	require.Equal(t, 1, summary.NotesTouched)
	require.Equal(t, 1, summary.PropertyChanges["old"])
	require.Equal(t, 1, summary.PropertyChanges["new"])

	data, err := os.ReadFile(notePath)
	require.NoError(t, err)
	content := string(data)
	require.Contains(t, content, "new:")
	require.Contains(t, content, "- a")
	require.Contains(t, content, "- b")
	require.NotContains(t, content, "old:")
}
