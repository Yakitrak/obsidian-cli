package actions

import (
	"reflect"
	"testing"

	"github.com/Yakitrak/obsidian-cli/mocks"
)

func TestTags(t *testing.T) {
	tests := []struct {
		name          string
		notes         map[string]string // notePath -> content
		expected      []TagSummary
		expectedError bool
		filterNotes   []string
	}{
		{
			name: "hierarchical tags example from user",
			notes: map[string]string{
				"note1.md":  "---\ntags: [a]\n---\nContent",
				"note2.md":  "#a/b",
				"note3.md":  "#a/b",
				"note4.md":  "#a/b\n#a/c",
				"note5.md":  "---\ntags: [a/c/c]\n---\nContent",
				"note6.md":  "#a/c/d",
				"note7.md":  "#a/c/d",
				"note8.md":  "#b",
				"note9.md":  "#b",
				"note10.md": "#b",
				"note11.md": "#b",
				"note12.md": "#b",
			},
			expected: []TagSummary{
				{Name: "a", IndividualCount: 1, AggregateCount: 7},
				{Name: "a/c", IndividualCount: 1, AggregateCount: 4},
				{Name: "a/c/d", IndividualCount: 2, AggregateCount: 2},
				{Name: "a/c/c", IndividualCount: 1, AggregateCount: 1},
				{Name: "a/b", IndividualCount: 3, AggregateCount: 3},
				{Name: "b", IndividualCount: 5, AggregateCount: 5},
			},
		},
		{
			name: "mixed frontmatter and hashtags",
			notes: map[string]string{
				"note1.md": "---\ntags: [project, work/meeting]\n---\nContent here",
				"note2.md": "#project/planning\nSome content",
				"note3.md": "---\ntags: work\n---\n\n#personal",
			},
			expected: []TagSummary{
				{Name: "project", IndividualCount: 1, AggregateCount: 2},
				{Name: "project/planning", IndividualCount: 1, AggregateCount: 1},
				{Name: "work", IndividualCount: 1, AggregateCount: 2},
				{Name: "work/meeting", IndividualCount: 1, AggregateCount: 1},
				{Name: "personal", IndividualCount: 1, AggregateCount: 1},
			},
		},
		{
			name: "case insensitive tags",
			notes: map[string]string{
				"note1.md": "#Project",
				"note2.md": "---\ntags: [PROJECT]\n---\nContent",
				"note3.md": "#project/Planning",
			},
			expected: []TagSummary{
				{Name: "project", IndividualCount: 2, AggregateCount: 3},
				{Name: "project/planning", IndividualCount: 1, AggregateCount: 1},
			},
		},
		{
			name:     "no tags",
			notes:    map[string]string{"note1.md": "Just content, no tags"},
			expected: []TagSummary{},
		},
		{
			name: "subset of notes",
			notes: map[string]string{
				"note1.md": "#a",
				"note2.md": "#a\n#b",
			},
			expected: []TagSummary{
				{Name: "a", IndividualCount: 1, AggregateCount: 1},
				{Name: "b", IndividualCount: 1, AggregateCount: 1},
			},
			filterNotes: []string{"note2.md"},
		},
		{
			name: "invalid tags filtered out",
			notes: map[string]string{
				"note1.md": "---\ntags: [\"1\", \"327\", \"person person\", \"valid-tag\"]\n---\nContent with #2 and #project tags.",
			},
			expected: []TagSummary{
				{Name: "project", IndividualCount: 1, AggregateCount: 1},
				{Name: "valid-tag", IndividualCount: 1, AggregateCount: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock managers
			vaultManager := &mocks.VaultManager{}
			noteManager := &mocks.NoteManager{}

			// Set up mock vault path
			vaultManager.On("Path").Return("/mock/vault", nil)

			// Set up mock notes list when not explicitly filtered
			if len(tt.filterNotes) == 0 {
				var notesList []string
				for notePath := range tt.notes {
					notesList = append(notesList, notePath)
				}
				noteManager.On("GetNotesList", "/mock/vault").Return(notesList, nil)
			}

			// Set up mock note contents
			if len(tt.filterNotes) > 0 {
				for _, notePath := range tt.filterNotes {
					if content, ok := tt.notes[notePath]; ok {
						noteManager.On("GetContents", "/mock/vault", notePath).Return(content, nil)
					}
				}
			} else {
				for notePath, content := range tt.notes {
					noteManager.On("GetContents", "/mock/vault", notePath).Return(content, nil)
				}
			}

			// Run the function
			var opts []TagsOptions
			if len(tt.filterNotes) > 0 {
				opts = append(opts, TagsOptions{Notes: tt.filterNotes})
			}
			result, err := Tags(vaultManager, noteManager, opts...)

			// Check error expectation
			if tt.expectedError && err == nil {
				t.Errorf("expected error but got none")
				return
			}
			if !tt.expectedError && err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Compare results
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Tags() result mismatch\nExpected: %+v\nActual:   %+v", tt.expected, result)

				// Print detailed comparison for debugging
				t.Logf("Expected %d tags, got %d", len(tt.expected), len(result))
				for i, expected := range tt.expected {
					if i < len(result) {
						actual := result[i]
						if expected != actual {
							t.Logf("  Tag %d: expected %+v, got %+v", i, expected, actual)
						}
					} else {
						t.Logf("  Tag %d: expected %+v, got <missing>", i, expected)
					}
				}
				for i := len(tt.expected); i < len(result); i++ {
					t.Logf("  Tag %d: expected <missing>, got %+v", i, result[i])
				}
			}

			// Verify mock expectations
			vaultManager.AssertExpectations(t)
			noteManager.AssertExpectations(t)
		})
	}
}

