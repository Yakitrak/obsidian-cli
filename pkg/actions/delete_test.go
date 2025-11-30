package actions_test

import (
	"errors"
	"testing"

	"github.com/atomicobject/obsidian-cli/mocks"
	"github.com/atomicobject/obsidian-cli/pkg/actions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeleteNote(t *testing.T) {
	t.Run("Successful delete note", func(t *testing.T) {
		// Arrange
		vault := &mocks.VaultManager{}
		note := &mocks.NoteManager{}

		vault.On("DefaultName").Return("myVault", nil)
		vault.On("Path").Return("/test/vault", nil)
		note.On("Delete", mock.AnythingOfType("string")).Return(nil)

		// Act
		err := actions.DeleteNote(vault, note, actions.DeleteParams{
			NotePath: "noteToDelete",
		})

		// Assert
		assert.NoError(t, err, "Expected no error")
		vault.AssertExpectations(t)
		note.AssertExpectations(t)
	})

	t.Run("vault.DefaultName returns an error", func(t *testing.T) {
		// Arrange
		vault := &mocks.VaultManager{}
		note := &mocks.NoteManager{}
		expectedErr := errors.New("Failed to get default vault name")

		vault.On("DefaultName").Return("", expectedErr)

		// Act
		err := actions.DeleteNote(vault, note, actions.DeleteParams{
			NotePath: "noteToDelete",
		})

		// Assert
		assert.Equal(t, expectedErr, err)
		vault.AssertExpectations(t)
	})

	t.Run("vault.Path returns an error", func(t *testing.T) {
		// Arrange
		vault := &mocks.VaultManager{}
		note := &mocks.NoteManager{}
		expectedErr := errors.New("Failed to get vault path")

		vault.On("DefaultName").Return("myVault", nil)
		vault.On("Path").Return("", expectedErr)

		// Act
		err := actions.DeleteNote(vault, note, actions.DeleteParams{
			NotePath: "noteToDelete",
		})

		// Assert
		assert.Equal(t, expectedErr, err)
		vault.AssertExpectations(t)
	})

	t.Run("note.Delete returns an error", func(t *testing.T) {
		// Arrange
		vault := &mocks.VaultManager{}
		note := &mocks.NoteManager{}
		expectedErr := errors.New("Could not delete")

		vault.On("DefaultName").Return("myVault", nil)
		vault.On("Path").Return("/test/vault", nil)
		note.On("Delete", mock.AnythingOfType("string")).Return(expectedErr)

		// Act
		err := actions.DeleteNote(vault, note, actions.DeleteParams{
			NotePath: "noteToDelete",
		})

		// Assert
		assert.Equal(t, expectedErr, err)
		vault.AssertExpectations(t)
		note.AssertExpectations(t)
	})
}
