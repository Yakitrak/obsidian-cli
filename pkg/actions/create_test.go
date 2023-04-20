package actions_test

import (
	"errors"
	"testing"

	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/stretchr/testify/assert"
)

// Define a struct to represent the test cases
type createNoteTestcase struct {
	name            string
	noteName        string
	content         string
	shouldAppend    bool
	shouldOverwrite bool
	expectedResult  map[string]string
	expectedError   error
	mockDefaultName func() (string, error)
}

// Define a table of test cases
var createNoteTestcases = []createNoteTestcase{
	{
		name:            "Note only",
		noteName:        "myNoteName",
		content:         "",
		shouldAppend:    false,
		shouldOverwrite: false,
		expectedResult: map[string]string{
			"file":  "myNoteName",
			"vault": "myVaultName",
		},
	},
	{
		name:            "Note with content",
		noteName:        "myNoteName",
		content:         "Note-content",
		shouldAppend:    false,
		shouldOverwrite: false,
		expectedResult: map[string]string{
			"file":    "myNoteName",
			"content": "Note-content",
			"vault":   "myVaultName",
		},
	},
	{
		name:            "Note with content and append",
		noteName:        "myNoteName",
		content:         "Note-content",
		shouldAppend:    true,
		shouldOverwrite: false,
		expectedResult: map[string]string{
			"file":    "myNoteName",
			"content": "Note-content",
			"append":  "true",
		},
	},
	{
		name:            "Note with content and overwrite",
		noteName:        "myNoteName",
		content:         "Note-content",
		shouldAppend:    false,
		shouldOverwrite: true,
		expectedResult: map[string]string{
			"file":      "myNoteName",
			"content":   "Note-content",
			"overwrite": "true",
		},
	},
	{
		name:            "Error getting default vault name",
		noteName:        "myNoteName",
		mockDefaultName: func() (string, error) { return "", errors.New("failed to get default vault name") },
		content:         "Note content",
		shouldAppend:    true,
		shouldOverwrite: false,
		expectedError:   errors.New("failed to get default vault name"),
	},
}

func TestCreateNote(t *testing.T) {
	// Iterate over the test cases
	for _, tc := range createNoteTestcases {
		// Define the test function
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock implementation of VaultInterface
			mockVault := &MockVault{
				DefaultNameFunc: tc.mockDefaultName,
			}

			// Call CreateNote with the mockVaultOperator
			result, err := actions.CreateNote(mockVault, tc.noteName, tc.content, tc.shouldAppend, tc.shouldOverwrite)

			// Assert that there are no errors
			if tc.expectedError != nil {
				assert.Error(t, err, "Expected error")
				assert.EqualError(t, err, tc.expectedError.Error(), "Unexpected error")
				return
			}
			assert.NoError(t, err, "Unexpected error")

			// Assert that the returned result is a map with expected values
			for k, v := range tc.expectedResult {
				assert.Contains(t, result, k, "Expected key not found")
				assert.Contains(t, result, v, "Expected value not found")
			}
		})
	}
}
