//go:build darwin || linux

package obsidian

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateLinksWriteLimitPreservesOriginal(t *testing.T) {
	if vault := os.Getenv("NOTESMD_TEST_WRITE_LIMIT_VAULT"); vault != "" {
		if err := syscall.Setrlimit(syscall.RLIMIT_FSIZE, &syscall.Rlimit{Cur: 0, Max: 0}); err != nil {
			os.Exit(25)
		}
		if err := (&Note{}).UpdateLinks(vault, "alpha", "renamed"); err != nil {
			os.Exit(24)
		}
		os.Exit(0)
	}
	vault := t.TempDir()
	path := filepath.Join(vault, "daily.md")
	original := "Worked on [[alpha]] today.\n"
	checkReplacementError(t, os.WriteFile(path, []byte(original), 0600))
	cmd := exec.Command(os.Args[0], "-test.run=^TestUpdateLinksWriteLimitPreservesOriginal$")
	cmd.Env = append(os.Environ(), "NOTESMD_TEST_WRITE_LIMIT_VAULT="+vault)
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected write to fail under zero file-size limit")
	}
	if cmd.ProcessState.ExitCode() == 25 {
		t.Fatal("failed to set file-size limit")
	}
	content, err := os.ReadFile(path)
	checkReplacementError(t, err)
	assert.Equal(t, original, string(content))
}
