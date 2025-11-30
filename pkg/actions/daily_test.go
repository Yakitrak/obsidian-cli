package actions_test

import (
	"errors"
	"testing"

	"github.com/atomicobject/obsidian-cli/mocks"
	"github.com/atomicobject/obsidian-cli/pkg/actions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDailyNote(t *testing.T) {
	t.Run("Successful creates / opens daily note", func(t *testing.T) {
		// Arrange
		vault := &mocks.VaultManager{}
		uri := &mocks.MockUriManager{}

		vault.On("DefaultName").Return("myVault", nil)
		uri.On("Construct", mock.AnythingOfType("string"), mock.AnythingOfType("map[string]string")).Return("obsidian://daily?vault=myVault", nil)
		uri.On("Execute", mock.AnythingOfType("string")).Return(nil)

		// Act
		err := actions.DailyNote(vault, uri)

		// Assert
		assert.NoError(t, err)
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
		err := actions.DailyNote(vault, uri)

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
		uri.On("Construct", mock.AnythingOfType("string"), mock.AnythingOfType("map[string]string")).Return("obsidian://daily?vault=myVault", nil)
		uri.On("Execute", mock.AnythingOfType("string")).Return(expectedErr)

		// Act
		err := actions.DailyNote(vault, uri)

		// Assert
		assert.Equal(t, expectedErr, err)
		vault.AssertExpectations(t)
		uri.AssertExpectations(t)
	})
}
