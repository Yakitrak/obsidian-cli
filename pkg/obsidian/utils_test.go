package obsidian_test

import (
	"github.com/Yakitrak/obsidian-cli/mocks"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAddMdSuffix(t *testing.T) {
	tests := []struct {
		testName string
		input    string
		want     string
	}{
		{testName: "Without existing .md", input: "myNote", want: "myNote.md"},
		{testName: "With existing .md", input: "myNote.md", want: "myNote.md"},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			// Act
			got := obsidian.AddMdSuffix(test.input)
			// Assert
			assert.Equal(t, test.want, got)
		})
	}
}

func TestRemoveMdSuffix(t *testing.T) {
	tests := []struct {
		testName string
		input    string
		want     string
	}{
		{testName: "Without existing .md", input: "myNote", want: "myNote"},
		{testName: "With existing .md", input: "myNote.md", want: "myNote"},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			// Act
			got := obsidian.RemoveMdSuffix(test.input)
			// Assert
			assert.Equal(t, test.want, got)
		})
	}
}

func TestGenerateNoteLinkTexts(t *testing.T) {
	var tests = []struct {
		testName string
		noteName string
		want     [3]string
	}{
		{"Note with .md", "note.md", [3]string{"[[note]]", "[[note|", "[[note#"}},
		{"Note without .md", "note", [3]string{"[[note]]", "[[note|", "[[note#"}},
		{"Note with .md and path", "path/to/note.md", [3]string{"[[note]]", "[[note|", "[[note#"}},
		{"Note without .md and path", "path/to/note", [3]string{"[[note]]", "[[note|", "[[note#"}},
	}
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			// Act
			got := obsidian.GenerateNoteLinkTexts(test.noteName)
			// Assert
			assert.Equal(t, test.want, got)
		})
	}
}

func TestGenerateLinkReplacements(t *testing.T) {
	t.Run("Simple note rename", func(t *testing.T) {
		replacements := obsidian.GenerateLinkReplacements("oldNote", "newNote")

		// Should have wikilink patterns
		assert.Equal(t, "[[newNote]]", replacements["[[oldNote]]"])
		assert.Equal(t, "[[newNote|", replacements["[[oldNote|"])
		assert.Equal(t, "[[newNote#", replacements["[[oldNote#"])

		// Should have markdown link patterns
		assert.Equal(t, "](newNote.md)", replacements["](oldNote.md)"])
		assert.Equal(t, "](newNote)", replacements["](oldNote)"])
	})

	t.Run("Note with path", func(t *testing.T) {
		replacements := obsidian.GenerateLinkReplacements("folder/oldNote", "folder/newNote")

		// Simple wikilinks (basename)
		assert.Equal(t, "[[newNote]]", replacements["[[oldNote]]"])
		assert.Equal(t, "[[newNote|", replacements["[[oldNote|"])
		assert.Equal(t, "[[newNote#", replacements["[[oldNote#"])

		// Path-based wikilinks
		assert.Equal(t, "[[folder/newNote]]", replacements["[[folder/oldNote]]"])
		assert.Equal(t, "[[folder/newNote|", replacements["[[folder/oldNote|"])
		assert.Equal(t, "[[folder/newNote#", replacements["[[folder/oldNote#"])

		// Markdown links
		assert.Equal(t, "](folder/newNote.md)", replacements["](folder/oldNote.md)"])
		assert.Equal(t, "](folder/newNote)", replacements["](folder/oldNote)"])

		// Relative markdown links
		assert.Equal(t, "](./folder/newNote.md)", replacements["](./folder/oldNote.md)"])
	})

	t.Run("Move to different folder", func(t *testing.T) {
		replacements := obsidian.GenerateLinkReplacements("folder1/note", "folder2/note")

		// Basename stays same
		assert.Equal(t, "[[note]]", replacements["[[note]]"])

		// Path-based wikilinks update
		assert.Equal(t, "[[folder2/note]]", replacements["[[folder1/note]]"])

		// Markdown links update
		assert.Equal(t, "](folder2/note.md)", replacements["](folder1/note.md)"])
	})

	t.Run("Note with .md extension", func(t *testing.T) {
		replacements := obsidian.GenerateLinkReplacements("folder/note.md", "folder/renamed.md")

		// Wikilinks don't include .md
		assert.Equal(t, "[[renamed]]", replacements["[[note]]"])
		assert.Equal(t, "[[folder/renamed]]", replacements["[[folder/note]]"])

		// Markdown links with .md
		assert.Equal(t, "](folder/renamed.md)", replacements["](folder/note.md)"])
	})

	t.Run("Nested path", func(t *testing.T) {
		replacements := obsidian.GenerateLinkReplacements("a/b/c/note", "x/y/note")

		assert.Equal(t, "[[x/y/note]]", replacements["[[a/b/c/note]]"])
		assert.Equal(t, "](x/y/note.md)", replacements["](a/b/c/note.md)"])
	})
}

