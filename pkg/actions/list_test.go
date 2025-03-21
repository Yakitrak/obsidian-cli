package actions

import (
	"errors"
	"strings"
	"testing"

	"github.com/Yakitrak/obsidian-cli/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestParseInputs tests the ParseInputs function
func TestParseInputs(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedInputs []ListInput
		expectedError  bool
		errorContains  string
	}{
		{
			name:           "Empty args",
			args:           []string{},
			expectedInputs: nil,
			expectedError:  false,
		},
		{
			name: "Single file path",
			args: []string{"path/to/file.md"},
			expectedInputs: []ListInput{
				{
					Type:  InputTypeFile,
					Value: "path/to/file.md",
				},
			},
			expectedError: false,
		},
		{
			name: "Multiple file paths",
			args: []string{"path/to/file1.md", "path/to/file2.md"},
			expectedInputs: []ListInput{
				{
					Type:  InputTypeFile,
					Value: "path/to/file1.md",
				},
				{
					Type:  InputTypeFile,
					Value: "path/to/file2.md",
				},
			},
			expectedError: false,
		},
		{
			name: "Single tag",
			args: []string{"tag:project"},
			expectedInputs: []ListInput{
				{
					Type:  InputTypeTag,
					Value: "project",
				},
			},
			expectedError: false,
		},
		{
			name: "Tag with quotes",
			args: []string{`tag:"multi word"`},
			expectedInputs: []ListInput{
				{
					Type:  InputTypeTag,
					Value: "multi word",
				},
			},
			expectedError: false,
		},
		{
			name: "Tag with single quotes",
			args: []string{"tag:'multi word'"},
			expectedInputs: []ListInput{
				{
					Type:  InputTypeTag,
					Value: "'multi word'",
				},
			},
			expectedError: false,
		},
		{
			name: "Multiple tags",
			args: []string{"tag:project", "tag:work"},
			expectedInputs: []ListInput{
				{
					Type:  InputTypeTag,
					Value: "project",
				},
				{
					Type:  InputTypeTag,
					Value: "work",
				},
			},
			expectedError: false,
		},
		{
			name: "Single find",
			args: []string{"find:note"},
			expectedInputs: []ListInput{
				{
					Type:  InputTypeFind,
					Value: "note",
				},
			},
			expectedError: false,
		},
		{
			name: "Find with quotes",
			args: []string{`find:"complex query"`},
			expectedInputs: []ListInput{
				{
					Type:  InputTypeFind,
					Value: "complex query",
				},
			},
			expectedError: false,
		},
		{
			name: "Find with single quotes",
			args: []string{"find:'complex query'"},
			expectedInputs: []ListInput{
				{
					Type:  InputTypeFind,
					Value: "'complex query'",
				},
			},
			expectedError: false,
		},
		{
			name: "Mixed inputs",
			args: []string{"path/to/file.md", "tag:project", "find:note"},
			expectedInputs: []ListInput{
				{
					Type:  InputTypeFile,
					Value: "path/to/file.md",
				},
				{
					Type:  InputTypeTag,
					Value: "project",
				},
				{
					Type:  InputTypeFind,
					Value: "note",
				},
			},
			expectedError: false,
		},
		{
			name: "Special characters in inputs",
			args: []string{"path/to/file-with-dashes.md", "tag:project-2023", "find:note_123"},
			expectedInputs: []ListInput{
				{
					Type:  InputTypeFile,
					Value: "path/to/file-with-dashes.md",
				},
				{
					Type:  InputTypeTag,
					Value: "project-2023",
				},
				{
					Type:  InputTypeFind,
					Value: "note_123",
				},
			},
			expectedError: false,
		},
		{
			name:          "Empty tag value",
			args:          []string{"tag:"},
			expectedError: true,
			errorContains: "tag cannot be empty",
		},
		{
			name:          "Empty find value",
			args:          []string{"find:"},
			expectedError: true,
			errorContains: "find cannot be empty",
		},
		{
			name:          "Tag with wildcard",
			args:          []string{"tag:*"},
			expectedError: true,
			errorContains: "tag cannot be empty or a wildcard",
		},
		{
			name:          "Find with wildcard",
			args:          []string{"find:*"},
			expectedError: true,
			errorContains: "find cannot be empty or a wildcard",
		},
		{
			name: "Malformed tag and find without colon",
			args: []string{"tagproject", "findnote"},
			expectedInputs: []ListInput{
				{
					Type:  InputTypeFile,
					Value: "tagproject",
				},
				{
					Type:  InputTypeFile,
					Value: "findnote",
				},
			},
			expectedError: false,
		},
		{
			name: "Wildcard in file path",
			args: []string{"*"},
			expectedInputs: []ListInput{
				{
					Type:  InputTypeFile,
					Value: "*",
				},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputs, err := ParseInputs(tt.args)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedInputs, inputs)
			}
		})
	}
}

