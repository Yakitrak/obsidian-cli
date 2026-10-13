package obsidian_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Yakitrak/notesmd-cli/mocks"
	"github.com/Yakitrak/notesmd-cli/pkg/obsidian"
	"github.com/stretchr/testify/assert"
)

func TestAddVault(t *testing.T) {
	originalObsidianConfigFile := obsidian.ObsidianConfigFile
	originalRunningInWSL := obsidian.RunningInWSL
	defer func() {
		obsidian.ObsidianConfigFile = originalObsidianConfigFile
		obsidian.RunningInWSL = originalRunningInWSL
	}()

	obsidian.RunningInWSL = func() bool { return false }

	t.Run("Adds vault to existing config", func(t *testing.T) {
		mockObsidianConfigFile := mocks.CreateMockObsidianConfigFile(t)
		obsidian.ObsidianConfigFile = func() (string, error) {
			return mockObsidianConfigFile, nil
		}

		err := os.WriteFile(mockObsidianConfigFile, []byte(`{"vaults":{}}`), 0644)
		assert.NoError(t, err)

		vaultDir := t.TempDir()
		absPath, err := obsidian.AddVault(vaultDir)
		assert.NoError(t, err)
		assert.Equal(t, vaultDir, absPath)

		// Verify vault was added
		content, err := os.ReadFile(mockObsidianConfigFile)
		assert.NoError(t, err)

		var cfg obsidian.ObsidianVaultConfig
		err = json.Unmarshal(content, &cfg)
		assert.NoError(t, err)
		assert.Len(t, cfg.Vaults, 1)

		for _, v := range cfg.Vaults {
			assert.Equal(t, vaultDir, v.Path)
		}
	})

	t.Run("Rejects non-existent path", func(t *testing.T) {
		mockObsidianConfigFile := mocks.CreateMockObsidianConfigFile(t)
		obsidian.ObsidianConfigFile = func() (string, error) {
			return mockObsidianConfigFile, nil
		}

		err := os.WriteFile(mockObsidianConfigFile, []byte(`{"vaults":{}}`), 0644)
		assert.NoError(t, err)

		_, err = obsidian.AddVault("/nonexistent/path/to/vault")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path does not exist")
	})

	t.Run("Rejects file path", func(t *testing.T) {
		mockObsidianConfigFile := mocks.CreateMockObsidianConfigFile(t)
		obsidian.ObsidianConfigFile = func() (string, error) {
			return mockObsidianConfigFile, nil
		}

		err := os.WriteFile(mockObsidianConfigFile, []byte(`{"vaults":{}}`), 0644)
		assert.NoError(t, err)

		tmpFile := filepath.Join(t.TempDir(), "file.txt")
		err = os.WriteFile(tmpFile, []byte("test"), 0644)
		assert.NoError(t, err)

		_, err = obsidian.AddVault(tmpFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a directory")
	})

	t.Run("Rejects duplicate vault", func(t *testing.T) {
		mockObsidianConfigFile := mocks.CreateMockObsidianConfigFile(t)
		obsidian.ObsidianConfigFile = func() (string, error) {
			return mockObsidianConfigFile, nil
		}

		vaultDir := t.TempDir()
		configContent := `{
			"vaults": {
				"abc123": {
					"path": "` + vaultDir + `"
				}
			}
		}`
		err := os.WriteFile(mockObsidianConfigFile, []byte(configContent), 0644)
		assert.NoError(t, err)

		_, err = obsidian.AddVault(vaultDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})

	t.Run("Adds multiple vaults", func(t *testing.T) {
		mockObsidianConfigFile := mocks.CreateMockObsidianConfigFile(t)
		obsidian.ObsidianConfigFile = func() (string, error) {
			return mockObsidianConfigFile, nil
		}

		err := os.WriteFile(mockObsidianConfigFile, []byte(`{"vaults":{}}`), 0644)
		assert.NoError(t, err)

		vault1 := t.TempDir()
		vault2 := t.TempDir()

		_, err = obsidian.AddVault(vault1)
		assert.NoError(t, err)

		_, err = obsidian.AddVault(vault2)
		assert.NoError(t, err)

		content, err := os.ReadFile(mockObsidianConfigFile)
		assert.NoError(t, err)

		var cfg obsidian.ObsidianVaultConfig
		err = json.Unmarshal(content, &cfg)
		assert.NoError(t, err)
		assert.Len(t, cfg.Vaults, 2)
	})
}

func TestRemoveVault(t *testing.T) {
	originalObsidianConfigFile := obsidian.ObsidianConfigFile
	originalRunningInWSL := obsidian.RunningInWSL
	defer func() {
		obsidian.ObsidianConfigFile = originalObsidianConfigFile
		obsidian.RunningInWSL = originalRunningInWSL
	}()

	obsidian.RunningInWSL = func() bool { return false }

	t.Run("Removes vault by name", func(t *testing.T) {
		mockObsidianConfigFile := mocks.CreateMockObsidianConfigFile(t)
		obsidian.ObsidianConfigFile = func() (string, error) {
			return mockObsidianConfigFile, nil
		}

		configContent := `{
			"vaults": {
				"abc123": {
					"path": "/Users/user/Documents/Personal"
				},
				"def456": {
					"path": "/Users/user/Documents/Work"
				}
			}
		}`
		err := os.WriteFile(mockObsidianConfigFile, []byte(configContent), 0644)
		assert.NoError(t, err)

		name, err := obsidian.RemoveVault("Personal")
		assert.NoError(t, err)
		assert.Equal(t, "Personal", name)

		content, err := os.ReadFile(mockObsidianConfigFile)
		assert.NoError(t, err)

		var cfg obsidian.ObsidianVaultConfig
		err = json.Unmarshal(content, &cfg)
		assert.NoError(t, err)
		assert.Len(t, cfg.Vaults, 1)

		for _, v := range cfg.Vaults {
			assert.Equal(t, "/Users/user/Documents/Work", v.Path)
		}
	})

	t.Run("Removes vault by path and returns resolved name", func(t *testing.T) {
		mockObsidianConfigFile := mocks.CreateMockObsidianConfigFile(t)
		obsidian.ObsidianConfigFile = func() (string, error) {
			return mockObsidianConfigFile, nil
		}

		configContent := `{
			"vaults": {
				"abc123": {
					"path": "/Users/user/Documents/Personal"
				}
			}
		}`
		err := os.WriteFile(mockObsidianConfigFile, []byte(configContent), 0644)
		assert.NoError(t, err)

		name, err := obsidian.RemoveVault("/Users/user/Documents/Personal")
		assert.NoError(t, err)
		assert.Equal(t, "Personal", name)

		content, err := os.ReadFile(mockObsidianConfigFile)
		assert.NoError(t, err)

		var cfg obsidian.ObsidianVaultConfig
		err = json.Unmarshal(content, &cfg)
		assert.NoError(t, err)
		assert.Len(t, cfg.Vaults, 0)
	})

	t.Run("Returns error for non-existent vault", func(t *testing.T) {
		mockObsidianConfigFile := mocks.CreateMockObsidianConfigFile(t)
		obsidian.ObsidianConfigFile = func() (string, error) {
			return mockObsidianConfigFile, nil
		}

		err := os.WriteFile(mockObsidianConfigFile, []byte(`{"vaults":{}}`), 0644)
		assert.NoError(t, err)

		_, err = obsidian.RemoveVault("NonExistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestRegistryPreservesObsidianFields(t *testing.T) {
	original := obsidian.ObsidianConfigFile
	t.Cleanup(func() { obsidian.ObsidianConfigFile = original })
	path := filepath.Join(t.TempDir(), "obsidian.json")
	obsidian.ObsidianConfigFile = func() (string, error) { return path, nil }
	originalJSON := `{"vaults":{"existing":{"path":"/vaults/keep","ts":9007199254740993,"open":true,"future":{"nested":[1,null,"x"]}}},"theme":"dark","futureSetting":{"enabled":true,"number":9007199254740993}}`
	if !assert.NoError(t, os.WriteFile(path, []byte(originalJSON), 0600)) {
		return
	}
	added := t.TempDir()
	if _, err := obsidian.AddVault(added); !assert.NoError(t, err) {
		return
	}
	content, err := os.ReadFile(path)
	if !assert.NoError(t, err) {
		return
	}
	var before, after map[string]json.RawMessage
	if !assert.NoError(t, json.Unmarshal([]byte(originalJSON), &before)) {
		return
	}
	if !assert.NoError(t, json.Unmarshal(content, &after)) {
		return
	}
	assert.JSONEq(t, string(before["futureSetting"]), string(after["futureSetting"]))
	assert.JSONEq(t, string(before["theme"]), string(after["theme"]))
	var vaults map[string]json.RawMessage
	if !assert.NoError(t, json.Unmarshal(after["vaults"], &vaults)) {
		return
	}
	assert.Len(t, vaults, 2)
	assert.JSONEq(t, `{"path":"/vaults/keep","ts":9007199254740993,"open":true,"future":{"nested":[1,null,"x"]}}`, string(vaults["existing"]))
	// Check exact numeric representation as JSONEq uses floating-point decoding.
	assert.Contains(t, string(vaults["existing"]), "9007199254740993")
	assert.Contains(t, string(after["futureSetting"]), "9007199254740993")
	if _, err := obsidian.RemoveVault(added); !assert.NoError(t, err) {
		return
	}
	content, err = os.ReadFile(path)
	if !assert.NoError(t, err) {
		return
	}
	assert.JSONEq(t, originalJSON, string(content))
	assert.Contains(t, string(content), "9007199254740993")
}

func TestAddVaultWithoutVaultMap(t *testing.T) {
	original := obsidian.ObsidianConfigFile
	t.Cleanup(func() { obsidian.ObsidianConfigFile = original })
	for _, contents := range []string{`{"theme":"dark"}`, `{"theme":"dark","vaults":null}`} {
		t.Run(contents, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "obsidian.json")
			obsidian.ObsidianConfigFile = func() (string, error) { return path, nil }
			if !assert.NoError(t, os.WriteFile(path, []byte(contents), 0600)) {
				return
			}
			_, err := obsidian.AddVault(t.TempDir())
			if !assert.NoError(t, err) {
				return
			}
			content, err := os.ReadFile(path)
			if !assert.NoError(t, err) {
				return
			}
			var cfg map[string]json.RawMessage
			if !assert.NoError(t, json.Unmarshal(content, &cfg)) {
				return
			}
			assert.JSONEq(t, `"dark"`, string(cfg["theme"]))
			var vaults map[string]json.RawMessage
			if !assert.NoError(t, json.Unmarshal(cfg["vaults"], &vaults)) {
				return
			}
			assert.Len(t, vaults, 1)
		})
	}
}
