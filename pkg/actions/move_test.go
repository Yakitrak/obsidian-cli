package actions_test

import (
	"errors"
	"github.com/Yakitrak/obsidian-cli/mocks"
	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMoveNote(t *testing.T) {
	t.Run("Successful execution", func(t *testing.T) {
		// Arrange
		vaultOp := &mocks.MockVaultOperator{Name: "myVault"}
		uriManager := &mocks.MockUriManager{}
		noteManager := &mocks.MockNoteManager{}
		// Act
		err := actions.MoveNote(vaultOp, noteManager, uriManager, actions.MoveParams{
			CurrentNoteName: "string",
			NewNoteName:     "string",
			ShouldOpen:      true,
		})
		// Assert
		assert.NoError(t, err, "Expected no error")
	})

	t.Run("VaultOperator.DefaultName returns an error", func(t *testing.T) {
		// Arrange
		vaultOpErr := errors.New("Failed to get vault name")
		vaultOp := &mocks.MockVaultOperator{
			ExecuteErr: vaultOpErr,
		}
		// Act
		err := actions.MoveNote(vaultOp, &mocks.MockNoteManager{}, &mocks.MockUriManager{}, actions.MoveParams{
			CurrentNoteName: "string",
			NewNoteName:     "string",
			ShouldOpen:      true,
		})
		// Assert
		assert.Error(t, err, "Expected error to occur")
		assert.EqualError(t, err, vaultOpErr.Error(), "Expected error to be %v", vaultOpErr)
	})

	t.Run("VaultOperator.Path returns an error", func(t *testing.T) {
		// Arrange
		vaultOpErr := errors.New("Failed to get vault path")
		vaultOp := &mocks.MockVaultOperator{
			PathError: vaultOpErr,
		}
		// Act
		err := actions.MoveNote(vaultOp, &mocks.MockNoteManager{}, &mocks.MockUriManager{}, actions.MoveParams{
			CurrentNoteName: "string",
			NewNoteName:     "string",
			ShouldOpen:      false,
		})
		// Assert
		assert.Error(t, err, "Expected error to occur")
		assert.EqualError(t, err, vaultOpErr.Error(), "Expected error to be %v", vaultOpErr)
	})

	t.Run("NoteManager move returns an error", func(t *testing.T) {
		// Arrange
		noteManager := &mocks.MockNoteManager{
			ExecuteErr: errors.New("Failed to execute URI"),
		}
		// Act
		err := actions.MoveNote(&mocks.MockVaultOperator{}, noteManager, &mocks.MockUriManager{}, actions.MoveParams{
			CurrentNoteName: "string",
			NewNoteName:     "string",
			ShouldOpen:      false,
		})
		// Assert
		assert.Error(t, err, "Expected error to occur")
		assert.EqualError(t, err, "Failed to execute URI", "Expected error to be 'Failed to execute URI'")
	})

	t.Run("NoteManager update links returns an error", func(t *testing.T) {
		// Arrange
		noteManager := &mocks.MockNoteManager{
			UpdateLinksError: errors.New("Failed to execute URI"),
		}
		// Act
		err := actions.MoveNote(&mocks.MockVaultOperator{}, noteManager, &mocks.MockUriManager{}, actions.MoveParams{
			CurrentNoteName: "string",
			NewNoteName:     "string",
			ShouldOpen:      false,
		})
		// Assert
		assert.Error(t, err, "Expected error to occur")
		assert.EqualError(t, err, "Failed to execute URI", "Expected error to be 'Failed to execute URI'")
	})

	t.Run("UriManager Execute returns an error", func(t *testing.T) {
		// Arrange
		uriManager := &mocks.MockUriManager{
			ExecuteErr: errors.New("Failed to execute URI"),
		}
		// Act
		err := actions.MoveNote(&mocks.MockVaultOperator{}, &mocks.MockNoteManager{}, uriManager, actions.MoveParams{
			CurrentNoteName: "string",
			NewNoteName:     "string",
			ShouldOpen:      true,
		})
		// Assert
		assert.Error(t, err, "Expected error to occur")
		assert.EqualError(t, err, "Failed to execute URI", "Expected error to be 'Failed to execute URI'")
	})
}
