package obsidian

import (
	"path/filepath"
	"strings"
)

// ShouldIgnorePath reports whether a vault-relative path should be skipped for link rewrites or scanning.
// It skips hidden entries (dotfiles/directories), the .git directory, and any path with a prefix in ignoredPaths.
// Paths are evaluated relative to the vault root; ignoredPaths should also be vault-relative prefixes using '/'.
func ShouldIgnorePath(vaultPath, candidate string, ignoredPaths []string) bool {
	rel := candidate
	if vaultPath != "" {
		if r, err := filepath.Rel(vaultPath, candidate); err == nil {
			rel = r
		}
	}

	rel = filepath.ToSlash(rel)
	rel = strings.TrimPrefix(rel, "./")

	if rel == "." {
		return false
	}

	base := filepath.Base(rel)
	if strings.HasPrefix(base, ".") {
		return true
	}
	if rel == ".git" || strings.HasPrefix(rel, ".git/") {
		return true
	}

	for _, ig := range ignoredPaths {
		ig = strings.TrimSpace(ig)
		if ig == "" {
			continue
		}
		ig = filepath.ToSlash(ig)
		ig = strings.TrimPrefix(ig, "./")
		if strings.HasPrefix(rel, ig) {
			return true
		}
	}

	return false
}
