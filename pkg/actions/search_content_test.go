package actions_test

import (
	"errors"
	"os"
	"testing"

	"github.com/Yakitrak/obsidian-cli/mocks"
	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/stretchr/testify/assert"
)

// CustomMockNoteForSingleMatch returns exactly one match for editor testing
type CustomMockNoteForSingleMatch struct{}

func (m *CustomMockNoteForSingleMatch) Delete(string) error                       { return nil }
func (m *CustomMockNoteForSingleMatch) Move(string, string) error                  { return nil }
func (m *CustomMockNoteForSingleMatch) UpdateLinks(string, string, string) error   { return nil }
func (m *CustomMockNoteForSingleMatch) GetContents(string, string) (string, error) { return "", nil }
func (m *CustomMockNoteForSingleMatch) SetContents(string, string, string) error   { return nil }
func (m *CustomMockNoteForSingleMatch) GetNotesList(string) ([]string, error)      { return nil, nil }
func (m *CustomMockNoteForSingleMatch) SearchNotesWithSnippets(string, string) ([]obsidian.NoteMatch, error) {
	return []obsidian.NoteMatch{
		{FilePath: "test-note.md", LineNumber: 5, MatchLine: "test content"},
	}, nil
}

func TestSearchNotesContent(t *testing.T) {
	t.Run("Successful content search with single match", func(t *testing.T) {
		vault := mocks.MockVaultOperator{Name: "myVault"}
		uri := mocks.MockUriManager{}
		note := mocks.MockNoteManager{}
		fuzzyFinder := mocks.MockFuzzyFinder{}

		err := actions.SearchNotesContent(&vault, &note, &uri, &fuzzyFinder, "test", false)
		assert.NoError(t, err)
	})

	t.Run("No matches found", func(t *testing.T) {
		vault := mocks.MockVaultOperator{Name: "myVault"}
		uri := mocks.MockUriManager{}
		note := mocks.MockNoteManager{NoMatches: true}
		fuzzyFinder := mocks.MockFuzzyFinder{}

		err := actions.SearchNotesContent(&vault, &note, &uri, &fuzzyFinder, "nonexistent", false)
		assert.NoError(t, err)
	})

	t.Run("SearchNotesWithSnippets returns error", func(t *testing.T) {
		vault := mocks.MockVaultOperator{Name: "myVault"}
		uri := mocks.MockUriManager{}
		note := mocks.MockNoteManager{
			GetContentsError: errors.New("search failed"),
		}
		fuzzyFinder := mocks.MockFuzzyFinder{}

		err := actions.SearchNotesContent(&vault, &note, &uri, &fuzzyFinder, "test", false)
		assert.Error(t, err)
	})

	t.Run("vault.DefaultName returns error", func(t *testing.T) {
		vault := mocks.MockVaultOperator{
			DefaultNameErr: errors.New("vault name error"),
		}
		uri := mocks.MockUriManager{}
		note := mocks.MockNoteManager{}
		fuzzyFinder := mocks.MockFuzzyFinder{}

		err := actions.SearchNotesContent(&vault, &note, &uri, &fuzzyFinder, "test", false)
		assert.Error(t, err)
	})

	t.Run("vault.Path returns error", func(t *testing.T) {
		vault := mocks.MockVaultOperator{
			PathError: errors.New("vault path error"),
		}
		uri := mocks.MockUriManager{}
		note := mocks.MockNoteManager{}
		fuzzyFinder := mocks.MockFuzzyFinder{}

		err := actions.SearchNotesContent(&vault, &note, &uri, &fuzzyFinder, "test", false)
		assert.Error(t, err)
	})

	t.Run("fuzzy finder returns error", func(t *testing.T) {
		vault := mocks.MockVaultOperator{Name: "myVault"}
		uri := mocks.MockUriManager{}
		note := mocks.MockNoteManager{}
		fuzzyFinder := mocks.MockFuzzyFinder{
			FindErr: errors.New("fuzzy finder error"),
		}

		err := actions.SearchNotesContent(&vault, &note, &uri, &fuzzyFinder, "test", false)
		assert.Error(t, err)
	})

	t.Run("uri execution returns error", func(t *testing.T) {
		vault := mocks.MockVaultOperator{Name: "myVault"}
		uri := mocks.MockUriManager{
			ExecuteErr: errors.New("uri execution error"),
		}
		note := mocks.MockNoteManager{}
		fuzzyFinder := mocks.MockFuzzyFinder{}

		err := actions.SearchNotesContent(&vault, &note, &uri, &fuzzyFinder, "test", false)
		assert.Error(t, err)
	})

	t.Run("Successful content search with editor flag - single match", func(t *testing.T) {
		// Set up mocks for single match scenario
		vault := mocks.MockVaultOperator{
			Name: "myVault",
		}
		uri := mocks.MockUriManager{}
		// Create a custom mock that returns exactly one match
		note := &CustomMockNoteForSingleMatch{}
		fuzzyFinder := mocks.MockFuzzyFinder{}

		// Set EDITOR to a command that will succeed
		originalEditor := os.Getenv("EDITOR")
		defer os.Setenv("EDITOR", originalEditor)
		os.Setenv("EDITOR", "true")

		// Act - test with editor flag enabled
		err := actions.SearchNotesContent(&vault, note, &uri, &fuzzyFinder, "test", true)
		
		// Assert - should succeed without calling URI execute
		assert.NoError(t, err)
	})

	t.Run("Successful content search with editor flag - multiple matches", func(t *testing.T) {
		// Set up mocks for multiple match scenario  
		vault := mocks.MockVaultOperator{
			Name: "myVault",
		}
		uri := mocks.MockUriManager{}
		note := mocks.MockNoteManager{} // This returns 2 matches by default
		fuzzyFinder := mocks.MockFuzzyFinder{
			SelectedIndex: 0,
		}

		// Set EDITOR to a command that will succeed
		originalEditor := os.Getenv("EDITOR")
		defer os.Setenv("EDITOR", originalEditor)
		os.Setenv("EDITOR", "true")

		// Act - test with editor flag enabled
		err := actions.SearchNotesContent(&vault, &note, &uri, &fuzzyFinder, "test", true)
		
		// Assert - should succeed without calling URI execute
		assert.NoError(t, err)
	})

	t.Run("Content search with editor flag fails when editor fails", func(t *testing.T) {
		// Set up mocks for single match scenario
		vault := mocks.MockVaultOperator{
			Name: "myVault",
		}
		uri := mocks.MockUriManager{}
		note := &CustomMockNoteForSingleMatch{}
		fuzzyFinder := mocks.MockFuzzyFinder{}

		// Set EDITOR to a command that will fail
		originalEditor := os.Getenv("EDITOR")
		defer os.Setenv("EDITOR", originalEditor)
		os.Setenv("EDITOR", "false") // 'false' command always fails

		// Act - test with editor flag enabled
		err := actions.SearchNotesContent(&vault, note, &uri, &fuzzyFinder, "test", true)
		
		// Assert - should fail due to editor failure
		assert.Error(t, err)
	})
}
