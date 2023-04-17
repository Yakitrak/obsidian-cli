package utils_test

import (
	"fmt"
	"github.com/Yakitrak/obsidian-cli/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCliConfigPath(t *testing.T) {
	originalUserConfigDirectory := utils.UserConfigDirectory
	defer func() { utils.UserConfigDirectory = originalUserConfigDirectory }()

	t.Run("userConfigDirectory func returns a directory", func(t *testing.T) {
		// Arrange
		utils.UserConfigDirectory = func() (string, error) {
			return "user/config/dir", nil
		}
		// Act
		obsConfigDir, obsConfigFile, err := utils.CliConfigPath()
		// Assert
		assert.Equal(t, nil, err)
		assert.Equal(t, "user/config/dir/obs", obsConfigDir)
		assert.Equal(t, "user/config/dir/obs/preferences.json", obsConfigFile)
	})

	t.Run("userConfigDirectory func returns an error", func(t *testing.T) {
		// Arrange
		utils.UserConfigDirectory = func() (string, error) {
			return "", fmt.Errorf("user config directory not found")
		}
		// Act
		obsConfigDir, obsConfigFile, err := utils.CliConfigPath()
		// Assert
		assert.Equal(t, "user config directory not found", err.Error())
		assert.Equal(t, "", obsConfigDir)
		assert.Equal(t, "", obsConfigFile)
	})

}
