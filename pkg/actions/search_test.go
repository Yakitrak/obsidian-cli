package actions_test

import (
	"errors"
	"os"
	"testing"

	"github.com/Yakitrak/obsidian-cli/mocks"
	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/stretchr/testify/assert"
)

func TestSearchNotes(t *testing.T) {
	t.Run("Successful search note", func(t *testing.T) {
		// Arrange
		vault := mocks.MockVaultOperator{Name: "myVault"}
		uri := mocks.MockUriManager{}
		note := mocks.MockNoteManager{}
		fuzzyFinder := mocks.MockFuzzyFinder{}
		// Act
		err := actions.SearchNotes(&vault, &note, &uri, &fuzzyFinder, false)
		// Assert
		assert.NoError(t, err, "Expected no error")
	})

	t.Run("fuzzy find returns error", func(t *testing.T) {
		// Arrange
		vault := mocks.MockVaultOperator{Name: "myVault"}
		uri := mocks.MockUriManager{}
		note := mocks.MockNoteManager{}
		fuzzyFinder := mocks.MockFuzzyFinder{
			FindErr: errors.New("Fuzzy find error"),
		}
		// Act
		err := actions.SearchNotes(&vault, &note, &uri, &fuzzyFinder, false)
		// Assert
		assert.Equal(t, err, fuzzyFinder.FindErr)
	})

	t.Run("vault.DefaultName returns an error", func(t *testing.T) {
		// Arrange
		vault := mocks.MockVaultOperator{
			DefaultNameErr: errors.New("Failed to get vault name"),
		}
		uri := mocks.MockUriManager{}
		note := mocks.MockNoteManager{}
		fuzzyFinder := mocks.MockFuzzyFinder{}
		// Act
		err := actions.SearchNotes(&vault, &note, &uri, &fuzzyFinder, false)
		// Assert
		assert.Equal(t, err, vault.DefaultNameErr)
	})

	t.Run("vault.Path returns an error", func(t *testing.T) {
		// Arrange
		vault := mocks.MockVaultOperator{
			PathError: errors.New("Failed to get vault path"),
		}
		uri := mocks.MockUriManager{}
		note := mocks.MockNoteManager{}
		fuzzyFinder := mocks.MockFuzzyFinder{}
		// Act
		err := actions.SearchNotes(&vault, &note, &uri, &fuzzyFinder, false)
		// Assert
		assert.Equal(t, err, vault.PathError)
	})

	t.Run("uri.Execute returns an error", func(t *testing.T) {
		// Arrange
		vault := mocks.MockVaultOperator{Name: "myVault"}
		note := mocks.MockNoteManager{}
		fuzzyFinder := mocks.MockFuzzyFinder{}
		uri := mocks.MockUriManager{
			ExecuteErr: errors.New("Failed to execute URI"),
		}
		// Act
		err := actions.SearchNotes(&vault, &note, &uri, &fuzzyFinder, false)
		// Assert
		assert.Equal(t, err, uri.ExecuteErr)
	})

	t.Run("Successful search with editor flag", func(t *testing.T) {
		// Set up mocks
		vault := mocks.MockVaultOperator{
			Name: "myVault",
		}
		uri := mocks.MockUriManager{}
		note := mocks.MockNoteManager{}
		fuzzyFinder := mocks.MockFuzzyFinder{
			SelectedIndex: 0,
		}

		// Set EDITOR to a command that will succeed
		originalEditor := os.Getenv("EDITOR")
		defer os.Setenv("EDITOR", originalEditor)
		os.Setenv("EDITOR", "true")

		// Act - test with editor flag enabled
		err := actions.SearchNotes(&vault, &note, &uri, &fuzzyFinder, true)
		
		// Assert - should succeed without calling URI execute
		assert.NoError(t, err)
	})

	t.Run("Search with editor flag fails when editor fails", func(t *testing.T) {
		// Set up mocks
		vault := mocks.MockVaultOperator{
			Name: "myVault",
		}
		uri := mocks.MockUriManager{}
		note := mocks.MockNoteManager{}
		fuzzyFinder := mocks.MockFuzzyFinder{
			SelectedIndex: 0,
		}

		// Set EDITOR to a command that will fail
		originalEditor := os.Getenv("EDITOR")
		defer os.Setenv("EDITOR", originalEditor)
		os.Setenv("EDITOR", "false") // 'false' command always fails

		// Act - test with editor flag enabled
		err := actions.SearchNotes(&vault, &note, &uri, &fuzzyFinder, true)
		
		// Assert - should fail due to editor failure
		assert.Error(t, err)
	})
}
