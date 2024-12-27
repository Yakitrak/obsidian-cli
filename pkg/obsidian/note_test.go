package obsidian_test

import (
	"fmt"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestDeleteNote(t *testing.T) {
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
			noteManager := obsidian.Note{}
			err = noteManager.Delete(notePath)
			// Assert
			assert.Equal(t, nil, err, "Expected no error while deleting note")

		})
	}

	t.Run("Delete non-existent note", func(t *testing.T) {
		// Arrange
		noteManager := obsidian.Note{}
		// Act
		err := noteManager.Delete("non-existent-note")
		// Assert
		assert.Equal(t, obsidian.NoteDoesNotExistError, err.Error(), "Expected error while deleting non-existent note")

	})
}
func TestNote_GetContents(t *testing.T) {
	tests := []struct {
		testName           string
		noteToCreate       string
		noteNameToRetrieve string
	}{
		{"Get contents of note", "note.md", "note.md"},
		{"Get contents of note without md", "note.md", "note"},
	}
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			// Arrange
			tempDir := t.TempDir()
			vaultPath := "vault-folder"
			notePath := filepath.Join(tempDir, vaultPath, test.noteToCreate)
			fileContents := "Example file contents here"

			err := os.MkdirAll(filepath.Join(tempDir, vaultPath), 0755)
			if err != nil {
				t.Fatal(err)
			}

			err = os.WriteFile(notePath, []byte(fileContents), 0644)
			if err != nil {
				t.Fatal(err)
			}

			// Act
			noteManager := obsidian.Note{}
			content, err := noteManager.GetContents(filepath.Join(tempDir, vaultPath), test.noteNameToRetrieve)

			// Assert
			assert.Equal(t, nil, err, "Expected no error while retrieving note contents")
			assert.Equal(t, fileContents, content, "Expected contents to match the file contents")
		})
	}

	t.Run("Get contents of non-existent note", func(t *testing.T) {
		// Arrange
		noteManager := obsidian.Note{}
		// Act
		contents, err := noteManager.GetContents("path", "non-existent-note")
		// Assert
		assert.Equal(t, obsidian.NoteDoesNotExistError, err.Error(), "Expected error while deleting non-existent note")
		assert.Equal(t, contents, "")

	})
}

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
			// Arrange
			tempDir := t.TempDir()
			existingNoteFullPathToCreate := filepath.Join(tempDir, test.existingNotePathToCreate)
			expectedNewPath := filepath.Join(tempDir, test.expectedNotePath)

			err := os.WriteFile(existingNoteFullPathToCreate, []byte(originalContent), 0644)
			if err != nil {
				t.Fatal(err)
			}

			fullOriginalNotePath := filepath.Join(tempDir, test.originalNotePath)
			fullNewNotePath := filepath.Join(tempDir, test.newNotePath)

			noteManager := obsidian.Note{}

			// Act
			err = noteManager.Move(fullOriginalNotePath, fullNewNotePath)

			// Assert
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
		// Arrange
		noteManager := obsidian.Note{}
		// Act
		err := noteManager.Move("filepath/that/does/not/exist", "newNote")
		// Assert
		assert.Equal(t, err.Error(), obsidian.NoteDoesNotExistError)
	})
}

func createTmpDirAndFiles(t *testing.T, perm os.FileMode, files []string, content []byte) string {
	t.Helper()
	// Create a temporary test directory
	tmpDir := t.TempDir()
	for _, file := range files {
		err := os.WriteFile(filepath.Join(tmpDir, file), content, perm)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}
	// create other non markdown files
	err := os.WriteFile(filepath.Join(tmpDir, "file4.txt"), []byte("This is a test file"), 0644)
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
	content := []byte(fmt.Sprintf("This is a test file with [[%s]] [[%s#section]] [[%s#section|text]]", oldNoteName, oldNoteName, oldNoteName))
	testFiles := []string{"file1.md", "file2.md", "file3.md"}

	t.Run("Update note links successfully", func(t *testing.T) {
		// Arrange
		tmpDir := createTmpDirAndFiles(t, 0644, testFiles, content)

		noteManager := obsidian.Note{}

		// Act
		err := noteManager.UpdateLinks(tmpDir, oldNoteName, newNoteName)
		assert.Equal(t, nil, err)

		// Assert
		for _, file := range testFiles {
			newContent, err := os.ReadFile(filepath.Join(tmpDir, file))
			if err != nil {
				t.Fatalf("Failed to read test file: %v", err)
			}
			expectedContent := fmt.Sprintf("This is a test file with [[%s]] [[%s#section]] [[%s#section|text]]", newNoteName, newNoteName, newNoteName)
			assert.Equal(t, expectedContent, string(newContent))

		}
	})

	t.Run("Error on incorrect vault", func(t *testing.T) {
		// Arrange
		noteManager := obsidian.Note{}
		// Act
		err := noteManager.UpdateLinks("", "oldNote", "newNote")
		// Assert
		assert.Equal(t, err.Error(), obsidian.VaultAccessError)
	})

	t.Run("Error reading files in vault", func(t *testing.T) {
		// Arrange
		tmpDir := createTmpDirAndFiles(t, 0000, testFiles, content)
		noteManager := obsidian.Note{}
		// Act
		err := noteManager.UpdateLinks(tmpDir, "oldNote", "newNote")
		// Assert
		assert.Equal(t, err.Error(), obsidian.VaultReadError)
	})

	t.Run("Error on writing to files in vault", func(t *testing.T) {
		// Arrange
		tmpDir := createTmpDirAndFiles(t, 0444, testFiles, content)
		noteManager := obsidian.Note{}
		// Act
		err := noteManager.UpdateLinks(tmpDir, "oldNote", "newNote")
		// Assert
		assert.Equal(t, err.Error(), obsidian.VaultWriteError)
	})
}

func TestNote_GetNotesList(t *testing.T) {
	t.Run("Retrieve list of notes successfully", func(t *testing.T) {
		// Arrange
		testFiles := []string{"file1.md", "file2.md", "file3.md"}
		content := []byte("This is a test note")
		tmpDir := createTmpDirAndFiles(t, 0644, testFiles, content)

		noteManager := obsidian.Note{}

		// Act
		notes, err := noteManager.GetNotesList(tmpDir)

		// Assert
		assert.NoError(t, err, "Expected no error while retrieving notes list")
		assert.ElementsMatch(t, testFiles, notes, "Expected notes list to match the created files")
	})

	t.Run("Empty vault directory", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		noteManager := obsidian.Note{}

		// Act
		notes, err := noteManager.GetNotesList(tmpDir)

		// Assert
		assert.NoError(t, err, "Expected no error for empty vault directory")
		assert.Empty(t, notes, "Expected empty notes list for empty vault directory")
	})

	t.Run("Vault directory with non-Markdown files", func(t *testing.T) {
		// Arrange
		tmpDir := createTmpDirAndFiles(t, 0644, []string{"file1.txt", "file2.jpg"}, []byte("Non-markdown content"))
		noteManager := obsidian.Note{}

		// Act
		notes, err := noteManager.GetNotesList(tmpDir)

		// Assert
		assert.NoError(t, err, "Expected no error when non-Markdown files are present")
		assert.Empty(t, notes, "Expected empty notes list when no Markdown files are present")
	})
}
