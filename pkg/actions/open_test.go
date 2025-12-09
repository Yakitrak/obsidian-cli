package actions_test

import (
	"errors"
	"os"
	"testing"

	"github.com/Yakitrak/obsidian-cli/mocks"
	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/stretchr/testify/assert"
)

func TestOpenNote(t *testing.T) {
	t.Run("Successful open note", func(t *testing.T) {
		// Arrange
		vault := mocks.MockVaultOperator{Name: "myVault"}
		uri := mocks.MockUriManager{}
		// Act
		err := actions.OpenNote(&vault, &uri, actions.OpenParams{
			NoteName:  "note.md",
			UseEditor: false,
		})
		// Assert
		assert.Equal(t, err, nil)
	})

	t.Run("vault.DefaultName returns an error", func(t *testing.T) {
		// Arrange
		vaultDefaultNameErr := errors.New("Failed to get vault name")
		vaultOp := &mocks.MockVaultOperator{
			DefaultNameErr: vaultDefaultNameErr,
		}
		// Act
		err := actions.OpenNote(vaultOp, &mocks.MockUriManager{}, actions.OpenParams{
			NoteName:  "note.md",
			UseEditor: false,
		})
		// Assert
		assert.Error(t, err, vaultDefaultNameErr)
	})

	t.Run("uri.Execute returns an error", func(t *testing.T) {
		// Arrange
		uri := mocks.MockUriManager{
			ExecuteErr: errors.New("Failed to execute URI"),
		}
		// Act
		err := actions.OpenNote(&mocks.MockVaultOperator{}, &uri, actions.OpenParams{
			NoteName:  "note1.md",
			UseEditor: false,
		})
		// Assert
		assert.Equal(t, err, uri.ExecuteErr)
	})

	t.Run("Successful open note with editor flag", func(t *testing.T) {
		// Arrange
		vault := mocks.MockVaultOperator{Name: "myVault"}
		uri := mocks.MockUriManager{}

		// Set EDITOR to a command that will succeed
		originalEditor := os.Getenv("EDITOR")
		defer os.Setenv("EDITOR", originalEditor)
		os.Setenv("EDITOR", "true")

		// Act
		err := actions.OpenNote(&vault, &uri, actions.OpenParams{
			NoteName:  "note.md",
			UseEditor: true,
		})

		// Assert
		assert.NoError(t, err)
	})

	t.Run("Open note with editor flag fails when editor fails", func(t *testing.T) {
		// Arrange
		vault := mocks.MockVaultOperator{Name: "myVault"}
		uri := mocks.MockUriManager{}

		// Set EDITOR to a command that will fail
		originalEditor := os.Getenv("EDITOR")
		defer os.Setenv("EDITOR", originalEditor)
		os.Setenv("EDITOR", "false") // 'false' command always fails

		// Act
		err := actions.OpenNote(&vault, &uri, actions.OpenParams{
			NoteName:  "note.md",
			UseEditor: true,
		})

		// Assert
		assert.Error(t, err)
	})

	t.Run("Open note with editor flag fails when vault.Path returns error", func(t *testing.T) {
		// Arrange
		vault := mocks.MockVaultOperator{
			Name:      "myVault",
			PathError: errors.New("Failed to get vault path"),
		}
		uri := mocks.MockUriManager{}

		// Act
		err := actions.OpenNote(&vault, &uri, actions.OpenParams{
			NoteName:  "note.md",
			UseEditor: true,
		})

		// Assert
		assert.Error(t, err)
	})
}
