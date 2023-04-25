package actions_test

import (
	"errors"
	"github.com/Yakitrak/obsidian-cli/mocks"
	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOpenNote(t *testing.T) {
	t.Run("Successful execution", func(t *testing.T) {
		// Arrange
		vaultOp := &mocks.MockVaultOperator{Name: "myVault"}
		uriManager := &mocks.MockUriManager{}
		// Act
		err := actions.OpenNote(vaultOp, uriManager, actions.OpenParams{
			NoteName: "note.md",
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
		err := actions.OpenNote(vaultOp, &mocks.MockUriManager{}, actions.OpenParams{
			NoteName: "note.md",
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
		err := actions.OpenNote(&mocks.MockVaultOperator{}, uriManager, actions.OpenParams{
			NoteName: "note1.md",
		})
		// Assert
		assert.Error(t, err, "Expected error to occur")
		assert.EqualError(t, err, "Failed to execute URI", "Expected error to be 'Failed to execute URI'")
	})
}
