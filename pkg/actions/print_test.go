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

	t.Run("IncludeMentions false returns just contents", func(t *testing.T) {
		// Arrange
		vault := mocks.MockVaultOperator{Name: "myVault"}
		note := mocks.MockNoteManager{Contents: "note content here"}
		// Act
		content, err := actions.PrintNote(&vault, &note, actions.PrintParams{
			NoteName:        "note-name",
			IncludeMentions: false,
		})
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "note content here", content)
		assert.NotContains(t, content, "Linked Mentions")
	})

	t.Run("IncludeMentions true appends mentions section", func(t *testing.T) {
		// Arrange
		vault := mocks.MockVaultOperator{Name: "myVault"}
		note := mocks.MockNoteManager{Contents: "note content here"}
		// Act
		content, err := actions.PrintNote(&vault, &note, actions.PrintParams{
			NoteName:        "note-name",
			IncludeMentions: true,
		})
		// Assert
		assert.NoError(t, err)
		assert.Contains(t, content, "note content here")
		assert.Contains(t, content, "## Linked Mentions")
		assert.Contains(t, content, "[[linking-note]]")
		assert.Contains(t, content, "[[another-note]]")
	})

	t.Run("IncludeMentions true with no backlinks omits mentions section", func(t *testing.T) {
		// Arrange
		vault := mocks.MockVaultOperator{Name: "myVault"}
		note := mocks.MockNoteManager{
			Contents:  "note content here",
			NoMatches: true,
		}
		// Act
		content, err := actions.PrintNote(&vault, &note, actions.PrintParams{
			NoteName:        "note-name",
			IncludeMentions: true,
		})
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "note content here", content)
		assert.NotContains(t, content, "Linked Mentions")
	})

	t.Run("FindBacklinks error is returned", func(t *testing.T) {
		// Arrange
		vault := mocks.MockVaultOperator{Name: "myVault"}
		note := mocks.MockNoteManager{
			Contents:         "note content",
			FindBacklinksErr: errors.New("failed to find backlinks"),
		}
		// Act
		_, err := actions.PrintNote(&vault, &note, actions.PrintParams{
			NoteName:        "note-name",
			IncludeMentions: true,
		})
		// Assert
		assert.Error(t, err)
		assert.Equal(t, "failed to find backlinks", err.Error())
	})
}
