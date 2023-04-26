package actions_test

import (
	"errors"
	"github.com/Yakitrak/obsidian-cli/mocks"
	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMoveNote(t *testing.T) {
	t.Run("Successful move note", func(t *testing.T) {
		// Arrange
		vault := mocks.MockVaultOperator{Name: "myVault"}
		uri := mocks.MockUriManager{}
		note := mocks.MockNoteManager{}
		// Act
		err := actions.MoveNote(&vault, &note, &uri, actions.MoveParams{
			CurrentNoteName: "string",
			NewNoteName:     "string",
			ShouldOpen:      true,
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
		err := actions.MoveNote(&vault, &mocks.MockNoteManager{}, &mocks.MockUriManager{}, actions.MoveParams{
			CurrentNoteName: "string",
			NewNoteName:     "string",
			ShouldOpen:      true,
		})
		// Assert
		assert.Equal(t, err, vault.DefaultNameErr)
	})

	t.Run("vault.Path returns an error", func(t *testing.T) {
		// Arrange
		vaultOp := &mocks.MockVaultOperator{
			PathError: errors.New("Failed to get vault path"),
		}
		// Act
		err := actions.MoveNote(vaultOp, &mocks.MockNoteManager{}, &mocks.MockUriManager{}, actions.MoveParams{
			CurrentNoteName: "string",
			NewNoteName:     "string",
			ShouldOpen:      false,
		})
		// Assert
		assert.Equal(t, err, vaultOp.PathError)
	})

	t.Run("note.Move returns an error", func(t *testing.T) {
		// Arrange
		note := mocks.MockNoteManager{
			MoveErr: errors.New("Failed to execute URI"),
		}
		// Act
		err := actions.MoveNote(&mocks.MockVaultOperator{}, &note, &mocks.MockUriManager{}, actions.MoveParams{
			CurrentNoteName: "string",
			NewNoteName:     "string",
			ShouldOpen:      false,
		})
		// Assert
		assert.Equal(t, err, note.MoveErr)
	})

	t.Run("note.UpdateLinks returns an error", func(t *testing.T) {
		// Arrange
		note := mocks.MockNoteManager{
			UpdateLinksError: errors.New("Failed to execute URI"),
		}
		// Act
		err := actions.MoveNote(&mocks.MockVaultOperator{}, &note, &mocks.MockUriManager{}, actions.MoveParams{
			CurrentNoteName: "string",
			NewNoteName:     "string",
			ShouldOpen:      false,
		})
		// Assert
		assert.Equal(t, err, note.UpdateLinksError)
	})

	t.Run("uri.Execute returns an error", func(t *testing.T) {
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
		assert.Equal(t, err, uriManager.ExecuteErr)
	})
}
