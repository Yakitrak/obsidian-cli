package config_test

import (
	"errors"
	"testing"

	"github.com/Yakitrak/obsidian-cli/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestConfigCliPath(t *testing.T) {
	originalUserConfigDirectory := config.UserConfigDirectory
	defer func() { config.UserConfigDirectory = originalUserConfigDirectory }()

	t.Run("UserConfigDir func returns a directory", func(t *testing.T) {
		// Arrange
		config.UserConfigDirectory = func() (string, error) {
			return "user/config/dir", nil
		}
		// Act
		obsConfigDir, obsConfigFile, err := config.CliPath()
		// Assert
		assert.Equal(t, nil, err)
		assert.Equal(t, "user/config/dir/obsidian-cli", obsConfigDir)
		assert.Equal(t, "user/config/dir/obsidian-cli/preferences.json", obsConfigFile)
	})

	t.Run("UserConfigDir func returns an error", func(t *testing.T) {
		// Arrange
		config.UserConfigDirectory = func() (string, error) {
			return "", errors.New(config.UserConfigDirectoryNotFoundErrorMessage)
		}
		// Act
		obsConfigDir, obsConfigFile, err := config.CliPath()
		// Assert
		assert.Equal(t, config.UserConfigDirectoryNotFoundErrorMessage, err.Error())
		assert.Equal(t, "", obsConfigDir)
		assert.Equal(t, "", obsConfigFile)
	})
}
