package actions_test

import (
	"errors"
	"github.com/Yakitrak/obsidian-cli/mocks"
	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeleteNote(t *testing.T) {
	t.Run("Successful delete note", func(t *testing.T) {
		// Arrange
		vault := mocks.MockVaultOperator{Name: "myVault"}
		note := mocks.MockNoteManager{}
		// Act
		err := actions.DeleteNote(&vault, &note, actions.DeleteParams{
			NotePath: "noteToDelete",
		})
		// Assert
		assert.NoError(t, err, "Expected no error")
	})

	t.Run("vault.DefaultName returns an error", func(t *testing.T) {
		// Arrange
		vault := mocks.MockVaultOperator{
			DefaultNameErr: errors.New("Failed to get default vault name"),
		}
		// Act
		err := actions.DeleteNote(&vault, &mocks.MockNoteManager{}, actions.DeleteParams{
			NotePath: "noteToDelete",
		})
		// Assert
		assert.Equal(t, vault.DefaultNameErr, err)
	})

	t.Run("vault.Path returns an error", func(t *testing.T) {
		// Arrange
		vault := mocks.MockVaultOperator{
			PathError: errors.New("Failed to get vault path"),
		}
		// Act
		err := actions.DeleteNote(&vault, &mocks.MockNoteManager{}, actions.DeleteParams{
			NotePath: "noteToDelete",
		})
		// Assert
		assert.Equal(t, vault.PathError, err)
	})

	t.Run("note.Delete returns an error", func(t *testing.T) {
		// Arrange
		note := mocks.MockNoteManager{
			DeleteErr: errors.New("Could not delete"),
		}
		// Act
		err := actions.DeleteNote(&mocks.MockVaultOperator{}, &note, actions.DeleteParams{
			NotePath: "noteToDelete",
		})
		// Assert
		assert.Equal(t, note.DeleteErr, err)
	})

	t.Run("rejects note paths that escape the vault", func(t *testing.T) {
		err := actions.DeleteNote(&mocks.MockVaultOperator{}, &mocks.MockNoteManager{}, actions.DeleteParams{
			NotePath: "../escape",
		})
		assert.Error(t, err)
	})
}
