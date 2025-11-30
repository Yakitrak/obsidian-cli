package actions_test

import (
	"errors"
	"testing"

	"github.com/atomicobject/obsidian-cli/mocks"
	"github.com/atomicobject/obsidian-cli/pkg/actions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateNote(t *testing.T) {
	t.Run("Successful create note", func(t *testing.T) {
		// Arrange
		vault := &mocks.VaultManager{}
		uri := &mocks.MockUriManager{}

		vault.On("DefaultName").Return("myVault", nil)
		uri.On("Construct", mock.AnythingOfType("string"), mock.AnythingOfType("map[string]string")).Return("obsidian://new?vault=myVault&file=note.md", nil)
		uri.On("Execute", mock.AnythingOfType("string")).Return(nil)

		// Act
		err := actions.CreateNote(vault, uri, actions.CreateParams{
			NoteName: "note.md",
		})

		// Assert
		assert.NoError(t, err, "Expected no error")
		vault.AssertExpectations(t)
		uri.AssertExpectations(t)
	})

	t.Run("vault.DefaultName returns an error", func(t *testing.T) {
		// Arrange
		vault := &mocks.VaultManager{}
		uri := &mocks.MockUriManager{}
		expectedErr := errors.New("Failed to get vault name")

		vault.On("DefaultName").Return("", expectedErr)

		// Act
		err := actions.CreateNote(vault, uri, actions.CreateParams{
			NoteName: "note.md",
		})

		// Assert
		assert.Equal(t, expectedErr, err)
		vault.AssertExpectations(t)
	})

	t.Run("uri.Execute returns an error", func(t *testing.T) {
		// Arrange
		vault := &mocks.VaultManager{}
		uri := &mocks.MockUriManager{}
		expectedErr := errors.New("Failed to execute URI")

		vault.On("DefaultName").Return("myVault", nil)
		uri.On("Construct", mock.AnythingOfType("string"), mock.AnythingOfType("map[string]string")).Return("obsidian://new?vault=myVault&file=note.md", nil)
		uri.On("Execute", mock.AnythingOfType("string")).Return(expectedErr)

		// Act
		err := actions.CreateNote(vault, uri, actions.CreateParams{
			NoteName: "note.md",
		})

		// Assert
		assert.Equal(t, expectedErr, err)
		vault.AssertExpectations(t)
		uri.AssertExpectations(t)
	})
}

func TestNormalizeContent(t *testing.T) {
	t.Run("Replaces escape sequences with actual characters", func(t *testing.T) {
		// Arrange
		input := "Hello\\nWorld\\tTabbed\\rReturn\\\"Quote\\'SingleQuote\\\\Backslash"
		expected := "Hello\nWorld\tTabbed\rReturn\"Quote'SingleQuote\\Backslash"

		// Act
		result := actions.NormalizeContent(input)

		// Assert
		assert.Equal(t, expected, result, "The content should have the escape sequences replaced correctly")
	})

	t.Run("Handles empty input", func(t *testing.T) {
		// Arrange
		input := ""
		expected := ""

		// Act
		result := actions.NormalizeContent(input)

		// Assert
		assert.Equal(t, expected, result, "Empty input should return empty output")
	})

	t.Run("No escape sequences in input", func(t *testing.T) {
		// Arrange
		input := "Plain text with no escapes"
		expected := "Plain text with no escapes"

		// Act
		result := actions.NormalizeContent(input)

		// Assert
		assert.Equal(t, expected, result, "Content without escape sequences should remain unchanged")
	})
}
