package note_test

import (
	"github.com/Yakitrak/obsidian-cli/utils/note"
	"github.com/stretchr/testify/assert"
	"testing"
)

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
