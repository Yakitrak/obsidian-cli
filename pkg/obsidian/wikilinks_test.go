package obsidian

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockNoteManager is a mock implementation of NoteManager for testing
type MockNoteManager struct {
	mock.Mock
}

func (m *MockNoteManager) Move(originalPath string, newPath string) error {
	args := m.Called(originalPath, newPath)
	return args.Error(0)
}

func (m *MockNoteManager) Delete(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

func (m *MockNoteManager) UpdateLinks(vaultPath string, oldNoteName string, newNoteName string) error {
	args := m.Called(vaultPath, oldNoteName, newNoteName)
	return args.Error(0)
}

func (m *MockNoteManager) GetContents(vaultPath, notePath string) (string, error) {
	args := m.Called(vaultPath, notePath)
	return args.String(0), args.Error(1)
}

func (m *MockNoteManager) GetNotesList(vaultPath string) ([]string, error) {
	args := m.Called(vaultPath)
	return args.Get(0).([]string), args.Error(1)
}

func TestExtractWikilinks(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name:    "multiple wikilinks",
			content: "Link to [[Project]] and [[Todo List]]",
			want:    []string{"Project", "Todo List"},
		},
		{
			name:    "no wikilinks",
			content: "Just regular text",
			want:    []string{},
		},
		{
			name:    "wikilink with spaces",
			content: "Link to [[Project Notes]] and [[Todo List 2024]]",
			want:    []string{"Project Notes", "Todo List 2024"},
		},
		{
			name:    "wikilink with path",
			content: "Link to [[folder/Project]] and [[subfolder/Todo List]]",
			want:    []string{"folder/Project", "subfolder/Todo List"},
		},
		{
			name:    "wikilink with alias",
			content: "Link to [[Project|My Project]] and [[Todo List|Tasks]]",
			want:    []string{"Project", "Todo List"},
		},
		{
			name:    "wikilink with both path and alias",
			content: "Link to [[folder/Project|My Project]] and [[subfolder/Todo List|Tasks]]",
			want:    []string{"folder/Project", "subfolder/Todo List"},
		},
		{
			name:    "wikilinks in code block",
			content: "```\n[[Project]]\n```\nOutside [[Real Link]]",
			want:    []string{"Project", "Real Link"}, // Note: Currently extracts from code blocks too
		},
		{
			name:    "wikilinks with file extension",
			content: "Link to [[Project.md]] and [[Todo List.md]]",
			want:    []string{"Project.md", "Todo List.md"},
		},
		{
			name:    "wikilinks with heading",
			content: "Link to [[Project#section]] and [[Todo List#details]]",
			want:    []string{"Project#section", "Todo List#details"},
		},
		{
			name:    "wikilinks with special characters",
			content: "Link to [[Project-2023]] and [[Todo_List]]",
			want:    []string{"Project-2023", "Todo_List"},
		},
		{
			name:    "wikilinks in complex text",
			content: "# Header\n\nParagraph with [[Link1]] and [[Link2]]\n\n> Quote with [[Link3]]\n\n- List item with [[Link4]]\n",
			want:    []string{"Link1", "Link2", "Link3", "Link4"},
		},
		{
			name:    "nested wikilinks don't exist in Obsidian",
			content: "Link to [[Outer [[Inner]]]]",
			want:    []string{"Outer [[Inner"}, // This is expected behavior in Obsidian
		},
		{
			name:    "adjacent wikilinks",
			content: "Link to [[Link1]][[Link2]]",
			want:    []string{"Link1", "Link2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractWikilinks(tt.content, DefaultWikilinkOptions)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestBuildNotePathCache(t *testing.T) {
	tests := []struct {
		name         string
		notes        []string
		expectedPath string
		lookupKey    string
	}{
		{
			name:         "basic path resolution",
			notes:        []string{"folder/note1.md", "folder/note2.md"},
			lookupKey:    "note1",
			expectedPath: "folder/note1.md",
		},
		{
			name:         "resolve with full path",
			notes:        []string{"folder/note1.md", "subfolder/note1.md"},
			lookupKey:    "folder/note1",
			expectedPath: "folder/note1.md",
		},
		{
			name:         "duplicate filename with shorter path wins",
			notes:        []string{"folder/subfolder/note1.md", "folder/note1.md"},
			lookupKey:    "note1",
			expectedPath: "folder/note1.md", // shorter path wins
		},
		{
			name:         "note with spaces",
			notes:        []string{"folder/My Note.md", "folder/Other Note.md"},
			lookupKey:    "My Note",
			expectedPath: "folder/My Note.md",
		},
		{
			name:         "lookup with extension",
			notes:        []string{"folder/note1.md", "folder/note2.md"},
			lookupKey:    "note1.md",
			expectedPath: "folder/note1.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := BuildNotePathCache(tt.notes)

			// Verify specific path resolution
			path, exists := cache.ResolveNote(tt.lookupKey)
			assert.True(t, exists, "Note %s should exist in cache", tt.lookupKey)
			assert.Equal(t, tt.expectedPath, path)
		})
	}
}

func TestNotePathCacheResolve(t *testing.T) {
	tests := []struct {
		name         string
		cachePaths   map[string]string
		lookupKey    string
		expectedPath string
		shouldExist  bool
	}{
		{
			name: "exact match",
			cachePaths: map[string]string{
				"note1":        "folder/note1.md",
				"folder/note1": "folder/note1.md",
			},
			lookupKey:    "note1",
			expectedPath: "folder/note1.md",
			shouldExist:  true,
		},
		{
			name: "match with path",
			cachePaths: map[string]string{
				"note1":        "folder/note1.md",
				"folder/note1": "folder/note1.md",
			},
			lookupKey:    "folder/note1",
			expectedPath: "folder/note1.md",
			shouldExist:  true,
		},
		{
			name: "match by basename when path not found",
			cachePaths: map[string]string{
				"note1":        "folder/note1.md",
				"folder/note1": "folder/note1.md",
			},
			lookupKey:    "subfolder/note1",
			expectedPath: "folder/note1.md",
			shouldExist:  true,
		},
		{
			name: "non-existent note",
			cachePaths: map[string]string{
				"note1":        "folder/note1.md",
				"folder/note1": "folder/note1.md",
			},
			lookupKey:    "note2",
			expectedPath: "",
			shouldExist:  false,
		},
		{
			name: "with extension in lookup",
			cachePaths: map[string]string{
				"note1":        "folder/note1.md",
				"folder/note1": "folder/note1.md",
			},
			lookupKey:    "note1.md",
			expectedPath: "folder/note1.md",
			shouldExist:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := &NotePathCache{
				Paths: tt.cachePaths,
			}

			path, exists := cache.ResolveNote(tt.lookupKey)
			assert.Equal(t, tt.shouldExist, exists)
			if tt.shouldExist {
				assert.Equal(t, tt.expectedPath, path)
			}
		})
	}
}

