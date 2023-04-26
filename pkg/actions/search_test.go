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
		uri := &mocks.MockUriManager{}
		// Act
		err := actions.SearchNotes(&vault, uri, actions.SearchParams{
			SearchText: "Search-text-here",
		})
		// Assert
		assert.NoError(t, err, "Expected no error")
	})

	t.Run("vault.DefaultName returns an error", func(t *testing.T) {
		// Arrange
		vault := mocks.MockVaultOperator{
			DefaultNameErr: errors.New("Failed to get vault name"),
		}
		// Act
		err := actions.SearchNotes(&vault, &mocks.MockUriManager{}, actions.SearchParams{
			SearchText: "Search-text-here",
		})
		// Assert
		assert.Equal(t, err, vault.DefaultNameErr)
	})

	t.Run("uri.Execute returns an error", func(t *testing.T) {
		// Arrange
		uri := mocks.MockUriManager{
			ExecuteErr: errors.New("Failed to execute URI"),
		}
		// Act
		err := actions.SearchNotes(&mocks.MockVaultOperator{}, &uri, actions.SearchParams{
			SearchText: "Search-text-here",
		})
		// Assert
		assert.Equal(t, err, uri.ExecuteErr)
	})
}
