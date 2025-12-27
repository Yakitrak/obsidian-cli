package obsidian_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/stretchr/testify/assert"
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

	t.Run("Get contents by full path", func(t *testing.T) {
		// Arrange
		tempDir := t.TempDir()
		vaultPath := "vault-folder"
		subDir := "02 Personal/Hobbies/Cooking"
		noteName := "Cookies.md"
		fullNotePath := filepath.Join(subDir, noteName)
		actualNotePath := filepath.Join(tempDir, vaultPath, fullNotePath)
		fileContents := "Cookie recipe content here"

		err := os.MkdirAll(filepath.Dir(actualNotePath), 0755)
		if err != nil {
			t.Fatal(err)
		}

		err = os.WriteFile(actualNotePath, []byte(fileContents), 0644)
		if err != nil {
			t.Fatal(err)
		}

		// Act - test with full path
		noteManager := obsidian.Note{}
		content, err := noteManager.GetContents(filepath.Join(tempDir, vaultPath), fullNotePath)

		// Assert
		assert.Equal(t, nil, err, "Expected no error while retrieving note contents by full path")
		assert.Equal(t, fileContents, content, "Expected contents to match the file contents")

		// Act - test with just filename (should still work for backward compatibility)
		content2, err2 := noteManager.GetContents(filepath.Join(tempDir, vaultPath), "Cookies")

		// Assert
		assert.Equal(t, nil, err2, "Expected no error while retrieving note contents by filename")
		assert.Equal(t, fileContents, content2, "Expected contents to match the file contents")
	})

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

func TestUpdateNoteLinks_PathBasedWikilinks(t *testing.T) {
	t.Run("Update path-based wikilinks", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a file with path-based wikilinks
		content := []byte("Link to [[folder/oldNote]] and [[folder/oldNote#section]] and [[folder/oldNote|alias]]")
		testFile := filepath.Join(tmpDir, "test.md")
		err := os.WriteFile(testFile, content, 0644)
		assert.NoError(t, err)

		noteManager := obsidian.Note{}
		err = noteManager.UpdateLinks(tmpDir, "folder/oldNote", "folder/newNote")
		assert.NoError(t, err)

		newContent, _ := os.ReadFile(testFile)
		contentStr := string(newContent)
		assert.Contains(t, contentStr, "[[folder/newNote]]")
		assert.Contains(t, contentStr, "[[folder/newNote#section]]")
		assert.Contains(t, contentStr, "[[folder/newNote|alias]]")
	})
}

func TestUpdateNoteLinks_MarkdownLinks(t *testing.T) {
	t.Run("Update markdown links", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a file with markdown links
		content := []byte("Link [here](folder/oldNote.md) and [also](./folder/oldNote.md) and [no ext](folder/oldNote)")
		testFile := filepath.Join(tmpDir, "test.md")
		err := os.WriteFile(testFile, content, 0644)
		assert.NoError(t, err)

		noteManager := obsidian.Note{}
		err = noteManager.UpdateLinks(tmpDir, "folder/oldNote", "folder/newNote")
		assert.NoError(t, err)

		newContent, _ := os.ReadFile(testFile)
		contentStr := string(newContent)
		assert.Contains(t, contentStr, "(folder/newNote.md)")
		assert.Contains(t, contentStr, "(./folder/newNote.md)")
		assert.Contains(t, contentStr, "(folder/newNote)")
	})
}

func TestUpdateNoteLinks_MixedFormats(t *testing.T) {
	t.Run("Update mixed link formats in same file", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a file with both wikilinks and markdown links
		content := []byte(`# Mixed Links
- Wikilink: [[folder/oldNote]]
- Wikilink with alias: [[folder/oldNote|click here]]
- Wikilink with section: [[folder/oldNote#intro]]
- Simple wikilink: [[oldNote]]
- Markdown link: [text](folder/oldNote.md)
- Relative markdown: [text](./folder/oldNote.md)
`)
		testFile := filepath.Join(tmpDir, "test.md")
		err := os.WriteFile(testFile, content, 0644)
		assert.NoError(t, err)

		noteManager := obsidian.Note{}
		err = noteManager.UpdateLinks(tmpDir, "folder/oldNote", "newFolder/newNote")
		assert.NoError(t, err)

		newContent, _ := os.ReadFile(testFile)
		contentStr := string(newContent)

		// All links should be updated
		assert.Contains(t, contentStr, "[[newFolder/newNote]]")
		assert.Contains(t, contentStr, "[[newFolder/newNote|click here]]")
		assert.Contains(t, contentStr, "[[newFolder/newNote#intro]]")
		assert.Contains(t, contentStr, "[[newNote]]")
		assert.Contains(t, contentStr, "(newFolder/newNote.md)")
		assert.Contains(t, contentStr, "(./newFolder/newNote.md)")
	})
}

