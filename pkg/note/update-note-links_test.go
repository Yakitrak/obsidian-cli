package note_test

import (
	"github.com/Yakitrak/obsidian-cli/pkg/note"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestUpdateNoteLinks(t *testing.T) {
	// Create a temporary test directory
	tmpDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files with oldNoteName in their content
	oldNoteName := "oldNote"
	newNoteName := "newNote"
	testFiles := []string{"file1.md", "file2.md", "file3.md"}
	for _, file := range testFiles {
		content := []byte("This is a test file with [[oldNote]] [[oldNote#section]] [[oldNote#section|text]]")
		err := os.WriteFile(filepath.Join(tmpDir, file), content, 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// create other non markdown files
	err = os.WriteFile(filepath.Join(tmpDir, "file4.txt"), []byte("This is a test file"), 0644)
	// create hidden directory
	err = os.Mkdir(filepath.Join(tmpDir, ".hidden"), 0755)

	// Call the function to be tested
	err = note.UpdateNoteLinks(tmpDir, oldNoteName, newNoteName)
	if err != nil {
		t.Fatalf("Failed to update note links: %v", err)
	}

	// Check if note links were updated in test files
	for _, file := range testFiles {
		content, err := os.ReadFile(filepath.Join(tmpDir, file))
		if err != nil {
			t.Fatalf("Failed to read test file: %v", err)
		}
		assert.Equal(t, "This is a test file with [[newNote]] [[newNote#section]] [[newNote#section|text]]", string(content))

	}
}
