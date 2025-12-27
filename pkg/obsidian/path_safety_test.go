package obsidian_test

import (
	"path/filepath"
	"testing"

	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/stretchr/testify/assert"
)

func TestSafeJoinVaultPath(t *testing.T) {
	vault := t.TempDir()

	t.Run("Rejects absolute paths", func(t *testing.T) {
		_, err := obsidian.SafeJoinVaultPath(vault, string(filepath.Separator)+"abs.md")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "absolute paths are not allowed:")
	})

	t.Run("Rejects empty paths", func(t *testing.T) {
		_, err := obsidian.SafeJoinVaultPath(vault, "  ")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "note path cannot be empty")
	})

	t.Run("Rejects escape paths", func(t *testing.T) {
		_, err := obsidian.SafeJoinVaultPath(vault, "../escape.md")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "note path escapes vault:")
	})

	t.Run("Allows normal relative paths", func(t *testing.T) {
		got, err := obsidian.SafeJoinVaultPath(vault, "Folder/Note.md")
		assert.NoError(t, err)
		assert.Contains(t, got, filepath.Join(vault, "Folder", "Note.md"))
	})
}
