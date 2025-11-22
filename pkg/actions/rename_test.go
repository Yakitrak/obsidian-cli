package actions

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

type stubVault struct {
	path string
}

func (s stubVault) DefaultName() (string, error) { return "", nil }
func (s stubVault) SetDefaultName(string) error  { return nil }
func (s stubVault) Path() (string, error)        { return s.path, nil }

func TestRenameNote_GitRenameWithBacklinks(t *testing.T) {
	vaultDir := t.TempDir()
	oldName := "Old Note.md"

	// Seed files
	if err := os.WriteFile(filepath.Join(vaultDir, oldName), []byte("# Old\n"), 0o644); err != nil {
		t.Fatalf("write old note: %v", err)
	}
	refContent := "Links [[Old Note]] [[Old Note|Alias]] [[Old Note#Heading]] [[Old Note#^block|Alias]] [md](Old Note.md#section) ![emb](Old Note.md)"
	if err := os.WriteFile(filepath.Join(vaultDir, "Ref.md"), []byte(refContent), 0o644); err != nil {
		t.Fatalf("write ref note: %v", err)
	}

	// Init git and track files for history-preserving rename.
	if err := exec.Command("git", "-C", vaultDir, "init").Run(); err != nil {
		t.Fatalf("git init: %v", err)
	}
	if err := exec.Command("git", "-C", vaultDir, "add", ".").Run(); err != nil {
		t.Fatalf("git add: %v", err)
	}

	params := RenameParams{
		Source:          "Old Note",
		Target:          "New Note",
		Overwrite:       false,
		UpdateBacklinks: true,
	}
	res, err := RenameNote(stubVault{path: vaultDir}, params)
	assert.NoError(t, err)
	assert.Equal(t, "New Note.md", res.RenamedPath)
	assert.True(t, res.GitHistoryPreserved)
	assert.GreaterOrEqual(t, res.LinkUpdates, 5)

	// Links should be rewritten to the new target
	updated, readErr := os.ReadFile(filepath.Join(vaultDir, "Ref.md"))
	assert.NoError(t, readErr)
	assert.NotContains(t, string(updated), "Old Note.md")
	assert.Contains(t, string(updated), "[[New Note]]")
	assert.Contains(t, string(updated), "[[New Note|Alias]]")
	assert.Contains(t, string(updated), "[[New Note#Heading]]")
	assert.Contains(t, string(updated), "[[New Note#^block|Alias]]")
	assert.Contains(t, string(updated), "(New Note.md#section)")

	// Git status should report a rename
	statusOut, statusErr := exec.Command("git", "-C", vaultDir, "status", "--porcelain").Output()
	assert.NoError(t, statusErr)
	assert.Contains(t, string(statusOut), "New Note.md")
	assert.NotContains(t, string(statusOut), "Old Note.md")
}

func TestRenameNote_TargetExistsBlocksWithoutOverwrite(t *testing.T) {
	vaultDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(vaultDir, "Old.md"), []byte("old"), 0o644); err != nil {
		t.Fatalf("write old: %v", err)
	}
	if err := os.WriteFile(filepath.Join(vaultDir, "Old2.md"), []byte("new"), 0o644); err != nil {
		t.Fatalf("write new: %v", err)
	}

	params := RenameParams{
		Source:          "Old.md",
		Target:          "Old2.md",
		Overwrite:       false,
		UpdateBacklinks: false,
	}
	_, err := RenameNote(stubVault{path: vaultDir}, params)
	assert.Error(t, err)
}