// MockVaultManager is a mock implementation of VaultManager
type MockVaultManager struct {
	mock.Mock
}

func (m *MockVaultManager) Path() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockVaultManager) DefaultName() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockVaultManager) SetDefaultName(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

// MockNoteManager is a mock implementation of NoteManager
type MockNoteManager struct {
	mock.Mock
}

func (m *MockNoteManager) GetContents(vaultPath, notePath string) (string, error) {
	args := m.Called(vaultPath, notePath)
	return args.String(0), args.Error(1)
}

func (m *MockNoteManager) GetNotesList(vaultPath string) ([]string, error) {
	args := m.Called(vaultPath)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockNoteManager) Delete(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

func (m *MockNoteManager) Move(originalPath, newPath string) error {
	args := m.Called(originalPath, newPath)
	return args.Error(0)
}

func (m *MockNoteManager) UpdateLinks(vaultPath, oldNoteName, newNoteName string) error {
	args := m.Called(vaultPath, oldNoteName, newNoteName)
	return args.Error(0)
}

func TestListFiles(t *testing.T) {
	tests := []struct {
		name             string
		mockVault        *mocks.VaultManager
		mockNote         *mocks.NoteManager
		params           ListParams
		expectedFiles    []string
		expectedErr      error
		setupMocks       func(*mocks.VaultManager, *mocks.NoteManager)
		validateResponse func(*testing.T, []string, error)
	}{
		{
			name:      "list with multiple input types",
			mockVault: &mocks.VaultManager{},
			mockNote:  &mocks.NoteManager{},
			params: ListParams{
				Inputs: []ListInput{
					{Type: InputTypeTag, Value: "tag1"},
					{Type: InputTypeFile, Value: "folder/note1.md"},
					{Type: InputTypeFind, Value: "note3"},
				},
			},
			expectedFiles: []string{"folder/note1.md", "note2.md", "note3.md", "note4.md"},
			setupMocks: func(v *mocks.VaultManager, n *mocks.NoteManager) {
				v.On("Path").Return("/test/vault", nil)
				n.On("GetNotesList", "/test/vault").Return([]string{"folder/note1.md", "note2.md", "note3.md", "note4.md"}, nil)
				n.On("GetContents", "/test/vault", "folder/note1.md").Return("Regular content", nil)
				n.On("GetContents", "/test/vault", "note2.md").Return("---\ntags: [tag1]\n---\nContent", nil)
				n.On("GetContents", "/test/vault", "note3.md").Return("Content", nil)
				n.On("GetContents", "/test/vault", "note4.md").Return("Content with #tag1", nil)
			},
			validateResponse: func(t *testing.T, files []string, err error) {
				assert.NoError(t, err)
				assert.ElementsMatch(t, []string{"folder/note1.md", "note2.md", "note3.md", "note4.md"}, files)
			},
		},
		{
			name:      "list with multiple tags",
			mockVault: &mocks.VaultManager{},
			mockNote:  &mocks.NoteManager{},
			params: ListParams{
				Inputs: []ListInput{
					{Type: InputTypeTag, Value: "tag1"},
					{Type: InputTypeTag, Value: "tag2"},
				},
			},
			expectedFiles: []string{"note1.md", "note2.md"},
			setupMocks: func(v *mocks.VaultManager, n *mocks.NoteManager) {
				v.On("Path").Return("/test/vault", nil)
				n.On("GetNotesList", "/test/vault").Return([]string{"note1.md", "note2.md", "note3.md"}, nil)
				n.On("GetContents", "/test/vault", "note1.md").Return("---\ntags: [tag1]\n---\nContent", nil)
				n.On("GetContents", "/test/vault", "note2.md").Return("Content with #tag2", nil)
				n.On("GetContents", "/test/vault", "note3.md").Return("Regular content", nil)
			},
			validateResponse: func(t *testing.T, files []string, err error) {
				assert.NoError(t, err)
				assert.ElementsMatch(t, []string{"note1.md", "note2.md"}, files)
			},
		},
		{
			name:      "list with quoted tag",
			mockVault: &mocks.VaultManager{},
			mockNote:  &mocks.NoteManager{},
			params: ListParams{
				Inputs: []ListInput{
					{Type: InputTypeTag, Value: "some-tag"},
				},
			},
			expectedFiles: []string{"note1.md"},
			setupMocks: func(v *mocks.VaultManager, n *mocks.NoteManager) {
				v.On("Path").Return("/test/vault", nil)
				n.On("GetNotesList", "/test/vault").Return([]string{"note1.md", "note2.md"}, nil)
				n.On("GetContents", "/test/vault", "note1.md").Return("---\ntags: [some-tag]\n---\nContent", nil)
				n.On("GetContents", "/test/vault", "note2.md").Return("Regular content", nil)
			},
			validateResponse: func(t *testing.T, files []string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, []string{"note1.md"}, files)
			},
		},
		{
			name:      "list with directory path",
			mockVault: &mocks.VaultManager{},
			mockNote:  &mocks.NoteManager{},
			params: ListParams{
				Inputs: []ListInput{
					{Type: InputTypeFile, Value: "folder"},
				},
			},
			expectedFiles: []string{"folder/note1.md", "folder/note2.md"},
			setupMocks: func(v *mocks.VaultManager, n *mocks.NoteManager) {
				v.On("Path").Return("/test/vault", nil)
				n.On("GetNotesList", "/test/vault").Return([]string{"folder/note1.md", "folder/note2.md", "other/note3.md"}, nil)
			},
			validateResponse: func(t *testing.T, files []string, err error) {
				assert.NoError(t, err)
				assert.ElementsMatch(t, []string{"folder/note1.md", "folder/note2.md"}, files)
			},
		},
		{
			name:      "list with no inputs",
			mockVault: &mocks.VaultManager{},
			mockNote:  &mocks.NoteManager{},
			params: ListParams{
				Inputs: []ListInput{},
			},
			expectedFiles: []string{},
			setupMocks: func(v *mocks.VaultManager, n *mocks.NoteManager) {
				v.On("Path").Return("/test/vault", nil)
				n.On("GetNotesList", "/test/vault").Return([]string{"note1.md", "note2.md"}, nil)
			},
			validateResponse: func(t *testing.T, files []string, err error) {
				assert.NoError(t, err)
				assert.Empty(t, files)
			},
		},
		{
			name:      "vault.Path returns error",
			mockVault: &mocks.VaultManager{},
			mockNote:  &mocks.NoteManager{},
			params: ListParams{
				Inputs: []ListInput{},
			},
			expectedErr: errors.New("Failed to get vault path"),
			setupMocks: func(v *mocks.VaultManager, n *mocks.NoteManager) {
				v.On("Path").Return("", errors.New("Failed to get vault path"))
			},
			validateResponse: func(t *testing.T, files []string, err error) {
				assert.Error(t, err)
				assert.Equal(t, "Failed to get vault path", err.Error())
				assert.Empty(t, files)
			},
		},
		{
			name:      "note.GetNotesList returns error",
			mockVault: &mocks.VaultManager{},
			mockNote:  &mocks.NoteManager{},
			params: ListParams{
				Inputs: []ListInput{},
			},
			expectedErr: errors.New("Failed to get notes list"),
			setupMocks: func(v *mocks.VaultManager, n *mocks.NoteManager) {
				v.On("Path").Return("/test/vault", nil)
				n.On("GetNotesList", "/test/vault").Return(nil, errors.New("Failed to get notes list"))
			},
			validateResponse: func(t *testing.T, files []string, err error) {
				assert.Error(t, err)
				assert.Equal(t, "Failed to get notes list", err.Error())
				assert.Empty(t, files)
			},
		},
		{
			name:      "list with wildcard pattern",
			mockVault: &mocks.VaultManager{},
			mockNote:  &mocks.NoteManager{},
			params: ListParams{
				Inputs: []ListInput{
					{Type: InputTypeFile, Value: "*"},
				},
			},
			expectedFiles: []string{"note1.md", "note2.md", "folder/note3.md"},
			setupMocks: func(v *mocks.VaultManager, n *mocks.NoteManager) {
				v.On("Path").Return("/test/vault", nil)
				n.On("GetNotesList", "/test/vault").Return([]string{"note1.md", "note2.md", "folder/note3.md"}, nil)
			},
			validateResponse: func(t *testing.T, files []string, err error) {
				assert.NoError(t, err)
				assert.ElementsMatch(t, []string{"note1.md", "note2.md", "folder/note3.md"}, files)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(tt.mockVault, tt.mockNote)

			files, err := ListFiles(tt.mockVault, tt.mockNote, tt.params)
			tt.validateResponse(t, files, err)

			tt.mockVault.AssertExpectations(t)
			tt.mockNote.AssertExpectations(t)
		})
	}
}

func TestListFilesWithFuzzySearch(t *testing.T) {
	tests := []struct {
		name     string
		inputs   []ListInput
		files    []string
		expected []string
	}{
		{
			name: "single character directory match",
			inputs: []ListInput{
				{Type: InputTypeFind, Value: "l/"},
			},
			files: []string{
				"Log/Sync with team.md",
				"Notes/Some file.md",
			},
			expected: []string{
				"Log/Sync with team.md",
			},
		},
		{
			name: "directory and content match",
			inputs: []ListInput{
				{Type: InputTypeFind, Value: "log/sync joe"},
			},
			files: []string{
				"Log/Sync with Joe.md",
				"Log/Meeting with Joe.md",
				"Notes/Sync with Joe.md",
			},
			expected: []string{
				"Log/Sync with Joe.md",
			},
		},
		{
			name: "multiple matches",
			inputs: []ListInput{
				{Type: InputTypeFind, Value: "log/sync"},
			},
			files: []string{
				"Log/Sync with team.md",
				"Log/Sync with Joe.md",
				"Notes/Sync.md",
			},
			expected: []string{
				"Log/Sync with team.md",
				"Log/Sync with Joe.md",
			},
		},
		{
			name: "combined tag and fuzzy search",
			inputs: []ListInput{
				{Type: InputTypeTag, Value: "meeting"},
				{Type: InputTypeFind, Value: "log/sync"},
			},
			files: []string{
				"Log/Sync meeting.md",   // Matches both tag and fuzzy search
				"Log/Sync with team.md", // Matches fuzzy search only
				"Notes/Sync meeting.md", // Matches tag only
			},
			expected: []string{
				"Log/Sync meeting.md",   // Include once even though it matches both
				"Log/Sync with team.md", // Matches fuzzy search
				"Notes/Sync meeting.md", // Matches tag
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock vault and note managers
			mockVault := &MockVaultManager{}
			mockNote := &MockNoteManager{}

			// Setup mock expectations
			mockVault.On("Path").Return("/test/vault", nil)
			mockNote.On("GetNotesList", "/test/vault").Return(tt.files, nil)

			// For tag tests, setup content expectations
			for _, input := range tt.inputs {
				if input.Type == InputTypeTag {
					for _, file := range tt.files {
						if strings.Contains(file, "meeting") {
							mockNote.On("GetContents", "/test/vault", file).Return("---\ntags: [meeting]\n---\n", nil).Once()
						} else {
							mockNote.On("GetContents", "/test/vault", file).Return("", nil).Once()
						}
					}
				}
			}

			// Run the test
			result, err := ListFiles(mockVault, mockNote, ListParams{Inputs: tt.inputs})

			// Verify results
			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.expected, result)

			// Verify all mock expectations were met
			mockVault.AssertExpectations(t)
			mockNote.AssertExpectations(t)
		})
	}
}

