package note_test

import (
	"github.com/Yakitrak/obsidian-cli/utils/note"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestDelete(t *testing.T) {
	tests := []struct {
		testName     string
		noteToCreate string
		noteArg      string
	}{
		{"Delete note with .md", "note.md", "note.md"},
		{"Delete note without .md", "note.md", "note"},
	}
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			// Arrange
			tempDir := t.TempDir()
			notePathToCreate := filepath.Join(tempDir, test.noteToCreate)
			notePath := filepath.Join(tempDir, test.noteArg)

			err := os.WriteFile(notePathToCreate, []byte(""), 0644)
			if err != nil {
				t.Fatal(err)
			}
			// Act
			err = note.Delete(notePath)
			// Assert
			assert.Equal(t, nil, err, "Expected no error while deleting note")

		})
	}

	t.Run("Delete non-existent note", func(t *testing.T) {
		// Act
		err := note.Delete("non-existent-note")
		// Assert
		assert.Equal(t, "note does not exist", err.Error(), "Expected error while deleting non-existent note")

	})
}
