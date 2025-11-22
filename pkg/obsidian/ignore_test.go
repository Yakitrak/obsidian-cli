package obsidian

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldIgnorePath(t *testing.T) {
	vault := "/tmp/vault"
	tests := []struct {
		name    string
		path    string
		ignored []string
		want    bool
	}{
		{name: "hidden file", path: filepath.Join(vault, ".hidden.md"), want: true},
		{name: ".git root", path: filepath.Join(vault, ".git"), want: true},
		{name: ".git child", path: filepath.Join(vault, ".git", "config"), want: true},
		{name: "ignored prefix match", path: filepath.Join(vault, "Archive", "note.md"), ignored: []string{"Archive"}, want: true},
		{name: "ignored prefix with leading ./", path: filepath.Join(vault, "Archive", "note.md"), ignored: []string{"./Archive/"}, want: true},
		{name: "not ignored normal file", path: filepath.Join(vault, "notes", "note.md"), ignored: []string{"Archive"}, want: false},
		{name: "relative path preserved", path: "notes/note.md", ignored: []string{"Archive"}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ShouldIgnorePath(vault, tt.path, tt.ignored))
		})
	}
}
