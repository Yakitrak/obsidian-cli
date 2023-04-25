package config

import (
	"errors"
	"path/filepath"
)

func ObsidianFile() (obsidianConfigFile string, err error) {
	userConfigDir, err := UserConfigDirectory()
	if err != nil {
		return "", errors.New(UserConfigDirectoryNotFoundErrorMessage)
	}
	obsidianConfigFile = filepath.Join(userConfigDir, ObsidianConfigDirectory, ObsidianConfigFile)
	return obsidianConfigFile, nil
}