func TestListFilesWithWikilinks(t *testing.T) {
	tests := []struct {
		name         string
		files        []string
		fileContents map[string]string
		inputs       []ListInput
		followLinks  bool
		maxDepth     int
		skipAnchors  bool
		expected     []string
	}{
		{
			name:  "follow wikilinks disabled",
			files: []string{"note1.md", "note2.md", "note3.md", "note4.md"},
			fileContents: map[string]string{
				"note1.md": "Content with link to [[note2]]",
				"note2.md": "Content with #tag1",
				"note3.md": "Content with link to [[note4]]",
				"note4.md": "Regular content",
			},
			inputs: []ListInput{
				{Type: InputTypeTag, Value: "tag1"},
			},
			followLinks: false,
			maxDepth:    0,
			skipAnchors: false,
			expected:    []string{"note2.md"}, // only direct tag match
		},
		{
			name:  "with filepath input and follow links",
			files: []string{"folder/note1.md", "folder/note2.md", "note3.md", "note4.md"},
			fileContents: map[string]string{
				"folder/note1.md": "Content with link to [[note3]]",
				"folder/note2.md": "Content with link to [[note4]]",
				"note3.md":        "Content with #important",
				"note4.md":        "Regular content",
			},
			inputs: []ListInput{
				{Type: InputTypeFile, Value: "folder"},
			},
			followLinks: true,
			maxDepth:    1,
			skipAnchors: false,
			expected:    []string{"folder/note1.md", "folder/note2.md", "note3.md", "note4.md"},
		},
		{
			name:  "find input with depth traversal - checks callback behavior",
			files: []string{"weekly/2025-W12.md", "daily/2025-03-17.md", "daily/2025-03-18.md", "daily/2025-03-19.md"},
			fileContents: map[string]string{
				"weekly/2025-W12.md": "Weekly note with links to [[daily/2025-03-17]], [[daily/2025-03-18]], and [[daily/2025-03-19]]",
				"daily/2025-03-17.md": "Daily note for Monday",
				"daily/2025-03-18.md": "Daily note for Tuesday",
				"daily/2025-03-19.md": "Daily note for Wednesday",
			},
			inputs: []ListInput{
				{Type: InputTypeFind, Value: "2025-W12"},
			},
			followLinks: true,
			maxDepth:    2,
			skipAnchors: false,
			expected:    []string{"weekly/2025-W12.md", "daily/2025-03-17.md", "daily/2025-03-18.md", "daily/2025-03-19.md"},
		},
		{
			name:  "skip anchored links",
			files: []string{"project.md", "tasks.md", "section1.md", "section2.md"},
			fileContents: map[string]string{
				"project.md": "Project with links to [[tasks]] and sections [[section1#details]], [[section2#summary]]",
				"tasks.md":   "Tasks related to the project",
				"section1.md": "Section 1 details",
				"section2.md": "Section 2 summary",
			},
			inputs: []ListInput{
				{Type: InputTypeFile, Value: "project.md"},
			},
			followLinks: true,
			maxDepth:    1,
			skipAnchors: true,
			expected:    []string{"project.md", "tasks.md"}, // should not include section1.md or section2.md
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock vault and note managers
			mockVault := &MockVaultManager{}
			mockNote := &MockNoteManager{}

			// Setup vault path
			mockVault.On("Path").Return("/test/vault", nil)

			// Setup notes list
			mockNote.On("GetNotesList", "/test/vault").Return(tt.files, nil).Once()

			// If following links, we need another GetNotesList call
			if tt.followLinks {
				mockNote.On("GetNotesList", "/test/vault").Return(tt.files, nil).Once()
			}

			// Setup content expectations for each file - this needs to be very specific
			for _, input := range tt.inputs {
				if input.Type == InputTypeTag {
					for _, file := range tt.files {
						if strings.Contains(tt.fileContents[file], "#"+input.Value) {
							mockNote.On("GetContents", "/test/vault", file).Return(tt.fileContents[file], nil).Once()
						} else {
							mockNote.On("GetContents", "/test/vault", file).Return(tt.fileContents[file], nil).Maybe()
						}
					}
				}
			}

			if tt.followLinks {
				for file, content := range tt.fileContents {
					if !strings.Contains(content, "#") || // Always mock files without tags
						(len(tt.inputs) > 0 && tt.inputs[0].Type == InputTypeFile) { // Always mock when input is a file path
						mockNote.On("GetContents", "/test/vault", file).Return(content, nil).Maybe()
					}
				}
			}

			// Run the test
			result, err := ListFiles(mockVault, mockNote, ListParams{
				Inputs:      tt.inputs,
				FollowLinks: tt.followLinks,
				MaxDepth:    tt.maxDepth,
				SkipAnchors: tt.skipAnchors,
			})

			// Verify results
			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.expected, result)

			// Verify all mock expectations were met
			mockVault.AssertExpectations(t)
			mockNote.AssertExpectations(t)
		})
	}
}
