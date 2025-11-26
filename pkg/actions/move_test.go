package actions

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Yakitrak/obsidian-cli/mocks"
	"github.com/stretchr/testify/assert"
)

type moveStubVault struct {
	path    string
	name    string
	err     error
	nameErr error
}

func (s moveStubVault) DefaultName() (string, error) {
	if s.nameErr != nil {
		return "", s.nameErr
	}
	if s.name != "" {
		return s.name, nil
	}
	return "", nil
}
func (s moveStubVault) SetDefaultName(string) error { return nil }
func (s moveStubVault) Path() (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return s.path, nil
}

func TestMoveNotes_SingleMoveNoBacklinks(t *testing.T) {
	vaultDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(vaultDir, "Note.md"), []byte("body"), 0o644); err != nil {
		t.Fatalf("seed note: %v", err)
	}

	uri := &mocks.MockUriManager{}
	summary, err := MoveNotes(moveStubVault{path: vaultDir}, uri, MoveParams{
		Moves: []MoveRequest{
			{Source: "Note.md", Target: "Folder/Note.md"},
		},
	})

	assert.NoError(t, err)
	assert.Len(t, summary.Results, 1)
	assert.Equal(t, "Folder/Note.md", summary.Results[0].Target)
	assert.Equal(t, 0, summary.TotalLinkUpdates)

	_, statErr := os.Stat(filepath.Join(vaultDir, "Folder/Note.md"))
	assert.NoError(t, statErr)
	_, oldErr := os.Stat(filepath.Join(vaultDir, "Note.md"))
	assert.Error(t, oldErr)
	uri.AssertNotCalled(t, "Construct")
	uri.AssertNotCalled(t, "Execute")
}

func TestMoveNotes_BulkMoveWithBacklinks(t *testing.T) {
	vaultDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(vaultDir, "One.md"), []byte("# One"), 0o644); err != nil {
		t.Fatalf("seed one: %v", err)
	}
	if err := os.WriteFile(filepath.Join(vaultDir, "Two.md"), []byte("# Two"), 0o644); err != nil {
		t.Fatalf("seed two: %v", err)
	}
	ref := "Links [[One]] and [[Two]] plus [md](One.md)"
	if err := os.WriteFile(filepath.Join(vaultDir, "Ref.md"), []byte(ref), 0o644); err != nil {
		t.Fatalf("seed ref: %v", err)
	}

	uri := &mocks.MockUriManager{}
	summary, err := MoveNotes(moveStubVault{path: vaultDir}, uri, MoveParams{
		Moves: []MoveRequest{
			{Source: "One", Target: "Dest/One"},
			{Source: "Two.md", Target: "Dest/Two.md"},
		},
		UpdateBacklinks: true,
	})

	assert.NoError(t, err)
	assert.Equal(t, 2, len(summary.Results))
	assert.GreaterOrEqual(t, summary.TotalLinkUpdates, 3)

	updated, readErr := os.ReadFile(filepath.Join(vaultDir, "Ref.md"))
	assert.NoError(t, readErr)
	assert.Contains(t, string(updated), "[[Dest/One]]")
	assert.Contains(t, string(updated), "[[Dest/Two]]")
	assert.Contains(t, string(updated), "(Dest/One.md)")
	uri.AssertNotCalled(t, "Construct")
	uri.AssertNotCalled(t, "Execute")
}

func TestMoveNotes_ShouldOpenSingle(t *testing.T) {
	vaultDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(vaultDir, "Note.md"), []byte("body"), 0o644); err != nil {
		t.Fatalf("seed note: %v", err)
	}

	uri := &mocks.MockUriManager{}
	uri.On("Construct", "obsidian://open", map[string]string{"file": "Renamed.md", "vault": "vault"}).Return("obsidian://open?vault=vault&file=Renamed.md")
	uri.On("Execute", "obsidian://open?vault=vault&file=Renamed.md").Return(nil)

	_, err := MoveNotes(moveStubVault{path: vaultDir, name: "vault"}, uri, MoveParams{
		Moves:      []MoveRequest{{Source: "Note.md", Target: "Renamed.md"}},
		ShouldOpen: true,
	})

	assert.NoError(t, err)
	uri.AssertExpectations(t)
}

func TestMoveNotes_DuplicateTargets(t *testing.T) {
	vaultDir := t.TempDir()
	uri := &mocks.MockUriManager{}

	_, err := MoveNotes(moveStubVault{path: vaultDir}, uri, MoveParams{
		Moves: []MoveRequest{
			{Source: "One", Target: "Dest/Note"},
			{Source: "Two", Target: "Dest/Note"},
		},
	})

	assert.Error(t, err)
}

func TestMoveNotes_BacklinkRewriteFolderSpecific(t *testing.T) {
	vaultDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(vaultDir, "Folder"), 0o755); err != nil {
		t.Fatalf("make folder: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(vaultDir, "Archive"), 0o755); err != nil {
		t.Fatalf("make archive: %v", err)
	}

	notePath := filepath.Join(vaultDir, "Folder", "Note.md")
	if err := os.WriteFile(notePath, []byte("# body"), 0o644); err != nil {
		t.Fatalf("seed note: %v", err)
	}
	refContent := strings.Join([]string{
		"[[Folder/Note]]",
		"[[Folder/Note#Heading]]",
		"[[Folder/Note#^block|Alias]]",
		"[md](Folder/Note.md#h)",
		"[[Note]]",
	}, "\n")
	refPath := filepath.Join(vaultDir, "Ref.md")
	if err := os.WriteFile(refPath, []byte(refContent), 0o644); err != nil {
		t.Fatalf("seed ref: %v", err)
	}

	uri := &mocks.MockUriManager{}
	summary, err := MoveNotes(moveStubVault{path: vaultDir}, uri, MoveParams{
		Moves:           []MoveRequest{{Source: "Folder/Note", Target: "Archive/Note"}},
		UpdateBacklinks: true,
	})
	assert.NoError(t, err)
	assert.Greater(t, summary.TotalLinkUpdates, 0)

	updated, readErr := os.ReadFile(refPath)
	assert.NoError(t, readErr)
	updatedStr := string(updated)
	assert.Contains(t, updatedStr, "[[Archive/Note]]")
	assert.Contains(t, updatedStr, "[[Archive/Note#Heading]]")
	assert.Contains(t, updatedStr, "[[Archive/Note#^block|Alias]]")
	assert.Contains(t, updatedStr, "(Archive/Note.md#h)")
	assert.Contains(t, updatedStr, "[[Note]]") // bare note name should remain untouched
}
