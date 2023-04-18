package note_test

import (
	"github.com/Yakitrak/obsidian-cli/utils/note"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

type mockFileInfo struct {
	name  string
	isDir bool
}

func (fi *mockFileInfo) Name() string {
	return fi.name
}

func (fi *mockFileInfo) Size() int64 {
	return 0
}

func (fi *mockFileInfo) Mode() os.FileMode {
	return 0
}

func (fi *mockFileInfo) ModTime() time.Time {
	return time.Now()
}

func (fi *mockFileInfo) IsDir() bool {
	return fi.isDir
}

func (fi *mockFileInfo) Sys() interface{} {
	return nil
}
func TestShouldSkipDirectoryOrFile(t *testing.T) {
	tests := []struct {
		testName string
		info     os.FileInfo
		want     bool
	}{
		{"markdown file", &mockFileInfo{"file.md", false}, false},
		{"text file", &mockFileInfo{"file.txt", false}, true},
		{"image file", &mockFileInfo{"file.jpg", false}, true},
		{"directory", &mockFileInfo{"directory", true}, true},
		{"hidden directory", &mockFileInfo{".hidden_directory", true}, true},
		{"hidden file", &mockFileInfo{".hidden_file", false}, true},
		{"file with no extension", &mockFileInfo{"file_with_no_extension", false}, true},
		{"file with dots", &mockFileInfo{"file.md.with.dots", false}, true},
		{"markdown file with dots", &mockFileInfo{"file.with.multiple.dots.md", false}, false},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := note.ShouldSkipDirectoryOrFile(tt.info)
			assert.Equal(t, tt.want, got)
		})
	}
}
