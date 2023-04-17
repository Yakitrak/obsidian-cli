package utils

import (
	"errors"
	"os"
)

var UserConfigDirectory = os.UserConfigDir

func CliConfigPath() (obsConfigDir string, obsConfigFile string, err error) {
	userConfigDir, err := UserConfigDirectory()
	if err != nil {
		return "", "", errors.New("user config directory not found")
	}
	obsConfigDir = userConfigDir + "/obs"
	obsConfigFile = obsConfigDir + "/preferences.json"
	return obsConfigDir, obsConfigFile, nil
}