func TestReplaceContent(t *testing.T) {
	tests := []struct {
		testName     string
		content      []byte
		replacements map[string]string
		want         []byte
	}{
		{"No replacements", []byte("This is the original content"), map[string]string{}, []byte("This is the original content")},
		{"Replace one word", []byte("This is the original content"), map[string]string{"original": "new"}, []byte("This is the new content")},
		{"Replace multiple words", []byte("This is the original content"), map[string]string{"original": "new", "content": "text"}, []byte("This is the new text")},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			// Act
			got := obsidian.ReplaceContent(test.content, test.replacements)
			// Assert
			assert.Equal(t, string(test.want), string(got))
		})
	}

}

func TestShouldSkipDirectoryOrFile(t *testing.T) {
	tests := []struct {
		testName string
		info     os.FileInfo
		want     bool
	}{
		{testName: "markdown file", info: &mocks.MockFileInfo{FileName: "file.md"}},
		{"text file", &mocks.MockFileInfo{FileName: "file.txt"}, true},
		{"image file", &mocks.MockFileInfo{FileName: "file.jpg"}, true},
		{"directory", &mocks.MockFileInfo{FileName: "directory", IsDirectory: true}, true},
		{"hidden directory", &mocks.MockFileInfo{FileName: ".hidden_directory", IsDirectory: true}, true},
		{"hidden file", &mocks.MockFileInfo{FileName: ".hidden_file"}, true},
		{"file with no extension", &mocks.MockFileInfo{FileName: "file_with_no_extension"}, true},
		{"file with dots", &mocks.MockFileInfo{FileName: "file.md.with.dots"}, true},
		{"markdown file with dots", &mocks.MockFileInfo{FileName: "file.with.multiple.dots.md"}, false},
	}
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			// Act
			got := obsidian.ShouldSkipDirectoryOrFile(test.info)
			// Assert
			assert.Equal(t, test.want, got)
		})
	}
}

func TestOpenInEditor(t *testing.T) {
	// Create a temporary file for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.md")
	err := os.WriteFile(testFile, []byte("# Test Note\n\nThis is a test note."), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	t.Run("Uses EDITOR environment variable", func(t *testing.T) {
		// Set EDITOR to a command that will succeed and exit quickly
		originalEditor := os.Getenv("EDITOR")
		defer os.Setenv("EDITOR", originalEditor)
		
		os.Setenv("EDITOR", "true") // 'true' command always succeeds and exits immediately
		
		err := obsidian.OpenInEditor(testFile)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("Falls back to vim when EDITOR not set", func(t *testing.T) {
		// Unset EDITOR environment variable
		originalEditor := os.Getenv("EDITOR")
		defer os.Setenv("EDITOR", originalEditor)
		
		os.Unsetenv("EDITOR")
		
		// TODO: We can't easily test vim opening in a test environment without it hanging.
		// Consider using a test helper that mocks the command execution or documenting
		// this limitation more explicitly to clarify this is intentional technical debt.
		t.Skip("Skipping vim fallback test to avoid hanging in test environment")
	})

	t.Run("Handles nonexistent file", func(t *testing.T) {
		originalEditor := os.Getenv("EDITOR")
		defer os.Setenv("EDITOR", originalEditor)
		
		// Use 'true' which succeeds regardless of arguments
		// Editors typically create new files when they don't exist
		os.Setenv("EDITOR", "true")
		
		nonexistentFile := filepath.Join(tempDir, "nonexistent.md")
		err := obsidian.OpenInEditor(nonexistentFile)
		
		// Should succeed - editors typically create new files
		if err != nil {
			t.Errorf("Expected no error for nonexistent file with 'true' editor, got: %v", err)
		}
	})

	t.Run("Handles editor command failure", func(t *testing.T) {
		originalEditor := os.Getenv("EDITOR")
		defer os.Setenv("EDITOR", originalEditor)
		
		// Use a command that will always fail
		os.Setenv("EDITOR", "false") // 'false' command always fails
		
		err := obsidian.OpenInEditor(testFile)
		
		// We expect this to fail and have error context
		if err == nil {
			t.Error("Expected error for failing editor command, got nil")
		}
		
		// Check that error message contains context
		if err != nil {
			errMsg := err.Error()
			if !strings.Contains(errMsg, "failed") {
				t.Errorf("Expected error message to contain 'failed', got: %s", errMsg)
			}
		}
	})
}
