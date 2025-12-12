package actions_test

import (
	"errors"
	"os"
	"testing"

	"github.com/Yakitrak/obsidian-cli/mocks"
	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/stretchr/testify/assert"
)

func TestCreateNote(t *testing.T) {
	t.Run("Successful create note", func(t *testing.T) {
		// Arrange
		vault := mocks.MockVaultOperator{Name: "myVault"}
		uri := mocks.MockUriManager{}
		// Act
		err := actions.CreateNote(&vault, &uri, actions.CreateParams{
			NoteName:  "note.md",
			UseEditor: false,
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
			NoteName:  "note-name",
			UseEditor: false,
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
			NoteName:  "note-name",
			UseEditor: false,
		})
		// Assert
		assert.Equal(t, err, uri.ExecuteErr)
	})

	t.Run("Successful create note with editor flag and open", func(t *testing.T) {
		// Arrange
		vault := mocks.MockVaultOperator{Name: "myVault"}
		uri := mocks.MockUriManager{}

		// Set EDITOR to a command that will succeed
		originalEditor := os.Getenv("EDITOR")
		defer os.Setenv("EDITOR", originalEditor)
		os.Setenv("EDITOR", "true")

		// Act
		err := actions.CreateNote(&vault, &uri, actions.CreateParams{
			NoteName:   "note.md",
			ShouldOpen: true,
			UseEditor:  true,
		})

		// Assert
		assert.NoError(t, err)
	})

	t.Run("Create note with editor flag fails when editor fails", func(t *testing.T) {
		// Arrange
		vault := mocks.MockVaultOperator{Name: "myVault"}
		uri := mocks.MockUriManager{}

		// Set EDITOR to a command that will fail
		originalEditor := os.Getenv("EDITOR")
		defer os.Setenv("EDITOR", originalEditor)
		os.Setenv("EDITOR", "false")

		// Act
		err := actions.CreateNote(&vault, &uri, actions.CreateParams{
			NoteName:   "note.md",
			ShouldOpen: true,
			UseEditor:  true,
		})

		// Assert
		assert.Error(t, err)
	})

	t.Run("Create note with editor flag without open does not use editor", func(t *testing.T) {
		// Arrange
		vault := mocks.MockVaultOperator{Name: "myVault"}
		uri := mocks.MockUriManager{}

		// Act - UseEditor is true but ShouldOpen is false
		err := actions.CreateNote(&vault, &uri, actions.CreateParams{
			NoteName:   "note.md",
			ShouldOpen: false,
			UseEditor:  true,
		})

		// Assert - should succeed via normal Obsidian path
		assert.NoError(t, err)
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
