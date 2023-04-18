package config

import "errors"

func ObsidianFile() (obsidianConfigFile string, err error) {
	userConfigDir, err := UserConfigDirectory()
	if err != nil {
		return "", errors.New("user config directory not found")
	}
	obsidianConfigFile = userConfigDir + "/obsidian/obsidian.json"
	return obsidianConfigFile, nil
}
