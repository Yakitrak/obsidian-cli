package config_test

import (
	"errors"
	"testing"

	"github.com/Yakitrak/obsidian-cli/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestTargetsPath(t *testing.T) {
	originalUserConfigDirectory := config.UserConfigDirectory
	defer func() { config.UserConfigDirectory = originalUserConfigDirectory }()

	t.Run("Returns targets path under obsidian-cli config directory", func(t *testing.T) {
		config.UserConfigDirectory = func() (string, error) {
			return "user/config/dir", nil
		}

		dir, file, err := config.TargetsPath()
		assert.NoError(t, err)
		assert.Equal(t, "user/config/dir/obsidian-cli", dir)
		assert.Equal(t, "user/config/dir/obsidian-cli/targets.yaml", file)
	})

	t.Run("Returns error when user config directory not found", func(t *testing.T) {
		config.UserConfigDirectory = func() (string, error) {
			return "", errors.New(config.UserConfigDirectoryNotFoundErrorMessage)
		}

		_, _, err := config.TargetsPath()
		assert.Equal(t, config.UserConfigDirectoryNotFoundErrorMessage, err.Error())
	})
}
