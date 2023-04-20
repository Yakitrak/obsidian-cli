package actions_test

import (
	"errors"
	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Define a struct to represent the test cases
type searchNoteTestcase struct {
	testName         string
	vaultName        string
	searchName       string
	expectedKV       map[string]string
	expectedErrorMsg string
	mockDefaultName  func() (string, error)
	expectedError    error
}

// Define a table of test cases
var searchNoteTestcases = []searchNoteTestcase{
	{
		testName:   "Happy path",
		vaultName:  "myVaultName",
		searchName: "search-text",
		expectedKV: map[string]string{"search": "search-text", "vault": "myVaultName"},
	},
	{
		testName:        "Error getting default vault name",
		vaultName:       "",
		searchName:      "search-text",
		mockDefaultName: func() (string, error) { return "", errors.New("failed to get default vault name") },
		expectedError:   errors.New("failed to get default vault name"),
	},
}

func TestSearchNotes(t *testing.T) {
	// Iterate over the test cases
	for _, tc := range searchNoteTestcases {
		// Define the test function
		t.Run(tc.testName, func(t *testing.T) {
			// Create a mock implementation of VaultInterface
			mockVault := &MockVault{
				DefaultNameFunc: tc.mockDefaultName,
			}

			// Call OpenNote with the mockVault
			result, err := actions.SearchNotes(mockVault, tc.searchName)

			// Assert that there are no errors
			if tc.expectedError != nil {
				assert.Error(t, err, "Expected error")
				return
			}
			assert.NoError(t, err, "Unexpected error")

			// Assert that the returned URI includes the expected keys and values
			for k, v := range tc.expectedKV {
				assert.Contains(t, result, k, "Expected key not found")
				assert.Contains(t, result, v, "Expected value not found")
			}
		})
	}
}
