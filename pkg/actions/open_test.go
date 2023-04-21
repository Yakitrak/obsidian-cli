package actions_test

import (
	"errors"
	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/stretchr/testify/assert"
	"testing"
)

type MockVaultOperator struct {
	ExecuteErr error
	PathError  error
	Name       string
}

func (m *MockVaultOperator) DefaultName() (string, error) {
	if m.ExecuteErr != nil {
		return "", m.ExecuteErr
	}
	return m.Name, nil
}

func (m *MockVaultOperator) SetDefaultName(_ string) error {
	if m.ExecuteErr != nil {
		return m.ExecuteErr
	}
	return nil
}

func (m *MockVaultOperator) Path() (string, error) {
	if m.PathError != nil {
		return "", m.PathError
	}
	return "path", nil
}

// MockUriManager is a mock implementation of the UriManager interface for testing.
type MockUriManager struct {
	ConstructedURI string
	ExecuteErr     error
}

// Construct mocks the Construct method of UriManager.
func (m *MockUriManager) Construct(base string, params map[string]string) string {
	return m.ConstructedURI
}

// Execute mocks the Execute method of UriManager.
func (m *MockUriManager) Execute(uri string) error {
	return m.ExecuteErr
}

func TestOpenNote(t *testing.T) {
	t.Run("Successful execution", func(t *testing.T) {
		// Mock dependencies
		vaultOp := &MockVaultOperator{Name: "myVault"}
		uriManager := &MockUriManager{}

		err := actions.OpenNote(vaultOp, uriManager, actions.OpenParams{
			NoteName: "note.md",
		})

		assert.NoError(t, err, "Expected no error")
	})

	t.Run("VaultOperator returns an error", func(t *testing.T) {
		vaultOpErr := errors.New("Failed to get vault name")
		vaultOp := &MockVaultOperator{
			ExecuteErr: vaultOpErr,
		}

		err := actions.OpenNote(vaultOp, &MockUriManager{}, actions.OpenParams{
			NoteName: "note.md",
		})

		assert.Error(t, err, "Expected error to occur")
		assert.EqualError(t, err, vaultOpErr.Error(), "Expected error to be %v", vaultOpErr)
	})

	t.Run("UriManager Execute returns an error", func(t *testing.T) {
		uriManager := &MockUriManager{
			ExecuteErr: errors.New("Failed to execute URI"),
		}

		err := actions.OpenNote(&MockVaultOperator{}, uriManager, actions.OpenParams{
			NoteName: "note1.md",
		})

		assert.Error(t, err, "Expected error to occur")
		assert.EqualError(t, err, "Failed to execute URI", "Expected error to be 'Failed to execute URI'")
	})
}
