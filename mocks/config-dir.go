package mocks

import (
	"testing"
)

func CreateMockObsidianConfigFile(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	return tmpDir + "/obsidian.json"
}

func CreateMockCliConfigDirectories(t *testing.T) (string, string) {
	t.Helper()
	tmpDir := t.TempDir()
	return tmpDir, tmpDir + "/preferences.json"
}
