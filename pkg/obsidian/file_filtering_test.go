package obsidian

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFuzzyMatch(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		path     string
		expected bool
	}{
		// Single character directory matching
		{
			name:     "single character directory match",
			pattern:  "l/",
			path:     "Log/Some file.md",
			expected: true,
		},
		{
			name:     "single character directory no match",
			pattern:  "n/",
			path:     "Log/Some file.md",
			expected: false,
		},

		// Exact directory matching
		{
			name:     "exact directory match",
			pattern:  "log/",
			path:     "Log/Some file.md",
			expected: true,
		},
		{
			name:     "directory with spaces match",
			pattern:  "my notes/",
			path:     "My Notes/Some file.md",
			expected: true,
		},

		// Content word matching
		{
			name:     "content word match",
			pattern:  "log/sync",
			path:     "Log/Sync with team.md",
			expected: true,
		},
		{
			name:     "content word no match",
			pattern:  "log/sync",
			path:     "Log/Meeting notes.md",
			expected: false,
		},

		// Multiple word content matching
		{
			name:     "multiple words match in order",
			pattern:  "log/sync joe",
			path:     "Log/Sync with Joe.md",
			expected: true,
		},
		{
			name:     "multiple words match out of order",
			pattern:  "log/sync joe",
			path:     "Log/Joe Sync.md",
			expected: false,
		},
		{
			name:     "multiple words partial match",
			pattern:  "log/sync joe",
			path:     "Log/Sync with team.md",
			expected: false,
		},

		// Content-only matching (no slash)
		{
			name:     "content word match without slash",
			pattern:  "joe",
			path:     "Log/Sync with Joe.md",
			expected: true,
		},
		{
			name:     "content word match in directory",
			pattern:  "log",
			path:     "Log/Some file.md",
			expected: true,
		},
		{
			name:     "multiple words match without slash",
			pattern:  "sync joe",
			path:     "Log/Sync with Joe.md",
			expected: true,
		},
		{
			name:     "multiple words no match without slash",
			pattern:  "sync team",
			path:     "Log/Sync with Joe.md",
			expected: false,
		},

		// Word boundary matching
		{
			name:     "word start match",
			pattern:  "i/s joe",
			path:     "Inbox/Start of Joe's file.md",
			expected: true,
		},
		{
			name:     "word start no match mid-word",
			pattern:  "i/s joe",
			path:     "Inbox/Decision not to hire Joey.md",
			expected: false,
		},
		{
			name:     "word start match with hyphen",
			pattern:  "i/s-t joe",
			path:     "Inbox/Start-Time with Joe.md",
			expected: true,
		},
		{
			name:     "word start match with underscore",
			pattern:  "i/s_t joe",
			path:     "Inbox/Start_Time with Joe.md",
			expected: true,
		},
		{
			name:     "word start match with numbers",
			pattern:  "i/s2 joe",
			path:     "Inbox/S2 meeting with Joe.md",
			expected: true,
		},
		{
			name:     "word start no match for mid-word numbers",
			pattern:  "i/s2 joe",
			path:     "Inbox/PS2 with Joe.md",
			expected: false,
		},

		// Edge cases
		{
			name:     "empty pattern",
			pattern:  "",
			path:     "Log/Some file.md",
			expected: false,
		},
		{
			name:     "empty path",
			pattern:  "log/",
			path:     "",
			expected: false,
		},
		{
			name:     "path with no slash",
			pattern:  "log/",
			path:     "Some file.md",
			expected: false,
		},
		{
			name:     "case insensitive match",
			pattern:  "LOG/SYNC",
			path:     "Log/Sync with team.md",
			expected: true,
		},
		{
			name:     "multiple slashes in pattern",
			pattern:  "log/sync/meeting",
			path:     "Log/Sync meeting.md",
			expected: false,
		},
		{
			name:     "multiple slashes in path",
			pattern:  "log/sync",
			path:     "Log/Team/Sync.md",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FuzzyMatch(tt.pattern, tt.path)
			assert.Equal(t, tt.expected, result, "pattern: %s, path: %s", tt.pattern, tt.path)
		})
	}
}

func TestSplitWords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple words",
			input:    "hello world",
			expected: []string{"hello", "world"},
		},
		{
			name:     "with hyphens",
			input:    "hello-world example-test",
			expected: []string{"hello", "world", "example", "test"},
		},
		{
			name:     "with underscores",
			input:    "hello_world example_test",
			expected: []string{"hello", "world", "example", "test"},
		},
		{
			name:     "mixed separators",
			input:    "hello-world_example test_case",
			expected: []string{"hello", "world", "example", "test", "case"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "single word",
			input:    "word",
			expected: []string{"word"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitWords(tt.input)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestIsWordBoundary(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		pos      int
		expected bool
	}{
		{
			name:     "start of string",
			text:     "hello",
			pos:      0,
			expected: true,
		},
		{
			name:     "after space",
			text:     "hello world",
			pos:      6,
			expected: true,
		},
		{
			name:     "after hyphen",
			text:     "hello-world",
			pos:      6,
			expected: true,
		},
		{
			name:     "after underscore",
			text:     "hello_world",
			pos:      6,
			expected: true,
		},
		{
			name:     "middle of word",
			text:     "hello",
			pos:      3,
			expected: false,
		},
		{
			name:     "end of string",
			text:     "hello",
			pos:      5,
			expected: false,
		},
		{
			name:     "beyond end of string",
			text:     "hello",
			pos:      10,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isWordBoundary(tt.text, tt.pos)
			assert.Equal(t, tt.expected, result)
		})
	}
}