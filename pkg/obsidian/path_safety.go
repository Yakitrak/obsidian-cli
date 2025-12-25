package obsidian

import (
	"fmt"
	"path/filepath"
	"strings"
)

// SafeJoinVaultPath joins a vault root and a relative note path and ensures the result stays within the vault.
func SafeJoinVaultPath(vaultPath string, relativePath string) (string, error) {
	if filepath.IsAbs(relativePath) {
		return "", fmt.Errorf("absolute paths are not allowed: %s", relativePath)
	}
	cleaned := filepath.Clean(strings.TrimSpace(relativePath))
	cleaned = strings.TrimPrefix(cleaned, string(filepath.Separator))
	cleaned = strings.TrimPrefix(cleaned, "./")
	if cleaned == "" || cleaned == "." {
		return "", fmt.Errorf("note path cannot be empty")
	}

	absVault, err := filepath.Abs(vaultPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve vault path: %w", err)
	}

	joined := filepath.Join(absVault, filepath.FromSlash(cleaned))
	absJoined, err := filepath.Abs(joined)
	if err != nil {
		return "", fmt.Errorf("failed to resolve note path: %w", err)
	}

	if absJoined != absVault && !strings.HasPrefix(absJoined, absVault+string(filepath.Separator)) {
		return "", fmt.Errorf("note path escapes vault: %s", relativePath)
	}

	return absJoined, nil
}
