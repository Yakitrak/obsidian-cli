package actions_test

import (
	"errors"
	"github.com/Yakitrak/obsidian-cli/mocks"
	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDailyNote(t *testing.T) {
	t.Run("Successful creates / opens daily note", func(t *testing.T) {
		// Arrange
		vault := mocks.MockVaultOperator{Name: "myVault"}
		uri := mocks.MockUriManager{}
		// Act
		err := actions.DailyNote(&vault, &uri)
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
		err := actions.DailyNote(vaultOp, &mocks.MockUriManager{})
		// Assert
		assert.Error(t, err, vaultDefaultNameErr)
	})

	t.Run("uri.Execute returns an error", func(t *testing.T) {
		// Arrange
		uri := mocks.MockUriManager{
			ExecuteErr: errors.New("Failed to execute URI"),
		}
		// Act
		err := actions.DailyNote(&mocks.MockVaultOperator{}, &uri)
		// Assert
		assert.Equal(t, err, uri.ExecuteErr)
	})
}
