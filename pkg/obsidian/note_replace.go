package obsidian

import (
	"io"
	"os"
	"path/filepath"
)

// replaceNoteFile stages the complete replacement beside the original. A failed
// write or an interruption before the rename leaves the original untouched.
func replaceNoteFile(path string, content []byte) error {
	return writeNoteReplacement(path, func(w io.Writer) error {
		_, err := w.Write(content)
		return err
	})
}

func writeNoteReplacement(path string, write func(io.Writer) error) error {
	// Follow symlinks as WriteFile did, rather than replacing the link itself.
	target, err := filepath.EvalSymlinks(path)
	if err != nil {
		return err
	}
	info, err := os.Stat(target)
	if err != nil {
		return err
	}
	// Replacement depends on directory permissions; still respect read-only notes.
	if info.Mode().Perm()&0222 == 0 {
		return os.ErrPermission
	}
	temp, err := os.CreateTemp(filepath.Dir(target), ".notesmd-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(temp.Name()) }()
	defer temp.Close()
	if err := write(temp); err != nil {
		return err
	}
	if err := temp.Chmod(info.Mode()); err != nil {
		return err
	}
	if err := temp.Sync(); err != nil {
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(temp.Name(), target)
}
