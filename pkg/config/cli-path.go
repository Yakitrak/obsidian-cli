package config

import (
	"errors"
	"os"
)

var UserConfigDirectory = os.UserConfigDir

func CliPath() (cliConfigDir string, cliConfigFile string, err error) {
	userConfigDir, err := UserConfigDirectory()
	if err != nil {
		return "", "", errors.New("user config directory not found")
	}
	cliConfigDir = userConfigDir + "/obs"
	cliConfigFile = cliConfigDir + "/preferences.json"
	return cliConfigDir, cliConfigFile, nil
}
