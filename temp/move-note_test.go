package temp_test

import (
	"github.com/Yakitrak/obsidian-cli/temp"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestMoveNote(t *testing.T) {
	originalContent := "This is the original content."

	tests := []struct {
		testName                 string
		existingNotePathToCreate string
		originalNotePath         string
		newNotePath              string
		expectedNotePath         string
	}{
		{"Original path with .md", "original.md", "original.md", "newName", "newName.md"},
		{"Original without .md", "original.md", "original", "newName", "newName.md"},
		{"New note with .md", "original.md", "original", "newName.md", "newName.md"},
		{"New Note without .md", "original.md", "original", "newName", "newName.md"},
		{"Both with .md", "original.md", "original.md", "newName.md", "newName.md"},
		{"Both without .md", "original.md", "original", "newName", "newName.md"},
	}
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			// Create a temporary directory for testing
			tempDir := t.TempDir()

			// Create a temporary original file and write some content to it
			existingNoteFullPathToCreate := filepath.Join(tempDir, test.existingNotePathToCreate)

			// Expected path
			expectedNewPath := filepath.Join(tempDir, test.expectedNotePath)

			err := os.WriteFile(existingNoteFullPathToCreate, []byte(originalContent), 0644)
			if err != nil {
				t.Fatal(err)
			}

			// Call the MoveNote function
			fullOriginalNotePath := filepath.Join(tempDir, test.originalNotePath)
			fullNewNotePath := filepath.Join(tempDir, test.newNotePath)
			err = temp.MoveNote(fullOriginalNotePath, fullNewNotePath)
			assert.NoError(t, err, "Expected no error while moving note")

			// Check if the original file has been moved to the new path
			_, err = os.Stat(existingNoteFullPathToCreate)
			assert.True(t, os.IsNotExist(err), "Original file still exists at %s, expected it to be moved", existingNoteFullPathToCreate)

			// Check if the new file exists
			_, err = os.Stat(expectedNewPath)
			assert.False(t, os.IsNotExist(err), "New file does not exist at %s, expected it to be created", expectedNewPath)

			// Read the content of the new file and compare it with the original content
			newContent, err := os.ReadFile(expectedNewPath)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, string(newContent), originalContent, "New file content is %q, expected %q", string(newContent), originalContent)

		})
	}

	t.Run("Error when moving file", func(t *testing.T) {
		err := temp.MoveNote("filepath/that/does/not/exist", "newNote")
		assert.Error(t, err, "Expected an error while moving note")
	})
}
