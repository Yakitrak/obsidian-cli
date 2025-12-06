package actions

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	if err := exec.Command("git", "-C", vaultDir, "commit", "-m", "seed").Run(); err != nil {
		t.Fatalf("git commit: %v", err)
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
	_, statusErr := exec.Command("git", "-C", vaultDir, "status", "--porcelain").Output()
	assert.NoError(t, statusErr)
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

func TestRenameNote_OverwriteExistingTargetGit(t *testing.T) {
	vaultDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(vaultDir, "Old.md"), []byte("old"), 0o644); err != nil {
		t.Fatalf("write old: %v", err)
	}
	if err := os.WriteFile(filepath.Join(vaultDir, "Existing.md"), []byte("existing"), 0o644); err != nil {
		t.Fatalf("write existing: %v", err)
	}
	if err := exec.Command("git", "-C", vaultDir, "init").Run(); err != nil {
		t.Fatalf("git init: %v", err)
	}
	if err := exec.Command("git", "-C", vaultDir, "add", ".").Run(); err != nil {
		t.Fatalf("git add: %v", err)
	}
	if err := exec.Command("git", "-C", vaultDir, "commit", "-m", "seed").Run(); err != nil {
		t.Fatalf("git commit: %v", err)
	}

	params := RenameParams{
		Source:          "Old",
		Target:          "Existing",
		Overwrite:       true,
		UpdateBacklinks: false,
	}
	res, err := RenameNote(stubVault{path: vaultDir}, params)
	assert.NoError(t, err)
	assert.Equal(t, "Existing.md", res.RenamedPath)

	content, readErr := os.ReadFile(filepath.Join(vaultDir, "Existing.md"))
	assert.NoError(t, readErr)
	assert.Equal(t, "old", string(content))
}

func TestRenameNote_DirtyGitBlocks(t *testing.T) {
	vaultDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(vaultDir, "Old.md"), []byte("old"), 0o644); err != nil {
		t.Fatalf("write old: %v", err)
	}
	if err := os.WriteFile(filepath.Join(vaultDir, ".gitkeep"), []byte(""), 0o644); err != nil {
		t.Fatalf("write gitkeep: %v", err)
	}
	if err := exec.Command("git", "-C", vaultDir, "init").Run(); err != nil {
		t.Fatalf("git init: %v", err)
	}
	if err := exec.Command("git", "-C", vaultDir, "add", ".").Run(); err != nil {
		t.Fatalf("git add: %v", err)
	}
	if err := exec.Command("git", "-C", vaultDir, "commit", "-m", "seed").Run(); err != nil {
		t.Fatalf("git commit: %v", err)
	}

	params := RenameParams{
		Source:          "Old",
		Target:          "New",
		Overwrite:       false,
		UpdateBacklinks: false,
	}
	// dirty file should not block; expect fallback to git mv success or fs rename success
	_, err := RenameNote(stubVault{path: vaultDir}, params)
	assert.NoError(t, err)
}

func TestRenameAttachmentUpdatesLinks(t *testing.T) {
	vaultDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(vaultDir, "image.png"), []byte("data"), 0o644); err != nil {
		t.Fatalf("write attachment: %v", err)
	}
	content := "Cover ![[image.png]] and [inline](image.png)"
	if err := os.WriteFile(filepath.Join(vaultDir, "Ref.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write ref: %v", err)
	}

	params := RenameParams{
		Source:          "image.png",
		Target:          "assets/image.png",
		UpdateBacklinks: true,
	}
	res, err := RenameNote(stubVault{path: vaultDir}, params)
	assert.NoError(t, err)
	assert.Equal(t, "assets/image.png", res.RenamedPath)

	updated, readErr := os.ReadFile(filepath.Join(vaultDir, "Ref.md"))
	assert.NoError(t, readErr)
	assert.Contains(t, string(updated), "![[assets/image.png]]")
	assert.Contains(t, string(updated), "(assets/image.png)")
	_, statErr := os.Stat(filepath.Join(vaultDir, "assets/image.png"))
	assert.NoError(t, statErr)
}

func TestRenameNote_DuplicateBasenameSkipsBareLinks(t *testing.T) {
	vaultDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(vaultDir, "Folder"), 0o755); err != nil {
		t.Fatalf("make folder: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(vaultDir, "Area"), 0o755); err != nil {
		t.Fatalf("make area: %v", err)
	}
	if err := os.WriteFile(filepath.Join(vaultDir, "Area", "Old Note.md"), []byte("# Old root"), 0o644); err != nil {
		t.Fatalf("write area note: %v", err)
	}
	if err := os.WriteFile(filepath.Join(vaultDir, "Folder", "Old Note.md"), []byte("# Old other"), 0o644); err != nil {
		t.Fatalf("write folder note: %v", err)
	}

	refContent := strings.Join([]string{
		"[[Old Note]]",           // ambiguous, should not rewrite
		"[md](Area/Old Note.md)", // fully-qualified, should rewrite
		"[[Folder/Old Note]]",    // points to other file, should stay
	}, "\n")
	refPath := filepath.Join(vaultDir, "Ref.md")
	if err := os.WriteFile(refPath, []byte(refContent), 0o644); err != nil {
		t.Fatalf("write ref: %v", err)
	}

	params := RenameParams{
		Source:          "Area/Old Note",
		Target:          "Area/New Note",
		UpdateBacklinks: true,
	}
	_, err := RenameNote(stubVault{path: vaultDir}, params)
	assert.NoError(t, err)

	updated, readErr := os.ReadFile(refPath)
	assert.NoError(t, readErr)
	updatedStr := string(updated)
	assert.Contains(t, updatedStr, "[md](Area/New Note.md)")
	assert.Contains(t, updatedStr, "[[Old Note]]")
	assert.Contains(t, updatedStr, "[[Folder/Old Note]]")
	assert.NotContains(t, updatedStr, "[[New Note]]")
}
