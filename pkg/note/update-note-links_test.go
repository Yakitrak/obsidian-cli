package note_test

import (
	"fmt"
	"github.com/Yakitrak/obsidian-cli/pkg/note"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func createTmpDirAndFiles(t *testing.T, perm os.FileMode, files []string, content []byte) string {
	t.Helper()
	// Create a temporary test directory
	tmpDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %s", err)
	}
	for _, file := range files {
		err := os.WriteFile(filepath.Join(tmpDir, file), content, perm)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}
	// create other non markdown files
	err = os.WriteFile(filepath.Join(tmpDir, "file4.txt"), []byte("This is a test file"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	// create hidden directory
	err = os.Mkdir(filepath.Join(tmpDir, ".hidden"), 0644)
	if err != nil {
		t.Fatalf("Failed to create hidden directory: %v", err)
	}
	return tmpDir
}
func TestUpdateNoteLinks(t *testing.T) {
	oldNoteName := "oldNote"
	newNoteName := "newNote"
	//content := []byte("This is a test file with [[oldNote]] [[oldNote#section]] [[oldNote#section|text]]")
	// content with string format to avoid escaping
	content := []byte(fmt.Sprintf("This is a test file with [[%s]] [[%s#section]] [[%s#section|text]]", oldNoteName, oldNoteName, oldNoteName))
	testFiles := []string{"file1.md", "file2.md", "file3.md"}

	t.Run("Update note links", func(t *testing.T) {
		// setup
		tmpDir := createTmpDirAndFiles(t, 0644, testFiles, content)
		defer os.RemoveAll(tmpDir)

		// Call the function to be tested
		err := note.UpdateNoteLinks(tmpDir, oldNoteName, newNoteName)
		assert.Equal(t, nil, err)

		// Check if note links were updated in test files
		for _, file := range testFiles {
			newContent, err := os.ReadFile(filepath.Join(tmpDir, file))
			if err != nil {
				t.Fatalf("Failed to read test file: %v", err)
			}
			expectedContent := fmt.Sprintf("This is a test file with [[%s]] [[%s#section]] [[%s#section|text]]", newNoteName, newNoteName, newNoteName)
			assert.Equal(t, expectedContent, string(newContent))

		}
	})

	t.Run("Error on incorrect vaultPath", func(t *testing.T) {
		err := note.UpdateNoteLinks("", "oldNote", "newNote")
		assert.ErrorContains(t, err, "Failed to access vault directory")
	})

	t.Run("Error reading files in vaultPath", func(t *testing.T) {
		tmpDir := createTmpDirAndFiles(t, 0000, testFiles, content)
		defer os.RemoveAll(tmpDir)
		err := note.UpdateNoteLinks(tmpDir, "oldNote", "newNote")
		assert.ErrorContains(t, err, "Failed to read files in vault")
	})

	t.Run("Error on writing to files in vaultPath", func(t *testing.T) {

		tmpDir := createTmpDirAndFiles(t, 0444, testFiles, content)
		defer os.RemoveAll(tmpDir)
		err := note.UpdateNoteLinks(tmpDir, "oldNote", "newNote")
		assert.ErrorContains(t, err, "Failed to write to files in vault")
	})

}
