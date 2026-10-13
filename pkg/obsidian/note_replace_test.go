package obsidian

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoteReplacement(t *testing.T) {
	for _, partial := range []bool{false, true} {
		name := "before write"
		if partial {
			name = "partial write"
		}
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "note.md")
			checkReplacementError(t, os.WriteFile(path, []byte("original content"), 0600))
			failure := errors.New("disk full")
			err := writeNoteReplacement(path, func(w io.Writer) error {
				if partial {
					_, err := io.WriteString(w, "partial")
					checkReplacementError(t, err)
				}
				return failure
			})
			assert.ErrorIs(t, err, failure)
			content, err := os.ReadFile(path)
			checkReplacementError(t, err)
			assert.Equal(t, "original content", string(content))
			entries, err := os.ReadDir(dir)
			checkReplacementError(t, err)
			assert.Len(t, entries, 1)
		})
	}
	t.Run("successful replacement", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "note.md")
		checkReplacementError(t, os.WriteFile(path, []byte("old"), 0600))
		checkReplacementError(t, replaceNoteFile(path, []byte("new")))
		content, err := os.ReadFile(path)
		checkReplacementError(t, err)
		assert.Equal(t, "new", string(content))
		info, err := os.Stat(path)
		checkReplacementError(t, err)
		assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
	})
	t.Run("symlink preserved", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "note.md")
		link := filepath.Join(dir, "link.md")
		checkReplacementError(t, os.WriteFile(path, []byte("old"), 0600))
		checkReplacementError(t, os.Symlink("note.md", link))
		checkReplacementError(t, replaceNoteFile(link, []byte("new")))
		target, err := os.Readlink(link)
		checkReplacementError(t, err)
		assert.Equal(t, "note.md", target)
		content, err := os.ReadFile(path)
		checkReplacementError(t, err)
		assert.Equal(t, "new", string(content))
	})
}

func checkReplacementError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func TestNoteReplacementInterrupted(t *testing.T) {
	if path := os.Getenv("NOTESMD_TEST_INTERRUPTED_PATH"); path != "" {
		err := writeNoteReplacement(path, func(w io.Writer) error {
			if _, err := io.WriteString(w, "partial replacement"); err != nil {
				return err
			}
			os.Exit(23) // Simulate abrupt termination without running deferred cleanup.
			return nil
		})
		t.Fatalf("expected process exit, got %v", err)
	}
	path := filepath.Join(t.TempDir(), "note.md")
	checkReplacementError(t, os.WriteFile(path, []byte("original content"), 0600))
	cmd := exec.Command(os.Args[0], "-test.run=^TestNoteReplacementInterrupted$")
	cmd.Env = append(os.Environ(), "NOTESMD_TEST_INTERRUPTED_PATH="+path)
	err := cmd.Run()
	var exitError *exec.ExitError
	if !errors.As(err, &exitError) {
		t.Fatalf("expected child exit, got %v", err)
	}
	assert.Equal(t, 23, exitError.ExitCode())
	content, err := os.ReadFile(path)
	checkReplacementError(t, err)
	assert.Equal(t, "original content", string(content))
}
