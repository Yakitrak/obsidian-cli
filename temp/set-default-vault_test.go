package temp_test

import (
	"github.com/Yakitrak/obsidian-cli/utils/temp"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestSetDefaultVault(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}

	// Tests
	t.Run("happy path", func(t *testing.T) {
		defer os.RemoveAll(tmpDir)
		err := temp.SetDefaultVault("my_vault", tmpDir)
		assert.Equal(t, nil, err)
		content, err := os.ReadFile(tmpDir + "/preferences.json")
		assert.Equal(t, nil, err)
		assert.Equal(t, `{"default_vault_name":"my_vault"}`, string(content))
	})

	t.Run("fail to json marshals", func(t *testing.T) {
		defer os.RemoveAll(tmpDir)
		err := temp.SetDefaultVault("", tmpDir)
		t.Logf(err.Error())
		assert.ErrorContains(t, err, "failed to save default vault to configuration")
	})

	t.Run("fail to create default vault configuration", func(t *testing.T) {
		err := temp.SetDefaultVault("my_vault", "")
		assert.ErrorContains(t, err, "failed to save default vault to configuration")
	})

}
