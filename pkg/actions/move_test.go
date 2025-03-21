package actions_test

import (
	"errors"
	"testing"

	"github.com/Yakitrak/obsidian-cli/mocks"
	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMoveNote(t *testing.T) {
	t.Run("Successful move note", func(t *testing.T) {
		// Arrange
		vault := &mocks.VaultManager{}
		note := &mocks.NoteManager{}
		uri := &mocks.MockUriManager{}

		vault.On("DefaultName").Return("myVault", nil)
		vault.On("Path").Return("/test/vault", nil)
		note.On("Move", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
		note.On("UpdateLinks", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
		uri.On("Construct", mock.AnythingOfType("string"), mock.AnythingOfType("map[string]string")).Return("obsidian://open?vault=myVault&file=newNote.md")
		uri.On("Execute", mock.AnythingOfType("string")).Return(nil)

		// Act
		err := actions.MoveNote(vault, note, uri, actions.MoveParams{
			CurrentNoteName: "oldNote.md",
			NewNoteName:     "newNote.md",
			ShouldOpen:      true,
		})

		// Assert
		assert.NoError(t, err, "Expected no error")
		vault.AssertExpectations(t)
		note.AssertExpectations(t)
		uri.AssertExpectations(t)
	})

	t.Run("vault.DefaultName returns an error", func(t *testing.T) {
		// Arrange
		vault := &mocks.VaultManager{}
		note := &mocks.NoteManager{}
		uri := &mocks.MockUriManager{}
		expectedErr := errors.New("Failed to get default vault name")

		vault.On("DefaultName").Return("", expectedErr)

		// Act
		err := actions.MoveNote(vault, note, uri, actions.MoveParams{
			CurrentNoteName: "oldNote.md",
			NewNoteName:     "newNote.md",
			ShouldOpen:      false,
		})

		// Assert
		assert.Equal(t, expectedErr, err)
		vault.AssertExpectations(t)
	})

	t.Run("vault.Path returns an error", func(t *testing.T) {
		// Arrange
		vault := &mocks.VaultManager{}
		note := &mocks.NoteManager{}
		uri := &mocks.MockUriManager{}
		expectedErr := errors.New("Failed to get vault path")

		vault.On("DefaultName").Return("myVault", nil)
		vault.On("Path").Return("", expectedErr)

		// Act
		err := actions.MoveNote(vault, note, uri, actions.MoveParams{
			CurrentNoteName: "oldNote.md",
			NewNoteName:     "newNote.md",
			ShouldOpen:      false,
		})

		// Assert
		assert.Equal(t, expectedErr, err)
		vault.AssertExpectations(t)
	})

	t.Run("note.Move returns an error", func(t *testing.T) {
		// Arrange
		vault := &mocks.VaultManager{}
		note := &mocks.NoteManager{}
		uri := &mocks.MockUriManager{}
		expectedErr := errors.New("Could not move")

		vault.On("DefaultName").Return("myVault", nil)
		vault.On("Path").Return("/test/vault", nil)
		note.On("Move", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(expectedErr)

		// Act
		err := actions.MoveNote(vault, note, uri, actions.MoveParams{
			CurrentNoteName: "oldNote.md",
			NewNoteName:     "newNote.md",
			ShouldOpen:      false,
		})

		// Assert
		assert.Equal(t, expectedErr, err)
		vault.AssertExpectations(t)
		note.AssertExpectations(t)
	})

	t.Run("note.UpdateLinks returns an error", func(t *testing.T) {
		// Arrange
		vault := &mocks.VaultManager{}
		note := &mocks.NoteManager{}
		uri := &mocks.MockUriManager{}
		expectedErr := errors.New("Could not update links")

		vault.On("DefaultName").Return("myVault", nil)
		vault.On("Path").Return("/test/vault", nil)
		note.On("Move", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
		note.On("UpdateLinks", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(expectedErr)

		// Act
		err := actions.MoveNote(vault, note, uri, actions.MoveParams{
			CurrentNoteName: "oldNote.md",
			NewNoteName:     "newNote.md",
			ShouldOpen:      false,
		})

		// Assert
		assert.Equal(t, expectedErr, err)
		vault.AssertExpectations(t)
		note.AssertExpectations(t)
	})

	t.Run("uri.Execute returns an error", func(t *testing.T) {
		// Arrange
		vault := &mocks.VaultManager{}
		note := &mocks.NoteManager{}
		uri := &mocks.MockUriManager{}
		expectedErr := errors.New("Could not execute URI")

		vault.On("DefaultName").Return("myVault", nil)
		vault.On("Path").Return("/test/vault", nil)
		note.On("Move", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
		note.On("UpdateLinks", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
		uri.On("Construct", mock.AnythingOfType("string"), mock.AnythingOfType("map[string]string")).Return("obsidian://open?vault=myVault&file=newNote.md")
		uri.On("Execute", mock.AnythingOfType("string")).Return(expectedErr)

		// Act
		err := actions.MoveNote(vault, note, uri, actions.MoveParams{
			CurrentNoteName: "oldNote.md",
			NewNoteName:     "newNote.md",
			ShouldOpen:      true,
		})

		// Assert
		assert.Equal(t, expectedErr, err)
		vault.AssertExpectations(t)
		note.AssertExpectations(t)
		uri.AssertExpectations(t)
	})
}
