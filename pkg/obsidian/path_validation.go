package obsidian

import (
	"errors"
	"path/filepath"
	"strings"
)

// ErrPathTraversal is returned when a path attempts to escape the base directory
var ErrPathTraversal = errors.New("path traversal detected: path must remain within vault directory")

// ValidatePath ensures the given relative path, when combined with basePath,
// stays within basePath. It returns the cleaned absolute path on success.
//
// This function:
// 1. Rejects absolute paths
// 2. Cleans both paths and joins them
// 3. Verifies the result starts with the base path
//
// Returns:
// - The validated absolute path
// - ErrPathTraversal if path escapes basePath
// - Other errors for filesystem issues
func ValidatePath(basePath, relativePath string) (string, error) {
	// Reject absolute paths
	if filepath.IsAbs(relativePath) {
		return "", ErrPathTraversal
	}

	// Clean and make base path absolute
	absBase, err := filepath.Abs(filepath.Clean(basePath))
	if err != nil {
		return "", err
	}

	// Clean the relative path and join with base
	cleanRelative := filepath.Clean(relativePath)
	joinedPath := filepath.Join(absBase, cleanRelative)

	// Get absolute path of joined result
	absJoined, err := filepath.Abs(joinedPath)
	if err != nil {
		return "", err
	}

	// Verify the joined path starts with the base path
	// Add trailing separator to prevent partial matches (e.g., /vault-backup matching /vault)
	if !strings.HasPrefix(absJoined, absBase+string(filepath.Separator)) && absJoined != absBase {
		return "", ErrPathTraversal
	}

	return absJoined, nil
}