func TestFollowWikilinks(t *testing.T) {
	// Create test mocks
	mockNote := &MockNoteManager{}

	// Setup vault path
	vaultPath := "/test/vault"

	// Notes list mock - used for building the cache
	mockNote.On("GetNotesList", vaultPath).Return([]string{
		"note1.md",
		"note2.md",
		"folder/note3.md",
		"folder/note4.md",
	}, nil)

	// Content mocks - different wikilink structures
	mockNote.On("GetContents", vaultPath, "note1.md").Return("Content with link to [[note2]]", nil)
	mockNote.On("GetContents", vaultPath, "note2.md").Return("Content with link to [[folder/note3]]", nil)
	mockNote.On("GetContents", vaultPath, "folder/note3.md").Return("Content with link to [[note4]]", nil)
	mockNote.On("GetContents", vaultPath, "folder/note4.md").Return("Content with link to [[note1]] and [[non-existent]]", nil)

	tests := []struct {
		name        string
		startFile   string
		maxDepth    int
		expected    []string
		expectedErr bool
	}{
		{
			name:        "follow one level",
			startFile:   "note1.md",
			maxDepth:    1,
			expected:    []string{"note1.md", "note2.md"},
			expectedErr: false,
		},
		{
			name:        "follow two levels",
			startFile:   "note1.md",
			maxDepth:    2,
			expected:    []string{"note1.md", "note2.md", "folder/note3.md"},
			expectedErr: false,
		},
		{
			name:        "follow all levels",
			startFile:   "note1.md",
			maxDepth:    3,
			expected:    []string{"note1.md", "note2.md", "folder/note3.md", "folder/note4.md"},
			expectedErr: false,
		},
		{
			name:        "start from middle",
			startFile:   "note2.md",
			maxDepth:    2,
			expected:    []string{"note2.md", "folder/note3.md", "folder/note4.md"},
			expectedErr: false,
		},
		{
			name:        "depth 0 returns only starting file",
			startFile:   "note1.md",
			maxDepth:    0,
			expected:    []string{"note1.md"},
			expectedErr: false,
		},
		{
			name:        "handle circular references",
			startFile:   "folder/note4.md",
			maxDepth:    3,
			expected:    []string{"folder/note4.md", "note1.md", "note2.md", "folder/note3.md"},
			expectedErr: false,
		},
	}

	// Get all notes to build the cache
	allNotes, _ := mockNote.GetNotesList(vaultPath)
	cache := BuildNotePathCache(allNotes)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			visited := make(map[string]bool)
			options := FollowWikilinksOptions{
				WikilinkOptions: DefaultWikilinkOptions,
				MaxDepth:        tt.maxDepth,
			}
			result, err := FollowWikilinks(vaultPath, mockNote, tt.startFile, visited, cache, options)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tt.expected, result)
			}
		})
	}

	// Test error case
	t.Run("error when getting content", func(t *testing.T) {
		mockNote := &MockNoteManager{}
		mockNote.On("GetContents", vaultPath, "error.md").Return("", errors.New("content error"))

		visited := make(map[string]bool)
		options := DefaultFollowWikilinksOptions
		result, err := FollowWikilinks(vaultPath, mockNote, "error.md", visited, cache, options)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestFollowWikilinksWithOptions(t *testing.T) {
	// Create test mocks
	mockNote := &MockNoteManager{}

	// Setup vault path
	vaultPath := "/test/vault"

	// Notes list mock - used for building the cache
	mockNote.On("GetNotesList", vaultPath).Return([]string{
		"note1.md",
		"note2.md",
		"note3.md",
		"folder/note4.md",
		"folder/note5.md",
	}, nil)

	// Content mocks - with some anchored links
	mockNote.On("GetContents", vaultPath, "note1.md").Return("Content with link to [[note2]] and anchored [[note3#section]]", nil)
	mockNote.On("GetContents", vaultPath, "note2.md").Return("Content with link to [[folder/note4#details]] and [[folder/note5]]", nil)
	mockNote.On("GetContents", vaultPath, "note3.md").Return("Content with link to [[note1]]", nil)
	mockNote.On("GetContents", vaultPath, "folder/note4.md").Return("Content with no links", nil)
	mockNote.On("GetContents", vaultPath, "folder/note5.md").Return("Content with link back to [[note2]]", nil)

	// Get all notes to build the cache
	allNotes, _ := mockNote.GetNotesList(vaultPath)
	cache := BuildNotePathCache(allNotes)

	t.Run("follow with skipAnchors=true", func(t *testing.T) {
		visited := make(map[string]bool)
		options := FollowWikilinksOptions{
			WikilinkOptions: WikilinkOptions{
				SkipAnchors: true,
				SkipEmbeds:  false,
			},
			MaxDepth: 3,
		}
		result, err := FollowWikilinks(vaultPath, mockNote, "note1.md", visited, cache, options)

		assert.NoError(t, err)
		expected := []string{"note1.md", "note2.md", "folder/note5.md"}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("follow with skipAnchors=false", func(t *testing.T) {
		visited := make(map[string]bool)
		options := FollowWikilinksOptions{
			WikilinkOptions: WikilinkOptions{
				SkipAnchors: false,
				SkipEmbeds:  false,
			},
			MaxDepth: 3,
		}
		result, err := FollowWikilinks(vaultPath, mockNote, "note1.md", visited, cache, options)

		assert.NoError(t, err)
		// When skipAnchors=false, we follow all links including anchored ones
		expected := []string{"note1.md", "note2.md", "note3.md", "folder/note4.md", "folder/note5.md"}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("follow with unlimited depth", func(t *testing.T) {
		visited := make(map[string]bool)
		options := FollowWikilinksOptions{
			WikilinkOptions: DefaultWikilinkOptions,
			MaxDepth:        -1,
		}
		result, err := FollowWikilinks(vaultPath, mockNote, "note1.md", visited, cache, options)

		assert.NoError(t, err)
		expected := []string{"note1.md", "note2.md", "note3.md", "folder/note4.md", "folder/note5.md"}
		assert.ElementsMatch(t, expected, result)
	})
}

func TestExtractWikilinksOptions(t *testing.T) {
	const content = "Link to [[Regular Link]], anchored [[Anchored Link#section]], and embed ![[Embedded Link]]"

	tests := []struct {
		name    string
		options WikilinkOptions
		want    []string
	}{
		{
			name:    "extract all links",
			options: WikilinkOptions{SkipAnchors: false, SkipEmbeds: false},
			want:    []string{"Regular Link", "Anchored Link#section", "Embedded Link"},
		},
		{
			name:    "skip anchors",
			options: WikilinkOptions{SkipAnchors: true, SkipEmbeds: false},
			want:    []string{"Regular Link", "Embedded Link"},
		},
		{
			name:    "skip embeds",
			options: WikilinkOptions{SkipAnchors: false, SkipEmbeds: true},
			want:    []string{"Regular Link", "Anchored Link#section"},
		},
		{
			name:    "skip both anchors and embeds",
			options: WikilinkOptions{SkipAnchors: true, SkipEmbeds: true},
			want:    []string{"Regular Link"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractWikilinks(content, tt.options)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestDeduplicateResults(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "no duplicates",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "with duplicates",
			input:    []string{"a", "b", "a", "c", "b"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "empty list",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "all duplicates",
			input:    []string{"a", "a", "a"},
			expected: []string{"a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DeduplicateResults(tt.input)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "already normalized",
			input:    "folder/note.md",
			expected: "folder/note.md",
		},
		{
			name:     "Windows path",
			input:    "folder\\note.md",
			expected: "folder/note.md",
		},
		{
			name:     "with leading ./",
			input:    "./folder/note.md",
			expected: "folder/note.md",
		},
		{
			name:     "with leading ../",
			input:    "../folder/note.md",
			expected: "folder/note.md",
		},
		{
			name:     "mixed path separators",
			input:    "folder\\subfolder/note.md",
			expected: "folder/subfolder/note.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizePath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