func TestGetAllPrefixes(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		expected []string
	}{
		{
			name:     "single level tag",
			tag:      "project",
			expected: []string{"project"},
		},
		{
			name:     "two level tag",
			tag:      "project/planning",
			expected: []string{"project", "project/planning"},
		},
		{
			name:     "three level tag",
			tag:      "work/project/planning",
			expected: []string{"work", "work/project", "work/project/planning"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getAllPrefixes(tt.tag)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("getAllPrefixes(%q) = %v, expected %v", tt.tag, result, tt.expected)
			}
		})
	}
}

func TestNormalizeTag(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		expected string
	}{
		{
			name:     "uppercase tag",
			tag:      "PROJECT",
			expected: "project",
		},
		{
			name:     "mixed case tag",
			tag:      "ProJeCt",
			expected: "project",
		},
		{
			name:     "tag with spaces",
			tag:      "  project  ",
			expected: "project",
		},
		{
			name:     "hierarchical tag",
			tag:      "Work/Project/Planning",
			expected: "work/project/planning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeTag(tt.tag)
			if result != tt.expected {
				t.Errorf("normalizeTag(%q) = %q, expected %q", tt.tag, result, tt.expected)
			}
		})
	}
}

func TestIsValidTag(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		expected bool
	}{
		// Valid tags
		{
			name:     "simple tag",
			tag:      "project",
			expected: true,
		},
		{
			name:     "hierarchical tag",
			tag:      "work/project",
			expected: true,
		},
		{
			name:     "tag with numbers and letters",
			tag:      "project2024",
			expected: true,
		},
		{
			name:     "tag with hyphens",
			tag:      "career-pathing",
			expected: true,
		},
		{
			name:     "tag with underscores",
			tag:      "career_pathing",
			expected: true,
		},

		// Invalid tags
		{
			name:     "empty tag",
			tag:      "",
			expected: false,
		},
		{
			name:     "purely numeric tag",
			tag:      "123",
			expected: false,
		},
		{
			name:     "single digit",
			tag:      "1",
			expected: false,
		},
		{
			name:     "tag with spaces",
			tag:      "person person",
			expected: false,
		},
		{
			name:     "tag with leading space",
			tag:      " project",
			expected: false,
		},
		{
			name:     "tag with trailing space",
			tag:      "project ",
			expected: false,
		},
		{
			name:     "tag with internal spaces",
			tag:      "my project",
			expected: false,
		},
		{
			name:     "purely symbols",
			tag:      "---",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidTag(tt.tag)
			if result != tt.expected {
				t.Errorf("isValidTag(%q) = %v, expected %v", tt.tag, result, tt.expected)
			}
		})
	}
}

func TestExtractAllTags(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "frontmatter and hashtags",
			content:  "---\ntags: [project, work]\n---\n\n# Content\n\nSome text with #personal and #project/planning tags.",
			expected: []string{"project", "work", "personal", "project/planning"},
		},
		{
			name:     "only hashtags",
			content:  "# Content\n\nThis has #tag1 and #tag2/subtag.",
			expected: []string{"tag1", "tag2/subtag"},
		},
		{
			name:     "only frontmatter",
			content:  "---\ntags: [solo-tag]\n---\n\nContent without hashtags.",
			expected: []string{"solo-tag"},
		},
		{
			name:     "no tags",
			content:  "Just plain content without any tags.",
			expected: []string{},
		},
		{
			name:     "hashtags with leading whitespace",
			content:  "Content with #coaching and #development-practice tags.",
			expected: []string{"coaching", "development-practice"},
		},
		{
			name:     "frontmatter tags with hashtag prefix",
			content:  "---\ntags:\n  - \"#career-pathing/feedback-syncs\"\n---\n\n# Content here",
			expected: []string{"career-pathing/feedback-syncs"},
		},
		{
			name:     "mixed frontmatter and hashtag with same tag",
			content:  "---\ntags: [\"#coaching\"]\n---\n\nContent with #coaching mentioned.",
			expected: []string{"coaching", "coaching"}, // extractAllTags doesn't deduplicate
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractAllTags(tt.content)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("extractAllTags() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
