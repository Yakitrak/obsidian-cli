package obsidian_test

import (
	"github.com/Yakitrak/obsidian-cli/mocks"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
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
			got := obsidian.AddMdSuffix(tt.input)
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
			got := obsidian.RemoveMdSuffix(tt.input)
			// Assert
			assert.Equal(t, tt.expected, got)
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
		expected     []byte
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
			assert.Equal(t, string(test.expected), string(got))
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
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			// Act
			got := obsidian.ShouldSkipDirectoryOrFile(tt.info)
			// Assert
			assert.Equal(t, tt.want, got)
		})
	}
}
