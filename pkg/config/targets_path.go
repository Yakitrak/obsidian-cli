package config

import "path/filepath"

const ObsidianCLITargetsFile = "targets.yaml"

// TargetsPath returns the directory and absolute file path for targets.yaml.
func TargetsPath() (cliConfigDir string, targetsFile string, err error) {
	cliConfigDir, _, err = CliPath()
	if err != nil {
		return "", "", err
	}
	targetsFile = filepath.Join(cliConfigDir, ObsidianCLITargetsFile)
	return cliConfigDir, targetsFile, nil
}
