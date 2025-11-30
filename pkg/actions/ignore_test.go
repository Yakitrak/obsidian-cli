package actions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/atomicobject/obsidian-cli/mocks"
	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallDefaultIgnoreWritesFile(t *testing.T) {
	tmp := t.TempDir()

	vault := &mocks.VaultManager{}
	vault.On("Path").Return(tmp, nil)

	path, err := InstallDefaultIgnore(vault, InstallIgnoreOptions{})
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(tmp, ".obsidianignore"), path)

	data, readErr := os.ReadFile(path)
	require.NoError(t, readErr)
	for _, pattern := range obsidian.DefaultIgnorePatterns() {
		assert.Contains(t, string(data), pattern)
	}
}

func TestInstallDefaultIgnoreRespectsExistingFile(t *testing.T) {
	tmp := t.TempDir()
	existing := filepath.Join(tmp, ".obsidianignore")
	require.NoError(t, os.WriteFile(existing, []byte("custom\n"), 0o644))

	vault := &mocks.VaultManager{}
	vault.On("Path").Return(tmp, nil)

	_, err := InstallDefaultIgnore(vault, InstallIgnoreOptions{})
	assert.Error(t, err)

	path, err := InstallDefaultIgnore(vault, InstallIgnoreOptions{Force: true})
	require.NoError(t, err)
	assert.Equal(t, existing, path)

	data, readErr := os.ReadFile(existing)
	require.NoError(t, readErr)
	assert.Contains(t, string(data), "node_modules/")
}
