package actions_test

import (
	"errors"
	"github.com/Yakitrak/obsidian-cli/mocks"
	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateNote(t *testing.T) {
	t.Run("Successful create note", func(t *testing.T) {
		// Arrange
		vault := mocks.MockVaultOperator{Name: "myVault"}
		uri := mocks.MockUriManager{}
		// Act
		err := actions.CreateNote(&vault, &uri, actions.CreateParams{
			NoteName: "note.md",
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
		err := actions.CreateNote(&vault, &mocks.MockUriManager{}, actions.CreateParams{
			NoteName: "note-name",
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
		err := actions.CreateNote(&mocks.MockVaultOperator{}, &uri, actions.CreateParams{
			NoteName: "note-name",
		})
		// Assert
		assert.Equal(t, err, uri.ExecuteErr)
	})
}

// todo test for create note with open flag
