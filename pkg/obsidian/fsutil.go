package obsidian

import (
	"io/fs"
	"os"
	"path/filepath"
)

// WriteFileAtomic writes data to a file atomically by writing to a temporary file
// and then renaming it over the target. This prevents partial writes if the
// process crashes or is interrupted.
func WriteFileAtomic(path string, data []byte, perm fs.FileMode) error {
	dir := filepath.Dir(path)
	base := filepath.Base(path)

	// Create temporary file in same directory to ensure atomic rename
	tmp, err := os.CreateTemp(dir, base+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	// Clean up temp file on any error
	defer func() {
		if tmp != nil {
			tmp.Close()
			os.Remove(tmpName)
		}
	}()

	// Write contents
	if _, err := tmp.Write(data); err != nil {
		return err
	}

	// Flush to disk for durability
	if err := tmp.Sync(); err != nil {
		return err
	}

	// Set permissions
	if err := tmp.Chmod(perm); err != nil {
		return err
	}

	// Close before rename
	if err := tmp.Close(); err != nil {
		return err
	}
	tmp = nil // Prevent double-close in defer

	// Atomic rename over original
	return os.Rename(tmpName, path)
}
