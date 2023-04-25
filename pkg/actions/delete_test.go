package actions_test

import (
	"errors"
	"github.com/Yakitrak/obsidian-cli/mocks"
	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeleteNote(t *testing.T) {
	t.Run("Successful execution", func(t *testing.T) {
		// Arrange
		vaultOp := &mocks.MockVaultOperator{Name: "myVault"}
		noteManager := &mocks.MockNoteManager{}
		// Act
		err := actions.DeleteNote(vaultOp, noteManager, actions.DeleteParams{
			NotePath: "Search-text-here",
		})
		// Assert
		assert.NoError(t, err, "Expected no error")
	})

	t.Run("VaultOperator returns an error", func(t *testing.T) {
		// Arrange
		vaultOpErr := errors.New("Failed to get vault name")
		vaultOp := &mocks.MockVaultOperator{
			PathError: vaultOpErr,
		}
		// Act
		err := actions.DeleteNote(vaultOp, &mocks.MockNoteManager{}, actions.DeleteParams{
			NotePath: "Search-text-here",
		})
		// Assert
		assert.Error(t, err, "Expected error to occur")
		assert.EqualError(t, err, vaultOpErr.Error(), "Expected error to be %v", vaultOpErr)
	})

	t.Run("NoteManager Execute returns an error", func(t *testing.T) {
		// Arrange
		noteManager := &mocks.MockNoteManager{
			ExecuteErr: errors.New("Failed to execute URI"),
		}
		// Act
		err := actions.DeleteNote(&mocks.MockVaultOperator{}, noteManager, actions.DeleteParams{
			NotePath: "Search-text-here",
		})
		// Assert
		assert.Error(t, err, "Expected error to occur")
		assert.EqualError(t, err, "Failed to execute URI", "Expected error to be 'Failed to execute URI'")
	})
}
