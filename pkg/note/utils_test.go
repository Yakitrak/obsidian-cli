package note_test

import (
	"github.com/Yakitrak/obsidian-cli/pkg/note"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestAddMdSuffix(t *testing.T) {
	tests := []struct {
		testName string
		input    string
		expected string
	}{
		{testName: "Without existing .md", input: "myNote", expected: "myNote.md"},
		{testName: "With existing .md", input: "myNote.md", expected: "myNote.md"},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			// Act
			got := note.AddMdSuffix(tt.input)
			// Assert
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestRemoveMdSuffix(t *testing.T) {
	tests := []struct {
		testName string
		input    string
		expected string
	}{
		{testName: "Without existing .md", input: "myNote", expected: "myNote"},
		{testName: "With existing .md", input: "myNote.md", expected: "myNote"},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			// Act
			got := note.RemoveMdSuffix(tt.input)
			// Assert
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestGenerateNoteLinkTexts(t *testing.T) {
	var tests = []struct {
		testName string
		noteName string
		expected [3]string
	}{
		{"Note with .md", "note.md", [3]string{"[[note]]", "[[note|", "[[note#"}},
		{"Note without .md", "note", [3]string{"[[note]]", "[[note|", "[[note#"}},
		{"Note with .md and path", "path/to/note.md", [3]string{"[[note]]", "[[note|", "[[note#"}},
		{"Note without .md and path", "path/to/note", [3]string{"[[note]]", "[[note|", "[[note#"}},
	}
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			actual := note.GenerateNoteLinkTexts(test.noteName)
			if actual != test.expected {
				t.Errorf("Expected %q, got %q", test.expected, actual)
			}
		})
	}
}

func TestReplaceContent(t *testing.T) {

	tests := []struct {
		testName     string
		content      []byte
		replacements map[string]string
		expected     []byte
	}{
		{"No replacements", []byte("This is the original content"), map[string]string{}, []byte("This is the original content")},
		{"Replace one word", []byte("This is the original content"), map[string]string{"original": "new"}, []byte("This is the new content")},
		{"Replace multiple words", []byte("This is the original content"), map[string]string{"original": "new", "content": "text"}, []byte("This is the new text")},
		{"Replace multiple words with multiple replacements", []byte("This is the original content"), map[string]string{"original": "new", "content": "text", "new": "old"}, []byte("This is the old text")},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			got := note.ReplaceContent(test.content, test.replacements)
			assert.Equal(t, string(test.expected), string(got))
		})
	}

}

type mockFileInfo struct {
	name  string
	isDir bool
}

func (fi *mockFileInfo) Name() string {
	return fi.name
}

func (fi *mockFileInfo) Size() int64 {
	return 0
}

func (fi *mockFileInfo) Mode() os.FileMode {
	return 0
}

func (fi *mockFileInfo) ModTime() time.Time {
	return time.Now()
}

func (fi *mockFileInfo) IsDir() bool {
	return fi.isDir
}

func (fi *mockFileInfo) Sys() interface{} {
	return nil
}
func TestShouldSkipDirectoryOrFile(t *testing.T) {
	tests := []struct {
		testName string
		info     os.FileInfo
		want     bool
	}{
		{"markdown file", &mockFileInfo{"file.md", false}, false},
		{"text file", &mockFileInfo{"file.txt", false}, true},
		{"image file", &mockFileInfo{"file.jpg", false}, true},
		{"directory", &mockFileInfo{"directory", true}, true},
		{"hidden directory", &mockFileInfo{".hidden_directory", true}, true},
		{"hidden file", &mockFileInfo{".hidden_file", false}, true},
		{"file with no extension", &mockFileInfo{"file_with_no_extension", false}, true},
		{"file with dots", &mockFileInfo{"file.md.with.dots", false}, true},
		{"markdown file with dots", &mockFileInfo{"file.with.multiple.dots.md", false}, false},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := note.ShouldSkipDirectoryOrFile(tt.info)
			assert.Equal(t, tt.want, got)
		})
	}
}
