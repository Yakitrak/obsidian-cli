package mocks

import (
	"os"
	"time"
)

type MockFileInfo struct {
	FileName    string
	IsDirectory bool
}

func (fi *MockFileInfo) Name() string {
	return fi.FileName
}

func (fi *MockFileInfo) Size() int64 {
	return 0
}

func (fi *MockFileInfo) Mode() os.FileMode {
	return 0
}

func (fi *MockFileInfo) ModTime() time.Time {
	return time.Now()
}

func (fi *MockFileInfo) IsDir() bool {
	return fi.IsDirectory
}

func (fi *MockFileInfo) Sys() interface{} {
	return nil
}
