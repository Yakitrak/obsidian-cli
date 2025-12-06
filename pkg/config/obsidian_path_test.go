package config_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/atomicobject/obsidian-cli/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestConfigObsidianPath(t *testing.T) {
	t.Run("UserConfigDir func successfully returns directory", func(t *testing.T) {
		// Arrange
		config.UserConfigDirectory = func() (string, error) {
			return filepath.Join("user", "config", "dir"), nil
		}
		// Act
		obsConfigFile, err := config.ObsidianFile()
		// Assert
		assert.Equal(t, nil, err)
		assert.Equal(t, filepath.Join("user", "config", "dir", "obsidian", "obsidian.json"), obsConfigFile)
	})

	t.Run("UserConfigDir func returns an error", func(t *testing.T) {
		// Arrange
		config.UserConfigDirectory = func() (string, error) {
			return "", errors.New(config.UserConfigDirectoryNotFoundErrorMessage)
		}
		// Act
		obsConfigFile, err := config.ObsidianFile()
		// Assert
		assert.Equal(t, config.UserConfigDirectoryNotFoundErrorMessage, err.Error())
		assert.Equal(t, "", obsConfigFile)
	})
}
