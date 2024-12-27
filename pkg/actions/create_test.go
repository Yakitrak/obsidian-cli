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
