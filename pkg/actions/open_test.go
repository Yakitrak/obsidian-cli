package actions_test

import (
	"errors"
	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Define a mock implementation of VaultInterface for testing
type MockVault struct {
	DefaultNameFunc    func() (string, error)
	SetDefaultNameFunc func(vaultName string) error
	GetPathFunc        func() (string, error)
}

// Implement the VaultInterface methods on the mock Vault
func (m *MockVault) DefaultName() (string, error) {
	if m.DefaultNameFunc != nil {
		return m.DefaultNameFunc()
	}
	return "myVaultName", nil
}

func (m *MockVault) SetDefaultName(_ string) error { return nil }

func (m *MockVault) Path() (string, error) { return "", nil }

// Define a struct to represent the test cases
type testcase struct {
	testName         string
	vaultName        string
	noteName         string
	expectedKV       map[string]string
	expectedErrorMsg string
	mockDefaultName  func() (string, error)
	expectedError    error
}

// Define a table of test cases
var testcases = []testcase{
	{
		testName:   "Happy path",
		vaultName:  "myVaultName",
		noteName:   "myNoteName",
		expectedKV: map[string]string{"file": "myNoteName", "vault": "myVaultName"},
	},
	{
		testName:        "Error getting default vault name",
		vaultName:       "",
		noteName:        "myNoteName",
		mockDefaultName: func() (string, error) { return "", errors.New("failed to get default vault name") },
		expectedError:   errors.New("failed to get default vault name"),
	},
}

func TestOpenNote(t *testing.T) {
	// Iterate over the test cases
	for _, tc := range testcases {
		// Define the test function
		t.Run(tc.testName, func(t *testing.T) {
			// Create a mock implementation of VaultInterface
			mockVault := &MockVault{
				DefaultNameFunc: tc.mockDefaultName,
			}

			// Call OpenNote with the mockVault
			result, err := actions.OpenNote(mockVault, tc.noteName)

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
