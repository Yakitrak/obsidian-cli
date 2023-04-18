package note_test

import (
	"github.com/Yakitrak/obsidian-cli/pkg/note"
	"github.com/stretchr/testify/assert"
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
