package actions_test

import (
	"errors"
	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMoveNote(t *testing.T) {
	t.Run("Successful execution", func(t *testing.T) {
		// Mock dependencies
		vaultOp := &MockVaultOperator{Name: "myVault"}
		uriManager := &MockUriManager{}
		noteManager := &MockNoteManager{}

		err := actions.MoveNote(vaultOp, noteManager, uriManager, actions.MoveParams{
			CurrentNoteName: "string",
			NewNoteName:     "string",
			ShouldOpen:      true,
		})

		assert.NoError(t, err, "Expected no error")
	})

	t.Run("VaultOperator.DefaultName returns an error", func(t *testing.T) {
		vaultOpErr := errors.New("Failed to get vault name")
		vaultOp := &MockVaultOperator{
			ExecuteErr: vaultOpErr,
		}

		err := actions.MoveNote(vaultOp, &MockNoteManager{}, &MockUriManager{}, actions.MoveParams{
			CurrentNoteName: "string",
			NewNoteName:     "string",
			ShouldOpen:      true,
		})

		assert.Error(t, err, "Expected error to occur")
		assert.EqualError(t, err, vaultOpErr.Error(), "Expected error to be %v", vaultOpErr)
	})

	t.Run("VaultOperator.Path returns an error", func(t *testing.T) {
		vaultOpErr := errors.New("Failed to get vault path")
		vaultOp := &MockVaultOperator{
			PathError: vaultOpErr,
		}

		err := actions.MoveNote(vaultOp, &MockNoteManager{}, &MockUriManager{}, actions.MoveParams{
			CurrentNoteName: "string",
			NewNoteName:     "string",
			ShouldOpen:      false,
		})

		assert.Error(t, err, "Expected error to occur")
		assert.EqualError(t, err, vaultOpErr.Error(), "Expected error to be %v", vaultOpErr)
	})

	t.Run("NoteManager move returns an error", func(t *testing.T) {
		noteManager := &MockNoteManager{
			ExecuteErr: errors.New("Failed to execute URI"),
		}

		err := actions.MoveNote(&MockVaultOperator{}, noteManager, &MockUriManager{}, actions.MoveParams{
			CurrentNoteName: "string",
			NewNoteName:     "string",
			ShouldOpen:      false,
		})

		assert.Error(t, err, "Expected error to occur")
		assert.EqualError(t, err, "Failed to execute URI", "Expected error to be 'Failed to execute URI'")
	})

	t.Run("NoteManager update links returns an error", func(t *testing.T) {
		noteManager := &MockNoteManager{
			updateLinksError: errors.New("Failed to execute URI"),
		}

		err := actions.MoveNote(&MockVaultOperator{}, noteManager, &MockUriManager{}, actions.MoveParams{
			CurrentNoteName: "string",
			NewNoteName:     "string",
			ShouldOpen:      false,
		})

		assert.Error(t, err, "Expected error to occur")
		assert.EqualError(t, err, "Failed to execute URI", "Expected error to be 'Failed to execute URI'")
	})

	t.Run("UriManager Execute returns an error", func(t *testing.T) {
		uriManager := &MockUriManager{
			ExecuteErr: errors.New("Failed to execute URI"),
		}

		err := actions.MoveNote(&MockVaultOperator{}, &MockNoteManager{}, uriManager, actions.MoveParams{
			CurrentNoteName: "string",
			NewNoteName:     "string",
			ShouldOpen:      true,
		})

		assert.Error(t, err, "Expected error to occur")
		assert.EqualError(t, err, "Failed to execute URI", "Expected error to be 'Failed to execute URI'")
	})
}
