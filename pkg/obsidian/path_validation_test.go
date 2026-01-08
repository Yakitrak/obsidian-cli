package obsidian_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/stretchr/testify/assert"
)

func TestValidatePath(t *testing.T) {
	// Setup temp directory for tests
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault")
	err := os.MkdirAll(vaultPath, 0755)
	assert.NoError(t, err)

	tests := []struct {
		name         string
		basePath     string
		relativePath string
		wantErr      bool
	}{
		{
			name:         "Valid simple note name",
			basePath:     vaultPath,
			relativePath: "note.md",
			wantErr:      false,
		},
		{
			name:         "Valid nested path",
			basePath:     vaultPath,
			relativePath: "subfolder/note.md",
			wantErr:      false,
		},
		{
			name:         "Valid deeply nested path",
			basePath:     vaultPath,
			relativePath: "a/b/c/note.md",
			wantErr:      false,
		},
		{
			name:         "Path traversal with ..",
			basePath:     vaultPath,
			relativePath: "../secret.txt",
			wantErr:      true,
		},
		{
			name:         "Path traversal with nested ..",
			basePath:     vaultPath,
			relativePath: "subfolder/../../secret.txt",
			wantErr:      true,
		},
		{
			name:         "Path traversal with multiple ..",
			basePath:     vaultPath,
			relativePath: "../../../etc/passwd",
			wantErr:      true,
		},
		{
			name:         "Current dir prefix is valid",
			basePath:     vaultPath,
			relativePath: "./note.md",
			wantErr:      false,
		},
		{
			name:         "Hidden directory traversal",
			basePath:     vaultPath,
			relativePath: "./../secret.txt",
			wantErr:      true,
		},
		{
			name:         "Empty relative path resolves to vault root",
			basePath:     vaultPath,
			relativePath: "",
			wantErr:      false,
		},
		{
			name:         "Dot only path resolves to vault root",
			basePath:     vaultPath,
			relativePath: ".",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := obsidian.ValidatePath(tt.basePath, tt.relativePath)

			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, obsidian.ErrPathTraversal)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)
			}
		})
	}
}

func TestValidatePath_AbsolutePaths(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name         string
		relativePath string
	}{
		{
			name:         "Unix absolute path",
			relativePath: "/etc/passwd",
		},
		{
			name:         "Root path",
			relativePath: "/",
		},
	}

	// Only run Windows tests on Windows
	if runtime.GOOS == "windows" {
		tests = append(tests, struct {
			name         string
			relativePath string
		}{
			name:         "Windows absolute path",
			relativePath: "C:\\Windows\\System32",
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := obsidian.ValidatePath(tempDir, tt.relativePath)
			assert.Error(t, err)
			assert.ErrorIs(t, err, obsidian.ErrPathTraversal)
		})
	}
}

func TestValidatePath_ResultWithinBase(t *testing.T) {
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault")
	err := os.MkdirAll(vaultPath, 0755)
	assert.NoError(t, err)

	t.Run("Result path is within vault", func(t *testing.T) {
		result, err := obsidian.ValidatePath(vaultPath, "subfolder/note.md")
		assert.NoError(t, err)

		// Verify result starts with vault path
		absVault, _ := filepath.Abs(vaultPath)
		assert.True(t, len(result) >= len(absVault), "Result should be at least as long as vault path")
		assert.Equal(t, absVault, result[:len(absVault)], "Result should start with vault path")
	})
}
