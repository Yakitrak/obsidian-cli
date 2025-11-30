package actions

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
)

// InstallIgnoreOptions controls writing the default .obsidianignore file.
type InstallIgnoreOptions struct {
	Force bool
}

// InstallDefaultIgnore writes the built-in .obsidianignore to the vault root.
// If the file already exists and Force is false, an error is returned.
func InstallDefaultIgnore(vault obsidian.VaultManager, opts InstallIgnoreOptions) (string, error) {
	vaultPath, err := vault.Path()
	if err != nil {
		return "", err
	}

	dest := filepath.Join(vaultPath, ".obsidianignore")
	if _, err := os.Stat(dest); err == nil && !opts.Force {
		return "", fmt.Errorf("%s already exists; re-run with --force to overwrite", dest)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	content := obsidian.DefaultIgnoreFile()
	if err := os.WriteFile(dest, []byte(content), 0o644); err != nil {
		return "", err
	}

	return dest, nil
}
