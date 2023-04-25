package config

import (
	"errors"
	"os"
	"path/filepath"
)

var UserConfigDirectory = os.UserConfigDir

func CliPath() (cliConfigDir string, cliConfigFile string, err error) {
	userConfigDir, err := UserConfigDirectory()
	if err != nil {
		return "", "", errors.New(UserConfigDirectoryNotFoundErrorMessage)
	}
	cliConfigDir = filepath.Join(userConfigDir, ObsidianCLIConfigDirectory)
	cliConfigFile = filepath.Join(cliConfigDir, ObsidianCLIConfigFile)
	return cliConfigDir, cliConfigFile, nil
}