func TestUpdateLinks_PreservesTimestamps(t *testing.T) {
	t.Run("Only writes files with actual link changes", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		oldTime := time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC)

		// Create files with different content
		fileWithLinks := filepath.Join(tmpDir, "with_links.md")
		fileWithoutLinks := filepath.Join(tmpDir, "without_links.md")
		fileWithOtherLinks := filepath.Join(tmpDir, "other_links.md")

		// File that contains the old note name - should be updated
		err := os.WriteFile(fileWithLinks, []byte("Content with [[OldNote]] reference"), 0644)
		if err != nil {
			t.Fatal(err)
		}

		// File with no relevant links - should NOT be updated
		err = os.WriteFile(fileWithoutLinks, []byte("Content with no links"), 0644)
		if err != nil {
			t.Fatal(err)
		}

		// File with other links - should NOT be updated
		err = os.WriteFile(fileWithOtherLinks, []byte("Content with [[SomeOtherNote]] reference"), 0644)
		if err != nil {
			t.Fatal(err)
		}

		// Set all files to old timestamp
		for _, file := range []string{fileWithLinks, fileWithoutLinks, fileWithOtherLinks} {
			err = os.Chtimes(file, oldTime, oldTime)
			if err != nil {
				t.Fatal(err)
			}
		}

		// Record original timestamps
		getModTime := func(path string) time.Time {
			info, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			return info.ModTime()
		}

		originalWithLinks := getModTime(fileWithLinks)
		originalWithoutLinks := getModTime(fileWithoutLinks)
		originalOtherLinks := getModTime(fileWithOtherLinks)

		// Act
		noteManager := obsidian.Note{}
		err = noteManager.UpdateLinks(tmpDir, "OldNote", "newnote")
		assert.NoError(t, err)

		// Assert timestamps
		newWithLinks := getModTime(fileWithLinks)
		newWithoutLinks := getModTime(fileWithoutLinks)
		newOtherLinks := getModTime(fileWithOtherLinks)

		// File with links should have new timestamp
		assert.True(t, newWithLinks.After(originalWithLinks), "File with links should have updated timestamp")

		// Files without relevant links should preserve timestamps
		assert.Equal(t, originalWithoutLinks, newWithoutLinks, "File without links should preserve timestamp")
		assert.Equal(t, originalOtherLinks, newOtherLinks, "File with other links should preserve timestamp")

		// Verify content was actually updated in the changed file
		content, err := os.ReadFile(fileWithLinks)
		assert.NoError(t, err)
		assert.Contains(t, string(content), "[[newnote]]", "Links should be updated in changed file")
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

func TestSearchNotesWithSnippets(t *testing.T) {
	t.Run("Search notes with content matches", func(t *testing.T) {
		// Arrange
		tempDir := t.TempDir()
		vaultPath := "vault-folder"
		fullVaultPath := filepath.Join(tempDir, vaultPath)

		err := os.MkdirAll(fullVaultPath, 0755)
		assert.NoError(t, err)

		// Create test files
		testFiles := map[string]string{
			"note1.md":   "This is a test file\nwith some content\nand more lines",
			"note2.md":   "Another test document\nwith different content",
			"readme.txt": "This should be ignored",
		}

		for filename, content := range testFiles {
			err = os.WriteFile(filepath.Join(fullVaultPath, filename), []byte(content), 0644)
			assert.NoError(t, err)
		}

		// Act
		note := obsidian.Note{}
		matches, err := note.SearchNotesWithSnippets(fullVaultPath, "test")

		// Assert
		assert.NoError(t, err)
		assert.Len(t, matches, 2) // Should find 2 matches (one in each .md file)

		// Check that matches contain expected data
		foundFiles := make(map[string]bool)
		for _, match := range matches {
			foundFiles[match.FilePath] = true
			assert.Greater(t, match.LineNumber, 0)
			assert.Contains(t, match.MatchLine, "test")
		}

		assert.True(t, foundFiles["note1.md"])
		assert.True(t, foundFiles["note2.md"])
	})

	t.Run("Search notes with filename matches", func(t *testing.T) {
		// Arrange
		tempDir := t.TempDir()
		vaultPath := "vault-folder"
		fullVaultPath := filepath.Join(tempDir, vaultPath)

		err := os.MkdirAll(fullVaultPath, 0755)
		assert.NoError(t, err)

		err = os.WriteFile(filepath.Join(fullVaultPath, "test-note.md"), []byte("Some content"), 0644)
		assert.NoError(t, err)

		// Act
		note := obsidian.Note{}
		matches, err := note.SearchNotesWithSnippets(fullVaultPath, "test")

		// Assert
		assert.NoError(t, err)
		assert.Len(t, matches, 1)
		assert.Equal(t, "test-note.md", matches[0].FilePath)
		assert.Equal(t, 0, matches[0].LineNumber) // 0 indicates filename match
		assert.Contains(t, matches[0].MatchLine, "filename match")
	})

	t.Run("Search notes prioritizes content over filename matches", func(t *testing.T) {
		// Arrange
		tempDir := t.TempDir()
		vaultPath := "vault-folder"
		fullVaultPath := filepath.Join(tempDir, vaultPath)

		err := os.MkdirAll(fullVaultPath, 0755)
		assert.NoError(t, err)

		// Create a file that matches both filename and content
		err = os.WriteFile(filepath.Join(fullVaultPath, "test-note.md"), []byte("This contains test content\nAnother line"), 0644)
		assert.NoError(t, err)

		// Act
		note := obsidian.Note{}
		matches, err := note.SearchNotesWithSnippets(fullVaultPath, "test")

		// Assert
		assert.NoError(t, err)
		assert.Len(t, matches, 1) // Should only have content match, not filename match
		assert.Equal(t, "test-note.md", matches[0].FilePath)
		assert.Equal(t, 1, matches[0].LineNumber) // Should be content match (line 1)
		assert.Contains(t, matches[0].MatchLine, "test content")
		assert.NotContains(t, matches[0].MatchLine, "filename match")
	})

	t.Run("Search with no matches", func(t *testing.T) {
		// Arrange
		tempDir := t.TempDir()
		vaultPath := "vault-folder"
		fullVaultPath := filepath.Join(tempDir, vaultPath)

		err := os.MkdirAll(fullVaultPath, 0755)
		assert.NoError(t, err)

		err = os.WriteFile(filepath.Join(fullVaultPath, "note.md"), []byte("Some content"), 0644)
		assert.NoError(t, err)

		// Act
		note := obsidian.Note{}
		matches, err := note.SearchNotesWithSnippets(fullVaultPath, "nonexistent")

		// Assert
		assert.NoError(t, err)
		assert.Empty(t, matches)
	})

	t.Run("Search with long lines gets truncated", func(t *testing.T) {
		// Arrange
		tempDir := t.TempDir()
		vaultPath := "vault-folder"
		fullVaultPath := filepath.Join(tempDir, vaultPath)

		err := os.MkdirAll(fullVaultPath, 0755)
		assert.NoError(t, err)

		longLine := "This is a very long line that contains the word test and should be truncated because it exceeds the maximum length limit"
		err = os.WriteFile(filepath.Join(fullVaultPath, "note.md"), []byte(longLine), 0644)
		assert.NoError(t, err)

		// Act
		note := obsidian.Note{}
		matches, err := note.SearchNotesWithSnippets(fullVaultPath, "test")

		// Assert
		assert.NoError(t, err)
		assert.Len(t, matches, 1)
		assert.Less(t, len(matches[0].MatchLine), len(longLine))
		assert.Contains(t, matches[0].MatchLine, "test")
	})
}
