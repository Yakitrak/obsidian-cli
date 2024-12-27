package actions_test

import (
	"errors"
	"github.com/Yakitrak/obsidian-cli/mocks"
	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPrintNote(t *testing.T) {
	t.Run("Successful get contents of note", func(t *testing.T) {
		// Arrange
		vault := mocks.MockVaultOperator{Name: "myVault"}
		note := mocks.MockNoteManager{}
		// Act
		content, err := actions.PrintNote(&vault, &note, actions.PrintParams{
			NoteName: "note-name",
		})
		// Assert
		assert.NoError(t, err, "Expected no error")
		assert.Equal(t, content, "example contents", "Expect matching file contents")
	})

	t.Run("vault.DefaultName returns an error", func(t *testing.T) {
		// Arrange
		vault := mocks.MockVaultOperator{
			DefaultNameErr: errors.New("Failed to get vault name"),
		}
		note := mocks.MockNoteManager{}
		// Act
		_, err := actions.PrintNote(&vault, &note, actions.PrintParams{
			NoteName: "note-name",
		})
		// Assert
		assert.Equal(t, err, vault.DefaultNameErr)
	})

	t.Run("GetContents returns an error", func(t *testing.T) {
		// Arrange
		vault := mocks.MockVaultOperator{}
		note := mocks.MockNoteManager{
			GetContentsError: errors.New("Failed to read note"),
		}
		// Act
		_, err := actions.PrintNote(&vault, &note, actions.PrintParams{
			NoteName: "note-name",
		})
		// Assert
		assert.Equal(t, err, note.GetContentsError)
	})
}
