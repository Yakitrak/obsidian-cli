package obsidian

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteFileAtomic(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "fsutil-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.txt")
	testContent := []byte("Hello, atomic world!")

	// Test basic functionality
	err = WriteFileAtomic(testFile, testContent, 0644)
	assert.NoError(t, err)

	// Verify content was written correctly
	readContent, err := os.ReadFile(testFile)
	assert.NoError(t, err)
	assert.Equal(t, testContent, readContent)

	// Verify file permissions
	info, err := os.Stat(testFile)
	assert.NoError(t, err)
	got := info.Mode().Perm()
	expected := os.FileMode(0644)
	if runtime.GOOS == "windows" {
		// Windows commonly reports 0666; ensure requested bits are present.
		assert.Equal(t, expected, got&expected)
	} else {
		assert.Equal(t, expected, got)
	}
}

func TestWriteFileAtomicOverwrite(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "fsutil-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.txt")

	// Write initial content
	initialContent := []byte("Initial content")
	err = os.WriteFile(testFile, initialContent, 0644)
	assert.NoError(t, err)

	// Overwrite with atomic write
	newContent := []byte("New atomic content")
	err = WriteFileAtomic(testFile, newContent, 0644)
	assert.NoError(t, err)

	// Verify new content
	readContent, err := os.ReadFile(testFile)
	assert.NoError(t, err)
	assert.Equal(t, newContent, readContent)
}

func TestWriteFileAtomicNoTempFileLeftBehind(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "fsutil-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.txt")
	testContent := []byte("Test content")

	// Write file
	err = WriteFileAtomic(testFile, testContent, 0644)
	assert.NoError(t, err)

	// Check that no temporary files are left behind
	entries, err := os.ReadDir(tempDir)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "test.txt", entries[0].Name())
}

func TestWriteFileAtomicInvalidDir(t *testing.T) {
	// Try to write to a non-existent directory
	invalidPath := "/nonexistent/directory/file.txt"
	testContent := []byte("Test content")

	err := WriteFileAtomic(invalidPath, testContent, 0644)
	assert.Error(t, err)
}

func TestWriteFileAtomicPermissions(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "fsutil-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.txt")
	testContent := []byte("Test content")

	// Write with specific permissions
	err = WriteFileAtomic(testFile, testContent, 0600)
	assert.NoError(t, err)

	// Verify permissions
	info, err := os.Stat(testFile)
	assert.NoError(t, err)
	got := info.Mode().Perm()
	expected := os.FileMode(0600)
	if runtime.GOOS == "windows" {
		assert.Equal(t, expected, got&expected)
	} else {
		assert.Equal(t, expected, got)
	}
}
