package actions_test

import (
	"errors"
	"github.com/Yakitrak/obsidian-cli/mocks"
	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSearchNotes(t *testing.T) {
	t.Run("Successful search note", func(t *testing.T) {
		// Arrange
		vault := mocks.MockVaultOperator{Name: "myVault"}
		uri := mocks.MockUriManager{}
		note := mocks.MockNoteManager{}
		fuzzyFinder := mocks.MockFuzzyFinder{}
		// Act
		err := actions.SearchNotes(&vault, &note, &uri, &fuzzyFinder)
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
		err := actions.SearchNotes(&vault, &note, &uri, &fuzzyFinder)
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
		err := actions.SearchNotes(&vault, &note, &uri, &fuzzyFinder)
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
		err := actions.SearchNotes(&vault, &note, &uri, &fuzzyFinder)
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
		err := actions.SearchNotes(&vault, &note, &uri, &fuzzyFinder)
		// Assert
		assert.Equal(t, err, uri.ExecuteErr)
	})
}
