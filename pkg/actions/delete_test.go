package actions_test

import (
	"errors"
	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/stretchr/testify/assert"
	"testing"
)

type MockNoteManager struct {
	ExecuteErr       error
	updateLinksError error
}

// Construct mocks the Construct method of UriManager.
func (m *MockNoteManager) Delete(path string) error {
	return m.ExecuteErr
}

func (m *MockNoteManager) Move(_ string, _ string) error {
	return m.ExecuteErr
}

func (m *MockNoteManager) UpdateLinks(string, string, string) error {
	return m.updateLinksError
}
func TestDeleteNote(t *testing.T) {
	t.Run("Successful execution", func(t *testing.T) {
		// Mock dependencies
		vaultOp := &MockVaultOperator{Name: "myVault"}
		noteManager := &MockNoteManager{}

		err := actions.DeleteNote(vaultOp, noteManager, actions.DeleteParams{
			NotePath: "Search-text-here",
		})

		assert.NoError(t, err, "Expected no error")
	})

	t.Run("VaultOperator returns an error", func(t *testing.T) {
		vaultOpErr := errors.New("Failed to get vault name")
		vaultOp := &MockVaultOperator{
			PathError: vaultOpErr,
		}

		err := actions.DeleteNote(vaultOp, &MockNoteManager{}, actions.DeleteParams{
			NotePath: "Search-text-here",
		})

		assert.Error(t, err, "Expected error to occur")
		assert.EqualError(t, err, vaultOpErr.Error(), "Expected error to be %v", vaultOpErr)
	})

	t.Run("NoteManager Execute returns an error", func(t *testing.T) {
		noteManager := &MockNoteManager{
			ExecuteErr: errors.New("Failed to execute URI"),
		}

		err := actions.DeleteNote(&MockVaultOperator{}, noteManager, actions.DeleteParams{
			NotePath: "Search-text-here",
		})

		assert.Error(t, err, "Expected error to occur")
		assert.EqualError(t, err, "Failed to execute URI", "Expected error to be 'Failed to execute URI'")
	})
}
