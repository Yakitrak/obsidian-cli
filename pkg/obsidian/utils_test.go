package obsidian_test

import (
	"github.com/Yakitrak/obsidian-cli/mocks"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
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
		
		// We can't easily test vim opening in a test environment without it hanging
		// So we'll skip this test if vim exists and would actually run
		// This test serves more as documentation of the expected behavior
		t.Skip("Skipping vim fallback test to avoid hanging in test environment")
	})

	t.Run("Handles nonexistent file", func(t *testing.T) {
		originalEditor := os.Getenv("EDITOR")
		defer os.Setenv("EDITOR", originalEditor)
		
		// Use a command that will fail when given a nonexistent file
		os.Setenv("EDITOR", "ls") // ls will fail on nonexistent file
		
		nonexistentFile := filepath.Join(tempDir, "nonexistent.md")
		err := obsidian.OpenInEditor(nonexistentFile)
		
		// We expect this to fail
		if err == nil {
			t.Error("Expected error for nonexistent file, got nil")
		}
	})

	t.Run("Handles editor command failure", func(t *testing.T) {
		originalEditor := os.Getenv("EDITOR")
		defer os.Setenv("EDITOR", originalEditor)
		
		// Use a command that will always fail
		os.Setenv("EDITOR", "false") // 'false' command always fails
		
		err := obsidian.OpenInEditor(testFile)
		
		// We expect this to fail
		if err == nil {
			t.Error("Expected error for failing editor command, got nil")
		}
	})
}
