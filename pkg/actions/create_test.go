package actions_test

import (
	"errors"
	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateNote(t *testing.T) {
	t.Run("Successful execution", func(t *testing.T) {
		// Mock dependencies
		vaultOp := &MockVaultOperator{Name: "myVault"}
		uriManager := &MockUriManager{}

		err := actions.CreateNote(vaultOp, uriManager, actions.CreateParams{
			NoteName: "note.md",
		})

		assert.NoError(t, err, "Expected no error")
	})

	t.Run("VaultOperator returns an error", func(t *testing.T) {
		vaultOpErr := errors.New("Failed to get vault name")
		vaultOp := &MockVaultOperator{
			ExecuteErr: vaultOpErr,
		}

		err := actions.CreateNote(vaultOp, &MockUriManager{}, actions.CreateParams{
			NoteName: "note-name",
		})

		assert.Error(t, err, "Expected error to occur")
		assert.EqualError(t, err, vaultOpErr.Error(), "Expected error to be %v", vaultOpErr)
	})

	t.Run("UriManager Execute returns an error", func(t *testing.T) {
		uriManager := &MockUriManager{
			ExecuteErr: errors.New("Failed to execute URI"),
		}

		err := actions.CreateNote(&MockVaultOperator{}, uriManager, actions.CreateParams{
			NoteName: "note-name",
		})

		assert.Error(t, err, "Expected error to occur")
		assert.EqualError(t, err, "Failed to execute URI", "Expected error to be 'Failed to execute URI'")
	})
}
