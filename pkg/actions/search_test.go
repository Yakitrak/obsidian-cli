package actions_test

import (
	"errors"
	"testing"

	"github.com/atomicobject/obsidian-cli/mocks"
	"github.com/atomicobject/obsidian-cli/pkg/actions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSearchNotes(t *testing.T) {
	t.Run("Successful search notes", func(t *testing.T) {
		// Arrange
		vault := &mocks.VaultManager{}
		note := &mocks.NoteManager{}
		uri := &mocks.MockUriManager{}
		fuzzyFinder := &mocks.MockFuzzyFinder{}

		vault.On("DefaultName").Return("myVault", nil)
		vault.On("Path").Return("/test/vault", nil)
		note.On("GetNotesList", "/test/vault").Return([]string{"note1.md", "note2.md"}, nil)
		fuzzyFinder.On("Find", mock.AnythingOfType("[]string"), mock.AnythingOfType("func(int) string"), mock.Anything).Return(0, nil)
		uri.On("Construct", mock.AnythingOfType("string"), mock.AnythingOfType("map[string]string")).Return("obsidian://open?vault=myVault&file=note1.md", nil)
		uri.On("Execute", mock.AnythingOfType("string")).Return(nil)

		// Act
		err := actions.SearchNotes(vault, note, uri, fuzzyFinder)

		// Assert
		assert.NoError(t, err, "Expected no error")
		vault.AssertExpectations(t)
		note.AssertExpectations(t)
		uri.AssertExpectations(t)
		fuzzyFinder.AssertExpectations(t)
	})

	t.Run("vault.DefaultName returns an error", func(t *testing.T) {
		// Arrange
		vault := &mocks.VaultManager{}
		note := &mocks.NoteManager{}
		uri := &mocks.MockUriManager{}
		fuzzyFinder := &mocks.MockFuzzyFinder{}
		expectedErr := errors.New("Failed to get default vault name")

		vault.On("DefaultName").Return("", expectedErr)

		// Act
		err := actions.SearchNotes(vault, note, uri, fuzzyFinder)

		// Assert
		assert.Equal(t, expectedErr, err)
		vault.AssertExpectations(t)
	})

	t.Run("vault.Path returns an error", func(t *testing.T) {
		// Arrange
		vault := &mocks.VaultManager{}
		note := &mocks.NoteManager{}
		uri := &mocks.MockUriManager{}
		fuzzyFinder := &mocks.MockFuzzyFinder{}
		expectedErr := errors.New("Failed to get vault path")

		vault.On("DefaultName").Return("myVault", nil)
		vault.On("Path").Return("", expectedErr)

		// Act
		err := actions.SearchNotes(vault, note, uri, fuzzyFinder)

		// Assert
		assert.Equal(t, expectedErr, err)
		vault.AssertExpectations(t)
	})

	t.Run("note.GetNotesList returns an error", func(t *testing.T) {
		// Arrange
		vault := &mocks.VaultManager{}
		note := &mocks.NoteManager{}
		uri := &mocks.MockUriManager{}
		fuzzyFinder := &mocks.MockFuzzyFinder{}
		expectedErr := errors.New("Could not get notes list")

		vault.On("DefaultName").Return("myVault", nil)
		vault.On("Path").Return("/test/vault", nil)
		note.On("GetNotesList", "/test/vault").Return(nil, expectedErr)

		// Act
		err := actions.SearchNotes(vault, note, uri, fuzzyFinder)

		// Assert
		assert.Equal(t, expectedErr, err)
		vault.AssertExpectations(t)
		note.AssertExpectations(t)
	})

	t.Run("fuzzyFinder.Find returns an error", func(t *testing.T) {
		// Arrange
		vault := &mocks.VaultManager{}
		note := &mocks.NoteManager{}
		uri := &mocks.MockUriManager{}
		fuzzyFinder := &mocks.MockFuzzyFinder{}
		expectedErr := errors.New("Could not find note")

		vault.On("DefaultName").Return("myVault", nil)
		vault.On("Path").Return("/test/vault", nil)
		note.On("GetNotesList", "/test/vault").Return([]string{"note1.md", "note2.md"}, nil)
		fuzzyFinder.On("Find", mock.AnythingOfType("[]string"), mock.AnythingOfType("func(int) string"), mock.Anything).Return(-1, expectedErr)

		// Act
		err := actions.SearchNotes(vault, note, uri, fuzzyFinder)

		// Assert
		assert.Equal(t, expectedErr, err)
		vault.AssertExpectations(t)
		note.AssertExpectations(t)
		fuzzyFinder.AssertExpectations(t)
	})

	t.Run("uri.Execute returns an error", func(t *testing.T) {
		// Arrange
		vault := &mocks.VaultManager{}
		note := &mocks.NoteManager{}
		uri := &mocks.MockUriManager{}
		fuzzyFinder := &mocks.MockFuzzyFinder{}
		expectedErr := errors.New("Could not execute URI")

		vault.On("DefaultName").Return("myVault", nil)
		vault.On("Path").Return("/test/vault", nil)
		note.On("GetNotesList", "/test/vault").Return([]string{"note1.md", "note2.md"}, nil)
		fuzzyFinder.On("Find", mock.AnythingOfType("[]string"), mock.AnythingOfType("func(int) string"), mock.Anything).Return(0, nil)
		uri.On("Construct", mock.AnythingOfType("string"), mock.AnythingOfType("map[string]string")).Return("obsidian://open?vault=myVault&file=note1.md", nil)
		uri.On("Execute", mock.AnythingOfType("string")).Return(expectedErr)

		// Act
		err := actions.SearchNotes(vault, note, uri, fuzzyFinder)

		// Assert
		assert.Equal(t, expectedErr, err)
		vault.AssertExpectations(t)
		note.AssertExpectations(t)
		uri.AssertExpectations(t)
		fuzzyFinder.AssertExpectations(t)
	})
}
