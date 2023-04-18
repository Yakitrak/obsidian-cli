package note_test

import (
	"github.com/Yakitrak/obsidian-cli/utils/note"
	"testing"
)

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
