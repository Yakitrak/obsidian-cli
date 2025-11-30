package actions_test

import (
	"errors"
	"testing"

	"github.com/atomicobject/obsidian-cli/mocks"
	"github.com/atomicobject/obsidian-cli/pkg/actions"
	"github.com/stretchr/testify/assert"
)

func TestPrintNote(t *testing.T) {
	t.Run("Successful print note", func(t *testing.T) {
		// Arrange
		vault := &mocks.VaultManager{}
		note := &mocks.NoteManager{}

		vault.On("DefaultName").Return("myVault", nil)
		vault.On("Path").Return("/test/vault", nil)
		note.On("GetContents", "/test/vault", "noteToPrint.md").Return("Note content", nil)

		// Act
		content, err := actions.PrintNote(vault, note, actions.PrintParams{
			NoteName: "noteToPrint.md",
		})

		// Assert
		assert.NoError(t, err, "Expected no error")
		assert.Equal(t, "Note content", content)
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
		content, err := actions.PrintNote(vault, note, actions.PrintParams{
			NoteName: "noteToPrint.md",
		})

		// Assert
		assert.Equal(t, expectedErr, err)
		assert.Empty(t, content)
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
		content, err := actions.PrintNote(vault, note, actions.PrintParams{
			NoteName: "noteToPrint.md",
		})

		// Assert
		assert.Equal(t, expectedErr, err)
		assert.Empty(t, content)
		vault.AssertExpectations(t)
	})

	t.Run("note.GetContents returns an error", func(t *testing.T) {
		// Arrange
		vault := &mocks.VaultManager{}
		note := &mocks.NoteManager{}
		expectedErr := errors.New("Could not get contents")

		vault.On("DefaultName").Return("myVault", nil)
		vault.On("Path").Return("/test/vault", nil)
		note.On("GetContents", "/test/vault", "noteToPrint.md").Return("", expectedErr)

		// Act
		content, err := actions.PrintNote(vault, note, actions.PrintParams{
			NoteName: "noteToPrint.md",
		})

		// Assert
		assert.Equal(t, expectedErr, err)
		assert.Empty(t, content)
		vault.AssertExpectations(t)
		note.AssertExpectations(t)
	})
}
