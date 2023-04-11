package utils_test

import (
	"github.com/Yakitrak/obsidian-cli/utils"
	"testing"
)

func TestAddMdSuffix(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{input: "myNote", expected: "myNote.md"},
		{input: "myNote.md", expected: "myNote.md"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			got := utils.AddMdSuffix(tc.input)
			if got != tc.expected {
				t.Errorf("got %s, want %s", got, tc.expected)
			}
		})
	}
}
