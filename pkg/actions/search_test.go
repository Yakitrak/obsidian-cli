package actions_test

import (
	"errors"
	"github.com/Yakitrak/obsidian-cli/mocks"
	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSearchNotes(t *testing.T) {
	t.Run("Successful execution", func(t *testing.T) {
		// Arrange
		vaultOp := &mocks.MockVaultOperator{Name: "myVault"}
		uriManager := &mocks.MockUriManager{}
		// Act
		err := actions.SearchNotes(vaultOp, uriManager, actions.SearchParams{
			SearchText: "Search-text-here",
		})
		// Assert
		assert.NoError(t, err, "Expected no error")
	})

	t.Run("VaultOperator returns an error", func(t *testing.T) {
		// Arrange
		vaultOpErr := errors.New("Failed to get vault name")
		vaultOp := &mocks.MockVaultOperator{
			ExecuteErr: vaultOpErr,
		}
		// Act
		err := actions.SearchNotes(vaultOp, &mocks.MockUriManager{}, actions.SearchParams{
			SearchText: "Search-text-here",
		})
		// Assert
		assert.Error(t, err, "Expected error to occur")
		assert.EqualError(t, err, vaultOpErr.Error(), "Expected error to be %v", vaultOpErr)
	})

	t.Run("UriManager Execute returns an error", func(t *testing.T) {
		// Arrange
		uriManager := &mocks.MockUriManager{
			ExecuteErr: errors.New("Failed to execute URI"),
		}
		// Act
		err := actions.SearchNotes(&mocks.MockVaultOperator{}, uriManager, actions.SearchParams{
			SearchText: "Search-text-here",
		})
		// Assert
		assert.Error(t, err, "Expected error to occur")
		assert.EqualError(t, err, "Failed to execute URI", "Expected error to be 'Failed to execute URI'")
	})
}
