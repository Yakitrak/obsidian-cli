package config_test

import (
	"fmt"
	"github.com/Yakitrak/obsidian-cli/utils/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfigObsidianPath(t *testing.T) {
	t.Run("userConfigDirectory func returns a directory", func(t *testing.T) {
		// Arrange
		config.UserConfigDirectory = func() (string, error) {
			return "user/config/dir", nil
		}
		// Act
		obsConfigFile, err := config.ObsidianFile()
		// Assert
		assert.Equal(t, nil, err)
		assert.Equal(t, "user/config/dir/obsidian/obsidian.json", obsConfigFile)
	})

	t.Run("userConfigDirectory func returns an error", func(t *testing.T) {
		// Arrange
		config.UserConfigDirectory = func() (string, error) {
			return "", fmt.Errorf("user config directory not found")
		}
		// Act
		obsConfigFile, err := config.ObsidianFile()
		// Assert
		assert.Equal(t, "user config directory not found", err.Error())
		assert.Equal(t, "", obsConfigFile)
	})
}
